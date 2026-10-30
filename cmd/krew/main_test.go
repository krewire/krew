package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/krewire/krew/pkg/memory"
	"github.com/krewire/krewire/pkg/ai"
)

func TestReplaySession(t *testing.T) {
	msgs := []ai.Message{
		ai.UserMessage("How do I parse JSON in Go?"),
		{
			Role: ai.RoleAssistant,
			ToolCalls: []ai.ToolCall{
				{
					ID:        "call_123",
					Name:      "read_file",
					Arguments: `{"path":"main.go"}`,
				},
			},
			Content: "Let me check main.go",
		},
		ai.ToolMessage("call_123", "package main\nimport \"encoding/json\""),
		ai.AssistantMessage("Use json.Unmarshal() to decode JSON into a struct."),
	}

	var buf bytes.Buffer
	replaySession(&buf, msgs)

	out := buf.String()
	if !strings.Contains(out, "How do I parse JSON in Go?") {
		t.Errorf("missing user turn in replay: %s", out)
	}
	if !strings.Contains(out, "read_file") || !strings.Contains(out, "main.go") {
		t.Errorf("missing tool call in replay: %s", out)
	}
	if !strings.Contains(out, "package main") {
		t.Errorf("missing tool result preview: %s", out)
	}
	if !strings.Contains(out, "Use json.Unmarshal() to decode JSON into a struct.") {
		t.Errorf("missing final response: %s", out)
	}
}

func TestPrintSessionList(t *testing.T) {
	sessions := []memory.SessionInfo{
		{
			ID:       "fix-auth-tokens",
			ModTime:  time.Date(2026, 10, 30, 20, 0, 0, 0, time.UTC),
			Messages: 6,
			Tokens:   1250,
		},
		{
			ID:       "init-sqlite-db",
			ModTime:  time.Date(2026, 10, 30, 19, 30, 0, 0, time.UTC),
			Messages: 4,
			Tokens:   840,
		},
	}

	var buf bytes.Buffer
	printSessionList(&buf, sessions)

	out := buf.String()
	if !strings.Contains(out, "Saved Sessions:") {
		t.Errorf("missing header: %s", out)
	}
	if !strings.Contains(out, "fix-auth-tokens") || !strings.Contains(out, "6 msgs, 1250 tokens") {
		t.Errorf("missing first session info: %s", out)
	}
	if !strings.Contains(out, "init-sqlite-db") || !strings.Contains(out, "4 msgs, 840 tokens") {
		t.Errorf("missing second session info: %s", out)
	}
}

func TestExportAndSessionsCLIIntegration(t *testing.T) {
	dir := t.TempDir()

	log, err := memory.OpenSessionLog(dir, "my-chat-session")
	if err != nil {
		t.Fatalf("OpenSessionLog: %v", err)
	}
	if err := log.Append(
		ai.UserMessage("Hello agent"),
		ai.AssistantMessage("Hello user, how can I help?"),
	); err != nil {
		t.Fatalf("Append: %v", err)
	}
	_ = log.Close()

	// Test sessions list retrieval
	sessions, err := memory.ListSessions(dir)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != "my-chat-session" {
		t.Fatalf("unexpected sessions: %+v", sessions)
	}

	// Test export markdown rendering
	rendered, err := memory.RenderTranscriptMarkdown(dir, "my-chat-session")
	if err != nil {
		t.Fatalf("RenderTranscriptMarkdown: %v", err)
	}
	if !strings.Contains(rendered, "# Session my-chat-session") {
		t.Errorf("missing markdown header: %s", rendered)
	}
	if !strings.Contains(rendered, "Hello agent") || !strings.Contains(rendered, "Hello user, how can I help?") {
		t.Errorf("missing messages in markdown: %s", rendered)
	}

	// Test file existence
	if _, err := memory.LoadSession(dir, "my-chat-session"); err != nil {
		t.Errorf("LoadSession failed: %v", err)
	}

	// Assert path safely confined
	absPath := filepath.Join(dir, "my-chat-session.jsonl")
	if log.Path() != absPath {
		t.Errorf("expected path %q, got %q", absPath, log.Path())
	}
}
