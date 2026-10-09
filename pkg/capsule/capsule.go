package capsule

import (
	"archive/tar"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	CurrentCapsuleVersion = "2.0.0"
	ManifestFileName      = "capsule.json"
)

// Manifest holds full build event metadata, inputs, outputs, toolchain, provenance, and digital signatures.
type Manifest struct {
	CapsuleVersion string            `json:"capsule_version"`
	BuildID        string            `json:"build_id"`
	TaskName       string            `json:"task_name"`
	Command        string            `json:"command"`
	Cwd            string            `json:"cwd"`
	Timestamp      time.Time         `json:"timestamp"`
	DurationMs     int64             `json:"duration_ms"`
	ExitCode       int               `json:"exit_code"`
	Platform       string            `json:"platform"`
	Toolchain      string            `json:"toolchain"`
	InputHashes    map[string]string `json:"input_hashes"`
	EnvVars        map[string]string `json:"env_vars"`
	OutputHashes   map[string]string `json:"output_hashes"`
	ProvenanceJSON string            `json:"provenance,omitempty"`
	Stdout         string            `json:"stdout,omitempty"`
	Stderr         string            `json:"stderr,omitempty"`
	SignerPubKey   string            `json:"signer_pubkey,omitempty"` // Hex-encoded Ed25519 public key
	Signature      string            `json:"signature,omitempty"`     // Hex-encoded Ed25519 signature
}

// CanonicalBytes returns the byte payload for digital signature computation, omitting signature fields.
func (m *Manifest) CanonicalBytes() ([]byte, error) {
	copyM := *m
	copyM.SignerPubKey = ""
	copyM.Signature = ""
	return json.Marshal(copyM)
}

// Sign calculates an Ed25519 signature over canonical manifest bytes.
func (m *Manifest) Sign(privKey ed25519.PrivateKey) error {
	if len(privKey) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid ed25519 private key size: expected %d, got %d", ed25519.PrivateKeySize, len(privKey))
	}
	pubKey := privKey.Public().(ed25519.PublicKey)
	m.SignerPubKey = hex.EncodeToString(pubKey)

	canonical, err := m.CanonicalBytes()
	if err != nil {
		return fmt.Errorf("failed to canonicalize manifest for signing: %w", err)
	}

	sig := ed25519.Sign(privKey, canonical)
	m.Signature = hex.EncodeToString(sig)
	return nil
}

// VerifySignature validates that the manifest signature is authentic and matches the public key.
func (m *Manifest) VerifySignature(trustedPubKey ed25519.PublicKey) (bool, error) {
	if m.Signature == "" {
		return false, fmt.Errorf("capsule manifest is not digitally signed")
	}

	sigBytes, err := hex.DecodeString(m.Signature)
	if err != nil {
		return false, fmt.Errorf("invalid signature hex encoding: %w", err)
	}

	var pubKey ed25519.PublicKey
	if len(trustedPubKey) > 0 {
		pubKey = trustedPubKey
		if m.SignerPubKey != "" {
			expectedHex := hex.EncodeToString(trustedPubKey)
			if m.SignerPubKey != expectedHex {
				return false, fmt.Errorf("signer public key '%s' does not match trusted public key '%s'", m.SignerPubKey, expectedHex)
			}
		}
	} else {
		if m.SignerPubKey == "" {
			return false, fmt.Errorf("no signer public key present in manifest or provided")
		}
		pubBytes, err := hex.DecodeString(m.SignerPubKey)
		if err != nil {
			return false, fmt.Errorf("invalid signer public key hex: %w", err)
		}
		pubKey = pubBytes
	}

	canonical, err := m.CanonicalBytes()
	if err != nil {
		return false, err
	}

	valid := ed25519.Verify(pubKey, canonical, sigBytes)
	if !valid {
		return false, fmt.Errorf("cryptographic signature mismatch: manifest has been modified or signature is invalid")
	}
	return true, nil
}

// Key generation and persistence utilities
func GenerateKeypair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

func SaveKeypair(privPath, pubPath string, priv ed25519.PrivateKey, pub ed25519.PublicKey) error {
	_ = os.MkdirAll(filepath.Dir(privPath), 0755)
	_ = os.MkdirAll(filepath.Dir(pubPath), 0755)
	if err := os.WriteFile(privPath, []byte(hex.EncodeToString(priv)), 0600); err != nil {
		return err
	}
	return os.WriteFile(pubPath, []byte(hex.EncodeToString(pub)), 0644)
}

func LoadPrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoded, err := hex.DecodeString(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key hex: %w", err)
	}
	if len(decoded) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid ed25519 private key size: expected %d, got %d", ed25519.PrivateKeySize, len(decoded))
	}
	return ed25519.PrivateKey(decoded), nil
}

func LoadPublicKey(path string) (ed25519.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoded, err := hex.DecodeString(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key hex: %w", err)
	}
	if len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid ed25519 public key size: expected %d, got %d", ed25519.PublicKeySize, len(decoded))
	}
	return ed25519.PublicKey(decoded), nil
}

