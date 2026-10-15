package stress

import (
	"os"
	"testing"
)

func TestGauntlet(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	report, err := RunGauntlet(wd)
	if err != nil {
		t.Fatalf("RunGauntlet returned unexpected error: %v", err)
	}

	if report.TotalFailed > 0 {
		for _, s := range report.Scenarios {
			if !s.Passed {
				t.Errorf("Scenario '%s' failed: %s", s.Name, s.Details)
			}
		}
	}

	if report.TotalPassed != len(report.Scenarios) {
		t.Errorf("expected %d passed scenarios, got %d", len(report.Scenarios), report.TotalPassed)
	}
}
