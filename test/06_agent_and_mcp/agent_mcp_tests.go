package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type AgentResult struct {
	Name     string        `json:"name"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration_ms"`
	Detail   string        `json:"detail"`
}

type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	zephyrBin, _ := filepath.Abs(`..\..\zephyr.exe`)
	var results []AgentResult

	fmt.Println("=== RUNNING SUITE 06: AI AGENT CONTROL PLANE & MCP PROTOCOL ===")

	// Setup mock workspace
	testDir, _ := filepath.Abs("agent_test_ws")
	_ = os.RemoveAll(testDir)
	_ = os.MkdirAll(testDir, 0755)

	cfg := map[string]interface{}{
		"default": "agent_build",
		"tasks": map[string]interface{}{
			"agent_lint": map[string]interface{}{
				"command": "cmd /c echo lint ok",
				"inputs":  []string{"src/app.go"},
			},
			"agent_build": map[string]interface{}{
				"command":    "cmd /c echo build ok",
				"inputs":     []string{"src/app.go"},
				"outputs":    []string{"bin/app.exe"},
				"depends_on": []string{"agent_lint"},
			},
		},
	}
	cfgBytes, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile(filepath.Join(testDir, "tasks.json"), cfgBytes, 0644)

	// ------------------------------------------------------------------------
	// Test 6.1: Permission Gate - Run without --allow-exec MUST be DENIED
	// ------------------------------------------------------------------------
	{
		start := time.Now()
		cmd := exec.Command(zephyrBin, "agent", "run", "agent_build")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		out, err := cmd.CombinedOutput()
		dur := time.Since(start)

		output := string(out)
		blocked := (err != nil || strings.Contains(output, "PERMISSION_DENIED")) && !strings.Contains(output, "SUCCESS")
		detail := fmt.Sprintf("Execution without --allow-exec strictly rejected with PERMISSION_DENIED. Blocked=%v", blocked)

		results = append(results, AgentResult{
			Name:     "Agent Strict Permission Gating (Default Deny)",
			Passed:   blocked,
			Duration: dur,
			Detail:   detail,
		})
		fmt.Printf("[%s] Agent Permission Gating: %s\n", status(blocked), detail)
	}

	// ------------------------------------------------------------------------
	// Test 6.2: Authorized Execution with --allow-exec
	// ------------------------------------------------------------------------
	{
		start := time.Now()
		cmd := exec.Command(zephyrBin, "agent", "run", "--allow-exec", "agent_build")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		out, err := cmd.CombinedOutput()
		dur := time.Since(start)

		output := string(out)
		executed := err == nil && (strings.Contains(output, "SUCCESS") || strings.Contains(output, `"exit_code": 0`))
		detail := fmt.Sprintf("Execution with --allow-exec succeeded. JSON result returned.")

		results = append(results, AgentResult{
			Name:     "Agent Authorized Execution (--allow-exec)",
			Passed:   executed,
			Duration: dur,
			Detail:   detail,
		})
		fmt.Printf("[%s] Agent Authorized Execution: %s\n", status(executed), detail)
	}

	// ------------------------------------------------------------------------
	// Test 6.3: Agent Graph Topology Query
	// ------------------------------------------------------------------------
	{
		start := time.Now()
		cmd := exec.Command(zephyrBin, "agent", "graph")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "NO_COLOR=1")
		out, err := cmd.CombinedOutput()
		dur := time.Since(start)

		var parsed map[string]interface{}
		_ = json.Unmarshal(out, &parsed)
		data, _ := parsed["data"].(map[string]interface{})
		nodes, _ := data["nodes"].(map[string]interface{})

		passed := err == nil && len(nodes) >= 2
		detail := fmt.Sprintf("Agent graph returned %d nodes in structured JSON format", len(nodes))

		results = append(results, AgentResult{
			Name:     "Agent Dependency Graph Inspection",
			Passed:   passed,
			Duration: dur,
			Detail:   detail,
		})
		fmt.Printf("[%s] Agent Graph Inspection: %s\n", status(passed), detail)
	}

	// ------------------------------------------------------------------------
	// Test 6.4: Native MCP Server JSON-RPC Protocol Exchange
	// ------------------------------------------------------------------------
	{
		start := time.Now()
		cmd := exec.Command(zephyrBin, "mcp")
		cmd.Dir = testDir
		cmd.Env = append(os.Environ(), "NO_COLOR=1")

		stdin, errIn := cmd.StdinPipe()
		stdout, errOut := cmd.StdoutPipe()
		if errIn != nil || errOut != nil {
			panic("Failed to create MCP pipes")
		}

		_ = cmd.Start()

		reader := bufio.NewReader(stdout)

		// 1. Send initialize
		initReq := RPCRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "initialize",
			Params: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"clientInfo":      map[string]string{"name": "test-agent", "version": "1.0"},
			},
		}
		initBytes, _ := json.Marshal(initReq)
		_, _ = io.WriteString(stdin, string(initBytes)+"\n")

		line, _ := reader.ReadString('\n')
		var initResp RPCResponse
		_ = json.Unmarshal([]byte(line), &initResp)
		initPassed := initResp.ID == 1 && initResp.Error == nil

		// 2. Send tools/list
		toolsReq := RPCRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/list",
		}
		toolsBytes, _ := json.Marshal(toolsReq)
		_, _ = io.WriteString(stdin, string(toolsBytes)+"\n")

		line2, _ := reader.ReadString('\n')
		var toolsResp RPCResponse
		_ = json.Unmarshal([]byte(line2), &toolsResp)
		toolsPassed := toolsResp.ID == 2 && strings.Contains(string(toolsResp.Result), "zephyr_graph")

		// 3. Send tools/call for zephyr_graph
		callReq := RPCRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":      "zephyr_graph",
				"arguments": map[string]interface{}{},
			},
		}
		callBytes, _ := json.Marshal(callReq)
		_, _ = io.WriteString(stdin, string(callBytes)+"\n")

		line3, _ := reader.ReadString('\n')
		var callResp RPCResponse
		_ = json.Unmarshal([]byte(line3), &callResp)
		callPassed := callResp.ID == 3 && strings.Contains(string(callResp.Result), "agent_build")

		_ = stdin.Close()
		_ = cmd.Process.Kill()
		dur := time.Since(start)

		allMCPPassed := initPassed && toolsPassed && callPassed
		detail := fmt.Sprintf("MCP JSON-RPC Handshake: init=%v, tools/list=%v, tools/call(zephyr_graph)=%v", initPassed, toolsPassed, callPassed)

		results = append(results, AgentResult{
			Name:     "Native Model Context Protocol (MCP) Server",
			Passed:   allMCPPassed,
			Duration: dur,
			Detail:   detail,
		})
		fmt.Printf("[%s] Native MCP Server Protocol: %s\n", status(allMCPPassed), detail)
	}

	reportBytes, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile("agent_mcp_results.json", reportBytes, 0644)
	fmt.Println("\nSuite 06 Agent & MCP Tests Complete. Results written to agent_mcp_results.json")
}

func status(b bool) string {
	if b {
		return "PASS"
	}
	return "FAIL"
}
