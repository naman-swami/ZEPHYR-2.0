package verify

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAuditor_DeterministicTask(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	_ = os.WriteFile(srcFile, []byte("deterministic content 123"), 0644)

	auditor := NewAuditor(tmpDir)

	task := TaskSpec{
		Name:    "compile-test",
		Command: "go version > out.txt",
		Outputs: []string{"out.txt"},
	}

	report, err := auditor.AuditTask(context.Background(), task, 2)
	if err != nil {
		t.Fatalf("AuditTask failed: %v", err)
	}

	if !report.IsDeterministic {
		t.Fatalf("expected task to be deterministic, got: %+v", report)
	}

	if len(report.Artifacts) != 1 || !report.Artifacts[0].Match {
		t.Fatalf("expected 1 matching artifact, got: %+v", report.Artifacts)
	}

	if report.Notice != ReproducibilityNotice {
		t.Fatalf("expected exact reproducibility notice")
	}
}

func TestAuditor_NonDeterministicTask(t *testing.T) {
	tmpDir := t.TempDir()
	auditor := NewAuditor(tmpDir)

	// A task that writes time-dependent output
	task := TaskSpec{
		Name:    "time-dependent",
		Command: "powershell -Command \"Get-Date -Format 'HH:mm:ss.ffffff' | Out-File -Encoding ascii out.txt; Start-Sleep -Milliseconds 10\"",
		Shell:   "powershell",
		Outputs: []string{"out.txt"},
	}

	report, err := auditor.AuditTask(context.Background(), task, 2)
	if err != nil {
		t.Fatalf("AuditTask failed: %v", err)
	}

	if report.IsDeterministic {
		t.Fatalf("expected task to be flagged as non-deterministic due to timestamps")
	}

	if len(report.DivergenceReasons) == 0 {
		t.Fatalf("expected divergence reasons to be populated")
	}
}

func TestCloneDir_HardlinkZeroCopy(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")
	_ = os.MkdirAll(srcDir, 0755)

	srcFile := filepath.Join(srcDir, "sample.bin")
	content := []byte("payload for hardlink verification")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if err := cloneDir(srcDir, dstDir); err != nil {
		t.Fatalf("cloneDir failed: %v", err)
	}

	dstFile := filepath.Join(dstDir, "sample.bin")
	dstContent, err := os.ReadFile(dstFile)
	if err != nil || string(dstContent) != string(content) {
		t.Fatalf("cloned content mismatch: %s != %s", string(dstContent), string(content))
	}

	// Verify that os.SameFile returns true if hardlinking succeeded on the same volume
	fiSrc, err := os.Stat(srcFile)
	if err != nil {
		t.Fatalf("stat src failed: %v", err)
	}
	fiDst, err := os.Stat(dstFile)
	if err != nil {
		t.Fatalf("stat dst failed: %v", err)
	}

	if !os.SameFile(fiSrc, fiDst) {
		t.Logf("note: hardlinks not supported on current volume or cross-device, fell back to stream copy")
	} else {
		t.Logf("verified: zero-copy hardlink forest successfully linked (same inode)")
	}
}

