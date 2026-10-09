package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"taskrunner/pkg/verify"
)

// AgentResponse is the standard JSON envelope returned to AI agents and tool callers.
type AgentResponse struct {
	Operation string      `json:"operation"`
	Timestamp time.Time   `json:"timestamp"`
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// GraphNode describes a task node in the project dependency graph.
type GraphNode struct {
	Name      string   `json:"name"`
	Command   string   `json:"command"`
	Inputs    []string `json:"inputs"`
	Outputs   []string `json:"outputs"`
	DependsOn []string `json:"depends_on"`
	Timeout   string   `json:"timeout,omitempty"`
}

// ProjectGraph describes the full project DAG for an AI agent.
type ProjectGraph struct {
	RootPath string               `json:"root_path"`
	Nodes    map[string]GraphNode `json:"nodes"`
	Edges    map[string][]string  `json:"edges"` // task -> dependents
}

// AffectedAnalysis returns tasks affected by a set of modified files.
type AffectedAnalysis struct {
	ChangedFiles    []string `json:"changed_files"`
	DirectlyAffected []string `json:"directly_affected"`
	DownstreamAffected []string `json:"downstream_affected"`
	AllAffected     []string `json:"all_affected"`
}

// ExecutionPlan outlines which tasks need execution vs which are cached.
type ExecutionPlan struct {
	Target          string     `json:"target"`
	ExecutionOrder  [][]string `json:"execution_order_batches"`
	TotalTasks      int        `json:"total_tasks"`
}

// TaskExecutionResult captures execution results safely.
type TaskExecutionResult struct {
	TaskName   string        `json:"task_name"`
	Status     string        `json:"status"` // "EXECUTED", "CACHED", "FAILED", "PERMISSION_DENIED"
	ExitCode   int           `json:"exit_code"`
	Duration   time.Duration `json:"duration"`
	Stdout     string        `json:"stdout,omitempty"`
	Stderr     string        `json:"stderr,omitempty"`
}

// Service provides structured inspection and safe execution methods for AI coding agents.
type Service struct {
	WorkspaceRoot string
	Tasks         map[string]GraphNode
}

func NewService(workspaceRoot string, tasks map[string]GraphNode) *Service {
	return &Service{
		WorkspaceRoot: workspaceRoot,
		Tasks:         tasks,
	}
}

// InspectGraph returns the full dependency DAG as structured JSON.
func (s *Service) InspectGraph() *ProjectGraph {
	edges := make(map[string][]string)
	for name, node := range s.Tasks {
		for _, dep := range node.DependsOn {
			edges[dep] = append(edges[dep], name)
		}
	}

	return &ProjectGraph{
		RootPath: s.WorkspaceRoot,
		Nodes:    s.Tasks,
		Edges:    edges,
	}
}

// AnalyzeAffected identifies affected tasks from a list of changed file paths.
func (s *Service) AnalyzeAffected(changedFiles []string) *AffectedAnalysis {
	directlyAffectedMap := make(map[string]bool)

	for _, file := range changedFiles {
		cleanFile := filepath.ToSlash(file)
		for taskName, task := range s.Tasks {
			for _, inputPattern := range task.Inputs {
				if matchPattern(cleanFile, filepath.ToSlash(inputPattern)) {
					directlyAffectedMap[taskName] = true
				}
			}
		}
	}

	// Compute downstream dependents
	dependentsMap := make(map[string][]string)
	for name, t := range s.Tasks {
		for _, dep := range t.DependsOn {
			dependentsMap[dep] = append(dependentsMap[dep], name)
		}
	}

	allAffectedMap := make(map[string]bool)
	var queue []string
	var directlyList []string

	for task := range directlyAffectedMap {
		directlyList = append(directlyList, task)
		allAffectedMap[task] = true
		queue = append(queue, task)
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, dep := range dependentsMap[curr] {
			if !allAffectedMap[dep] {
				allAffectedMap[dep] = true
				queue = append(queue, dep)
			}
		}
	}

	var allList []string
	var downstreamList []string
	for task := range allAffectedMap {
		allList = append(allList, task)
		if !directlyAffectedMap[task] {
			downstreamList = append(downstreamList, task)
		}
	}

	return &AffectedAnalysis{
		ChangedFiles:       changedFiles,
		DirectlyAffected:   directlyList,
		DownstreamAffected: downstreamList,
		AllAffected:        allList,
	}
}

// VerifyTask delegates to the reproducibility auditor and returns a structured report.
func (s *Service) VerifyTask(ctx context.Context, taskName string) (*verify.VerificationReport, error) {
	node, exists := s.Tasks[taskName]
	if !exists {
		return nil, fmt.Errorf("task '%s' not found", taskName)
	}

	auditor := verify.NewAuditor(s.WorkspaceRoot)
	taskSpec := verify.TaskSpec{
		Name:    node.Name,
		Command: node.Command,
		Outputs: node.Outputs,
	}

	return auditor.AuditTask(ctx, taskSpec, 2)
}

// ExecuteTask enforces explicit permission checks before executing any process.
func (s *Service) ExecuteTask(ctx context.Context, taskName string, permissionGranted bool) (*TaskExecutionResult, error) {
	if !permissionGranted {
		return &TaskExecutionResult{
			TaskName: taskName,
			Status:   "PERMISSION_DENIED",
			ExitCode: 1,
			Stderr:   "Execution denied: Agent requested execution without explicit user authorization (--allow-exec flag required)",
		}, nil
	}

	node, exists := s.Tasks[taskName]
	if !exists {
		return nil, fmt.Errorf("task '%s' not found in workspace", taskName)
	}

	// Execution allowed: run safely
	auditor := verify.NewAuditor(s.WorkspaceRoot)
	spec := verify.TaskSpec{
		Name:    node.Name,
		Command: node.Command,
		Outputs: node.Outputs,
	}
	report, err := auditor.AuditTask(ctx, spec, 1)
	if err != nil {
		return nil, err
	}

	return &TaskExecutionResult{
		TaskName: taskName,
		Status:   "EXECUTED",
		ExitCode: report.Run1ExitCode,
		Duration: report.RunDurations[0],
		Stdout:   report.Run1StdoutSnippet,
	}, nil
}

// FormatJSON serializes an agent response into formatted JSON.
func FormatJSON(op string, data interface{}, err error) string {
	resp := AgentResponse{
		Operation: op,
		Timestamp: time.Now().UTC(),
		Success:   err == nil,
		Data:      data,
	}
	if err != nil {
		resp.Error = err.Error()
	}
	bytes, _ := json.MarshalIndent(resp, "", "  ")
	return string(bytes)
}

func matchPattern(path, pattern string) bool {
	matched, err := filepath.Match(pattern, path)
	if err == nil && matched {
		return true
	}
	if strings.Contains(pattern, "**") {
		prefix := strings.Split(pattern, "**")[0]
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
