package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/krewire/krew/pkg/agent"
)

// MCPRequest represents a JSON-RPC 2.0 message for the Model Context Protocol.
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response.
type MCPResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

// MCPServer wraps Krew Agent and exposes it via Model Context Protocol (stdio).
type MCPServer struct {
	agent *agent.Agent
	in    io.Reader
	out   io.Writer
}

// NewMCPServer initializes an MCP stdio server.
func NewMCPServer(ag *agent.Agent, in io.Reader, out io.Writer) *MCPServer {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	return &MCPServer{agent: ag, in: in, out: out}
}

// Serve handles the MCP JSON-RPC loop.
func (s *MCPServer) Serve(ctx context.Context) error {
	scanner := bufio.NewScanner(s.in)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}

		s.handleRPC(ctx, req)
	}
	return scanner.Err()
}

func (s *MCPServer) handleRPC(ctx context.Context, req MCPRequest) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    "krew-agent",
				"version": "0.1.0",
			},
			"capabilities": map[string]any{
				"tools": map[string]bool{"listChanged": true},
			},
		}

	case "tools/list":
		var mcpTools []map[string]any
		for _, t := range s.agent.Tools().List() {
			mcpTools = append(mcpTools, map[string]any{
				"name":        t.Name(),
				"description": t.Description(),
				"inputSchema": t.Parameters(),
			})
		}
		resp.Result = map[string]any{
			"tools": mcpTools,
		}

	case "tools/call":
		var callParams struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			resp.Error = map[string]any{"code": -32602, "message": "Invalid params"}
			break
		}

		out, err := s.agent.Tools().Execute(ctx, callParams.Name, callParams.Arguments)
		if err != nil {
			resp.Result = map[string]any{
				"isError": true,
				"content": []map[string]string{
					{"type": "text", "text": fmt.Sprintf("Error: %v", err)},
				},
			}
		} else {
			text := ""
			if str, ok := out.(string); ok {
				text = str
			} else {
				b, _ := json.Marshal(out)
				text = string(b)
			}
			resp.Result = map[string]any{
				"isError": false,
				"content": []map[string]string{
					{"type": "text", "text": text},
				},
			}
		}

	case "agent/run":
		var runParams struct {
			Prompt string `json:"prompt"`
		}
		if err := json.Unmarshal(req.Params, &runParams); err != nil {
			resp.Error = map[string]any{"code": -32602, "message": "Invalid params"}
			break
		}
		agentRes, err := s.agent.Run(ctx, runParams.Prompt)
		if err != nil {
			resp.Error = map[string]any{"code": -32000, "message": err.Error()}
		} else {
			resp.Result = agentRes
		}

	default:
		// Unknown or notification
		if req.ID == nil {
			return
		}
		resp.Error = map[string]any{"code": -32601, "message": "Method not found"}
	}

	b, _ := json.Marshal(resp)
	_, _ = fmt.Fprintln(s.out, string(b))
}
