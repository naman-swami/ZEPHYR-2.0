package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// sendRequest encodes a request, runs Serve for one message, and returns the response line.
func roundTrip(t *testing.T, s *Server, req interface{}) Response {
	t.Helper()

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	in := strings.NewReader(string(data) + "\n")
	var out bytes.Buffer
	srv := New(in, &out)
	srv.tools = s.tools
	srv.handlers = s.handlers

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = srv.Serve(ctx) // returns after EOF

	var resp Response
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &resp); err != nil {
		t.Fatalf("unmarshal response: %v\nraw: %s", err, out.String())
	}
	return resp
}

func TestInitialize(t *testing.T) {
	s := New(nil, nil)
	resp := roundTrip(t, s, Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{}`),
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	raw, _ := json.Marshal(resp.Result)
	var init InitializeResult
	if err := json.Unmarshal(raw, &init); err != nil {
		t.Fatalf("unmarshal InitializeResult: %v", err)
	}
	if init.ServerInfo.Name != "zephyr-mcp" {
		t.Errorf("got server name %q, want zephyr-mcp", init.ServerInfo.Name)
	}
	if init.ProtocolVersion == "" {
		t.Error("expected non-empty protocol version")
	}
}

func TestToolsList(t *testing.T) {
	s := New(nil, nil)
	s.RegisterTool("zephyr_graph", "Return DAG", `{"type":"object","properties":{}}`, func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return map[string]string{"status": "ok"}, nil
	})

	resp := roundTrip(t, s, Request{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	raw, _ := json.Marshal(resp.Result)
	var result map[string]interface{}
	json.Unmarshal(raw, &result)
	tools, ok := result["tools"].([]interface{})
	if !ok || len(tools) != 1 {
		t.Errorf("expected 1 tool, got result: %s", raw)
	}
}

func TestToolCall_Success(t *testing.T) {
	s := New(nil, nil)
	s.RegisterTool("zephyr_graph", "Return DAG", `{"type":"object","properties":{}}`, func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return map[string]string{"nodes": "3"}, nil
	})

	params, _ := json.Marshal(ToolCallParams{Name: "zephyr_graph", Arguments: map[string]interface{}{}})
	resp := roundTrip(t, s, Request{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  json.RawMessage(params),
	})
	if resp.Error != nil {
		t.Fatalf("unexpected RPC error: %+v", resp.Error)
	}
	raw, _ := json.Marshal(resp.Result)
	var result ToolCallResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("unmarshal ToolCallResult: %v", err)
	}
	if result.IsError {
		t.Error("expected IsError=false")
	}
	if len(result.Content) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestToolCall_UnknownTool(t *testing.T) {
	s := New(nil, nil)
	params, _ := json.Marshal(ToolCallParams{Name: "nonexistent_tool", Arguments: map[string]interface{}{}})
	resp := roundTrip(t, s, Request{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params:  json.RawMessage(params),
	})
	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
	if resp.Error.Code != ErrMethodNotFound {
		t.Errorf("expected code %d, got %d", ErrMethodNotFound, resp.Error.Code)
	}
}

func TestPing(t *testing.T) {
	s := New(nil, nil)
	resp := roundTrip(t, s, Request{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "ping",
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestParseError(t *testing.T) {
	in := strings.NewReader("not json\n")
	var out bytes.Buffer
	srv := New(in, &out)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = srv.Serve(ctx)

	var resp Response
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != ErrParseError {
		t.Errorf("expected parse error, got: %+v", resp.Error)
	}
}

func TestMethodNotFound(t *testing.T) {
	s := New(nil, nil)
	resp := roundTrip(t, s, Request{
		JSONRPC: "2.0",
		ID:      6,
		Method:  "resources/list", // valid MCP method but not implemented
	})
	if resp.Error == nil || resp.Error.Code != ErrMethodNotFound {
		t.Errorf("expected method not found error, got: %+v", resp.Error)
	}
}
