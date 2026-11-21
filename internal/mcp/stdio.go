package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// RunStdio reads newline-delimited JSON-RPC 2.0 messages from in, dispatches them
// to the MCP server, and writes responses to out.
func RunStdio(ctx context.Context, srv *Server, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	const maxCapacity = 10 * 1024 * 1024 // 10MB max buffer for large project structures
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	var outMu sync.Mutex
	writeResponse := func(res *Response) error {
		if res == nil {
			return nil
		}
		b, err := json.Marshal(res)
		if err != nil {
			return err
		}
		outMu.Lock()
		defer outMu.Unlock()
		_, err = fmt.Fprintf(out, "%s\n", string(b))
		return err
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			_ = writeResponse(&Response{
				JSONRPC: "2.0",
				Error: &Error{
					Code:    CodeParseError,
					Message: fmt.Sprintf("parse error: %v", err),
				},
			})
			continue
		}

		resp := srv.HandleRequest(ctx, req)
		if resp != nil {
			if err := writeResponse(resp); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}
