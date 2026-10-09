package verify

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ReproducibilityNotice provides an honest, precise claim regarding reproducibility.
const ReproducibilityNotice = "Notice: Two identical output hashes from repeated runs under tested conditions are evidence of reproducibility under those conditions, not proof that every environmental dependency has been discovered."

// ArtifactComparison holds byte and hash comparison between runs for a specific output file.
type ArtifactComparison struct {
	Path        string `json:"path"`
	Match       bool   `json:"match"`
	Run1Hash    string `json:"run1_hash"`
	Run2Hash    string `json:"run2_hash"`
	Run1Size    int64  `json:"run1_size"`
	Run2Size    int64  `json:"run2_size"`
	FirstDiffAt int64  `json:"first_diff_offset,omitempty"` // -1 if sizes differ or matched
}

// VerificationReport captures the full reproducibility audit across isolated runs.
type VerificationReport struct {
	TaskName          string                `json:"task_name"`
	Runs              int                   `json:"runs"`
	Timestamp         time.Time             `json:"timestamp"`
	IsDeterministic   bool                  `json:"is_deterministic"`
	ExitCodeMatch     bool                  `json:"exit_code_match"`
	StdoutMatch       bool                  `json:"stdout_match"`
	StderrMatch       bool                  `json:"stderr_match"`
	ArtifactsMatch    bool                  `json:"artifacts_match"`
	DeclaredOutputs   []string              `json:"declared_outputs"`
	Artifacts         []ArtifactComparison  `json:"artifacts"`
	DivergenceReasons []string              `json:"divergence_reasons,omitempty"`
	RunDurations      []time.Duration       `json:"run_durations"`
	Notice            string                `json:"notice"`
	Run1ExitCode      int                   `json:"run1_exit_code"`
	Run2ExitCode      int                   `json:"run2_exit_code"`
	Run1StdoutSnippet string                `json:"run1_stdout_snippet,omitempty"`
	Run2StdoutSnippet string                `json:"run2_stdout_snippet,omitempty"`
	Run1StderrSnippet string                `json:"run1_stderr_snippet,omitempty"`
	Run2StderrSnippet string                `json:"run2_stderr_snippet,omitempty"`
}

// TaskSpec defines minimal requirements for running a verification audit.
type TaskSpec struct {
	Name    string
	Command string
	Outputs []string
	Cwd     string
	Shell   string
	Env     []string
}

// Auditor executes repeated runs in independent workspace copies without reusing cache.
type Auditor struct {
	WorkspaceRoot string
}

func NewAuditor(workspaceRoot string) *Auditor {
	return &Auditor{WorkspaceRoot: workspaceRoot}
}

