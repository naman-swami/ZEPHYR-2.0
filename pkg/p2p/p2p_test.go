package p2p

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestP2PSync_ValidAndTamperedBlobs(t *testing.T) {
	peerDir := t.TempDir()
	clientDir := t.TempDir()

	// Create 1 valid CAS blob in peerDir
	validContent := []byte("reproducible binary artifact payload v2.0")
	sum := sha256.Sum256(validContent)
	validDigest := hex.EncodeToString(sum[:])

	validBlobPath := filepath.Join(peerDir, "cas", "objects", validDigest[:2], validDigest[2:])
	if err := os.MkdirAll(filepath.Dir(validBlobPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(validBlobPath, validContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Create 1 tampered CAS blob (content does NOT match its filename digest!)
	fakeDigest := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	fakeBlobPath := filepath.Join(peerDir, "cas", "objects", fakeDigest[:2], fakeDigest[2:])
	_ = os.MkdirAll(filepath.Dir(fakeBlobPath), 0755)
	_ = os.WriteFile(fakeBlobPath, []byte("malicious poisoned payload"), 0644)

	// Create 1 valid Action Cache manifest
	acHash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	_ = os.WriteFile(filepath.Join(peerDir, acHash+".json"), []byte(`{"task":"build"}`), 0644)

	clusterSecret := "zephyr-lan-secret-key"
	srv := httptest.NewServer(NewPeerHTTPHandler(peerDir, "node-alpha", clusterSecret))
	defer srv.Close()

	// Run P2P Sync from peer to clientDir
	rep, err := SyncFromPeer(srv.URL, clientDir, clusterSecret)
	if err != nil {
		t.Fatalf("SyncFromPeer failed: %v", err)
	}

	if rep.BlobsSynced != 1 {
		t.Errorf("expected 1 valid blob synced, got %d", rep.BlobsSynced)
	}
	if rep.TamperedRejected != 1 {
		t.Errorf("expected 1 tampered blob rejected, got %d", rep.TamperedRejected)
	}
	if rep.ManifestsSynced != 1 {
		t.Errorf("expected 1 manifest synced, got %d", rep.ManifestsSynced)
	}

	// Verify valid blob exists on client and tampered blob does NOT exist
	clientValidPath := filepath.Join(clientDir, "cas", "objects", validDigest[:2], validDigest[2:])
	gotBytes, err := os.ReadFile(clientValidPath)
	if err != nil || string(gotBytes) != string(validContent) {
		t.Errorf("synced blob content mismatch")
	}

	clientFakePath := filepath.Join(clientDir, "cas", "objects", fakeDigest[:2], fakeDigest[2:])
	if _, err := os.Stat(clientFakePath); err == nil {
		t.Errorf("SECURITY FAILURE: tampered blob was written to client CAS!")
	}
}

func TestBeaconHMACVerification(t *testing.T) {
	b := PeerBeacon{
		Protocol:  "ZEPHYR-P2P/2.0",
		NodeID:    "workstation-01",
		HTTPAddr:  "192.168.1.42:9090",
		Timestamp: 1700000000,
	}
	b.Signature = SignBeacon(b, "team-key")

	if !VerifyBeacon(b, "team-key") {
		t.Errorf("expected valid beacon signature to pass verification")
	}
	if VerifyBeacon(b, "wrong-key") {
		t.Errorf("expected wrong cluster key to reject beacon")
	}
}
