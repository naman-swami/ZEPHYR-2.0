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

type MonoResult struct {
	Name     string        `json:"name"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration_ms"`
	Detail   string        `json:"detail"`
}

func main() {
	zephyrBin, _ := filepath.Abs(`..\..\zephyr.exe`)
	var results []MonoResult

	fmt.Println("=== RUNNING BRUTAL SUITE 14: MULTI-WORKSPACE MONOREPO CROSS-PACKAGE CYCLES ===")

	testDir, _ := filepath.Abs("monorepo_ws")
	_ = os.RemoveAll(testDir)
	_ = os.MkdirAll(testDir, 0755)

	pkgADir := filepath.Join(testDir, "packages", "pkg-a")
	pkgBDir := filepath.Join(testDir, "packages", "pkg-b")
	pkgCDir := filepath.Join(testDir, "packages", "pkg-c")

	_ = os.MkdirAll(pkgADir, 0755)
	_ = os.MkdirAll(pkgBDir, 0755)
	_ = os.MkdirAll(pkgCDir, 0755)

	// pkg-a: build depends on pkg-b#build
	cfgA := map[string]interface{}{
		"tasks": map[string]interface{}{
			"build": map[string]interface{}{
				"command":    "echo building A",
				"depends_on": []string{"pkg-b#build"},
			},
		},
	}
	bA, _ := json.MarshalIndent(cfgA, "", "  ")
	_ = os.WriteFile(filepath.Join(pkgADir, "tasks.json"), bA, 0644)

	// pkg-b: build depends on pkg-c#build
	cfgB := map[string]interface{}{
		"tasks": map[string]interface{}{
			"build": map[string]interface{}{
				"command":    "echo building B",
				"depends_on": []string{"pkg-c#build"},
			},
		},
	}
	bB, _ := json.MarshalIndent(cfgB, "", "  ")
	_ = os.WriteFile(filepath.Join(pkgBDir, "tasks.json"), bB, 0644)

	// pkg-c: build depends on pkg-a#build (Creates cross-package cycle: A -> B -> C -> A)
	cfgC := map[string]interface{}{
		"tasks": map[string]interface{}{
			"build": map[string]interface{}{
				"command":    "echo building C",
				"depends_on": []string{"pkg-a#build"},
			},
		},
	}
	bC, _ := json.MarshalIndent(cfgC, "", "  ")
	_ = os.WriteFile(filepath.Join(pkgCDir, "tasks.json"), bC, 0644)

	// Root config with workspaces glob
	rootCfg := map[string]interface{}{
		"workspaces": []string{"packages/*"},
		"default":    "pkg-a#build",
	}
	bRoot, _ := json.MarshalIndent(rootCfg, "", "  ")
	_ = os.WriteFile(filepath.Join(testDir, "tasks.json"), bRoot, 0644)

	start := time.Now()
	cmd := exec.Command(zephyrBin, "-quiet")
	cmd.Dir = testDir
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	dur := time.Since(start)

	output := string(out)
	cycleCaught := err != nil && strings.Contains(output, "cyclic dependency detected")

	passed := cycleCaught
	detail := fmt.Sprintf("Cross-package cycle detected: %v. Output: %s", cycleCaught, strings.TrimSpace(output))

	results = append(results, MonoResult{
		Name:     "Cross-Package Monorepo Cycle Detection (A -> B -> C -> A)",
		Passed:   passed,
		Duration: dur,
		Detail:   detail,
	})
	fmt.Printf("[%s] Monorepo Cross-Package Cycle: %s\n", status(passed), detail)

	reportBytes, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile("brutal_monorepo_results.json", reportBytes, 0644)
	fmt.Println("Brutal Suite 14 Complete. Results written to brutal_monorepo_results.json")
}

func status(b bool) string {
	if b {
		return "PASS"
	}
	return "FAIL"
}