// AuditTask runs the task twice in isolated temporary workspaces and checks byte-for-byte reproducibility.
func (a *Auditor) AuditTask(ctx context.Context, task TaskSpec, numRuns int) (*VerificationReport, error) {
	if numRuns < 2 {
		numRuns = 2
	}

	report := &VerificationReport{
		TaskName:        task.Name,
		Runs:            numRuns,
		Timestamp:       time.Now().UTC(),
		DeclaredOutputs: task.Outputs,
		Notice:          ReproducibilityNotice,
		RunDurations:    make([]time.Duration, 0, numRuns),
	}

	// Execution results for run 1 and run 2
	type runResult struct {
		exitCode int
		stdout   string
		stderr   string
		duration time.Duration
		artDir   string
	}

	results := make([]runResult, numRuns)

	for i := 0; i < numRuns; i++ {
		// Create isolated clean workspace copy
		tempDir, err := os.MkdirTemp("", fmt.Sprintf("zephyr-verify-run%d-*", i+1))
		if err != nil {
			return nil, fmt.Errorf("failed to create isolated temp workspace: %w", err)
		}
		defer os.RemoveAll(tempDir)

		// Copy workspace files to temp directory (skipping .taskcache and .git)
		if err := copyDir(a.WorkspaceRoot, tempDir); err != nil {
			return nil, fmt.Errorf("failed to populate isolated workspace copy: %w", err)
		}

		// Prepare execution
		cmdCwd := tempDir
		if task.Cwd != "" && task.Cwd != "." {
			cmdCwd = filepath.Join(tempDir, task.Cwd)
		}

		startTime := time.Now()
		var stdoutBuf, stderrBuf bytes.Buffer

		cmd := buildCommand(ctx, task.Command, task.Shell)
		cmd.Dir = cmdCwd
		cmd.Env = os.Environ()
		for _, e := range task.Env {
			if strings.Contains(e, "=") {
				cmd.Env = append(cmd.Env, e)
			} else {
				if val, set := os.LookupEnv(e); set {
					cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", e, val))
				}
			}
		}
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf

		err = cmd.Run()
		duration := time.Since(startTime)
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 1
			}
		}

		results[i] = runResult{
			exitCode: exitCode,
			stdout:   stdoutBuf.String(),
			stderr:   stderrBuf.String(),
			duration: duration,
			artDir:   cmdCwd,
		}
		report.RunDurations = append(report.RunDurations, duration)
	}

	r1 := results[0]
	r2 := results[1]

	report.Run1ExitCode = r1.exitCode
	report.Run2ExitCode = r2.exitCode
	report.ExitCodeMatch = (r1.exitCode == r2.exitCode)
	report.StdoutMatch = (r1.stdout == r2.stdout)
	report.StderrMatch = (r1.stderr == r2.stderr)

	if !report.StdoutMatch {
		report.Run1StdoutSnippet = truncateString(r1.stdout, 300)
		report.Run2StdoutSnippet = truncateString(r2.stdout, 300)
	}

	if !report.StderrMatch || r1.exitCode != 0 {
		report.Run1StderrSnippet = truncateString(r1.stderr, 300)
		report.Run2StderrSnippet = truncateString(r2.stderr, 300)
	}

	// Compare declared outputs byte-for-byte
	report.ArtifactsMatch = true
	if len(task.Outputs) > 0 {
		for _, outPattern := range task.Outputs {
			// Resolve files matching pattern in both workspaces
			p1Files, _ := filepath.Glob(filepath.Join(r1.artDir, outPattern))
			p2Files, _ := filepath.Glob(filepath.Join(r2.artDir, outPattern))

			fileSet := make(map[string]bool)
			for _, f := range p1Files {
				rel, _ := filepath.Rel(r1.artDir, f)
				fileSet[rel] = true
			}
			for _, f := range p2Files {
				rel, _ := filepath.Rel(r2.artDir, f)
				fileSet[rel] = true
			}

			if len(fileSet) == 0 {
				// Pattern produced no files; treat exact pattern name if it exists directly
				target1 := filepath.Join(r1.artDir, outPattern)
				target2 := filepath.Join(r2.artDir, outPattern)
				comp := compareFiles(outPattern, target1, target2)
				report.Artifacts = append(report.Artifacts, comp)
				if !comp.Match {
					report.ArtifactsMatch = false
				}
				continue
			}

			for relPath := range fileSet {
				target1 := filepath.Join(r1.artDir, relPath)
				target2 := filepath.Join(r2.artDir, relPath)
				comp := compareFiles(relPath, target1, target2)
				report.Artifacts = append(report.Artifacts, comp)
				if !comp.Match {
					report.ArtifactsMatch = false
				}
			}
		}
	}

	report.IsDeterministic = report.ExitCodeMatch && report.ArtifactsMatch

	if !report.IsDeterministic {
		if !report.ArtifactsMatch {
			report.DivergenceReasons = append(report.DivergenceReasons,
				"Artifact byte digests differed across runs under identical declared inputs.")
			report.DivergenceReasons = append(report.DivergenceReasons,
				"Common causes: embedded compile timestamps (__DATE__, time.Now()), unpinned random seeds/UUIDs, or non-deterministic archive ordering.")
		}
		if !report.ExitCodeMatch {
			report.DivergenceReasons = append(report.DivergenceReasons,
				fmt.Sprintf("Exit code mismatch: Run 1 returned %d, Run 2 returned %d.", r1.exitCode, r2.exitCode))
		}
	} else if !report.StdoutMatch {
		report.DivergenceReasons = append(report.DivergenceReasons,
			"Outputs are byte-identical, but process stdout contained non-deterministic lines (e.g. log timestamps or elapsed times).")
	}

	return report, nil
}

func compareFiles(relPath, f1, f2 string) ArtifactComparison {
	comp := ArtifactComparison{
		Path:        relPath,
		FirstDiffAt: -1,
	}

	info1, err1 := os.Stat(f1)
	info2, err2 := os.Stat(f2)

	if os.IsNotExist(err1) && os.IsNotExist(err2) {
		comp.Match = true
		comp.Run1Hash = "NONE"
		comp.Run2Hash = "NONE"
		return comp
	}

	if err1 != nil || err2 != nil {
		comp.Match = false
		if err1 == nil {
			comp.Run1Size = info1.Size()
			comp.Run1Hash = hashFile(f1)
			comp.Run2Hash = "MISSING"
		} else {
			comp.Run1Hash = "MISSING"
			comp.Run2Size = info2.Size()
			comp.Run2Hash = hashFile(f2)
		}
		return comp
	}

	comp.Run1Size = info1.Size()
	comp.Run2Size = info2.Size()
	comp.Run1Hash = hashFile(f1)
	comp.Run2Hash = hashFile(f2)

	if comp.Run1Hash == comp.Run2Hash {
		comp.Match = true
		return comp
	}

	comp.Match = false
	// Find first differing byte offset if both exist
	comp.FirstDiffAt = findFirstByteDiff(f1, f2)
	return comp
}

func hashFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "ERROR"
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "ERROR"
	}
	return hex.EncodeToString(h.Sum(nil))
}

func findFirstByteDiff(p1, p2 string) int64 {
	f1, err := os.Open(p1)
	if err != nil {
		return -1
	}
	defer f1.Close()

	f2, err := os.Open(p2)
	if err != nil {
		return -1
	}
	defer f2.Close()

	buf1 := make([]byte, 4096)
	buf2 := make([]byte, 4096)
	var offset int64 = 0

	for {
		n1, err1 := f1.Read(buf1)
		n2, err2 := f2.Read(buf2)

		minN := n1
		if n2 < minN {
			minN = n2
		}

		for i := 0; i < minN; i++ {
			if buf1[i] != buf2[i] {
				return offset + int64(i)
			}
		}

		offset += int64(minN)

		if n1 != n2 || err1 != nil || err2 != nil {
			if n1 != n2 {
				return offset
			}
			break
		}
	}
	return -1
}

