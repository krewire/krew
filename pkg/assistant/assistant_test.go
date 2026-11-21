package assistant_test

import (
	"context"
	"testing"

	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/ai/providers"
	"github.com/krewire/krew/pkg/assistant"
	"github.com/krewire/krew/pkg/embed"
	"github.com/krewire/krew/pkg/persona"
)

func TestAssistantPoolAndStreaming(t *testing.T) {
	tempDir := t.TempDir()
	mock := providers.NewMock("Code refactored successfully.")

	pers := persona.CodingAssistant
	inst, err := embed.New(embed.Options{
		Workspace: tempDir,
		Persona:   &pers,
		Provider:  mock,
	})
	if err != nil {
		t.Fatalf("failed to create embedded instance: %v", err)
	}

	pool := assistant.NewAssistantPool()
	pool.SetInstance("ws-1", inst)

	// 1. Sync Chat
	res, err := pool.Chat(context.Background(), "ws-1", tempDir, "refactor this code")
	if err != nil {
		t.Fatalf("pool chat failed: %v", err)
	}
	if res != "Code refactored successfully." {
		t.Errorf("unexpected chat response: %s", res)
	}

	// 2. Stream Chat
	var events []agent.Event
	streamResp, err := pool.ChatStream(context.Background(), "ws-1", tempDir, "explain this function", func(ev agent.Event) {
		events = append(events, ev)
	})
	if err != nil {
		t.Fatalf("pool stream chat failed: %v", err)
	}
	if streamResp == nil || streamResp.Content != "Code refactored successfully." {
		t.Errorf("unexpected stream response: %+v", streamResp)
	}
	if len(events) == 0 {
		t.Errorf("expected stream events, got none")
	}
}
