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
