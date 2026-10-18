package prune

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestPruner_ReferentialIntegrity(t *testing.T) {
	tempDir := t.TempDir()
	objectsDir := filepath.Join(tempDir, "objects", "ab")
	_ = os.MkdirAll(objectsDir, 0755)

	// Create 2 fake 64-hex hash blobs
	liveHash := "ab11111111111111111111111111111111111111111111111111111111111111"
	deadHash := "ab22222222222222222222222222222222222222222222222222222222222222"

	liveBlob := filepath.Join(objectsDir, liveHash)
	deadBlob := filepath.Join(objectsDir, deadHash)

	_ = os.WriteFile(liveBlob, []byte("live-blob-data"), 0644)
	_ = os.WriteFile(deadBlob, []byte("dead-blob-data"), 0644)

	// Set mtime to 1 hour ago so grace period doesn't block eviction
	oldTime := time.Now().Add(-1 * time.Hour)
	_ = os.Chtimes(liveBlob, oldTime, oldTime)
	_ = os.Chtimes(deadBlob, oldTime, oldTime)

	// Write a manifest that references liveHash
	manifestPath := filepath.Join(tempDir, "task_build.manifest.json")
	manifestJSON := fmt.Sprintf(`{
		"task_name": "build",
		"output_hashes": {
			"dist/app": "%s"
		}
	}`, liveHash)
	_ = os.WriteFile(manifestPath, []byte(manifestJSON), 0644)

	pruner := NewPruner(tempDir, &sync.RWMutex{})
	ctx := context.Background()

	// Prune with a tiny max size (1 byte) to force eviction
	stats, err := pruner.Prune(ctx, PruneOptions{
		MaxSizeBytes: 1,
		Emergency:    true,
	})
	if err != nil {
		t.Fatalf("prune failed: %v", err)
	}

	if stats.BlobsLive != 1 {
		t.Errorf("expected 1 live blob, got %d", stats.BlobsLive)
	}
	if stats.BlobsEvicted != 1 {
		t.Errorf("expected 1 evicted blob, got %d", stats.BlobsEvicted)
	}

	// Verify live blob STILL exists
	if _, err := os.Stat(liveBlob); err != nil {
		t.Errorf("critical referential integrity failure: live blob was evicted!")
	}

	// Verify dead blob was evicted
	if _, err := os.Stat(deadBlob); !os.IsNotExist(err) {
		t.Errorf("expected dead blob to be evicted, but it still exists")
	}
}

func TestPruner_LRUBudgetEviction(t *testing.T) {
	tempDir := t.TempDir()
	objectsDir := filepath.Join(tempDir, "objects", "cc")
	_ = os.MkdirAll(objectsDir, 0755)

	// Create 3 unreferenced blobs with distinct mtimes
	b1Hash := "cc11111111111111111111111111111111111111111111111111111111111111"
	b2Hash := "cc22222222222222222222222222222222222222222222222222222222222222"
	b3Hash := "cc33333333333333333333333333333333333333333333333333333333333333"

	b1 := filepath.Join(objectsDir, b1Hash)
	b2 := filepath.Join(objectsDir, b2Hash)
	b3 := filepath.Join(objectsDir, b3Hash)

	payload := make([]byte, 1000) // 1000 bytes each
	_ = os.WriteFile(b1, payload, 0644)
	_ = os.WriteFile(b2, payload, 0644)
	_ = os.WriteFile(b3, payload, 0644)

	// Set LRU timestamps: b1 oldest, b2 medium, b3 newest
	tOldest := time.Now().Add(-3 * time.Hour)
	tMedium := time.Now().Add(-2 * time.Hour)
	tNewest := time.Now().Add(-1 * time.Hour)

	_ = os.Chtimes(b1, tOldest, tOldest)
	_ = os.Chtimes(b2, tMedium, tMedium)
	_ = os.Chtimes(b3, tNewest, tNewest)

	pruner := NewPruner(tempDir, &sync.RWMutex{})
	ctx := context.Background()

	// Budget of 1500 bytes (with 10% headroom -> target 1350 bytes).
	// Total size = 3000 bytes.
	// Evicting b1 (oldest) leaves 2000 bytes > 1350 bytes.
	// Evicting b2 leaves 1000 bytes <= 1350 bytes.
	// b3 (newest) should remain intact!
	stats, err := pruner.Prune(ctx, PruneOptions{
		MaxSizeBytes: 1500,
		Headroom:     0.10,
		Emergency:    true,
	})
	if err != nil {
		t.Fatalf("prune failed: %v", err)
	}

	if stats.BlobsEvicted != 2 {
		t.Fatalf("expected 2 blobs evicted, got %d", stats.BlobsEvicted)
	}

	if _, err := os.Stat(b1); !os.IsNotExist(err) {
		t.Errorf("expected b1 (oldest) to be evicted")
	}
	if _, err := os.Stat(b2); !os.IsNotExist(err) {
		t.Errorf("expected b2 (medium) to be evicted")
	}
	if _, err := os.Stat(b3); err != nil {
		t.Errorf("expected b3 (newest) to be preserved, err: %v", err)
	}
}

func TestPruner_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	objectsDir := filepath.Join(tempDir, "objects", "dd")
	_ = os.MkdirAll(objectsDir, 0755)

	hash := "dd11111111111111111111111111111111111111111111111111111111111111"
	blobPath := filepath.Join(objectsDir, hash)
	_ = os.WriteFile(blobPath, []byte("dry-run-data"), 0644)
	oldTime := time.Now().Add(-1 * time.Hour)
	_ = os.Chtimes(blobPath, oldTime, oldTime)

	pruner := NewPruner(tempDir, &sync.RWMutex{})
	ctx := context.Background()

	stats, err := pruner.Prune(ctx, PruneOptions{
		MaxSizeBytes: 1,
		DryRun:       true,
		Emergency:    true,
	})
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}

	if stats.BlobsEvicted != 1 {
		t.Errorf("expected 1 blob marked evicted in stats, got %d", stats.BlobsEvicted)
	}

	// In dry run, file must NOT be deleted
	if _, err := os.Stat(blobPath); err != nil {
		t.Errorf("dry run deleted file from disk: %v", err)
	}
}
