// Package prune provides an enterprise-grade, crash-safe, referentially-intact
// Content-Addressable Storage (CAS) LRU eviction and cache pruning engine.
//
// Key architectural guarantees:
//   1. Referential Integrity (Two-Phase Mark/Sweep): Manifests are scanned
//      under read lock to build the live digest set. No blob referenced by
//      any active manifest is EVER evicted, eliminating broken action caches.
//   2. Anti-Thrashing Headroom: Pruning reduces CAS disk usage to
//      Budget * (1 - Headroom), preventing continuous immediate re-eviction
//      on subsequent task runs.
//   3. Manifest-Write Race Protection: Blobs created within the PruneGrace
//      window (default 15 minutes) are protected from eviction during standard
//      pruning runs, closing the window where an in-flight build stores blobs
//      prior to writing its manifest.
//   4. Windows Sharing Violation Durability: File removals retry with
//      exponential backoff to handle transient antivirus, indexing, or read locks.
//   5. Zero External Dependencies: Implemented strictly with the Go standard library.
package prune

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var digestRe = regexp.MustCompile(`\b[0-9a-f]{64}\b`)

const (
	PruneGrace      = 15 * time.Minute
	DefaultHeadroom = 0.10
)

type PruneOptions struct {
	MaxSizeBytes int64         // Hard CAS storage budget (0 = disabled)
	MaxAge       time.Duration // Evict unreferenced blobs older than this (0 = disabled)
	Headroom     float64       // Fraction below budget to prune to (default 0.10)
	DryRun       bool          // Simulate pruning without deleting files
	Emergency    bool          // Waive grace period (under severe disk pressure)
}

type PruneStats struct {
	BlobsScanned  int    `json:"blobs_scanned"`
	BlobsLive     int    `json:"blobs_live"`
	BlobsEvicted  int    `json:"blobs_evicted"`
	BytesScanned  int64  `json:"bytes_scanned"`
	BytesEvicted  int64  `json:"bytes_evicted"`
	DurationMS    int64  `json:"duration_ms"`
	GraceWaived   bool   `json:"grace_waived"`
	StoppedReason string `json:"stopped_reason,omitempty"`
}

type blobInfo struct {
	hash string
	path string
	size int64
	mod  time.Time
}

type Pruner struct {
	cacheRoot  string
	manifestMu *sync.RWMutex
}

func NewPruner(cacheRoot string, manifestMu *sync.RWMutex) *Pruner {
	if manifestMu == nil {
		manifestMu = &sync.RWMutex{}
	}
	return &Pruner{
		cacheRoot:  cacheRoot,
		manifestMu: manifestMu,
	}
}

func (p *Pruner) Prune(ctx context.Context, opt PruneOptions) (*PruneStats, error) {
	started := time.Now()
	stats := &PruneStats{}

	if opt.Headroom <= 0 || opt.Headroom >= 1.0 {
		opt.Headroom = DefaultHeadroom
	}

	// Phase 1: Mark (scan manifests to build referentially intact live set)
	live, err := p.scanLiveSet(ctx)
	if err != nil {
		return stats, fmt.Errorf("prune manifest scan: %w", err)
	}
	stats.BlobsLive = len(live)

	// Phase 2: Collect blob inventory
	blobs, totalBytes, err := p.collectBlobs(ctx)
	if err != nil {
		return stats, fmt.Errorf("prune blob walk: %w", err)
	}
	stats.BlobsScanned = len(blobs)
	stats.BytesScanned = totalBytes

	// Phase 3: Filter eviction candidates
	graceCutoff := time.Now().Add(-PruneGrace)
	var candidates []blobInfo
	for _, b := range blobs {
		if _, isLive := live[b.hash]; isLive {
			continue // Never evict a blob referenced by an active manifest
		}
		if !opt.Emergency && b.mod.After(graceCutoff) {
			continue // Grace period: protect recently stored blobs from in-flight builds
		}
		candidates = append(candidates, b)
	}

	// Sort candidates by LRU (least recently used / oldest mtime first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].mod.Before(candidates[j].mod)
	})

	targetBytes := int64(float64(opt.MaxSizeBytes) * (1.0 - opt.Headroom))
	needSize := opt.MaxSizeBytes > 0 && totalBytes > targetBytes
	needAge := opt.MaxAge > 0

	if !needSize && !needAge {
		stats.StoppedReason = "within budget"
		stats.DurationMS = time.Since(started).Milliseconds()
		return stats, nil
	}

	// Phase 4: Sweep candidates until under budget and age policy
	for _, b := range candidates {
		if err := ctx.Err(); err != nil {
			stats.StoppedReason = "cancelled"
			break
		}

		ageExpired := needAge && time.Since(b.mod) > opt.MaxAge
		overBudget := needSize && totalBytes > targetBytes

		if !ageExpired && !overBudget {
			break
		}

		if opt.DryRun {
			stats.BlobsEvicted++
			stats.BytesEvicted += b.size
			totalBytes -= b.size
			continue
		}

		if err := removeWithRetry(b.path); err != nil {
			stats.StoppedReason = fmt.Sprintf("delete %s: %v", b.hash[:12], err)
			continue
		}

		stats.BlobsEvicted++
		stats.BytesEvicted += b.size
		totalBytes -= b.size
	}

	if stats.StoppedReason == "" {
		stats.StoppedReason = "complete"
	}
	stats.GraceWaived = opt.Emergency
	stats.DurationMS = time.Since(started).Milliseconds()
	return stats, nil
}

func (p *Pruner) scanLiveSet(ctx context.Context) (map[string]struct{}, error) {
	live := make(map[string]struct{})
	p.manifestMu.RLock()
	defer p.manifestMu.RUnlock()

	err := filepath.WalkDir(p.cacheRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			if d.Name() == "objects" || d.Name() == "quarantine" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only inspect JSON files (manifests, cache entries, stats)
		if !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil // raced with file update
			}
			return err
		}

		for _, m := range digestRe.FindAllString(string(data), -1) {
			live[strings.ToLower(m)] = struct{}{}
		}
		return nil
	})

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	return live, nil
}

func (p *Pruner) collectBlobs(ctx context.Context) ([]blobInfo, int64, error) {
	var blobs []blobInfo
	var total int64
	objectsRoot := filepath.Join(p.cacheRoot, "objects")

	err := filepath.WalkDir(objectsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".zephyr-tmp-") || strings.HasPrefix(d.Name(), "blob.") {
			return nil // Never touch in-flight temp files
		}

		hash := d.Name()
		if len(hash) != 64 {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		blobs = append(blobs, blobInfo{
			hash: strings.ToLower(hash),
			path: path,
			size: info.Size(),
			mod:  info.ModTime(),
		})
		total += info.Size()
		return nil
	})

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, 0, err
	}
	return blobs, total, nil
}

func removeWithRetry(path string) error {
	var lastErr error
	for attempt := 0; attempt < 6; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<attempt) * 5 * time.Millisecond) // 10..320ms exponential backoff
		}
		if err := os.Remove(path); err == nil || errors.Is(err, fs.ErrNotExist) {
			return nil
		} else {
			lastErr = err
		}
	}
	return lastErr
}
