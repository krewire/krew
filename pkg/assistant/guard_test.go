package assistant_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/krewire/krew/pkg/agent"
	kai "github.com/krewire/krew/pkg/ai"
	"github.com/krewire/krew/pkg/assistant"
	"github.com/krewire/krew/pkg/embed"
)

type scriptedProvider struct {
	call  kai.ToolCall
	turns atomic.Int32
}

func (s *scriptedProvider) Name() string { return "scripted" }

func (s *scriptedProvider) Complete(_ context.Context, _ kai.CompletionRequest) (*kai.CompletionResponse, error) {
	if s.turns.Add(1) == 1 {
		return &kai.CompletionResponse{ToolCalls: []kai.ToolCall{s.call}, FinishReason: "tool_calls"}, nil
	}
	return &kai.CompletionResponse{Content: "done", FinishReason: "stop"}, nil
}

func (s *scriptedProvider) Stream(ctx context.Context, req kai.CompletionRequest) (<-chan kai.StreamChunk, error) {
	resp, err := s.Complete(ctx, req)
	if err != nil {
		return nil, err
	}
	ch := make(chan kai.StreamChunk, 4)
	if resp.Content != "" {
		ch <- kai.StreamChunk{Delta: resp.Content}
	}
	for i := range resp.ToolCalls {
		ch <- kai.StreamChunk{ToolCall: &resp.ToolCalls[i]}
	}
	ch <- kai.StreamChunk{Done: true}
	close(ch)
	return ch, nil
}

func TestAssistantPool_ApprovalFlow(t *testing.T) {
	ws := t.TempDir()
	target := filepath.Join(ws, "created.txt")

	args, _ := json.Marshal(map[string]any{"path": "created.txt", "content": "approved content"})
	prov := &scriptedProvider{call: kai.ToolCall{ID: "c1", Name: "write_file", Arguments: string(args)}}

	pool := assistant.NewAssistantPoolWithOptions(assistant.PoolOptions{
		Approval:        assistant.ApprovalRequired,
		ApprovalTimeout: 2 * time.Second,
		Configure: func(o *embed.Options) {
			o.Provider = prov
		},
	})

	var evMu sync.Mutex
	var gotApprovalEvent bool
	var seenID string

	done := make(chan error, 1)
	go func() {
		_, err := pool.ChatStream(context.Background(), "ws-appr", ws, "write a file", func(ev agent.Event) {
			if ev.Type == assistant.EventApprovalRequired {
				evMu.Lock()
				gotApprovalEvent = true
				seenID = ev.ToolID
				evMu.Unlock()
			}
		})
		done <- err
	}()

	// Wait for pending approval to show up
	deadline := time.Now().Add(1 * time.Second)
	var foundID string
	for time.Now().Before(deadline) {
		pending := pool.PendingApprovals("ws-appr")
		if len(pending) > 0 {
			foundID = pending[0].ID
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if foundID == "" {
		t.Fatal("expected pending approval, none listed")
	}

	if err := pool.Decide("ws-appr", foundID, true); err != nil {
		t.Fatalf("decide error: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("chatStream failed: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("chatStream timed out after approval")
	}

	evMu.Lock()
	defer evMu.Unlock()
	if !gotApprovalEvent {
		t.Errorf("expected EventApprovalRequired event, got none")
	}
	if seenID != foundID {
		t.Errorf("event ToolID %q != pending ID %q", seenID, foundID)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("target file was not created: %v", err)
	}
	if string(content) != "approved content" {
		t.Errorf("unexpected content: %q", string(content))
	}
}

func TestAssistantPool_ApprovalDenial(t *testing.T) {
	ws := t.TempDir()
	target := filepath.Join(ws, "rejected.txt")

	args, _ := json.Marshal(map[string]any{"path": "rejected.txt", "content": "should not exist"})
	prov := &scriptedProvider{call: kai.ToolCall{ID: "c2", Name: "write_file", Arguments: string(args)}}

	pool := assistant.NewAssistantPoolWithOptions(assistant.PoolOptions{
		Approval:        assistant.ApprovalRequired,
		ApprovalTimeout: 2 * time.Second,
		Configure: func(o *embed.Options) {
			o.Provider = prov
		},
	})

	done := make(chan error, 1)
	go func() {
		_, err := pool.Chat(context.Background(), "ws-deny", ws, "write file")
		done <- err
	}()

	deadline := time.Now().Add(1 * time.Second)
	var foundID string
	for time.Now().Before(deadline) {
		pending := pool.PendingApprovals("ws-deny")
		if len(pending) > 0 {
			foundID = pending[0].ID
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if foundID == "" {
		t.Fatal("expected pending approval")
	}

	if err := pool.Decide("ws-deny", foundID, false); err != nil {
		t.Fatalf("decide error: %v", err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("chat timed out")
	}

	if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected target to not exist after denial, stat err: %v", err)
	}
}

func TestCommandTool(t *testing.T) {
	ws := t.TempDir()
	tool := assistant.NewCommandTool("ws-cmd", ws, assistant.HostCommandRunner{})

	args, _ := json.Marshal(map[string]any{"command": "echo hello-krew"})
	res, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("command tool execute error: %v", err)
	}
	m, ok := res.(map[string]any)
	if !ok {
		t.Fatalf("unexpected result type: %T", res)
	}
	if !strings.Contains(m["stdout"].(string), "hello-krew") {
		t.Errorf("unexpected stdout: %v", m["stdout"])
	}
}
