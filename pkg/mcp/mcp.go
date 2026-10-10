// Package mcp implements a native Model Context Protocol (MCP) server over stdio.
//
// ZEPHYR exposes build intelligence to AI coding agents (Claude Code, Cursor, Copilot,
// OpenHands) through the MCP protocol. All I/O is over stdin/stdout using newline-delimited
// JSON-RPC 2.0 messages. Zero network sockets. Zero external dependencies.
//
// Exposed Tools:
//   - zephyr_graph     : Return full task DAG as structured JSON
//   - zephyr_affected  : Blast-radius analysis for a set of modified files
//   - zephyr_why       : Explain why a specific task would be a cache miss
//   - zephyr_verify    : Run reproducibility audit on a task (read-only)
//   - zephyr_run       : Execute a task (requires explicit allow_exec=true permission)
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// JSON-RPC 2.0 wire types
// ─────────────────────────────────────────────────────────────────────────────

// Request is an incoming MCP JSON-RPC 2.0 message from the AI agent host.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is an outgoing MCP JSON-RPC 2.0 message to the AI agent host.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError follows JSON-RPC 2.0 error object spec.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Standard JSON-RPC error codes.
const (
	ErrParseError     = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternalError  = -32603
)

// ─────────────────────────────────────────────────────────────────────────────
// MCP Protocol types
// ─────────────────────────────────────────────────────────────────────────────

// ToolDefinition describes a tool exposed to the AI agent.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ToolCallParams are the params for a tools/call request.
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCallResult is the result of a tools/call request.
type ToolCallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock is a single piece of content in a tool result.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ServerInfo is returned in the initialize response.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult is the result of the initialize method.
type InitializeResult struct {
	ProtocolVersion string          `json:"protocolVersion"`
	Capabilities    map[string]bool `json:"capabilities"`
	ServerInfo      ServerInfo      `json:"serverInfo"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Tool handler interface
// ─────────────────────────────────────────────────────────────────────────────

// Handler is called by the Server to handle individual tool invocations.
// Implementations live in main.go (they close over the loaded config).
type Handler func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// ─────────────────────────────────────────────────────────────────────────────
// Server
// ─────────────────────────────────────────────────────────────────────────────

// Server is a stdio MCP server. It reads JSON-RPC requests from r and writes
// responses to w. Both r and w are typically os.Stdin / os.Stdout.
type Server struct {
	tools    []ToolDefinition
	handlers map[string]Handler
	in       io.Reader
	out      io.Writer
}

// New creates a new Server wired to the given reader/writer.
func New(in io.Reader, out io.Writer) *Server {
	return &Server{
		handlers: make(map[string]Handler),
		in:       in,
		out:      out,
	}
}

// RegisterTool adds a tool to the server's exposed tool list and registers its handler.
// schemaJSON must be a valid JSON Schema object literal (as a raw JSON string).
func (s *Server) RegisterTool(name, description, schemaJSON string, h Handler) {
	s.tools = append(s.tools, ToolDefinition{
		Name:        name,
		Description: description,
		InputSchema: json.RawMessage(schemaJSON),
	})
	s.handlers[name] = h
}

// Serve runs the MCP server loop: reads requests from s.in, dispatches to handlers,
// writes responses to s.out. Blocks until ctx is cancelled or s.in returns io.EOF.
func (s *Server) Serve(ctx context.Context) error {
	scanner := bufio.NewScanner(s.in)
	// MCP messages can be large (e.g. full DAG JSON). Increase scan buffer.
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	enc := json.NewEncoder(s.out)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("mcp: stdin read error: %w", err)
			}
			return nil // EOF — agent disconnected
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			_ = enc.Encode(Response{
				JSONRPC: "2.0",
				Error:   &RPCError{Code: ErrParseError, Message: "Parse error: " + err.Error()},
			})
			continue
		}

		resp := s.dispatch(ctx, &req)
		_ = enc.Encode(resp)
	}
}

// dispatch routes a JSON-RPC request to the appropriate method handler.
func (s *Server) dispatch(ctx context.Context, req *Request) Response {
	base := Response{JSONRPC: "2.0", ID: req.ID}

	switch req.Method {
	case "initialize":
		base.Result = InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]bool{"tools": true},
			ServerInfo:      ServerInfo{Name: "zephyr-mcp", Version: "2.0.0"},
		}

	case "initialized":
		// Notification — no response needed, but we still need to return something
		// when called from dispatch; caller ignores it if ID is nil.
		base.Result = map[string]string{}

	case "tools/list":
		base.Result = map[string]interface{}{
			"tools": s.tools,
		}

	case "tools/call":
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			base.Error = &RPCError{Code: ErrInvalidParams, Message: "Invalid params: " + err.Error()}
			return base
		}

		handler, ok := s.handlers[params.Name]
		if !ok {
			base.Error = &RPCError{Code: ErrMethodNotFound, Message: "Unknown tool: " + params.Name}
			return base
		}

		result, err := handler(ctx, params.Arguments)
		if err != nil {
			// MCP spec: tool errors are reported in result.isError, not in JSON-RPC error
			base.Result = ToolCallResult{
				IsError: true,
				Content: []ContentBlock{{Type: "text", Text: err.Error()}},
			}
			return base
		}

		// Marshal result to JSON text for the content block
		resultBytes, marshalErr := json.MarshalIndent(result, "", "  ")
		if marshalErr != nil {
			base.Result = ToolCallResult{
				IsError: true,
				Content: []ContentBlock{{Type: "text", Text: marshalErr.Error()}},
			}
			return base
		}

		base.Result = ToolCallResult{
			Content: []ContentBlock{{Type: "text", Text: string(resultBytes)}},
		}

	case "ping":
		base.Result = map[string]string{"status": "pong"}

	default:
		base.Error = &RPCError{Code: ErrMethodNotFound, Message: "Method not found: " + req.Method}
	}

	return base
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper: write startup banner to stderr (not stdout, which is MCP wire)
// ─────────────────────────────────────────────────────────────────────────────

// PrintStartupBanner writes a human-readable startup message to stderr.
// This is safe because MCP communication happens over stdout only.
func PrintStartupBanner(toolCount int) {
	fmt.Fprintf(os.Stderr, "[ZEPHYR MCP] Server v2.0.0 ready — %d tools registered\n", toolCount)
	fmt.Fprintf(os.Stderr, "[ZEPHYR MCP] Protocol: JSON-RPC 2.0 over stdio (MCP 2024-11-05)\n")
	fmt.Fprintf(os.Stderr, "[ZEPHYR MCP] Waiting for AI agent connection...\n")
}