// VerificationResult contains the result of verifying a Build Capsule's integrity and authenticity.
type VerificationResult struct {
	CapsulePath       string            `json:"capsule_path"`
	BuildID           string            `json:"build_id"`
	TaskName          string            `json:"task_name"`
	Valid             bool              `json:"valid"`
	SignatureVerified bool              `json:"signature_verified"`
	SignerPubKey      string            `json:"signer_pubkey,omitempty"`
	ArtifactCount     int               `json:"artifact_count"`
	VerifiedFiles     map[string]string `json:"verified_files"` // path -> sha256
	Errors            []string          `json:"errors,omitempty"`
}

// ReplayResult captures the outcome of restoring artifacts from a Build Capsule.
type ReplayResult struct {
	CapsulePath   string        `json:"capsule_path"`
	BuildID       string        `json:"build_id"`
	TaskName      string        `json:"task_name"`
	RestoredFiles []string      `json:"restored_files"`
	Duration      time.Duration `json:"duration"`
	TargetDir     string        `json:"target_dir"`
	VerifiedMatch bool          `json:"verified_match"`
}

// Create builds a portable .zcap archive bundling manifest, logs, and declared output files.
func Create(outputPath string, manifest Manifest, workspaceRoot string, outputFiles []string) error {
	return CreateSigned(outputPath, manifest, workspaceRoot, outputFiles, nil)
}

// CreateSigned builds a portable .zcap archive and digitally signs it with an Ed25519 private key.
func CreateSigned(outputPath string, manifest Manifest, workspaceRoot string, outputFiles []string, signingKey ed25519.PrivateKey) error {
	if manifest.CapsuleVersion == "" {
		manifest.CapsuleVersion = CurrentCapsuleVersion
	}
	if manifest.Platform == "" {
		manifest.Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	}
	if manifest.Toolchain == "" {
		manifest.Toolchain = runtime.Version()
	}
	if manifest.OutputHashes == nil {
		manifest.OutputHashes = make(map[string]string)
	}

	// Calculate and record hashes for all output files if not already done
	for _, relPath := range outputFiles {
		fullPath := filepath.Join(workspaceRoot, relPath)
		if hash, err := hashFile(fullPath); err == nil {
			manifest.OutputHashes[filepath.ToSlash(relPath)] = hash
		}
	}

	// Sign manifest if key provided
	if len(signingKey) > 0 {
		if err := manifest.Sign(signingKey); err != nil {
			return fmt.Errorf("failed to digitally sign capsule manifest: %w", err)
		}
	}

	_ = os.MkdirAll(filepath.Dir(outputPath), 0755)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create capsule file '%s': %w", outputPath, err)
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// 1. Write capsule.json manifest
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize manifest: %w", err)
	}

	manifestHeader := &tar.Header{
		Name:    ManifestFileName,
		Mode:    0644,
		Size:    int64(len(manifestBytes)),
		ModTime: manifest.Timestamp,
	}
	if err := tw.WriteHeader(manifestHeader); err != nil {
		return fmt.Errorf("failed to write manifest header: %w", err)
	}
	if _, err := tw.Write(manifestBytes); err != nil {
		return fmt.Errorf("failed to write manifest body: %w", err)
	}

	// 2. Write output artifacts
	for _, relPath := range outputFiles {
		fullPath := filepath.Join(workspaceRoot, relPath)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue // Skip missing files
		}
		if info.IsDir() {
			continue
		}

		f, err := os.Open(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read artifact '%s': %w", relPath, err)
		}

		header := &tar.Header{
			Name:    filepath.ToSlash(filepath.Join("artifacts", relPath)),
			Mode:    int64(info.Mode()),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		if err := tw.WriteHeader(header); err != nil {
			f.Close()
			return fmt.Errorf("failed to write tar header for '%s': %w", relPath, err)
		}

		if _, err := io.Copy(tw, f); err != nil {
			f.Close()
			return fmt.Errorf("failed to write artifact data for '%s': %w", relPath, err)
		}
		f.Close()
	}

	return nil
}

// Inspect reads the manifest from a .zcap archive without writing files to disk.
func Inspect(archivePath string) (*Manifest, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open capsule: %w", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("invalid capsule format (not gzip): %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading capsule archive: %w", err)
		}

		if header.Name == ManifestFileName {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed to read manifest: %w", err)
			}
			var m Manifest
			if err := json.Unmarshal(data, &m); err != nil {
				return nil, fmt.Errorf("corrupt manifest JSON: %w", err)
			}
			return &m, nil
		}
	}

	return nil, fmt.Errorf("capsule does not contain a valid %s", ManifestFileName)
}

// Verify checks the cryptographic checksums of all bundled artifacts against the manifest.
func Verify(archivePath string) (*VerificationResult, error) {
	return VerifyWithKey(archivePath, nil)
}

