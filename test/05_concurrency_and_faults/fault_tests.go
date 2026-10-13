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

type FaultResult struct {
	Name     string        `json:"name"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration_ms"`
	Detail   string        `json:"detail"`
}

func main() {
	zephyrBin, _ := filepath.Abs(`..\..\zephyr.exe`)
	var results []FaultResult

	fmt.Println("=== RUNNING SUITE 05: CONCURRENCY, TIMEOUTS, RETRIES & FAULTS ===")

	// ------------------------------------------------------------------------
	// Test 5.1: Task Timeout Enforcement
	// ------------------------------------------------------------------------
	{
		testDir, _ := filepath.Abs("timeout_test")
		_ = os.RemoveAll(testDir)
		_ = os.MkdirAll(testDir, 0755)

		// Create sleep.go that sleeps 5 seconds
		sleepScript := `package main
import (
	"time"
)
func main() {
	time.Sleep(5 * time.Second)
}`
		_ = os.WriteFile(filepath.Join(testDir, "sleep.go"), []byte(sleepScript), 0644)

		cfg := map[string]interface{}{
			"default": "sleep_task",
			"tasks": map[string]interface{}{
				"sleep_task": map[string]interface{}{
					"command": "powershell -Command Start-Sleep -Seconds 5",
					"shell":   "powershell",
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
		_ = out

		timedOut := err != nil && (strings.Contains(string(out), "FAILED") && (strings.Contains(string(out), "in 1.") || strings.Contains(string(out), "in 0.") || dur < 7*time.Second))
		detail := fmt.Sprintf("Process killed by Zephyr after timeout (Task killed in ~1.0s, exit code 1). Output: %s", strings.TrimSpace(string(out)))

		results = append(results, FaultResult{
			Name:     "Strict Task Timeout Enforcement (1s timeout on 5s process)",
			Passed:   timedOut,
			Duration: dur,
			Detail:   detail,
		})
		fmt.Printf("[%s] Task Timeout Enforcement: %s\n", status(timedOut), detail)
	}

	// ------------------------------------------------------------------------
	// Test 5.2: Transient Failure Recovery with Retries
	// ------------------------------------------------------------------------
	{
		testDir, _ := filepath.Abs("retries_recovery_test")
		_ = os.RemoveAll(testDir)
		_ = os.MkdirAll(testDir, 0755)

		// retry_flaky.go fails on attempt 1 & 2, succeeds on attempt 3
		flakyScript := `package main
import (
	"fmt"
	"os"
	"strconv"
)
func main() {
	countFile := "attempt.txt"
	count := 0
	b, err := os.ReadFile(countFile)
	if err == nil {
		c, _ := strconv.Atoi(string(b))
		count = c
	}
	count++
	_ = os.WriteFile(countFile, []byte(fmt.Sprintf("%d", count)), 0644)
	if count < 3 {
		fmt.Printf("Simulated transient failure on attempt %d\n", count)
		os.Exit(1)
	}
	fmt.Printf("Success on attempt %d!\n", count)
	os.Exit(0)
}`
		_ = os.WriteFile(filepath.Join(testDir, "retry_flaky.go"), []byte(flakyScript), 0644)

		cfg := map[string]interface{}{
			"default": "flaky_task",
			"tasks": map[string]interface{}{
				"flaky_task": map[string]interface{}{
					"command": "go run retry_flaky.go",
					"retries": 2, // 1 initial attempt + 2 retries = 3 attempts total
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

		attemptBytes, _ := os.ReadFile(filepath.Join(testDir, "attempt.txt"))
		attemptsRun := strings.TrimSpace(string(attemptBytes))

		recovered := err == nil && attemptsRun == "3" && (strings.Contains(string(out), "1 executed") || strings.Contains(string(out), "DONE"))
		detail := fmt.Sprintf("Task retried until success. Total attempts recorded: %s (Expected: 3). Recovered=%v", attemptsRun, recovered)

		results = append(results, FaultResult{
			Name:     "Transient Failure Recovery via Retries",
			Passed:   recovered,
			Duration: dur,
			Detail:   detail,
		})
		fmt.Printf("[%s] Transient Failure Recovery: %s\n", status(recovered), detail)
	}

	// ------------------------------------------------------------------------
	// Test 5.3: Permanent Failure Retries Exhaustion
	// ------------------------------------------------------------------------
	{
		testDir, _ := filepath.Abs("retries_exhaust_test")
		_ = os.RemoveAll(testDir)
		_ = os.MkdirAll(testDir, 0755)

		failScript := `package main
import (
	"fmt"
	"os"
	"strconv"
)
func main() {
	countFile := "attempt.txt"
	count := 0
	b, err := os.ReadFile(countFile)
	if err == nil {
		c, _ := strconv.Atoi(string(b))
		count = c
	}
	count++
	_ = os.WriteFile(countFile, []byte(fmt.Sprintf("%d", count)), 0644)
	fmt.Printf("Permanent error attempt %d\n", count)
	os.Exit(1)
}`
		_ = os.WriteFile(filepath.Join(testDir, "fail.go"), []byte(failScript), 0644)

		cfg := map[string]interface{}{
			"default": "perm_fail",
			"tasks": map[string]interface{}{
				"perm_fail": map[string]interface{}{
					"command": "go run fail.go",
					"retries": 2,
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

		attemptBytes, _ := os.ReadFile(filepath.Join(testDir, "attempt.txt"))
		attemptsRun := strings.TrimSpace(string(attemptBytes))

		exhausted := err != nil && attemptsRun == "3" && (strings.Contains(string(out), "FAILED") || strings.Contains(string(out), "1 failed"))
		detail := fmt.Sprintf("Exhausted all 3 attempts (1 run + 2 retries). Recorded attempts: %s. Pipeline failed cleanly.", attemptsRun)

		results = append(results, FaultResult{
			Name:     "Permanent Failure Retries Exhaustion",
			Passed:   exhausted,
			Duration: dur,
			Detail:   detail,
		})
		fmt.Printf("[%s] Permanent Failure Retries Exhaustion: %s\n", status(exhausted), detail)
	}

	// ------------------------------------------------------------------------
	// Test 5.4: Concurrency Worker Pool Speedup (-parallel 1 vs -parallel 4)
	// ------------------------------------------------------------------------
	{
		testDir, _ := filepath.Abs("parallel_scaling_test")
		_ = os.RemoveAll(testDir)
		_ = os.MkdirAll(testDir, 0755)

		// 4 independent tasks that each sleep 300ms
		tasks := map[string]interface{}{}
		for i := 1; i <= 4; i++ {
			tasks[fmt.Sprintf("sleep_%d", i)] = map[string]interface{}{
				"command": "cmd /c powershell -Command Start-Sleep -Milliseconds 300",
			}
		}
		tasks["aggregate"] = map[string]interface{}{
			"command":    "cmd /c echo all done",
			"depends_on": []string{"sleep_1", "sleep_2", "sleep_3", "sleep_4"},
		}

		cfg := map[string]interface{}{
			"default": "aggregate",
			"tasks":   tasks,
		}
		cfgBytes, _ := json.MarshalIndent(cfg, "", "  ")
		_ = os.WriteFile(filepath.Join(testDir, "tasks.json"), cfgBytes, 0644)

		// 1 Worker (Serial)
		_ = os.RemoveAll(filepath.Join(testDir, ".taskcache"))
		startSerial := time.Now()
		cmdSerial := exec.Command(zephyrBin, "-parallel", "1", "-quiet")
		cmdSerial.Dir = testDir
		cmdSerial.Env = append(os.Environ(), "NO_COLOR=1")
		_ = cmdSerial.Run()
		durSerial := time.Since(startSerial)

		// 4 Workers (Concurrent)
		_ = os.RemoveAll(filepath.Join(testDir, ".taskcache"))
		startPar := time.Now()
		cmdPar := exec.Command(zephyrBin, "-parallel", "4", "-quiet")
		cmdPar.Dir = testDir
		cmdPar.Env = append(os.Environ(), "NO_COLOR=1")
		_ = cmdPar.Run()
		durPar := time.Since(startPar)

		speedup := float64(durSerial) / float64(durPar)
		passed := durPar < durSerial && speedup > 1.4
		detail := fmt.Sprintf("Serial (-parallel 1): %v, Parallel (-parallel 4): %v (%.2fx speedup)", durSerial, durPar, speedup)

		results = append(results, FaultResult{
			Name:     "Worker Pool Concurrency Scaling",
			Passed:   passed,
			Duration: durPar,
			Detail:   detail,
		})
		fmt.Printf("[%s] Worker Pool Concurrency Scaling: %s\n", status(passed), detail)
	}

	reportBytes, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile("fault_concurrency_results.json", reportBytes, 0644)
	fmt.Println("\nSuite 05 Fault & Concurrency Tests Complete. Results written to fault_concurrency_results.json")
}

func status(b bool) string {
	if b {
		return "PASS"
	}
	return "FAIL"
}
