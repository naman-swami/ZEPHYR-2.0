// Package lsp implements a native Language Server Protocol (LSP) server over stdio
// for ZEPHYR tasks.json configuration files.
//
// Compatible with VS Code, Cursor, Zed, Neovim, and Helix.
// Uses standard HTTP Content-Length header framing + JSON-RPC 2.0 payloads.
// Crafted with 100% Go Standard Library — Zero External Dependencies.
package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// ─────────────────────────────────────────────────────────────────────────────
// LSP JSON-RPC 2.0 Wire Types
// ─────────────────────────────────────────────────────────────────────────────

type RequestMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type ResponseMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *ResponseError `json:"error,omitempty"`
}

type NotificationMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ─────────────────────────────────────────────────────────────────────────────
// LSP Document & Diagnostic Types
// ─────────────────────────────────────────────────────────────────────────────

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"` // 1=Error, 2=Warning, 3=Information, 4=Hint
	Source   string `json:"source"`
	Message  string `json:"message"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"` // "markdown" or "plaintext"
	Value string `json:"value"`
}

type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind"` // 14=Keyword, 10=Property, 6=Variable
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Task Schema for Static Analysis
// ─────────────────────────────────────────────────────────────────────────────

type TaskDef struct {
	Command   string   `json:"command"`
	Inputs    []string `json:"inputs"`
	Outputs   []string `json:"outputs"`
	DependsOn []string `json:"depends_on"`
	EnvVars   []string `json:"env_vars"`
	Timeout   string   `json:"timeout"`
	Retries   int      `json:"retries"`
	Weight    int      `json:"weight"`
	Cwd       string   `json:"cwd"`
}

type ConfigDoc struct {
	Default string             `json:"default"`
	Tasks   map[string]TaskDef `json:"tasks"`
}

// ─────────────────────────────────────────────────────────────────────────────
// LSP Server Implementation
// ─────────────────────────────────────────────────────────────────────────────

type Server struct {
	in        io.Reader
	out       io.Writer
	mu        sync.Mutex
	documents map[string]string
	stats     map[string]int64 // Optional historical durations (ms)
}

func NewServer(in io.Reader, out io.Writer, stats map[string]int64) *Server {
	if stats == nil {
		stats = make(map[string]int64)
	}
	return &Server{
		in:        in,
		out:       out,
		documents: make(map[string]string),
		stats:     stats,
	}
}

// Serve reads LSP framed messages (Content-Length: ...\r\n\r\n{...}) until EOF or exit.
func (s *Server) Serve(ctx context.Context) error {
	reader := bufio.NewReader(s.in)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		body, err := readLSPFrame(reader)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		var req RequestMessage
		if err := json.Unmarshal(body, &req); err != nil {
			continue
		}

		if req.Method == "exit" {
			return nil
		}

		s.handleMessage(&req)
	}
}

func readLSPFrame(r *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break // End of headers
		}
		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			val := strings.TrimSpace(line[len("content-length:"):])
			n, err := strconv.Atoi(val)
			if err == nil {
				contentLength = n
			}
		}
	}

	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	buf := make([]byte, contentLength)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (s *Server) writeLSPFrame(payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	_, _ = s.out.Write([]byte(header))
	_, _ = s.out.Write(data)
}

