package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krewire/pkg/ai"
)

func TestSessionName(t *testing.T) {
	cases := []struct {
		prompt   string
		expected string
	}{
		{"", "session"},
		{"   ", "session"},
		{"Hello World!", "hello-world"},
		{"Build a simple CLI tool in Go for parsing JSON", "build-a-simple-cli-tool-in-go-for-parsing-json"},
		{"---special!!! characters??? only---", "special-characters-only"},
		{"What is 1 + 1?", "what-is-1-1"},
	}

	for _, tc := range cases {
		got := SessionName(tc.prompt)
		if got != tc.expected {
			t.Errorf("SessionName(%q) = %q; want %q", tc.prompt, got, tc.expected)
		}
	}
}

func TestStartSession(t *testing.T) {
	dir := t.TempDir()

	log1, name1, err := StartSession(dir, "Implement auth tokens")
	if err != nil {
		t.Fatalf("StartSession 1 failed: %v", err)
	}
	defer log1.Close()

	if name1 != "implement-auth-tokens" {
		t.Errorf("expected name 'implement-auth-tokens', got %q", name1)
	}

	// Starting another session with the same prompt should resolve conflict
	log2, name2, err := StartSession(dir, "Implement auth tokens")
	if err != nil {
		t.Fatalf("StartSession 2 failed: %v", err)
	}
	defer log2.Close()

	if name2 != "implement-auth-tokens-2" {
		t.Errorf("expected collision resolution 'implement-auth-tokens-2', got %q", name2)
	}
}

func TestRenderTranscriptMarkdown(t *testing.T) {
	dir := t.TempDir()
	log, err := OpenSessionLog(dir, "test-session")
	if err != nil {
		t.Fatalf("OpenSessionLog failed: %v", err)
	}

	msgs := []ai.Message{
		ai.SystemMessage("You are a helpful assistant"),
		ai.UserMessage("List files"),
		{
			Role: ai.RoleAssistant,
			ToolCalls: []ai.ToolCall{
				{
					ID:        "tc_1",
					Name:      "list_dir",
					Arguments: `{"path":"."}`,
				},
			},
			Content: "Looking up directory...",
		},
		ai.ToolMessage("tc_1", "main.go\ngo.mod"),
		ai.AssistantMessage("Here are the files: main.go, go.mod"),
	}

	if err := log.Append(msgs...); err != nil {
		t.Fatalf("Append failed: %v", err)
	}
	if err := log.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	md, err := RenderTranscriptMarkdown(dir, "test-session")
	if err != nil {
		t.Fatalf("RenderTranscriptMarkdown failed: %v", err)
	}

	if !strings.Contains(md, "# Session test-session") {
		t.Errorf("missing header in markdown: %s", md)
	}
	if !strings.Contains(md, "## You") || !strings.Contains(md, "List files") {
		t.Errorf("missing user turn in markdown: %s", md)
	}
	if !strings.Contains(md, "### Tool call: list_dir (`tc_1`)") {
		t.Errorf("missing tool call in markdown: %s", md)
	}
	if !strings.Contains(md, "### Tool result (`tc_1`)") {
		t.Errorf("missing tool result in markdown: %s", md)
	}
	if !strings.Contains(md, "## Krew") || !strings.Contains(md, "Here are the files: main.go, go.mod") {
		t.Errorf("missing assistant response in markdown: %s", md)
	}

	// Missing session error handling
	if _, err := RenderTranscriptMarkdown(dir, "nonexistent"); err == nil {
		t.Error("expected error for nonexistent session")
	}

	// Unsafe path rejection
	if _, err := RenderTranscriptMarkdown(dir, "../escaped"); err == nil {
		t.Error("expected error for unsafe session id")
	}

	// Ensure file was created in dir
	if _, err := os.Stat(filepath.Join(dir, "test-session.jsonl")); err != nil {
		t.Errorf("session file not found: %v", err)
	}
}
