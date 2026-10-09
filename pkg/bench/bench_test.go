package bench

import (
	"context"
	"testing"
)

func TestBenchmarkSuite_Execution(t *testing.T) {
	runner := NewRunner(t.TempDir())
	runner.Trials = 3 // Run 3 trials for quick test
	report, err := runner.RunBenchmark(context.Background())
	if err != nil {
		t.Fatalf("RunBenchmark failed: %v", err)
	}

	if report.CleanBuildStats.Median <= 0 {
		t.Errorf("expected clean build median > 0, got %v", report.CleanBuildStats.Median)
	}
	if report.TotalTasks != 5 || len(report.SkippedTasks) != 3 {
		t.Errorf("unexpected task skip stats: %+v", report)
	}
	if !report.ReproducibilityPassed {
		t.Errorf("expected reproducibility experiment to pass")
	}
	if !report.CapsuleIntegrityPassed || report.CapsuleFilesVerified != 1 {
		t.Errorf("expected capsule integrity to pass, got %+v", report)
	}

	formatted := FormatTerminalReport(report)
	if len(formatted) == 0 {
		t.Errorf("expected formatted report")
	}
}
