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

type FsEdgeResult struct {
	Name     string        `json:"name"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration_ms"`
	Detail   string        `json:"detail"`
}

func main() {
	zephyrBin, _ := filepath.Abs(`..\..\zephyr.exe`)
	var results []FsEdgeResult

	fmt.Println("=== RUNNING BRUTAL SUITE 13: FILESYSTEM EDGE CASES (ZERO-BYTE & UNICODE) ===")

	testDir, _ := filepath.Abs("fs_edge_ws")
	_ = os.RemoveAll(testDir)
	_ = os.MkdirAll(testDir, 0755)

	// 1. Zero-byte artifact task
	cfg := map[string]interface{}{
		"default": "empty_task",
		"tasks": map[string]interface{}{
			"empty_task": map[string]interface{}{
				"command": `cmd /c type nul > empty.bin`,
				"outputs": []string{"empty.bin"},
				"inputs":  []string{"nonexistent_wildcard/**/*.xyz"}, // Nonexistent glob
			},
			"unicode_⚡_task": map[string]interface{}{
				"command": `cmd /c echo UTF8_UNICODE_PAYLOAD > utf8.txt`,
				"outputs": []string{"utf8.txt"},
			},
		},
	}
	cfgBytes, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile(filepath.Join(testDir, "tasks.json"), cfgBytes, 0644)

	// Cold run on empty_task
	start := time.Now()
	cmd1 := exec.Command(zephyrBin, "-quiet", "empty_task")
	cmd1.Dir = testDir
	cmd1.Env = append(os.Environ(), "NO_COLOR=1")
	out1, err1 := cmd1.CombinedOutput()
	dur1 := time.Since(start)

	// Verify zero-byte output was created
	fi, statErr := os.Stat(filepath.Join(testDir, "empty.bin"))
	zeroByteCreated := statErr == nil && fi.Size() == 0

	// Remove empty.bin and test warm CAS restore of zero-byte file
	_ = os.Remove(filepath.Join(testDir, "empty.bin"))
	cmd2 := exec.Command(zephyrBin, "-quiet", "empty_task")
	cmd2.Dir = testDir
	cmd2.Env = append(os.Environ(), "NO_COLOR=1")
	out2, _ := cmd2.CombinedOutput()

	fi2, statErr2 := os.Stat(filepath.Join(testDir, "empty.bin"))
	zeroByteRestored := statErr2 == nil && fi2.Size() == 0
	cached := strings.Contains(string(out2), "CACHED")

	zeroBytePassed := err1 == nil && zeroByteCreated && zeroByteRestored && cached
	results = append(results, FsEdgeResult{
		Name:     "Zero-Byte Artifact CAS Hashing & Restoration",
		Passed:   zeroBytePassed,
		Duration: dur1,
		Detail:   fmt.Sprintf("Zero-byte file created: %v; Restored from CAS: %v; CACHED: %v (Out1: %s)", zeroByteCreated, zeroByteRestored, cached, strings.TrimSpace(string(out1))),
	})
	fmt.Printf("[%s] Zero-Byte Artifact CAS: %s\n", status(zeroBytePassed), results[len(results)-1].Detail)

	// Test Unicode task name
	cmdUni := exec.Command(zephyrBin, "-quiet", "unicode_⚡_task")
	cmdUni.Dir = testDir
	cmdUni.Env = append(os.Environ(), "NO_COLOR=1")
	outUni, errUni := cmdUni.CombinedOutput()
	uniPassed := errUni == nil && strings.Contains(string(outUni), "EXECUTED")

	results = append(results, FsEdgeResult{
		Name:     "Unicode UTF-8 Task Identifiers (unicode_⚡_task)",
		Passed:   uniPassed,
		Duration: 0,
		Detail:   fmt.Sprintf("Unicode task execution: %v (Out: %s)", uniPassed, strings.TrimSpace(string(outUni))),
	})
	fmt.Printf("[%s] Unicode Task Identifiers: %s\n", status(uniPassed), results[len(results)-1].Detail)

	reportBytes, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile("brutal_fs_edge_results.json", reportBytes, 0644)
	fmt.Println("Brutal Suite 13 Complete. Results written to brutal_fs_edge_results.json")
}

func status(b bool) string {
	if b {
		return "PASS"
	}
	return "FAIL"
}
