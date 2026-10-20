package organize

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClassifyFile(t *testing.T) {
	tests := []struct {
		path     string
		isRoot   bool
		expected FileCategory
	}{
		{"go.mod", true, CategoryRootAnchor},
		{"tasks.json", true, CategoryRootAnchor},
		{"README.md", true, CategoryRootAnchor},
		{"LICENSE", true, CategoryRootAnchor},
		{"architecture.md", true, CategoryDocs},
		{"banner.png", true, CategoryAssets},
		{"deploy.sh", true, CategoryScripts},
		{"docker-compose.yml", true, CategoryConfigs},
		{"payload.mock.json", true, CategoryTestData},
		{"server.go", true, CategoryCode},
		{"app.py", true, CategoryCode},
		{"client.ts", true, CategoryCode},
	}

	for _, tt := range tests {
		got := ClassifyFile(tt.path, tt.isRoot)
		if got != tt.expected {
			t.Errorf("ClassifyFile(%q, %v) = %v, want %v", tt.path, tt.isRoot, got, tt.expected)
		}
	}
}

func TestTarjanDetectCycles(t *testing.T) {
	// 1. Acyclic Graph
	nodes := []string{"models", "services", "api"}
	edges := map[string][]string{
		"api":      {"services"},
		"services": {"models"},
		"models":   {},
	}
	cycles := DetectCycles(nodes, edges)
	if len(cycles) > 0 {
		t.Fatalf("expected 0 cycles in acyclic graph, got %v", cycles)
	}

	// 2. Cyclic Graph: A -> B -> A
	cyclicNodes := []string{"pkgA", "pkgB"}
	cyclicEdges := map[string][]string{
		"pkgA": {"pkgB"},
		"pkgB": {"pkgA"},
	}
	cycles = DetectCycles(cyclicNodes, cyclicEdges)
	if len(cycles) != 1 {
		t.Fatalf("expected 1 cycle in cyclic graph, got %v", cycles)
	}
}

func TestAnalyzeAndPlan(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "zephyr_org_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create sample project structure with sprawl
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testmod\n\ngo 1.22\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# Test Project\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "notes.md"), []byte("Some notes\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "logo.png"), []byte("fake_image_bytes"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "run.sh"), []byte("#!/bin/bash\necho ok\n"), 0755)

	mainGoContent := `package main

import "fmt"

func main() {
	fmt.Println("hello world")
}
`
	_ = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(mainGoContent), 0644)

	// 1. Analyze
	report, err := Analyze(tempDir)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if report.TotalFiles < 5 {
		t.Errorf("expected at least 5 files, got %d", report.TotalFiles)
	}
	if report.TokenOverheadListing <= 0 {
		t.Errorf("expected positive token overhead, got %d", report.TokenOverheadListing)
	}

	// 2. Plan
	plan, err := Plan(tempDir, "standard-go")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}
	if len(plan.Relocations) == 0 {
		t.Fatalf("expected planned relocations, got 0")
	}

	// Verify root anchors were preserved
	for _, op := range plan.Relocations {
		if op.SourcePath == "go.mod" || op.SourcePath == "README.md" {
			t.Errorf("root anchor %s was planned for relocation!", op.SourcePath)
		}
	}
}

func TestApplyAndRollback(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "zephyr_org_apply_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module testmod\n\ngo 1.22\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "guide.md"), []byte("# Guide\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "icon.png"), []byte("icon_bytes"), 0644)

	plan, err := Plan(tempDir, "standard-go")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	// Apply without verification (synthetic test files)
	err = Apply(plan, false)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Verify files moved to target directories
	destDocs := filepath.Join(tempDir, "docs", "guide.md")
	if _, err := os.Stat(destDocs); os.IsNotExist(err) {
		t.Errorf("expected %s to exist after apply", destDocs)
	}

	destAssets := filepath.Join(tempDir, "assets", "icon.png")
	if _, err := os.Stat(destAssets); os.IsNotExist(err) {
		t.Errorf("expected %s to exist after apply", destAssets)
	}

	// Now Rollback!
	journal, err := Undo(tempDir)
	if err != nil {
		t.Fatalf("Undo failed: %v", err)
	}
	if journal.Status != "REVERTED" {
		t.Errorf("expected status REVERTED, got %s", journal.Status)
	}

	// Verify original files are back
	origDocs := filepath.Join(tempDir, "guide.md")
	if _, err := os.Stat(origDocs); os.IsNotExist(err) {
		t.Errorf("expected %s to be restored after rollback", origDocs)
	}

	origAssets := filepath.Join(tempDir, "icon.png")
	if _, err := os.Stat(origAssets); os.IsNotExist(err) {
		t.Errorf("expected %s to be restored after rollback", origAssets)
	}
}
