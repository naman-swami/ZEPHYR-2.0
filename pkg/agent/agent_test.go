package agent

import (
	"context"
	"testing"
)

func TestAgent_InspectAndAffected(t *testing.T) {
	tasks := map[string]GraphNode{
		"lint": {
			Name:    "lint",
			Command: "echo lint",
			Inputs:  []string{"src/**/*.ts"},
		},
		"test": {
			Name:      "test",
			Command:   "echo test",
			Inputs:    []string{"src/**/*.ts"},
			DependsOn: []string{"lint"},
		},
		"build": {
			Name:      "build",
			Command:   "echo build",
			Inputs:    []string{"src/**/*.ts"},
			DependsOn: []string{"test"},
		},
	}

	service := NewService("/workspace", tasks)

	// 1. Test InspectGraph
	graph := service.InspectGraph()
	if len(graph.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(graph.Nodes))
	}
	if len(graph.Edges["lint"]) != 1 || graph.Edges["lint"][0] != "test" {
		t.Fatalf("expected lint -> test edge, got %v", graph.Edges["lint"])
	}

	// 2. Test AnalyzeAffected
	affected := service.AnalyzeAffected([]string{"src/auth/token.ts"})
	if len(affected.AllAffected) != 3 {
		t.Fatalf("expected all 3 tasks affected, got %v", affected.AllAffected)
	}
	if len(affected.DirectlyAffected) != 3 {
		t.Fatalf("expected 3 directly affected tasks, got %v", affected.DirectlyAffected)
	}
}

func TestAgent_PermissionGating(t *testing.T) {
	tasks := map[string]GraphNode{
		"risky-task": {
			Name:    "risky-task",
			Command: "echo format-drive",
		},
	}

	service := NewService(t.TempDir(), tasks)

	// Without permission -> PERMISSION_DENIED
	resDenied, err := service.ExecuteTask(context.Background(), "risky-task", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resDenied.Status != "PERMISSION_DENIED" {
		t.Fatalf("expected PERMISSION_DENIED status, got: %s", resDenied.Status)
	}

	// With permission -> EXECUTED
	resAllowed, err := service.ExecuteTask(context.Background(), "risky-task", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resAllowed.Status != "EXECUTED" {
		t.Fatalf("expected EXECUTED status, got: %s", resAllowed.Status)
	}
}
