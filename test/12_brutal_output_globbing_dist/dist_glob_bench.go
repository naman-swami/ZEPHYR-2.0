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

type GlobResult struct {
	Name     string        `json:"name"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration_ms"`
	Detail   string        `json:"detail"`
}

func main() {
	zephyrBin, _ := filepath.Abs(`..\..\zephyr.exe`)
	var results []GlobResult

	fmt.Println("=== RUNNING BRUTAL SUITE 12: OUTPUT GLOBBING IN 'dist/' DIRECTORY ===")

	testDir, _ := filepath.Abs("dist_glob_ws")
	_ = os.RemoveAll(testDir)
	_ = os.MkdirAll(testDir, 0755)

	genScript := `package main
import (
	"fmt"
	"os"
	"path/filepath"
)
func main() {
	_ = os.MkdirAll("dist", 0755)
	for i := 1; i <= 10; i++ {
		_ = os.WriteFile(filepath.Join("dist", fmt.Sprintf("bundle_%02d.js", i)), []byte(fmt.Sprintf("console.log('bundle %d');", i)), 0644)
	}
}`
	_ = os.WriteFile(filepath.Join(testDir, "gen.go"), []byte(genScript), 0644)

	cfg := map[string]interface{}{
		"default": "bundle_task",
		"tasks": map[string]interface{}{
			"bundle_task": map[string]interface{}{
				"command": "go run gen.go",
				"outputs": []string{"dist/*.js"},
			},
		},
	}
	cfgBytes, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile(filepath.Join(testDir, "tasks.json"), cfgBytes, 0644)

	start := time.Now()
	// Cold Run: executes gen.go and harvests outputs under dist/*.js
	cmdCold := exec.Command(zephyrBin, "-quiet")
	cmdCold.Dir = testDir
	cmdCold.Env = append(os.Environ(), "NO_COLOR=1")
	outCold, errCold := cmdCold.CombinedOutput()
	durCold := time.Since(start)

	// Inspect Action Cache manifest to verify output_hashes is populated!
	acDir := filepath.Join(testDir, ".taskcache", "ac")
	acEntries, _ := os.ReadDir(acDir)
	var outputHashesCount int
	if len(acEntries) > 0 {
		data, _ := os.ReadFile(filepath.Join(acDir, acEntries[0].Name()))
		var manifest map[string]interface{}
		_ = json.Unmarshal(data, &manifest)
		hashes, _ := manifest["output_hashes"].(map[string]interface{})
		outputHashesCount = len(hashes)
	}

	// Remove dist files to test warm restoration from CAS
	_ = os.RemoveAll(filepath.Join(testDir, "dist"))
	_ = os.MkdirAll(filepath.Join(testDir, "dist"), 0755)

	cmdWarm := exec.Command(zephyrBin, "-quiet")
	cmdWarm.Dir = testDir
	cmdWarm.Env = append(os.Environ(), "NO_COLOR=1")
	outWarm, _ := cmdWarm.CombinedOutput()

	files, _ := os.ReadDir(filepath.Join(testDir, "dist"))
	restoredCount := len(files)
	cached := strings.Contains(string(outWarm), "CACHED")

	passed := errCold == nil && outputHashesCount == 10 && restoredCount == 10 && cached
	detail := fmt.Sprintf("dist/*.js outputs discovered: %d/10. Restored from CAS: %d/10. Warm CACHED: %v. Output: %s",
		outputHashesCount, restoredCount, cached, strings.TrimSpace(string(outCold)))

	results = append(results, GlobResult{
		Name:     "Declared Output Globbing in 'dist/' (patternMentionsDir patch)",
		Passed:   passed,
		Duration: durCold,
		Detail:   detail,
	})
	fmt.Printf("[%s] dist/ Output Globbing: %s\n", status(passed), detail)

	reportBytes, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile("brutal_dist_glob_results.json", reportBytes, 0644)
	fmt.Println("Brutal Suite 12 Complete. Results written to brutal_dist_glob_results.json")
}

func status(b bool) string {
	if b {
		return "PASS"
	}
	return "FAIL"
}
