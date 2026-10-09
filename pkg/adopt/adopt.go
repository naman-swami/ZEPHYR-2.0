package adopt

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiscoveredTask represents an auto-discovered task ready to be converted into ZEPHYR configuration.
type DiscoveredTask struct {
	Name      string   `json:"name"`
	Command   string   `json:"command"`
	Inputs    []string `json:"inputs"`
	Outputs   []string `json:"outputs"`
	DependsOn []string `json:"depends_on,omitempty"`
	EnvVars   []string `json:"env_vars,omitempty"`
}

// DiscoveredProject captures all detected tasks and metadata for a repository.
type DiscoveredProject struct {
	ProjectType string                    `json:"project_type"`
	Root        string                    `json:"root"`
	Tasks       map[string]DiscoveredTask `json:"tasks"`
	DefaultTask string                    `json:"default_task"`
}

// Adapter defines the interface for language and framework project discovery.
type Adapter interface {
	Name() string
	Detect(root string) bool
	Discover(root string) (*DiscoveredProject, error)
}

// Built-in adapters
var RegisteredAdapters = []Adapter{
	&GoAdapter{},
	&NodeAdapter{},
	&RustAdapter{},
	&PythonAdapter{},
	&MakefileAdapter{},
}

// AutoDiscover inspects the workspace root and returns a discovered project configuration.
func AutoDiscover(root string) (*DiscoveredProject, error) {
	for _, adapter := range RegisteredAdapters {
		if adapter.Detect(root) {
			return adapter.Discover(root)
		}
	}
	return nil, fmt.Errorf("no recognized project manifest found in '%s' (checked Go, Node.js, Rust, Python, Makefile)", root)
}

// ============================================================================
// GO ADAPTER
// ============================================================================

type GoAdapter struct{}

func (a *GoAdapter) Name() string { return "Go" }

func (a *GoAdapter) Detect(root string) bool {
	_, err := os.Stat(filepath.Join(root, "go.mod"))
	return err == nil
}

func (a *GoAdapter) Discover(root string) (*DiscoveredProject, error) {
	modFile := filepath.Join(root, "go.mod")
	moduleName := "app"

	if data, err := os.ReadFile(modFile); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "module ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					moduleName = filepath.Base(parts[1])
				}
				break
			}
		}
	}

	tasks := map[string]DiscoveredTask{
		"lint": {
			Name:      "lint",
			Command:   "go vet ./...",
			Inputs:    []string{"**/*.go"},
			Outputs:   []string{},
			DependsOn: []string{},
		},
		"test": {
			Name:      "test",
			Command:   "go test -v ./...",
			Inputs:    []string{"**/*.go", "go.mod"},
			Outputs:   []string{},
			DependsOn: []string{"lint"},
		},
		"build": {
			Name:      "build",
			Command:   fmt.Sprintf("go build -o bin/%s .", moduleName),
			Inputs:    []string{"**/*.go", "go.mod"},
			Outputs:   []string{fmt.Sprintf("bin/%s*", moduleName)},
			DependsOn: []string{"test"},
		},
	}

	return &DiscoveredProject{
		ProjectType: "Go",
		Root:        root,
		Tasks:       tasks,
		DefaultTask: "build",
	}, nil
}

// ============================================================================
// NODE.JS ADAPTER
// ============================================================================

type NodeAdapter struct{}

func (a *NodeAdapter) Name() string { return "Node.js" }

func (a *NodeAdapter) Detect(root string) bool {
	_, err := os.Stat(filepath.Join(root, "package.json"))
	return err == nil
}

type packageJSON struct {
	Name    string            `json:"name"`
	Scripts map[string]string `json:"scripts"`
}

func (a *NodeAdapter) Discover(root string) (*DiscoveredProject, error) {
	pkgPath := filepath.Join(root, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("invalid package.json: %w", err)
	}

	tasks := make(map[string]DiscoveredTask)
	defaultTask := ""

	hasLint := false
	if _, ok := pkg.Scripts["lint"]; ok {
		hasLint = true
		tasks["lint"] = DiscoveredTask{
			Name:      "lint",
			Command:   "npm run lint",
			Inputs:    []string{"src/**", "package.json"},
			Outputs:   []string{},
			DependsOn: []string{},
		}
	}

	hasTest := false
	if _, ok := pkg.Scripts["test"]; ok {
		hasTest = true
		deps := []string{}
		if hasLint {
			deps = append(deps, "lint")
		}
		tasks["test"] = DiscoveredTask{
			Name:      "test",
			Command:   "npm test",
			Inputs:    []string{"src/**", "test/**", "tests/**", "package.json"},
			Outputs:   []string{},
			DependsOn: deps,
		}
	}

	if _, ok := pkg.Scripts["build"]; ok {
		deps := []string{}
		if hasTest {
			deps = append(deps, "test")
		} else if hasLint {
			deps = append(deps, "lint")
		}
		tasks["build"] = DiscoveredTask{
			Name:      "build",
			Command:   "npm run build",
			Inputs:    []string{"src/**", "package.json", "tsconfig.json"},
			Outputs:   []string{"dist/**", "build/**"},
			DependsOn: deps,
		}
		defaultTask = "build"
	} else if hasTest {
		defaultTask = "test"
	} else if hasLint {
		defaultTask = "lint"
	}

	return &DiscoveredProject{
		ProjectType: "Node.js",
		Root:        root,
		Tasks:       tasks,
		DefaultTask: defaultTask,
	}, nil
}

