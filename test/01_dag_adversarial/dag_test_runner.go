package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type TestResult struct {
	Name        string        `json:"name"`
	Passed      bool          `json:"passed"`
	ExitCode    int           `json:"exit_code"`
	Duration    time.Duration `json:"duration_ms"`
	Expected    string        `json:"expected"`
	Observed    string        `json:"observed"`
	ErrorDetail string        `json:"error_detail,omitempty"`
}

func main() {
	zephyrBin := `..\..\zephyr.exe`

	tests := []struct {
		name          string
		args          []string
		expectedExit  int
		expectContent string
	}{
		{
			name:          "Direct Cyclic Dependency (A -> A)",
			args:          []string{"-config", "cycle_direct.json"},
			expectedExit:  1,
			expectContent: "cyclic dependency detected: taskA -> taskA",
		},
		{
			name:          "Indirect Cyclic Dependency (A -> B -> C -> A)",
			args:          []string{"-config", "cycle_indirect.json"},
			expectedExit:  1,
			expectContent: "cyclic dependency detected",
		},
		{
			name:          "Missing/Ghost Dependency Detection",
			args:          []string{"-config", "ghost_dep.json"},
			expectedExit:  1,
			expectContent: "depends on unknown task 'phantom_task'",
		},
		{
			name:          "Corrupt JSON Syntax Parsing",
			args:          []string{"-config", "corrupt_syntax.json"},
			expectedExit:  1,
			expectContent: "failed to parse JSON",
		},
		{
			name:          "Ultra-Deep Linear Chain (50 Levels) Cold Run",
			args:          []string{"-config", "deep_chain.json", "-quiet"},
			expectedExit:  0,
			expectContent: "50 executed, 0 failed",
		},
		{
			name:          "Ultra-Deep Linear Chain (50 Levels) Warm Cache Replay",
			args:          []string{"-config", "deep_chain.json", "-quiet"},
			expectedExit:  0,
			expectContent: "50 cached, 0 executed, 0 failed",
		},
		{
			name:          "Wide Parallel Fan-Out (20 Workers -> 1 Aggregator)",
			args:          []string{"-config", "fanout_wide.json", "-quiet", "-parallel", "8"},
			expectedExit:  0,
			expectContent: "21 executed, 0 failed",
		},
		{
			name:          "Disconnected Subgraph Isolated Execution",
			args:          []string{"-config", "islands.json", "-quiet", "island2_y"},
			expectedExit:  0,
			expectContent: "2 executed, 0 failed",
		},
	}

	var results []TestResult
	allPassed := true

	fmt.Println("=== RUNNING SUITE 01: DAG & TOPOLOGY STRESS TESTS ===")

	for _, tc := range tests {
		if strings.Contains(tc.name, "Cold Run") || strings.Contains(tc.name, "Wide Parallel Fan-Out") || strings.Contains(tc.name, "Disconnected Subgraph") {
			_ = os.RemoveAll(".taskcache")
		}

		start := time.Now()
		cmd := exec.Command(zephyrBin, tc.args...)
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		outBytes, err := cmd.CombinedOutput()
		dur := time.Since(start)
		output := string(outBytes)

		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = -1
			}
		}

		passed := exitCode == tc.expectedExit && strings.Contains(output, tc.expectContent)
		if !passed {
			allPassed = false
		}

		res := TestResult{
			Name:        tc.name,
			Passed:      passed,
			ExitCode:    exitCode,
			Duration:    dur,
			Expected:    fmt.Sprintf("exit=%d, content contains '%s'", tc.expectedExit, tc.expectContent),
			Observed:    fmt.Sprintf("exit=%d", exitCode),
			ErrorDetail: "",
		}
		if !passed {
			res.ErrorDetail = fmt.Sprintf("Output preview: %s", truncate(output, 200))
		}
		results = append(results, res)

		status := "PASS"
		if !passed {
			status = "FAIL"
		}
		fmt.Printf("[%s] %s (Exit: %d, Time: %v)\n", status, tc.name, exitCode, dur)
	}

	reportBytes, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile("dag_stress_results.json", reportBytes, 0644)
	fmt.Printf("\nSuite 01 Complete. All Passed: %v. Results saved to dag_stress_results.json\n", allPassed)
}

func truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
