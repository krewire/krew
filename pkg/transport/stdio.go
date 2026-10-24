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

// StdioServer handles line-delimited JSON-RPC or text prompt I/O over stdin and stdout.
type StdioServer struct {
	agent *agent.Agent
	in    io.Reader
	out   io.Writer
}

// NewStdioServer creates a standard input/output transport.
func NewStdioServer(ag *agent.Agent, in io.Reader, out io.Writer) *StdioServer {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	return &StdioServer{
		agent: ag,
		in:    in,
		out:   out,
	}
}

// Serve reads input lines from in, executes the agent, and outputs events.
func (s *StdioServer) Serve(ctx context.Context) error {
	scanner := bufio.NewScanner(s.in)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		// Parse either a JSON request or plain text
		prompt := line
		var jsonReq struct {
			Prompt string `json:"prompt"`
		}
		if err := json.Unmarshal([]byte(line), &jsonReq); err == nil && jsonReq.Prompt != "" {
			prompt = jsonReq.Prompt
		}

		res, err := s.agent.Run(ctx, prompt)
		if err != nil {
			errResp, _ := json.Marshal(map[string]any{"error": err.Error()})
			_, _ = fmt.Fprintln(s.out, string(errResp))
			continue
		}

		respBytes, _ := json.Marshal(res)
		_, _ = fmt.Fprintln(s.out, string(respBytes))
	}
	return scanner.Err()
}
