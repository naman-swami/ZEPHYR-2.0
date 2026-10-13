package lsp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func formatLSPMessage(payload interface{}) string {
	b, _ := json.Marshal(payload)
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(b), string(b))
}

func TestAnalyzeConfig_ValidAndInvalid(t *testing.T) {
	validJSON := `{
		"default": "build",
		"tasks": {
			"lint": {"command": "go vet ./..."},
			"build": {"command": "go build .", "depends_on": ["lint"]}
		}
	}`
	diags := AnalyzeConfig(validJSON)
	if len(diags) != 0 {
		t.Fatalf("expected 0 diagnostics for valid config, got %d: %+v", len(diags), diags)
	}

	cycleJSON := `{
		"default": "a",
		"tasks": {
			"a": {"command": "echo a", "depends_on": ["b"]},
			"b": {"command": "echo b", "depends_on": ["a"]}
		}
	}`
	diags = AnalyzeConfig(cycleJSON)
	if len(diags) == 0 {
		t.Fatalf("expected cycle diagnostic, got 0")
	}
	foundCycle := false
	for _, d := range diags {
		if strings.Contains(d.Message, "Cycle") {
			foundCycle = true
		}
	}
	if !foundCycle {
		t.Fatalf("expected cycle message in diagnostics: %+v", diags)
	}
}

func TestLSPServer_InitializeHoverAndCompletion(t *testing.T) {
	var inBuf bytes.Buffer
	var outBuf bytes.Buffer

	docText := "{\n  \"default\": \"build\",\n  \"tasks\": {\n    \"build\": {\n      \"command\": \"go build .\"\n    }\n  }\n}"

	inBuf.WriteString(formatLSPMessage(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]interface{}{},
	}))

	inBuf.WriteString(formatLSPMessage(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "textDocument/didOpen",
		"params": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"uri":  "file:///workspace/tasks.json",
				"text": docText,
			},
		},
	}))

	inBuf.WriteString(formatLSPMessage(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "textDocument/hover",
		"params": map[string]interface{}{
			"textDocument": map[string]interface{}{"uri": "file:///workspace/tasks.json"},
			"position":     map[string]interface{}{"line": 3, "character": 5},
		},
	}))

	inBuf.WriteString(formatLSPMessage(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "textDocument/completion",
		"params": map[string]interface{}{
			"textDocument": map[string]interface{}{"uri": "file:///workspace/tasks.json"},
		},
	}))

	inBuf.WriteString(formatLSPMessage(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "exit",
	}))

	stats := map[string]int64{"build": 1420}
	srv := NewServer(&inBuf, &outBuf, stats)
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("LSP Serve failed: %v", err)
	}

	outStr := outBuf.String()
	if !strings.Contains(outStr, "zephyr-lsp") {
		t.Errorf("expected initialize response with zephyr-lsp, got: %s", outStr)
	}
	if !strings.Contains(outStr, "1420ms") {
		t.Errorf("expected hover response with historical duration 1420ms, got: %s", outStr)
	}
	if !strings.Contains(outStr, "depends_on") {
		t.Errorf("expected completion response with schema fields, got: %s", outStr)
	}
}
