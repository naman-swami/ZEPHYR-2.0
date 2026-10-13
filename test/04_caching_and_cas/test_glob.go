//go:build ignore

package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func matchGlob(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	if pattern == path {
		return true
	}

	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}

			remaining := path
			if prefix != "" {
				remaining = strings.TrimPrefix(path, prefix)
			}
			remaining = strings.TrimPrefix(remaining, "/")

			if suffix == "" {
				return true
			}

			if strings.HasPrefix(suffix, "*") {
				ext := strings.TrimPrefix(suffix, "*")
				return strings.HasSuffix(remaining, ext)
			}

			if matched, _ := filepath.Match(suffix, remaining); matched {
				return true
			}
			if matched, _ := filepath.Match(suffix, filepath.Base(remaining)); matched {
				return true
			}
			return strings.HasSuffix(remaining, suffix)
		}
	}

	if matched, _ := filepath.Match(pattern, path); matched {
		return true
	}
	if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
		return true
	}

	return false
}

func main() {
	pattern := "dist/**/*.txt"
	path := "dist/file_001.txt"
	fmt.Printf("Pattern: %s, Path: %s, Result: %v\n", pattern, path, matchGlob(pattern, path))
	
	// What if pattern is dist/* or dist/*.txt?
	fmt.Printf("Pattern: dist/*, Path: %s, Result: %v\n", path, matchGlob("dist/*", path))
	fmt.Printf("Pattern: dist/*.txt, Path: %s, Result: %v\n", path, matchGlob("dist/*.txt", path))
	fmt.Printf("Pattern: **/*.txt, Path: %s, Result: %v\n", path, matchGlob("**/*.txt", path))
}
