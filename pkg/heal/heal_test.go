package heal

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProposePureGoRepair_MissingColon(t *testing.T) {
	code := `def calculate_tax(income, status)
    rate = 0.2
    return income * rate
`
	diag := "SyntaxError: expected ':' at line 1"
	res, err := ProposePureGoRepair(diag, code, "python")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(res.ReplaceBlock, "def calculate_tax(income, status):") {
		t.Fatalf("expected colon appended, got: %q", res.ReplaceBlock)
	}
}

func TestProposePureGoRepair_Indentation(t *testing.T) {
	code := `def compute():
  return 42
`
	diag := "IndentationError: unexpected indent at line 2"
	res, err := ProposePureGoRepair(diag, code, "python")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(res.ReplaceBlock, "    return 42") {
		t.Fatalf("expected 4-space indent, got: %q", res.ReplaceBlock)
	}
}

func TestVerifyDeterministicGate_GoValid(t *testing.T) {
	validGo := `package main

import "fmt"

func main() {
	fmt.Println("hello world")
}
`
	passed, err := verifyDeterministicGate(context.Background(), "test.go", validGo, "go")
	if !passed || err != nil {
		t.Fatalf("expected gate pass for valid Go code, got passed=%v, err=%v", passed, err)
	}
}

func TestVerifyDeterministicGate_GoInvalid(t *testing.T) {
	invalidGo := `package main

func main() {
	fmt.Println("missing closing brace"
`
	passed, err := verifyDeterministicGate(context.Background(), "test.go", invalidGo, "go")
	if passed || err == nil {
		t.Fatalf("expected gate rejection for invalid Go code, got passed=%v", passed)
	}
}

func TestRepairFile_EndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sample.py")
	brokenContent := `def greet(name)
    return "Hello " + name
`
	if err := os.WriteFile(testFile, []byte(brokenContent), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	req := HealRequest{
		FilePath:   testFile,
		Diagnostic: "SyntaxError: expected ':' at line 1",
		Language:   "python",
		DryRun:     false,
	}

	res, err := RepairFile(context.Background(), req)
	if err != nil {
		t.Fatalf("RepairFile failed: %v", err)
	}

	if !res.Success || !res.GatePassed {
		t.Fatalf("expected repair success and gate pass, got success=%v, gate=%v, err=%s", res.Success, res.GatePassed, res.GateError)
	}

	// Verify file on disk was repaired
	repairedBytes, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("read repaired file: %v", err)
	}

	if !strings.Contains(string(repairedBytes), "def greet(name):") {
		t.Fatalf("expected repaired content, got:\n%s", string(repairedBytes))
	}
}
