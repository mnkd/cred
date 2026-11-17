// mcpserver is a minimal stdio MCP server for checking cred run by hand.
//
// It speaks newline-delimited JSON-RPC 2.0 on stdin and stdout (the MCP
// stdio transport). Tool show_token returns the DEMO_TOKEN environment
// variable, so a masked cred run answers with "<concealed by cred>". The
// token is also written to stderr at start, so the client's MCP log shows
// whether stderr is masked too.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type request struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params struct {
		ProtocolVersion string `json:"protocolVersion"`
		Name            string `json:"name"`
	} `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

var tools = []map[string]any{
	{
		"name":        "show_token",
		"description": "Return the DEMO_TOKEN environment variable injected by cred run.",
		"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
	},
	{
		"name":        "ping_cred",
		"description": "Return a fixed string, to check the connection.",
		"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
	},
}

func text(s string) map[string]any {
	return map[string]any{"content": []map[string]string{{"type": "text", "text": s}}}
}

func handle(req request, token string) response {
	resp := response{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		version := req.Params.ProtocolVersion
		if version == "" {
			version = "2025-06-18"
		}
		resp.Result = map[string]any{
			"protocolVersion": version,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]string{"name": "cred-manual-check", "version": "1.0.0"},
		}
	case "tools/list":
		resp.Result = map[string]any{"tools": tools}
	case "tools/call":
		switch req.Params.Name {
		case "show_token":
			resp.Result = text("DEMO_TOKEN=" + token)
		case "ping_cred":
			resp.Result = text("pong from cred-manual-check")
		default:
			resp.Error = &rpcError{Code: -32602, Message: "unknown tool: " + req.Params.Name}
		}
	case "ping":
		resp.Result = map[string]any{}
	default:
		resp.Error = &rpcError{Code: -32601, Message: "method not found: " + req.Method}
	}
	return resp
}

func main() {
	token := os.Getenv("DEMO_TOKEN")
	fmt.Fprintf(os.Stderr, "cred-manual-check: started, DEMO_TOKEN=%s\n", token)

	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	out := json.NewEncoder(os.Stdout) // Encode writes one line per message
	for in.Scan() {
		var req request
		if err := json.Unmarshal(in.Bytes(), &req); err != nil {
			continue
		}
		// A request without an id is a notification and gets no response.
		if len(req.ID) == 0 || string(req.ID) == "null" {
			continue
		}
		if err := out.Encode(handle(req, token)); err != nil {
			fmt.Fprintf(os.Stderr, "cred-manual-check: write failed: %v\n", err)
			os.Exit(1)
		}
	}
}
