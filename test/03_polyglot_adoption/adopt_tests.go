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

type AdoptResult struct {
	Name        string        `json:"name"`
	Passed      bool          `json:"passed"`
	ProjectType string        `json:"project_type"`
	Duration    time.Duration `json:"duration_ms"`
	Detail      string        `json:"detail"`
}

func main() {
	zephyrBin, _ := filepath.Abs(`..\..\zephyr.exe`)
	var results []AdoptResult

	fmt.Println("=== RUNNING SUITE 03: POLYGLOT ADOPTION & AUTO-DISCOVERY ===")

	// ------------------------------------------------------------------------
	// Test 3.1: Node.js package.json auto-adoption
	// ------------------------------------------------------------------------
	{
		testDir, _ := filepath.Abs("node_proj")
		_ = os.RemoveAll(testDir)
		_ = os.MkdirAll(testDir, 0755)

		pkgJSON := `{
  "name": "sample-node-app",
  "version": "1.0.0",
  "scripts": {
    "lint": "eslint src/",
    "test": "jest",
    "build": "esbuild src/index.ts --outdir=dist"
  }
}`
		_ = os.WriteFile(filepath.Join(testDir, "package.json"), []byte(pkgJSON), 0644)

		start := time.Now()
		cmd := exec.Command(zephyrBin, "adopt", "--json")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		out, err := cmd.CombinedOutput()
		dur := time.Since(start)

		var parsed map[string]interface{}
		_ = json.Unmarshal(out, &parsed)

		tasks, _ := parsed["tasks"].(map[string]interface{})
		_, hasBuild := tasks["build"]
		_, hasTest := tasks["test"]
		_, hasLint := tasks["lint"]

		passed := err == nil && hasBuild && hasTest && hasLint
		detail := fmt.Sprintf("Discovered Node.js project. Tasks found: build=%v, test=%v, lint=%v", hasBuild, hasTest, hasLint)

		results = append(results, AdoptResult{
			Name:        "Node.js package.json Auto-Discovery",
			Passed:      passed,
			ProjectType: "Node.js",
			Duration:    dur,
			Detail:      detail,
		})
		fmt.Printf("[%s] Node.js Adoption: %s (Time: %v)\n", status(passed), detail, dur)
	}

	// ------------------------------------------------------------------------
	// Test 3.2: Python pyproject.toml auto-adoption
	// ------------------------------------------------------------------------
	{
		testDir, _ := filepath.Abs("python_proj")
		_ = os.RemoveAll(testDir)
		_ = os.MkdirAll(testDir, 0755)

		pyproject := `[project]
name = "sample-python-pkg"
version = "0.1.0"
dependencies = [
    "pytest>=7.0",
    "flake8>=5.0"
]
`
		_ = os.WriteFile(filepath.Join(testDir, "pyproject.toml"), []byte(pyproject), 0644)

		start := time.Now()
		cmd := exec.Command(zephyrBin, "adopt", "--json")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		out, err := cmd.CombinedOutput()
		dur := time.Since(start)

		var parsed map[string]interface{}
		_ = json.Unmarshal(out, &parsed)
		tasks, _ := parsed["tasks"].(map[string]interface{})

		passed := err == nil && (len(tasks) > 0)
		detail := fmt.Sprintf("Discovered Python project with %d generated tasks", len(tasks))

		results = append(results, AdoptResult{
			Name:        "Python pyproject.toml Auto-Discovery",
			Passed:      passed,
			ProjectType: "Python",
			Duration:    dur,
			Detail:      detail,
		})
		fmt.Printf("[%s] Python Adoption: %s (Time: %v)\n", status(passed), detail, dur)
	}

	// ------------------------------------------------------------------------
	// Test 3.3: Rust Cargo.toml auto-adoption
	// ------------------------------------------------------------------------
	{
		testDir, _ := filepath.Abs("rust_proj")
		_ = os.RemoveAll(testDir)
		_ = os.MkdirAll(testDir, 0755)

		cargo := `[package]
name = "sample_rust_crate"
version = "0.1.0"
edition = "2021"

[dependencies]
`
		_ = os.WriteFile(filepath.Join(testDir, "Cargo.toml"), []byte(cargo), 0644)

		start := time.Now()
		cmd := exec.Command(zephyrBin, "adopt", "--json")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		out, err := cmd.CombinedOutput()
		dur := time.Since(start)

		var parsed map[string]interface{}
		_ = json.Unmarshal(out, &parsed)
		tasks, _ := parsed["tasks"].(map[string]interface{})

		passed := err == nil && (len(tasks) > 0)
		detail := fmt.Sprintf("Discovered Rust project with %d generated cargo tasks", len(tasks))

		results = append(results, AdoptResult{
			Name:        "Rust Cargo.toml Auto-Discovery",
			Passed:      passed,
			ProjectType: "Rust",
			Duration:    dur,
			Detail:      detail,
		})
		fmt.Printf("[%s] Rust Adoption: %s (Time: %v)\n", status(passed), detail, dur)
	}

	// ------------------------------------------------------------------------
	// Test 3.4: Go go.mod auto-adoption
	// ------------------------------------------------------------------------
	{
		testDir, _ := filepath.Abs("go_proj")
		_ = os.RemoveAll(testDir)
		_ = os.MkdirAll(testDir, 0755)

		goMod := `module testgo

go 1.27
`
		_ = os.WriteFile(filepath.Join(testDir, "go.mod"), []byte(goMod), 0644)

		start := time.Now()
		cmd := exec.Command(zephyrBin, "adopt", "--json")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		out, err := cmd.CombinedOutput()
		dur := time.Since(start)

		var parsed map[string]interface{}
		_ = json.Unmarshal(out, &parsed)
		tasks, _ := parsed["tasks"].(map[string]interface{})

		_, hasBuild := tasks["build"]
		_, hasTest := tasks["test"]
		_, hasLint := tasks["lint"]

		passed := err == nil && hasBuild && hasTest && hasLint
		detail := fmt.Sprintf("Discovered Go module with standard lint, test, build targets")

		results = append(results, AdoptResult{
			Name:        "Go go.mod Auto-Discovery",
			Passed:      passed,
			ProjectType: "Go",
			Duration:    dur,
			Detail:      detail,
		})
		fmt.Printf("[%s] Go Adoption: %s (Time: %v)\n", status(passed), detail, dur)
	}

	// ------------------------------------------------------------------------
	// Test 3.5: Makefile auto-adoption
	// ------------------------------------------------------------------------
	{
		testDir, _ := filepath.Abs("make_proj")
		_ = os.RemoveAll(testDir)
		_ = os.MkdirAll(testDir, 0755)

		makefile := `all: build

build: compile
	echo "Building"

compile:
	echo "Compiling"
`
		_ = os.WriteFile(filepath.Join(testDir, "Makefile"), []byte(makefile), 0644)

		start := time.Now()
		cmd := exec.Command(zephyrBin, "adopt", "--json")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		out, err := cmd.CombinedOutput()
		dur := time.Since(start)

		var parsed map[string]interface{}
		_ = json.Unmarshal(out, &parsed)
		tasks, _ := parsed["tasks"].(map[string]interface{})

		passed := err == nil && len(tasks) > 0
		detail := fmt.Sprintf("Discovered Makefile project with %d parsed targets", len(tasks))

		results = append(results, AdoptResult{
			Name:        "Legacy Makefile Auto-Discovery",
			Passed:      passed,
			ProjectType: "Makefile",
			Duration:    dur,
			Detail:      detail,
		})
		fmt.Printf("[%s] Makefile Adoption: %s (Time: %v)\n", status(passed), detail, dur)
	}

	// ------------------------------------------------------------------------
	// Test 3.6: Overwrite Safety Protection
	// ------------------------------------------------------------------------
	{
		testDir, _ := filepath.Abs("safety_proj")
		_ = os.RemoveAll(testDir)
		_ = os.MkdirAll(testDir, 0755)

		_ = os.WriteFile(filepath.Join(testDir, "go.mod"), []byte("module safetytest\n\ngo 1.27\n"), 0644)
		existingTasks := `{"tasks":{"existing":{"command":"echo safe"}}}`
		_ = os.WriteFile(filepath.Join(testDir, "tasks.json"), []byte(existingTasks), 0644)

		// 1. Run adopt --write without --force
		cmdSafe := exec.Command(zephyrBin, "adopt", "--write")
		cmdSafe.Dir = testDir
		cmdSafe.Env = append(os.Environ(), "NO_COLOR=1")
		outSafe, errSafe := cmdSafe.CombinedOutput()

		blocked := errSafe != nil && strings.Contains(string(outSafe), "Safety Protection")

		// 2. Run adopt --write --force
		cmdForce := exec.Command(zephyrBin, "adopt", "--write", "--force")
		cmdForce.Dir = testDir
		cmdForce.Env = append(os.Environ(), "NO_COLOR=1")
		outForce, errForce := cmdForce.CombinedOutput()

		forcePassed := errForce == nil && strings.Contains(string(outForce), "Successfully wrote proposed configuration")

		passed := blocked && forcePassed
		detail := fmt.Sprintf("Blocked overwrite without force: %v; Allowed overwrite with --force: %v", blocked, forcePassed)

		results = append(results, AdoptResult{
			Name:        "Safe Overwrite Protection",
			Passed:      passed,
			ProjectType: "Policy Check",
			Duration:    0,
			Detail:      detail,
		})
		fmt.Printf("[%s] Overwrite Safety Protection: %s\n", status(passed), detail)
	}

	// ------------------------------------------------------------------------
	// Test 3.7: Corrupted Project Manifest Handling
	// ------------------------------------------------------------------------
	{
		testDir, _ := filepath.Abs("corrupt_proj")
		_ = os.RemoveAll(testDir)
		_ = os.MkdirAll(testDir, 0755)

		_ = os.WriteFile(filepath.Join(testDir, "package.json"), []byte(`{ "name": "broken", invalid json`), 0644)

		cmd := exec.Command(zephyrBin, "adopt")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		out, err := cmd.CombinedOutput()

		// Should report error cleanly without crashing/panicking
		graceful := err != nil && !strings.Contains(string(out), "panic:")
		detail := fmt.Sprintf("Handled corrupted package.json without unhandled panic (Graceful=%v)", graceful)

		results = append(results, AdoptResult{
			Name:        "Corrupted Manifest Graceful Resilience",
			Passed:      graceful,
			ProjectType: "Error Handling",
			Duration:    0,
			Detail:      detail,
		})
		fmt.Printf("[%s] Corrupted Manifest Resilience: %s\n", status(graceful), detail)
	}

	reportBytes, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile("polyglot_adoption_results.json", reportBytes, 0644)
	fmt.Println("\nSuite 03 Polyglot Adoption Tests Complete. Results written to polyglot_adoption_results.json")
}

func status(b bool) string {
	if b {
		return "PASS"
	}
	return "FAIL"
}