func (s *Server) handleMessage(req *RequestMessage) {
	switch req.Method {
	case "initialize":
		res := map[string]interface{}{
			"capabilities": map[string]interface{}{
				"textDocumentSync": 1, // Full sync
				"hoverProvider":    true,
				"completionProvider": map[string]interface{}{
					"triggerCharacters": []string{"\"", ":"},
				},
			},
			"serverInfo": map[string]string{
				"name":    "zephyr-lsp",
				"version": "2.0.0",
			},
		}
		s.writeLSPFrame(ResponseMessage{JSONRPC: "2.0", ID: req.ID, Result: res})

	case "initialized", "$/cancelRequest":
		// No response needed for notification

	case "shutdown":
		s.writeLSPFrame(ResponseMessage{JSONRPC: "2.0", ID: req.ID, Result: nil})

	case "textDocument/didOpen":
		var params struct {
			TextDocument struct {
				URI  string `json:"uri"`
				Text string `json:"text"`
			} `json:"textDocument"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			s.mu.Lock()
			s.documents[params.TextDocument.URI] = params.TextDocument.Text
			s.mu.Unlock()
			s.analyzeAndPublish(params.TextDocument.URI, params.TextDocument.Text)
		}

	case "textDocument/didChange":
		var params struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
			ContentChanges []struct {
				Text string `json:"text"`
			} `json:"contentChanges"`
		}
		if json.Unmarshal(req.Params, &params) == nil && len(params.ContentChanges) > 0 {
			text := params.ContentChanges[len(params.ContentChanges)-1].Text
			s.mu.Lock()
			s.documents[params.TextDocument.URI] = text
			s.mu.Unlock()
			s.analyzeAndPublish(params.TextDocument.URI, text)
		}

	case "textDocument/hover":
		var params struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
			Position Position `json:"position"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			s.mu.Lock()
			text := s.documents[params.TextDocument.URI]
			s.mu.Unlock()
			hover := s.computeHover(text, params.Position)
			s.writeLSPFrame(ResponseMessage{JSONRPC: "2.0", ID: req.ID, Result: hover})
		}

	case "textDocument/completion":
		var params struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
		}
		_ = json.Unmarshal(req.Params, &params)
		s.mu.Lock()
		text := s.documents[params.TextDocument.URI]
		s.mu.Unlock()
		items := s.computeCompletions(text)
		s.writeLSPFrame(ResponseMessage{JSONRPC: "2.0", ID: req.ID, Result: items})
	}
}

// AnalyzeConfig validates a tasks.json string and returns LSP diagnostics.
func AnalyzeConfig(text string) []Diagnostic {
	diags := make([]Diagnostic, 0)
	lines := strings.Split(text, "\n")

	var cfg ConfigDoc
	if err := json.Unmarshal([]byte(text), &cfg); err != nil {
		diags = append(diags, Diagnostic{
			Range:    Range{Start: Position{0, 0}, End: Position{0, 1}},
			Severity: 1,
			Source:   "zephyr-lsp",
			Message:  "JSON Syntax Error: " + err.Error(),
		})
		return diags
	}

	if cfg.Default != "" && cfg.Tasks != nil {
		if _, ok := cfg.Tasks[cfg.Default]; !ok {
			line := findLineForToken(lines, `"default"`)
			diags = append(diags, Diagnostic{
				Range:    Range{Start: Position{line, 0}, End: Position{line, 25}},
				Severity: 1,
				Source:   "zephyr-lsp",
				Message:  fmt.Sprintf("Default target '%s' does not exist in tasks map", cfg.Default),
			})
		}
	}

	for name, task := range cfg.Tasks {
		taskLine := findLineForToken(lines, `"`+name+`"`)

		if strings.TrimSpace(task.Command) == "" {
			diags = append(diags, Diagnostic{
				Range:    Range{Start: Position{taskLine, 0}, End: Position{taskLine, len(name) + 6}},
				Severity: 1,
				Source:   "zephyr-lsp",
				Message:  fmt.Sprintf("Task '%s' has an empty or missing 'command'", name),
			})
		}

		if strings.Contains(task.Cwd, "..") {
			diags = append(diags, Diagnostic{
				Range:    Range{Start: Position{taskLine, 0}, End: Position{taskLine, len(name) + 6}},
				Severity: 1,
				Source:   "zephyr-lsp",
				Message:  fmt.Sprintf("Security Violation: Task '%s' cwd '%s' escapes workspace root", name, task.Cwd),
			})
		}

		for _, dep := range task.DependsOn {
			if dep == name {
				diags = append(diags, Diagnostic{
					Range:    Range{Start: Position{taskLine, 0}, End: Position{taskLine, len(name) + 6}},
					Severity: 1,
					Source:   "zephyr-lsp",
					Message:  fmt.Sprintf("Self-dependency cycle in task '%s'", name),
				})
				continue
			}
			if _, exists := cfg.Tasks[dep]; !exists {
				diags = append(diags, Diagnostic{
					Range:    Range{Start: Position{taskLine, 0}, End: Position{taskLine, len(name) + 6}},
					Severity: 1,
					Source:   "zephyr-lsp",
					Message:  fmt.Sprintf("Task '%s' depends on undefined task '%s'", name, dep),
				})
			}
		}
	}

	// 3-Color DFS Cycle Detection
	if cycle := detectCycle(cfg.Tasks); len(cycle) > 0 {
		first := cycle[0]
		line := findLineForToken(lines, `"`+first+`"`)
		diags = append(diags, Diagnostic{
			Range:    Range{Start: Position{line, 0}, End: Position{line, 20}},
			Severity: 1,
			Source:   "zephyr-lsp",
			Message:  fmt.Sprintf("DAG Cycle Detected: %s", strings.Join(cycle, " -> ")),
		})
	}

	return diags
}

