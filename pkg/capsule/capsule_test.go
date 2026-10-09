package capsule

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCapsule_CreateInspectVerifyReplay(t *testing.T) {
	wsDir := t.TempDir()

	binDir := filepath.Join(wsDir, "bin")
	_ = os.MkdirAll(binDir, 0755)
	appFile := filepath.Join(binDir, "app.bin")
	metaFile := filepath.Join(wsDir, "version.txt")

	_ = os.WriteFile(appFile, []byte("compiled-binary-payload-98765"), 0755)
	_ = os.WriteFile(metaFile, []byte("v2.0.0-release"), 0644)

	capsuleFile := filepath.Join(t.TempDir(), "build.zcap")

	manifest := Manifest{
		BuildID:    "bld-test-12345",
		TaskName:   "build:app",
		Command:    "go build -o bin/app.bin",
		Timestamp:  time.Now().UTC(),
		DurationMs: 142,
		ExitCode:   0,
		InputHashes: map[string]string{
			"main.go": "hash123",
		},
		EnvVars: map[string]string{
			"GOOS": "windows",
		},
	}

	outputFiles := []string{
		filepath.ToSlash(filepath.Join("bin", "app.bin")),
		"version.txt",
	}

	err := Create(capsuleFile, manifest, wsDir, outputFiles)
	if err != nil {
		t.Fatalf("Create capsule failed: %v", err)
	}

	inspected, err := Inspect(capsuleFile)
	if err != nil {
		t.Fatalf("Inspect capsule failed: %v", err)
	}
	if inspected.BuildID != "bld-test-12345" || inspected.TaskName != "build:app" {
		t.Fatalf("Inspected manifest mismatch: %+v", inspected)
	}

	vRes, err := Verify(capsuleFile)
	if err != nil {
		t.Fatalf("Verify capsule failed: %v", err)
	}
	if !vRes.Valid || vRes.ArtifactCount != 2 {
		t.Fatalf("Verify failed or wrong artifact count: %+v", vRes)
	}

	replayDir := t.TempDir()
	rRes, err := Replay(capsuleFile, replayDir)
	if err != nil {
		t.Fatalf("Replay capsule failed: %v", err)
	}

	if !rRes.VerifiedMatch || len(rRes.RestoredFiles) != 2 {
		t.Fatalf("Replay failed verification match: %+v", rRes)
	}
}

func TestCapsule_Ed25519SignedAndAuthenticity(t *testing.T) {
	wsDir := t.TempDir()
	appFile := filepath.Join(wsDir, "server.exe")
	_ = os.WriteFile(appFile, []byte("production-signed-executable"), 0755)

	pub, priv, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("failed to generate ed25519 keypair: %v", err)
	}

	capsuleFile := filepath.Join(t.TempDir(), "signed_build.zcap")
	manifest := Manifest{
		BuildID:   "bld-signed-777",
		TaskName:  "build:prod",
		Timestamp: time.Now().UTC(),
		ExitCode:  0,
	}

	// 1. Create with digital signature
	err = CreateSigned(capsuleFile, manifest, wsDir, []string{"server.exe"}, priv)
	if err != nil {
		t.Fatalf("CreateSigned failed: %v", err)
	}

	// 2. Verify with matching trusted public key
	vRes, err := VerifyWithKey(capsuleFile, pub)
	if err != nil {
		t.Fatalf("VerifyWithKey failed: %v", err)
	}
	if !vRes.Valid || !vRes.SignatureVerified {
		t.Fatalf("expected valid signature, got: %+v", vRes)
	}

	// 3. Verify with wrong public key -> must fail signature check
	otherPub, _, _ := GenerateKeypair()
	vWrong, err := VerifyWithKey(capsuleFile, otherPub)
	if err != nil {
		t.Fatalf("unexpected error on VerifyWithKey: %v", err)
	}
	if vWrong.Valid || vWrong.SignatureVerified {
		t.Fatalf("verification should have failed with wrong public key!")
	}
}

func TestCapsule_TamperedManifestFails(t *testing.T) {
	wsDir := t.TempDir()
	appFile := filepath.Join(wsDir, "app.bin")
	_ = os.WriteFile(appFile, []byte("legit-binary"), 0755)

	pub, priv, _ := GenerateKeypair()
	capsuleFile := filepath.Join(t.TempDir(), "tamper_test.zcap")

	manifest := Manifest{
		BuildID:  "bld-original",
		TaskName: "build",
	}

	_ = CreateSigned(capsuleFile, manifest, wsDir, []string{"app.bin"}, priv)

	// Tamper: unpack, modify manifest exit_code to 1, re-pack with original signature
	inspected, _ := Inspect(capsuleFile)
	inspected.ExitCode = 1 // Tampered!

	// Check that signature verification fails on tampered manifest
	valid, err := inspected.VerifySignature(pub)
	if valid || err == nil {
		t.Fatalf("expected tampered manifest to fail Ed25519 verification!")
	}
}

func TestCapsule_TarSlipSecurityRejection(t *testing.T) {
	// Craft a malicious archive containing "artifacts/../../evil.txt"
	maliciousZcap := filepath.Join(t.TempDir(), "malicious.zcap")

	f, _ := os.Create(maliciousZcap)
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	manifestBytes, _ := json.Marshal(Manifest{
		BuildID:  "evil-build",
		TaskName: "attack",
	})

	_ = tw.WriteHeader(&tar.Header{
		Name: ManifestFileName,
		Mode: 0644,
		Size: int64(len(manifestBytes)),
	})
	_, _ = tw.Write(manifestBytes)

	// Malicious slip entry
	slipPayload := []byte("pwned!")
	_ = tw.WriteHeader(&tar.Header{
		Name: "artifacts/../../evil.txt",
		Mode: 0644,
		Size: int64(len(slipPayload)),
	})
	_, _ = tw.Write(slipPayload)

	tw.Close()
	gw.Close()
	f.Close()

	// Replay must detect and reject Tar Slip attempt!
	replayDir := t.TempDir()
	_, err := Replay(maliciousZcap, replayDir)
	if err == nil {
		t.Fatalf("SECURITY FAILURE: Tar Slip path traversal was permitted!")
	}

	// Verify evil.txt was NOT written outside replayDir
	evilPath := filepath.Join(filepath.Dir(replayDir), "evil.txt")
	if _, err := os.Stat(evilPath); err == nil {
		t.Fatalf("SECURITY FAILURE: Malicious file was written outside target dir: %s", evilPath)
	}
}

func TestCapsule_KeypairSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	privPath := filepath.Join(tmpDir, "zephyr.key")
	pubPath := filepath.Join(tmpDir, "zephyr.pub")

	pub, priv, err := GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair failed: %v", err)
	}

	if err := SaveKeypair(privPath, pubPath, priv, pub); err != nil {
		t.Fatalf("SaveKeypair failed: %v", err)
	}

	loadedPriv, err := LoadPrivateKey(privPath)
	if err != nil {
		t.Fatalf("LoadPrivateKey failed: %v", err)
	}

	loadedPub, err := LoadPublicKey(pubPath)
	if err != nil {
		t.Fatalf("LoadPublicKey failed: %v", err)
	}

	if string(loadedPriv) != string(priv) || string(loadedPub) != string(pub) {
		t.Fatalf("loaded key mismatch")
	}
}
