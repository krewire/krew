package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/krewire/krewire/pkg/ai"
)

func TestSessionLog_AppendLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()

	log, err := OpenSessionLog(dir, "sess_abc")
	if err != nil {
		t.Fatalf("OpenSessionLog failed: %v", err)
	}
	if !strings.HasSuffix(log.Path(), "sess_abc.jsonl") {
		t.Errorf("unexpected path %q", log.Path())
	}

	turn := []ai.Message{
		ai.UserMessage("hello"),
		ai.AssistantMessage("world"),
		ai.ToolMessage("call_1", "output"),
	}
	if err := log.Append(turn...); err != nil {
		t.Fatalf("Append failed: %v", err)
	}
	if err := log.Append(ai.UserMessage("again")); err != nil {
		t.Fatalf("Append failed: %v", err)
	}
	if err := log.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	msgs, err := LoadSession(dir, "sess_abc")
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(msgs))
	}
	if msgs[0].Role != ai.RoleUser || msgs[0].Content != "hello" {
		t.Errorf("unexpected first message: %+v", msgs[0])
	}
	if msgs[2].ToolCallID != "call_1" {
		t.Errorf("expected tool call id round-trip, got %+v", msgs[2])
	}
	if msgs[3].Content != "again" {
		t.Errorf("unexpected last message: %+v", msgs[3])
	}
}

func TestSessionLog_RejectsUnsafeIDs(t *testing.T) {
	for _, id := range []string{"", "../escape", "sub/dir", " spaced", "a b"} {
		if _, err := OpenSessionLog(t.TempDir(), id); err == nil {
			t.Errorf("expected invalid session id %q to be rejected", id)
		}
	}
	if _, err := LoadSession(t.TempDir(), "../escape"); err == nil {
		t.Error("expected invalid session id to be rejected on load")
	}
}

func TestSessionLog_MissingAndCorrupt(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadSession(dir, "absent"); err == nil {
		t.Error("expected missing session to fail")
	}

	path := filepath.Join(dir, "broken.jsonl")
	if err := os.WriteFile(path, []byte("{\"message\":{}}\nnot-json\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if _, err := LoadSession(dir, "broken"); err == nil {
		t.Error("expected corrupt transcript to fail")
	}

	// Empty lines are tolerated.
	if err := os.WriteFile(path, []byte("\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	msgs, err := LoadSession(dir, "broken")
	if err != nil || len(msgs) != 0 {
		t.Errorf("expected empty transcript to load clean, got %d msgs err=%v", len(msgs), err)
	}
}

func TestListSessions(t *testing.T) {
	dir := t.TempDir()

	for _, id := range []string{"older", "newer"} {
		log, err := OpenSessionLog(dir, id)
		if err != nil {
			t.Fatalf("open %s: %v", id, err)
		}
		if err := log.Append(ai.UserMessage(id), ai.AssistantMessage("r")); err != nil {
			t.Fatalf("append: %v", err)
		}
		if err := log.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
	}

	// Ensure stable mtimes ordering for the assertion.
	newerPath := filepath.Join(dir, "newer.jsonl")
	later := time.Now().Add(2 * time.Minute)
	if err := os.Chtimes(newerPath, later, later); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	sessions, err := ListSessions(dir)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if sessions[0].ID != "newer" || sessions[1].ID != "older" {
		t.Errorf("expected newest first, got %+v", sessions)
	}
	if sessions[0].Messages != 2 || sessions[0].Tokens <= 0 {
		t.Errorf("expected counts computed, got %+v", sessions[0])
	}

	// Missing directory yields no sessions, not an error.
	empty, err := ListSessions(filepath.Join(dir, "none"))
	if err != nil || len(empty) != 0 {
		t.Errorf("expected empty list for missing dir, got %v err=%v", empty, err)
	}

	// Non-transcript files are ignored.
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if sessions2, _ := ListSessions(dir); len(sessions2) != 2 {
		t.Errorf("expected non-jsonl files ignored, got %+v", sessions2)
	}
}

func TestDefaultSessionsDir(t *testing.T) {
	dir, err := DefaultSessionsDir()
	if err != nil {
		t.Fatalf("DefaultSessionsDir failed: %v", err)
	}
	if !strings.Contains(dir, filepath.Join(".krew", "sessions")) {
		t.Errorf("unexpected canonical sessions dir %q", dir)
	}
}
