// Package p2p implements Peer-to-Peer (P2P) Local Area Network cache discovery
// and cryptographically verified CAS blob synchronization across developer subnets.
//
// Crafted with 100% Go Standard Library — Zero External Dependencies.
package p2p

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Data Types
// ─────────────────────────────────────────────────────────────────────────────

type PeerBeacon struct {
	Protocol  string `json:"protocol"` // "ZEPHYR-P2P/2.0"
	NodeID    string `json:"node_id"`
	HTTPAddr  string `json:"http_addr"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature,omitempty"`
}

type NodeInventory struct {
	NodeID     string   `json:"node_id"`
	ACHashes   []string `json:"ac_hashes"`
	CASDigests []string `json:"cas_digests"`
}

type SyncReport struct {
	PeerURL          string `json:"peer_url"`
	ManifestsSynced  int    `json:"manifests_synced"`
	BlobsSynced      int    `json:"blobs_synced"`
	BlobsSkipped     int    `json:"blobs_skipped"`
	TamperedRejected int    `json:"tampered_rejected"`
	BytesTransferred int64  `json:"bytes_transferred"`
	DurationMs       int64  `json:"duration_ms"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Cryptographic Beacon Signing
// ─────────────────────────────────────────────────────────────────────────────

func SignBeacon(b PeerBeacon, clusterKey string) string {
	if clusterKey == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(clusterKey))
	payload := fmt.Sprintf("%s|%s|%s|%d", b.Protocol, b.NodeID, b.HTTPAddr, b.Timestamp)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyBeacon(b PeerBeacon, clusterKey string) bool {
	if clusterKey == "" {
		return true
	}
	expected := SignBeacon(b, clusterKey)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(b.Signature)) == 1
}

// ─────────────────────────────────────────────────────────────────────────────
// Local Cache Inventory Scanner
// ─────────────────────────────────────────────────────────────────────────────

func ScanLocalInventory(cacheDir, nodeID string) (*NodeInventory, error) {
	inv := &NodeInventory{
		NodeID:     nodeID,
		ACHashes:   make([]string, 0),
		CASDigests: make([]string, 0),
	}

	// Scan Action Cache manifests (.taskcache/*.json or .taskcache/ac/*.json)
	entries, err := os.ReadDir(cacheDir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") && e.Name() != "stats.json" {
				hash := strings.TrimSuffix(e.Name(), ".json")
				if len(hash) == 64 {
					inv.ACHashes = append(inv.ACHashes, hash)
				}
			}
		}
	}

	// Scan 2-level sharded CAS (.taskcache/cas/objects/ab/cdef...)
	casRoot := filepath.Join(cacheDir, "cas", "objects")
	_ = filepath.WalkDir(casRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(casRoot, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		parts := strings.Split(rel, "/")
		if len(parts) == 2 && len(parts[0]) == 2 && len(parts[1]) == 62 {
			inv.CASDigests = append(inv.CASDigests, parts[0]+parts[1])
		} else if len(parts) == 1 && len(parts[0]) == 64 {
			inv.CASDigests = append(inv.CASDigests, parts[0])
		}
		return nil
	})

	return inv, nil
}

func casBlobPath(cacheDir, digest string) string {
	if len(digest) >= 64 {
		return filepath.Join(cacheDir, "cas", "objects", digest[:2], digest[2:])
	}
	return filepath.Join(cacheDir, "cas", "objects", digest)
}

// ─────────────────────────────────────────────────────────────────────────────
// P2P HTTP Sync Handler
// ─────────────────────────────────────────────────────────────────────────────

func NewPeerHTTPHandler(cacheDir, nodeID, clusterKey string) http.Handler {
	mux := http.NewServeMux()

	authCheck := func(w http.ResponseWriter, r *http.Request) bool {
		if clusterKey == "" {
			return true
		}
		got := r.Header.Get("X-Zephyr-Cluster-Key")
		if subtle.ConstantTimeCompare([]byte(got), []byte(clusterKey)) != 1 {
			http.Error(w, "Unauthorized P2P cluster key", http.StatusUnauthorized)
			return false
		}
		return true
	}

	mux.HandleFunc("/p2p/v1/manifest", func(w http.ResponseWriter, r *http.Request) {
		if !authCheck(w, r) {
			return
		}
		inv, err := ScanLocalInventory(cacheDir, nodeID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(inv)
	})

	mux.HandleFunc("/p2p/v1/cas/", func(w http.ResponseWriter, r *http.Request) {
		if !authCheck(w, r) {
			return
		}
		digest := strings.TrimPrefix(r.URL.Path, "/p2p/v1/cas/")
		if len(digest) != 64 || strings.ContainsAny(digest, "./\\") {
			http.Error(w, "Invalid digest", http.StatusBadRequest)
			return
		}
		p := casBlobPath(cacheDir, digest)
		f, err := os.Open(p)
		if err != nil {
			http.Error(w, "Blob not found", http.StatusNotFound)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = io.Copy(w, f)
	})

	mux.HandleFunc("/p2p/v1/ac/", func(w http.ResponseWriter, r *http.Request) {
		if !authCheck(w, r) {
			return
		}
		hash := strings.TrimPrefix(r.URL.Path, "/p2p/v1/ac/")
		if len(hash) != 64 || strings.ContainsAny(hash, "./\\") {
			http.Error(w, "Invalid hash", http.StatusBadRequest)
			return
		}
		p := filepath.Join(cacheDir, hash+".json")
		f, err := os.Open(p)
		if err != nil {
			http.Error(w, "Manifest not found", http.StatusNotFound)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.Copy(w, f)
	})

	return mux
}

// ─────────────────────────────────────────────────────────────────────────────
// Verified P2P Delta Sync Client
// ─────────────────────────────────────────────────────────────────────────────

func SyncFromPeer(peerURL, localCacheDir, clusterKey string) (*SyncReport, error) {
	start := time.Now()
	peerURL = strings.TrimRight(peerURL, "/")
	report := &SyncReport{PeerURL: peerURL}

	client := &http.Client{Timeout: 15 * time.Second}

	doGet := func(url string) (*http.Response, error) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		if clusterKey != "" {
			req.Header.Set("X-Zephyr-Cluster-Key", clusterKey)
		}
		return client.Do(req)
	}

	resp, err := doGet(peerURL + "/p2p/v1/manifest")
	if err != nil {
		return nil, fmt.Errorf("failed to contact peer %s: %w", peerURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("peer returned HTTP %d", resp.StatusCode)
	}

	var peerInv NodeInventory
	if err := json.NewDecoder(io.LimitReader(resp.Body, 16*1024*1024)).Decode(&peerInv); err != nil {
		return nil, fmt.Errorf("invalid peer manifest: %w", err)
	}

	// Sync missing CAS blobs with strict SHA-256 content verification
	for _, digest := range peerInv.CASDigests {
		if len(digest) != 64 {
			continue
		}
		localPath := casBlobPath(localCacheDir, digest)
		if _, err := os.Stat(localPath); err == nil {
			report.BlobsSkipped++
			continue
		}

		bResp, err := doGet(peerURL + "/p2p/v1/cas/" + digest)
		if err != nil {
			continue
		}
		rawBytes, err := io.ReadAll(io.LimitReader(bResp.Body, 512*1024*1024))
		bResp.Body.Close()
		if err != nil || bResp.StatusCode != http.StatusOK {
			continue
		}

		// Cryptographic zero-trust verification: hash must match digest!
		sum := sha256.Sum256(rawBytes)
		gotDigest := hex.EncodeToString(sum[:])
		if gotDigest != digest {
			report.TamperedRejected++
			continue
		}

		if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
			continue
		}
		if err := os.WriteFile(localPath, rawBytes, 0444); err == nil {
			report.BlobsSynced++
			report.BytesTransferred += int64(len(rawBytes))
		}
	}

	// Sync missing Action Cache manifests
	if err := os.MkdirAll(localCacheDir, 0755); err == nil {
		for _, acHash := range peerInv.ACHashes {
			if len(acHash) != 64 {
				continue
			}
			localAC := filepath.Join(localCacheDir, acHash+".json")
			if _, err := os.Stat(localAC); err == nil {
				continue
			}
			aResp, err := doGet(peerURL + "/p2p/v1/ac/" + acHash)
			if err != nil {
				continue
			}
			acBytes, err := io.ReadAll(io.LimitReader(aResp.Body, 16*1024*1024))
			aResp.Body.Close()
			if err != nil || aResp.StatusCode != http.StatusOK {
				continue
			}
			if os.WriteFile(localAC, acBytes, 0644) == nil {
				report.ManifestsSynced++
				report.BytesTransferred += int64(len(acBytes))
			}
		}
	}

	report.DurationMs = time.Since(start).Milliseconds()
	return report, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// UDP Subnet Discovery
// ─────────────────────────────────────────────────────────────────────────────

func BroadcastBeacon(udpAddr string, beacon PeerBeacon, clusterKey string) error {
	beacon.Protocol = "ZEPHYR-P2P/2.0"
	beacon.Timestamp = time.Now().Unix()
	beacon.Signature = SignBeacon(beacon, clusterKey)

	data, err := json.Marshal(beacon)
	if err != nil {
		return err
	}

	conn, err := net.Dial("udp4", udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write(data)
	return err
}

func ListenForBeacons(listenAddr string, timeout time.Duration, clusterKey string) ([]PeerBeacon, error) {
	pc, err := net.ListenPacket("udp4", listenAddr)
	if err != nil {
		return nil, err
	}
	defer pc.Close()

	_ = pc.SetReadDeadline(time.Now().Add(timeout))
	seen := make(map[string]PeerBeacon)
	buf := make([]byte, 4096)

	for {
		n, _, err := pc.ReadFrom(buf)
		if err != nil {
			break
		}
		var b PeerBeacon
		if json.Unmarshal(buf[:n], &b) == nil && b.Protocol == "ZEPHYR-P2P/2.0" {
			if VerifyBeacon(b, clusterKey) {
				seen[b.HTTPAddr] = b
			}
		}
	}

	peers := make([]PeerBeacon, 0, len(seen))
	for _, p := range seen {
		peers = append(peers, p)
	}
	return peers, nil
}