// ============================================================================
// RUST ADAPTER
// ============================================================================

type RustAdapter struct{}

func (a *RustAdapter) Name() string { return "Rust" }

func (a *RustAdapter) Detect(root string) bool {
	_, err := os.Stat(filepath.Join(root, "Cargo.toml"))
	return err == nil
}

func (a *RustAdapter) Discover(root string) (*DiscoveredProject, error) {
	tasks := map[string]DiscoveredTask{
		"check": {
			Name:      "check",
			Command:   "cargo check",
			Inputs:    []string{"src/**/*.rs", "Cargo.toml"},
			Outputs:   []string{},
			DependsOn: []string{},
		},
		"test": {
			Name:      "test",
			Command:   "cargo test",
			Inputs:    []string{"src/**/*.rs", "tests/**/*.rs", "Cargo.toml"},
			Outputs:   []string{},
			DependsOn: []string{"check"},
		},
		"build": {
			Name:      "build",
			Command:   "cargo build --release",
			Inputs:    []string{"src/**/*.rs", "Cargo.toml"},
			Outputs:   []string{"target/release/*"},
			DependsOn: []string{"test"},
		},
	}

	return &DiscoveredProject{
		ProjectType: "Rust",
		Root:        root,
		Tasks:       tasks,
		DefaultTask: "build",
	}, nil
}

// ============================================================================
// PYTHON ADAPTER
// ============================================================================

type PythonAdapter struct{}

func (a *PythonAdapter) Name() string { return "Python" }

func (a *PythonAdapter) Detect(root string) bool {
	for _, f := range []string{"pyproject.toml", "requirements.txt", "setup.py"} {
		if _, err := os.Stat(filepath.Join(root, f)); err == nil {
			return true
		}
	}
	return false
}

func (a *PythonAdapter) Discover(root string) (*DiscoveredProject, error) {
	tasks := map[string]DiscoveredTask{
		"lint": {
			Name:      "lint",
			Command:   "python -m ruff check .",
			Inputs:    []string{"**/*.py"},
			Outputs:   []string{},
			DependsOn: []string{},
		},
		"test": {
			Name:      "test",
			Command:   "pytest",
			Inputs:    []string{"**/*.py"},
			Outputs:   []string{},
			DependsOn: []string{"lint"},
		},
	}

	return &DiscoveredProject{
		ProjectType: "Python",
		Root:        root,
		Tasks:       tasks,
		DefaultTask: "test",
	}, nil
}

// ============================================================================
// MAKEFILE ADAPTER
// ============================================================================

type MakefileAdapter struct{}

func (a *MakefileAdapter) Name() string { return "Makefile" }

func (a *MakefileAdapter) Detect(root string) bool {
	_, err := os.Stat(filepath.Join(root, "Makefile"))
	return err == nil
}

func (a *MakefileAdapter) Discover(root string) (*DiscoveredProject, error) {
	makePath := filepath.Join(root, "Makefile")
	data, err := os.ReadFile(makePath)
	if err != nil {
		return nil, err
	}

	tasks := make(map[string]DiscoveredTask)
	defaultTask := ""

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ".") {
			continue
		}
		if idx := strings.Index(line, ":"); idx > 0 {
			target := strings.TrimSpace(line[:idx])
			if !strings.Contains(target, " ") && !strings.Contains(target, "$") {
				if defaultTask == "" {
					defaultTask = target
				}
				tasks[target] = DiscoveredTask{
					Name:      target,
					Command:   fmt.Sprintf("make %s", target),
					Inputs:    []string{"**/*"},
					Outputs:   []string{},
					DependsOn: []string{},
				}
			}
		}
	}

	return &DiscoveredProject{
		ProjectType: "Makefile",
		Root:        root,
		Tasks:       tasks,
		DefaultTask: defaultTask,
	}, nil
}

// RenderProposedConfig converts the discovered project into valid tasks.json JSON bytes.
func (p *DiscoveredProject) RenderProposedConfig() ([]byte, error) {
	type rawConfig struct {
		DefaultTask string                    `json:"default,omitempty"`
		Tasks       map[string]DiscoveredTask `json:"tasks"`
	}

	cfg := rawConfig{
		DefaultTask: p.DefaultTask,
		Tasks:       p.Tasks,
	}

	return json.MarshalIndent(cfg, "", "  ")
}
