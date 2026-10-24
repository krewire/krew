package memory

import (
	"testing"

	"github.com/krewire/krewire/pkg/ai"
)

func TestBuffer(t *testing.T) {
	buf := NewBuffer("You are Krew Agent.", 3)

	buf.Add(ai.UserMessage("msg 1"))
	buf.Add(ai.AssistantMessage("resp 1"))
	buf.Add(ai.UserMessage("msg 2"))

	if buf.Len() != 3 {
		t.Fatalf("expected 3 messages, got %d", buf.Len())
	}

	// Adding 4th message should slide window (maxMsgs = 3)
	buf.Add(ai.AssistantMessage("resp 2"))
	if buf.Len() != 3 {
		t.Fatalf("expected 3 messages after slide, got %d", buf.Len())
	}

	msgs := buf.Messages()
	// System message + 3 messages = 4
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages with system, got %d", len(msgs))
	}
	if msgs[0].Role != ai.RoleSystem || msgs[0].Content != "You are Krew Agent." {
		t.Errorf("system message mismatch: %+v", msgs[0])
	}
	if msgs[1].Content != "resp 1" {
		t.Errorf("expected oldest msg to be 'resp 1', got %q", msgs[1].Content)
	}

	buf.Compact(2)
	if buf.Len() != 2 {
		t.Errorf("expected 2 messages after compact, got %d", buf.Len())
	}

	buf.Clear()
	if buf.Len() != 0 {
		t.Errorf("expected 0 messages after clear, got %d", buf.Len())
	}
}
