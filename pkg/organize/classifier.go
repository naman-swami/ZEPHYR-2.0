package organize

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// RootAnchorFiles lists invariant configuration and anchor files that must never be moved from workspace root.
var RootAnchorFiles = map[string]bool{
	"go.mod":             true,
	"go.sum":             true,
	"go.work":            true,
	"tasks.json":         true,
	"package.json":       true,
	"package-lock.json":  true,
	"pnpm-lock.yaml":     true,
	"yarn.lock":          true,
	"tsconfig.json":      true,
	"pyproject.toml":     true,
	"setup.py":           true,
	"setup.cfg":          true,
	"requirements.txt":   true,
	"makefile":           true,
	"readme.md":          true,
	"license":            true,
	"license.md":         true,
	"license.txt":        true,
	".gitignore":         true,
	".gitattributes":     true,
	".editorconfig":      true,
}

// IgnoredDirs lists directories skipped during workspace scanning.
var IgnoredDirs = map[string]bool{
	".git":         true,
	".taskcache":   true,
	"node_modules": true,
	"vendor":       true,
	".idea":        true,
	".vscode":      true,
	"bin":          true,
	"dist":         true,
	"build":        true,
	"__pycache__":  true,
	".venv":        true,
	"venv":         true,
}

// ClassifyFile evaluates path and returns its primary category.
func ClassifyFile(relPath string, isRoot bool) FileCategory {
	base := strings.ToLower(filepath.Base(relPath))
	ext := strings.ToLower(filepath.Ext(relPath))

	// 1. Root Anchors
	if isRoot && RootAnchorFiles[base] {
		return CategoryRootAnchor
	}

	// 2. Documentation
	if base == "contributing.md" || base == "changelog.md" || strings.HasPrefix(base, "rfc") {
		return CategoryDocs
	}
	switch ext {
	case ".md", ".rst", ".adoc", ".pdf":
		if isRoot && (base == "readme.md" || base == "license.md") {
			return CategoryRootAnchor
		}
		return CategoryDocs
	}

	// 3. Media & Assets
	switch ext {
	case ".png", ".jpg", ".jpeg", ".svg", ".ico", ".gif", ".webp", ".bmp", ".ttf", ".woff", ".woff2":
		return CategoryAssets
	}

	// 4. Automation & Tooling Scripts
	switch ext {
	case ".sh", ".bash", ".zsh", ".ps1", ".bat", ".cmd":
		return CategoryScripts
	}

	// 5. Test Fixtures & Mock Data
	if strings.Contains(base, "mock") || strings.Contains(base, "fixture") || strings.Contains(base, "sample") {
		return CategoryTestData
	}

	// 6. Configurations
	if strings.HasPrefix(base, "dockerfile") || strings.HasPrefix(base, "docker-compose") {
		return CategoryConfigs
	}
	switch ext {
	case ".toml", ".yaml", ".yml", ".ini", ".conf", ".env.example":
		return CategoryConfigs
	}

	// 7. Source Code
	switch ext {
	case ".go", ".py", ".js", ".ts", ".jsx", ".tsx", ".rs", ".c", ".cpp", ".h", ".java":
		return CategoryCode
	}

	return CategoryDocs
}

// DetectLanguage maps file extensions to language names.
func DetectLanguage(relPath string) string {
	ext := strings.ToLower(filepath.Ext(relPath))
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".rs":
		return "rust"
	case ".sh", ".bash", ".ps1":
		return "shell"
	default:
		return "other"
	}
}

// ComputeSHA256 returns hex encoded SHA-256 digest of a file.
func ComputeSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ScanWorkspace inspects all files within the workspace root directory.
func ScanWorkspace(root string) ([]*FileNode, error) {
	var nodes []*FileNode

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			base := info.Name()
			if IgnoredDirs[base] {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		isRoot := filepath.Dir(rel) == "."

		category := ClassifyFile(rel, isRoot)
		lang := DetectLanguage(rel)

		node := &FileNode{
			RelPath:    rel,
			AbsPath:    path,
			Category:   category,
			Language:   lang,
			Size:       info.Size(),
			IsRootFile: isRoot,
		}

		// Compute hash only for files under 10MB to avoid excessive latency
		if info.Size() < 10*1024*1024 {
			hash, err := ComputeSHA256(path)
			if err == nil {
				node.SHA256 = hash
			}
		}

		nodes = append(nodes, node)
		return nil
	})

	return nodes, err
}