func (s *Server) analyzeAndPublish(uri, text string) {
	diags := AnalyzeConfig(text)
	s.writeLSPFrame(NotificationMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/publishDiagnostics",
		Params: PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diags,
		},
	})
}

func (s *Server) computeHover(text string, pos Position) *Hover {
	var cfg ConfigDoc
	if err := json.Unmarshal([]byte(text), &cfg); err != nil {
		return nil
	}

	lines := strings.Split(text, "\n")
	if pos.Line < 0 || pos.Line >= len(lines) {
		return nil
	}
	lineText := lines[pos.Line]

	for name, t := range cfg.Tasks {
		if strings.Contains(lineText, `"`+name+`"`) {
			durInfo := "No prior execution telemetry"
			if ms, ok := s.stats[name]; ok {
				durInfo = fmt.Sprintf("%dms (Cached in .taskcache/stats.json)", ms)
			}
			md := fmt.Sprintf("### ⚡ ZEPHYR Task: `%s`\n\n"+
				"- **Command:** `%s`\n"+
				"- **Inputs:** `%v`\n"+
				"- **Outputs:** `%v`\n"+
				"- **Depends On:** `%v`\n"+
				"- **Historical Duration:** %s\n",
				name, t.Command, t.Inputs, t.Outputs, t.DependsOn, durInfo)
			return &Hover{
				Contents: MarkupContent{
					Kind:  "markdown",
					Value: md,
				},
			}
		}
	}
	return nil
}

func (s *Server) computeCompletions(text string) []CompletionItem {
	items := []CompletionItem{
		{Label: "command", Kind: 10, Detail: "Shell command to execute", InsertText: `"command": ""`},
		{Label: "inputs", Kind: 10, Detail: "Input file glob patterns", InsertText: `"inputs": ["**/*"]`},
		{Label: "outputs", Kind: 10, Detail: "Output artifact paths for CAS", InsertText: `"outputs": []`},
		{Label: "depends_on", Kind: 10, Detail: "Upstream task dependencies", InsertText: `"depends_on": []`},
		{Label: "env_vars", Kind: 10, Detail: "Environment variables tracked in hash", InsertText: `"env_vars": []`},
		{Label: "timeout", Kind: 10, Detail: "Maximum task execution timeout", InsertText: `"timeout": "2m"`},
		{Label: "retries", Kind: 10, Detail: "Retry attempts on failure", InsertText: `"retries": 1`},
		{Label: "weight", Kind: 10, Detail: "Worker pool concurrency weight", InsertText: `"weight": 1`},
		{Label: "cwd", Kind: 10, Detail: "Task working directory", InsertText: `"cwd": "."`},
	}

	var cfg ConfigDoc
	if json.Unmarshal([]byte(text), &cfg) == nil {
		for name, t := range cfg.Tasks {
			items = append(items, CompletionItem{
				Label:         name,
				Kind:          6,
				Detail:        "ZEPHYR Task Target",
				Documentation: "Command: " + t.Command,
				InsertText:    name,
			})
		}
	}
	return items
}

func findLineForToken(lines []string, token string) int {
	for i, l := range lines {
		if strings.Contains(l, token) {
			return i
		}
	}
	return 0
}

func detectCycle(tasks map[string]TaskDef) []string {
	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)
	state := make(map[string]int)
	var path []string

	var dfs func(node string) []string
	dfs = func(node string) []string {
		state[node] = visiting
		path = append(path, node)

		if t, ok := tasks[node]; ok {
			for _, dep := range t.DependsOn {
				if state[dep] == visiting {
					return append(path, dep)
				}
				if state[dep] == unvisited {
					if cyc := dfs(dep); len(cyc) > 0 {
						return cyc
					}
				}
			}
		}

		path = path[:len(path)-1]
		state[node] = visited
		return nil
	}

	for name := range tasks {
		if state[name] == unvisited {
			if cyc := dfs(name); len(cyc) > 0 {
				return cyc
			}
		}
	}
	return nil
}