func buildCommand(ctx context.Context, command, explicitShell string) *exec.Cmd {
	shell := explicitShell
	if shell == "" {
		if runtime.GOOS == "windows" {
			shell = "cmd"
		} else {
			shell = "sh"
		}
	}

	switch shell {
	case "powershell":
		return exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", command)
	case "cmd":
		return exec.CommandContext(ctx, "cmd.exe", "/C", command)
	case "bash":
		return exec.CommandContext(ctx, "bash", "-c", command)
	default:
		return exec.CommandContext(ctx, "sh", "-c", command)
	}
}

func copyDir(src, dst string) error {
	skipDirs := map[string]bool{
		".git":          true,
		".taskcache":    true,
		".cache-server": true,
		"bin":           true,
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) > 0 && skipDirs[parts[0]] {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		targetPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Regular file copy
		return copyFile(path, targetPath)
	})
}

func copyFile(src, dst string) error {
	_ = os.MkdirAll(filepath.Dir(dst), 0755)
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func truncateString(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// FormatTerminalReport renders an ANSI audit report for human developers.
func FormatTerminalReport(r *VerificationReport, colorEnabled bool) string {
	var b strings.Builder

	border := strings.Repeat("═", 78)
	b.WriteString(fmt.Sprintf("%s\n", border))
	b.WriteString(fmt.Sprintf("  ZEPHYR REPRODUCIBILITY AUDIT REPORT\n"))
	b.WriteString(fmt.Sprintf("  Task: %s | Runs: %d | Tested Environment: %s/%s\n", r.TaskName, r.Runs, runtime.GOOS, runtime.GOARCH))
	b.WriteString(fmt.Sprintf("%s\n", border))

	if r.IsDeterministic {
		b.WriteString(fmt.Sprintf("  VERDICT: [PASS] 100%% REPRODUCIBLE UNDER TESTED CONDITIONS\n"))
	} else {
		b.WriteString(fmt.Sprintf("  VERDICT: [FAIL] NON-DETERMINISTIC BEHAVIOR DETECTED\n"))
	}

	b.WriteString(fmt.Sprintf("\n  Detailed Checks:\n"))
	if r.ExitCodeMatch {
		b.WriteString(fmt.Sprintf("    ✓ Exit Codes Match: %d (identical across runs)\n", r.Run1ExitCode))
		if r.Run1ExitCode != 0 && r.Run1StderrSnippet != "" {
			b.WriteString(fmt.Sprintf("    ! Process Stderr: %s\n", r.Run1StderrSnippet))
		}
	} else {
		b.WriteString(fmt.Sprintf("    ✗ Exit Code Divergence: Run 1 = %d vs Run 2 = %d\n", r.Run1ExitCode, r.Run2ExitCode))
	}

	if r.StdoutMatch {
		b.WriteString(fmt.Sprintf("    ✓ Stdout Stream: Byte-identical across runs\n"))
	} else {
		b.WriteString(fmt.Sprintf("    ! Stdout Stream: Minor variance (output logs differ)\n"))
	}

	if len(r.Artifacts) == 0 {
		b.WriteString(fmt.Sprintf("    - Declared Artifacts: (No file outputs declared for this task)\n"))
	} else {
		b.WriteString(fmt.Sprintf("    Declared Output Artifacts (%d checked):\n", len(r.Artifacts)))
		for _, art := range r.Artifacts {
			if art.Match {
				b.WriteString(fmt.Sprintf("      ✓ %s\n", art.Path))
				b.WriteString(fmt.Sprintf("        SHA-256: %s (%d bytes)\n", art.Run1Hash, art.Run1Size))
			} else {
				b.WriteString(fmt.Sprintf("      ✗ %s [HASH DIVERGENCE]\n", art.Path))
				b.WriteString(fmt.Sprintf("        Run 1 SHA-256: %s (%d bytes)\n", art.Run1Hash, art.Run1Size))
				b.WriteString(fmt.Sprintf("        Run 2 SHA-256: %s (%d bytes)\n", art.Run2Hash, art.Run2Size))
				if art.FirstDiffAt >= 0 {
					b.WriteString(fmt.Sprintf("        First Byte Divergence: Offset 0x%X (%d)\n", art.FirstDiffAt, art.FirstDiffAt))
				}
			}
		}
	}

	if len(r.DivergenceReasons) > 0 {
		b.WriteString(fmt.Sprintf("\n  Audit Findings & Potential Causes:\n"))
		for _, reason := range r.DivergenceReasons {
			b.WriteString(fmt.Sprintf("    • %s\n", reason))
		}
	}

	b.WriteString(fmt.Sprintf("\n  %s\n", r.Notice))
	b.WriteString(fmt.Sprintf("%s\n", border))

	return b.String()
}
