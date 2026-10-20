package organize

import (
	"bytes"
	"encoding/json"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ExtractGoModuleName parses go.mod to extract the current module path.
func ExtractGoModuleName(root string) string {
	goModPath := filepath.Join(root, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "module"))
		}
	}
	return ""
}

// RewriteGoPackageClause updates the package declaration of a Go file.
func RewriteGoPackageClause(filePath string, newPackage string) error {
	if newPackage == "" {
		return nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return err
	}

	if f.Name.Name == newPackage {
		return nil
	}

	f.Name.Name = newPackage

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return err
	}

	return os.WriteFile(filePath, buf.Bytes(), 0644)
}

// EnsurePythonPackageMarkers writes an empty __init__.py in newly created Python directories.
func EnsurePythonPackageMarkers(dirPath string) error {
	initPath := filepath.Join(dirPath, "__init__.py")
	if _, err := os.Stat(initPath); os.IsNotExist(err) {
		return os.WriteFile(initPath, []byte("# Auto-generated package marker\n"), 0644)
	}
	return nil
}

// SyncTasksJSON updates tasks.json globs and target commands to reflect canonical layout.
func SyncTasksJSON(root string) error {
	tasksPath := filepath.Join(root, "tasks.json")
	content, err := os.ReadFile(tasksPath)
	if err != nil {
		return nil // tasks.json optional
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(content, &raw); err != nil {
		return nil
	}

	tasksRaw, ok := raw["tasks"].(map[string]interface{})
	if !ok {
		return nil
	}

	modified := false
	for _, tRaw := range tasksRaw {
		taskMap, ok := tRaw.(map[string]interface{})
		if !ok {
			continue
		}

		// Update inputs globs
		if inputsRaw, ok := taskMap["inputs"].([]interface{}); ok {
			var newInputs []interface{}
			for _, in := range inputsRaw {
				inStr, ok := in.(string)
				if !ok {
					continue
				}
				if inStr == "*.go" || inStr == "**/*.go" {
					newInputs = append(newInputs, "cmd/**/*.go", "internal/**/*.go", "pkg/**/*.go", "go.mod")
					modified = true
				} else {
					newInputs = append(newInputs, in)
				}
			}
			taskMap["inputs"] = newInputs
		}

		// Update build commands targeting root main.go
		if cmdStr, ok := taskMap["command"].(string); ok {
			if strings.Contains(cmdStr, "go build") && strings.Contains(cmdStr, "main.go") {
				taskMap["command"] = strings.ReplaceAll(cmdStr, "main.go", "./cmd/...")
				modified = true
			}
		}
	}

	if modified {
		data, err := json.MarshalIndent(raw, "", "  ")
		if err == nil {
			return os.WriteFile(tasksPath, data, 0644)
		}
	}

	return nil
}
