package shell

import (
	"os"
	"testing"
)

func TestShellInitialization(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if wd == "" {
		t.Errorf("working directory is empty")
	}
}
