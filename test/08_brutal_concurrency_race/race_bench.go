package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type RaceResult struct {
	Name     string        `json:"name"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration_ms"`
	Detail   string        `json:"detail"`
}

func main() {
	zephyrBin, _ := filepath.Abs(`..\..\zephyr.exe`)
	var results []RaceResult

	fmt.Println("=== RUNNING BRUTAL SUITE 08: HIGH-CONCURRENCY CACHE RACE CONDITIONS ===")

	testDir, _ := filepath.Abs("race_ws")
	_ = os.RemoveAll(testDir)
	_ = os.MkdirAll(testDir, 0755)

	genScript := `package main
import (
	"os"
	"time"
)
func main() {
	time.Sleep(100 * time.Millisecond)
	_ = os.WriteFile("output.dat", []byte("CONCURRENT_HEAVY_PAYLOAD_DATA"), 0644)
}`
	_ = os.WriteFile(filepath.Join(testDir, "work.go"), []byte(genScript), 0644)

	cfg := map[string]interface{}{
		"default": "race_task",
		"tasks": map[string]interface{}{
			"race_task": map[string]interface{}{
				"command": "go run work.go",
				"outputs": []string{"output.dat"},
			},
		},
	}
	cfgBytes, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile(filepath.Join(testDir, "tasks.json"), cfgBytes, 0644)

	// Launch 10 simultaneous processes racing on cold cache
	const concurrency = 10
	var wg sync.WaitGroup
	startBarrier := make(chan struct{})
	errorsCh := make(chan error, concurrency)
	durations := make([]time.Duration, concurrency)

	startTime := time.Now()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-startBarrier

			t0 := time.Now()
			cmd := exec.Command(zephyrBin, "-quiet")
			cmd.Dir = testDir
			cmd.Env = append(os.Environ(), "NO_COLOR=1")
			out, err := cmd.CombinedOutput()
			durations[idx] = time.Since(t0)

			if err != nil {
				errorsCh <- fmt.Errorf("worker %d error: %v, out: %s", idx, err, strings.TrimSpace(string(out)))
			}
		}(i)
	}

	// Release all 10 processes simultaneously
	close(startBarrier)
	wg.Wait()
	close(errorsCh)
	totalDur := time.Since(startTime)

	var errList []string
	for e := range errorsCh {
		errList = append(errList, e.Error())
	}

	// Verify artifact integrity
	outBytes, _ := os.ReadFile(filepath.Join(testDir, "output.dat"))
	intact := string(outBytes) == "CONCURRENT_HEAVY_PAYLOAD_DATA"

	passed := len(errList) == 0 && intact
	detail := fmt.Sprintf("10 concurrent processes completed in %v. Errors: %d/%d. Output intact: %v",
		totalDur, len(errList), concurrency, intact)

	results = append(results, RaceResult{
		Name:     "10-Way Simultaneous Cold Cache Write Race",
		Passed:   passed,
		Duration: totalDur,
		Detail:   detail,
	})
	fmt.Printf("[%s] Cache Concurrency Race: %s\n", status(passed), detail)

	reportBytes, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile("brutal_race_results.json", reportBytes, 0644)
	fmt.Println("Brutal Suite 08 Complete. Results written to brutal_race_results.json")
}

func status(b bool) string {
	if b {
		return "PASS"
	}
	return "FAIL"
}
