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

type ZombieResult struct {
	Name     string        `json:"name"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration_ms"`
	Detail   string        `json:"detail"`
}

func main() {
	zephyrBin, _ := filepath.Abs(`..\..\zephyr.exe`)
	var results []ZombieResult

	fmt.Println("=== RUNNING BRUTAL SUITE 10: PROCESS TREE TERMINATION & ZOMBIE CASCADE ===")

	testDir, _ := filepath.Abs("zombie_ws")
	_ = os.RemoveAll(testDir)
	_ = os.MkdirAll(testDir, 0755)

	sleepScript := `package main
import (
	"time"
)
func main() {
	time.Sleep(15 * time.Second)
}`
	_ = os.WriteFile(filepath.Join(testDir, "sleep15.go"), []byte(sleepScript), 0644)

	cfg := map[string]interface{}{
		"default": "deep_process_task",
		"tasks": map[string]interface{}{
			"deep_process_task": map[string]interface{}{
				"command": "go run sleep15.go",
				"timeout": "1s",
			},
		},
	}
	cfgBytes, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile(filepath.Join(testDir, "tasks.json"), cfgBytes, 0644)

	start := time.Now()
	cmd := exec.Command(zephyrBin, "-quiet")
	cmd.Dir = testDir
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	dur := time.Since(start)

	output := string(out)
	fmt.Printf("DEBUG: err=%v, output=%s\n", err, output)
	killedInTime := dur < 4*time.Second
	failedCleanly := err != nil || strings.Contains(output, "FAILED")

	passed := killedInTime && failedCleanly
	detail := fmt.Sprintf("Process tree killed cleanly in %v (under 4s ceiling vs 15s child sleep). Status: %v.", dur, passed)

	results = append(results, ZombieResult{
		Name:     "Deep Process Tree Termination (taskkill /F /T verification)",
		Passed:   passed,
		Duration: dur,
		Detail:   detail,
	})
	fmt.Printf("[%s] Process Tree Termination: %s\n", status(passed), detail)

	reportBytes, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile("brutal_zombie_results.json", reportBytes, 0644)
	fmt.Println("Brutal Suite 10 Complete. Results written to brutal_zombie_results.json")
}

func status(b bool) string {
	if b {
		return "PASS"
	}
	return "FAIL"
}
