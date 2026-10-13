package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ReproResult struct {
	Name     string        `json:"name"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration_ms"`
	Detail   string        `json:"detail"`
}

func main() {
	zephyrBin, _ := filepath.Abs(`..\..\zephyr.exe`)
	var results []ReproResult

	fmt.Println("=== RUNNING SUITE 07: REPRODUCIBILITY AUDITOR & BYTE-DIFF ENGINE ===")

	// ------------------------------------------------------------------------
	// Test 7.1: Deterministic Build Verification (Passes Audit)
	// ------------------------------------------------------------------------
	{
		testDir, _ := filepath.Abs("repro_clean_ws")
		_ = os.RemoveAll(testDir)
		_ = os.MkdirAll(testDir, 0755)

		_ = os.MkdirAll(filepath.Join(testDir, "bin"), 0755)

		cfg := map[string]interface{}{
			"default": "repro_task",
			"tasks": map[string]interface{}{
				"repro_task": map[string]interface{}{
					"command": `cmd /c if not exist bin mkdir bin & echo CONSTANT_REPRODUCIBLE_BYTE_STREAM > bin\app.bin`,
					"outputs": []string{"bin/app.bin"},
				},
			},
		}
		cfgBytes, _ := json.MarshalIndent(cfg, "", "  ")
		_ = os.WriteFile(filepath.Join(testDir, "tasks.json"), cfgBytes, 0644)

		start := time.Now()
		cmd := exec.Command(zephyrBin, "verify", "-runs", "2", "-json", "repro_task")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		out, err := cmd.CombinedOutput()
		dur := time.Since(start)

		output := string(out)
		var report map[string]interface{}
		_ = json.Unmarshal(out, &report)
		isDeterministic, _ := report["is_deterministic"].(bool)

		passed := err == nil && isDeterministic
		detail := fmt.Sprintf("Auditor confirmed byte-identical determinism across runs (is_deterministic=%v)", isDeterministic)
		if !passed {
			detail = fmt.Sprintf("Verification failed: %s", output)
		}

		results = append(results, ReproResult{
			Name:     "Deterministic Reproducibility Verification",
			Passed:   passed,
			Duration: dur,
			Detail:   detail,
		})
		fmt.Printf("[%s] Deterministic Build Verification: %s\n", status(passed), detail)
	}

	// ------------------------------------------------------------------------
	// Test 7.2: Non-Deterministic Build Detection (Byte Divergence Pinpointed)
	// ------------------------------------------------------------------------
	{
		testDir, _ := filepath.Abs("non_det_ws")
		_ = os.RemoveAll(testDir)
		_ = os.MkdirAll(testDir, 0755)

		_ = os.MkdirAll(filepath.Join(testDir, "bin"), 0755)

		randScript := `package main
import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
)
func main() {
	_ = os.MkdirAll("bin", 0755)
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	_ = os.WriteFile(filepath.Join("bin", "rand.bin"), []byte(hex.EncodeToString(b)), 0644)
}`
		_ = os.WriteFile(filepath.Join(testDir, "rand.go"), []byte(randScript), 0644)

		cfg := map[string]interface{}{
			"default": "flaky_task",
			"tasks": map[string]interface{}{
				"flaky_task": map[string]interface{}{
					"command": "go run rand.go",
					"outputs": []string{"bin/rand.bin"},
				},
			},
		}
		cfgBytes, _ := json.MarshalIndent(cfg, "", "  ")
		_ = os.WriteFile(filepath.Join(testDir, "tasks.json"), cfgBytes, 0644)

		start := time.Now()
		cmd := exec.Command(zephyrBin, "verify", "-runs", "2", "flaky_task")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		out, err := cmd.CombinedOutput()
		dur := time.Since(start)

		output := string(out)
		divergenceDetected := err != nil && (strings.Contains(output, "NON-DETERMINISTIC") || strings.Contains(output, "Divergence") || strings.Contains(output, "FAIL") || strings.Contains(output, "Offset"))
		detail := fmt.Sprintf("Auditor detected non-deterministic artifact divergence and reported byte diff (Detected=%v)", divergenceDetected)

		results = append(results, ReproResult{
			Name:     "Non-Deterministic Artifact Divergence Detection",
			Passed:   divergenceDetected,
			Duration: dur,
			Detail:   detail,
		})
		fmt.Printf("[%s] Non-Deterministic Divergence Detection: %s\n", status(divergenceDetected), detail)
	}

	reportBytes, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile("repro_auditor_results.json", reportBytes, 0644)
	fmt.Println("\nSuite 07 Reproducibility Auditor Tests Complete. Results written to repro_auditor_results.json")
}

func status(b bool) string {
	if b {
		return "PASS"
	}
	return "FAIL"
}