// VerifyWithKey checks checksums and verifies Ed25519 digital signature against trustedKey (if provided).
func VerifyWithKey(archivePath string, trustedKey ed25519.PublicKey) (*VerificationResult, error) {
	manifest, err := Inspect(archivePath)
	if err != nil {
		return nil, err
	}

	res := &VerificationResult{
		CapsulePath:   archivePath,
		BuildID:       manifest.BuildID,
		TaskName:      manifest.TaskName,
		Valid:         true,
		SignerPubKey:  manifest.SignerPubKey,
		VerifiedFiles: make(map[string]string),
	}

	// 1. Digital Signature Authenticity Verification
	if manifest.Signature != "" {
		validSig, sigErr := manifest.VerifySignature(trustedKey)
		if sigErr != nil || !validSig {
			res.Valid = false
			res.SignatureVerified = false
			res.Errors = append(res.Errors, fmt.Sprintf("Ed25519 signature verification failed: %v", sigErr))
		} else {
			res.SignatureVerified = true
		}
	} else if len(trustedKey) > 0 {
		res.Valid = false
		res.SignatureVerified = false
		res.Errors = append(res.Errors, "trusted public key provided, but capsule has no digital signature")
	}

	// 2. SHA-256 Artifact Integrity Verification
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			res.Valid = false
			res.Errors = append(res.Errors, err.Error())
			return res, nil
		}

		if strings.HasPrefix(header.Name, "artifacts/") {
			relPath := filepath.ToSlash(strings.TrimPrefix(header.Name, "artifacts/"))
			h := sha256.New()
			if _, err := io.Copy(h, tr); err != nil {
				res.Valid = false
				res.Errors = append(res.Errors, fmt.Sprintf("failed to read artifact '%s': %v", relPath, err))
				continue
			}
			computedHash := hex.EncodeToString(h.Sum(nil))
			res.VerifiedFiles[relPath] = computedHash
			res.ArtifactCount++

			expectedHash, exists := manifest.OutputHashes[relPath]
			if exists && expectedHash != computedHash {
				res.Valid = false
				res.Errors = append(res.Errors, fmt.Sprintf("checksum mismatch for '%s': expected %s, got %s", relPath, expectedHash, computedHash))
			}
		}
	}

	// Verify that all declared expected outputs were actually present in the archive
	for expectedPath := range manifest.OutputHashes {
		if _, found := res.VerifiedFiles[expectedPath]; !found {
			res.Valid = false
			res.Errors = append(res.Errors, fmt.Sprintf("missing declared artifact '%s' in capsule archive", expectedPath))
		}
	}

	return res, nil
}

// Replay restores output artifacts into targetDir and validates their integrity with strict Zip/Tar Slip protection.
func Replay(archivePath string, targetDir string) (*ReplayResult, error) {
	startTime := time.Now()
	manifest, err := Inspect(archivePath)
	if err != nil {
		return nil, err
	}

	targetAbs, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("invalid replay destination path '%s': %w", targetDir, err)
	}

	f, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	res := &ReplayResult{
		CapsulePath:   archivePath,
		BuildID:       manifest.BuildID,
		TaskName:      manifest.TaskName,
		TargetDir:     targetAbs,
		VerifiedMatch: true,
	}

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading capsule during replay: %w", err)
		}

		if strings.HasPrefix(header.Name, "artifacts/") {
			rawRel := strings.TrimPrefix(header.Name, "artifacts/")

			// CRITICAL SECURITY JAIL: Prevent Zip Slip / Tar Slip directory traversal
			cleanRel := filepath.Clean(filepath.FromSlash(rawRel))
			if strings.HasPrefix(cleanRel, "..") || filepath.IsAbs(cleanRel) || strings.HasPrefix(cleanRel, string(filepath.Separator)) {
				return nil, fmt.Errorf("security violation: capsule contains illegal path traversal artifact '%s'", rawRel)
			}

			destPath := filepath.Join(targetAbs, cleanRel)
			relToTarget, err := filepath.Rel(targetAbs, destPath)
			if err != nil || strings.HasPrefix(relToTarget, "..") || strings.HasPrefix(relToTarget, "/..") || strings.HasPrefix(relToTarget, "\\..") {
				return nil, fmt.Errorf("security violation: extracted artifact '%s' escapes destination directory '%s'", rawRel, targetAbs)
			}

			_ = os.MkdirAll(filepath.Dir(destPath), 0755)
			outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return nil, fmt.Errorf("failed to create destination artifact '%s': %w", destPath, err)
			}

			hasher := sha256.New()
			mw := io.MultiWriter(outFile, hasher)
			if _, err := io.Copy(mw, tr); err != nil {
				outFile.Close()
				return nil, fmt.Errorf("failed to write destination artifact '%s': %w", destPath, err)
			}
			outFile.Close()

			computedHash := hex.EncodeToString(hasher.Sum(nil))
			if expectedHash, ok := manifest.OutputHashes[filepath.ToSlash(cleanRel)]; ok {
				if expectedHash != computedHash {
					res.VerifiedMatch = false
				}
			}

			res.RestoredFiles = append(res.RestoredFiles, cleanRel)
		}
	}

	res.Duration = time.Since(startTime)
	return res, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
