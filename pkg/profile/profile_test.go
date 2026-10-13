package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProfileCommandAndFlameGraphGeneration(t *testing.T) {
	start := time.Now()
	tp1 := ProfileCommand("compile:core", "go version", 1, start)
	if tp1.Status != "EXECUTED" {
		t.Fatalf("expected EXECUTED status, got %s", tp1.Status)
	}
	if tp1.WallDurationMs <= 0 {
		t.Errorf("expected positive WallDurationMs, got %f", tp1.WallDurationMs)
	}

	tp2 := TaskProfile{
		Name:              "bundle:dist",
		Command:           "go build .",
		Batch:             2,
		Status:            "CACHED",
		StartOffsetUs:     tp1.WallDurationUs,
		WallDurationUs:    1200,
		WallDurationMs:    1.2,
		UserCPUMs:         0.0,
		SysCPUMs:          0.1,
		TotalCPUMs:        0.1,
		CPUUtilizationPct: 8.3,
		PeakMemoryKB:      4096,
	}

	prof := &PipelineProfile{
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Platform:       "windows/amd64",
		CPUCores:       12,
		TotalWallMs:    tp1.WallDurationMs + 1.2,
		TotalUserCPUMs: tp1.UserCPUMs,
		TotalSysCPUMs:  tp1.SysCPUMs + 0.1,
		PeakMemoryKB:   tp1.PeakMemoryKB,
		Tasks:          []TaskProfile{tp1, tp2},
	}

	tmpDir := t.TempDir()
	svgPath := filepath.Join(tmpDir, "flamegraph.svg")
	tracePath := filepath.Join(tmpDir, "trace.json")

	if err := GenerateFlameGraphSVG(prof, svgPath); err != nil {
		t.Fatalf("GenerateFlameGraphSVG failed: %v", err)
	}
	if err := GenerateChromeTraceJSON(prof, tracePath); err != nil {
		t.Fatalf("GenerateChromeTraceJSON failed: %v", err)
	}

	svgBytes, err := os.ReadFile(svgPath)
	if err != nil {
		t.Fatal(err)
	}
	svgStr := string(svgBytes)
	if !strings.Contains(svgStr, "<svg") || !strings.Contains(svgStr, "compile:core") || !strings.Contains(svgStr, "bundle:dist") {
		t.Errorf("SVG flamegraph missing expected elements: %s", svgStr)
	}

	traceBytes, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(traceBytes), "traceEvents") {
		t.Errorf("Chrome trace JSON missing traceEvents")
	}
}
