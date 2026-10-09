package adopt

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAdopt_GoProject(t *testing.T) {
	tmpDir := t.TempDir()
	goMod := filepath.Join(tmpDir, "go.mod")
	_ = os.WriteFile(goMod, []byte("module github.com/user/mycoolapp\n\ngo 1.22\n"), 0644)

	proj, err := AutoDiscover(tmpDir)
	if err != nil {
		t.Fatalf("AutoDiscover failed: %v", err)
	}

	if proj.ProjectType != "Go" {
		t.Fatalf("expected project type Go, got %s", proj.ProjectType)
	}

	if len(proj.Tasks) != 3 {
		t.Fatalf("expected 3 tasks for Go project, got %d", len(proj.Tasks))
	}

	buildTask, exists := proj.Tasks["build"]
	if !exists || buildTask.Command != "go build -o bin/mycoolapp ." {
		t.Fatalf("unexpected build task: %+v", buildTask)
	}

	configJSON, err := proj.RenderProposedConfig()
	if err != nil {
		t.Fatalf("RenderProposedConfig failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(configJSON, &parsed); err != nil {
		t.Fatalf("generated config is invalid JSON: %v", err)
	}
}

func TestAdopt_NodeProject(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJSON := filepath.Join(tmpDir, "package.json")
	content := `{
  "name": "frontend-web",
  "scripts": {
    "lint": "eslint .",
    "test": "jest",
    "build": "vite build"
  }
}`
	_ = os.WriteFile(pkgJSON, []byte(content), 0644)

	proj, err := AutoDiscover(tmpDir)
	if err != nil {
		t.Fatalf("AutoDiscover failed: %v", err)
	}

	if proj.ProjectType != "Node.js" {
		t.Fatalf("expected Node.js, got %s", proj.ProjectType)
	}

	if len(proj.Tasks) != 3 {
		t.Fatalf("expected 3 tasks for Node project, got %d", len(proj.Tasks))
	}

	if proj.Tasks["build"].DependsOn[0] != "test" {
		t.Fatalf("expected build to depend on test, got %v", proj.Tasks["build"].DependsOn)
	}
}

func TestAdopt_RustProject(t *testing.T) {
	tmpDir := t.TempDir()
	cargoToml := filepath.Join(tmpDir, "Cargo.toml")
	content := `[package]
name = "zephyr-core"
version = "0.1.0"
edition = "2021"

[dependencies]
`
	_ = os.WriteFile(cargoToml, []byte(content), 0644)

	proj, err := AutoDiscover(tmpDir)
	if err != nil {
		t.Fatalf("AutoDiscover failed: %v", err)
	}

	if proj.ProjectType != "Rust" {
		t.Fatalf("expected Rust project, got %s", proj.ProjectType)
	}

	expectedTasks := []string{"check", "test", "build"}
	for _, name := range expectedTasks {
		if _, exists := proj.Tasks[name]; !exists {
			t.Fatalf("expected task '%s' for Rust project", name)
		}
	}

	if proj.Tasks["build"].DependsOn[0] != "test" {
		t.Fatalf("expected build to depend on test in Rust project")
	}
}

func TestAdopt_PythonProject(t *testing.T) {
	tmpDir := t.TempDir()
	pyproject := filepath.Join(tmpDir, "pyproject.toml")
	content := `[build-system]
requires = ["setuptools"]
build-backend = "setuptools.build_meta"

[project]
name = "ml-pipeline"
version = "1.0.0"
`
	_ = os.WriteFile(pyproject, []byte(content), 0644)

	proj, err := AutoDiscover(tmpDir)
	if err != nil {
		t.Fatalf("AutoDiscover failed: %v", err)
	}

	if proj.ProjectType != "Python" {
		t.Fatalf("expected Python project, got %s", proj.ProjectType)
	}

	if _, exists := proj.Tasks["test"]; !exists {
		t.Fatalf("expected test task for Python project")
	}
	if _, exists := proj.Tasks["lint"]; !exists {
		t.Fatalf("expected lint task for Python project")
	}
}

func TestAdopt_MakefileProject(t *testing.T) {
	tmpDir := t.TempDir()
	makefile := filepath.Join(tmpDir, "Makefile")
	content := `# Sample Makefile
all: build test

build:
	echo "building binary"

test:
	echo "running tests"

clean:
	echo "cleaning up"
`
	_ = os.WriteFile(makefile, []byte(content), 0644)

	proj, err := AutoDiscover(tmpDir)
	if err != nil {
		t.Fatalf("AutoDiscover failed: %v", err)
	}

	if proj.ProjectType != "Makefile" {
		t.Fatalf("expected Makefile project, got %s", proj.ProjectType)
	}

	expectedTargets := []string{"all", "build", "test", "clean"}
	for _, target := range expectedTargets {
		if _, exists := proj.Tasks[target]; !exists {
			t.Fatalf("expected target '%s' in discovered Makefile tasks", target)
		}
	}
}
