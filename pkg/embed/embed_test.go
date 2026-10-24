package embed_test

import (
	"context"
	"testing"

	"github.com/krewire/krew/pkg/agent"
	"github.com/krewire/krew/pkg/embed"
	"github.com/krewire/krew/pkg/persona"
	"github.com/krewire/krewire/pkg/ai/providers"
)

func TestEmbed_NewAndRun(t *testing.T) {
	mock := providers.NewMock("Task executed successfully via embedded agent.")

	var events []agent.Event
	pers := persona.DevOpsRunner

	inst, err := embed.New(embed.Options{
		Workspace: t.TempDir(),
		Persona:   &pers,
		Provider:  mock,
		OnEvent: func(ev agent.Event) {
			events = append(events, ev)
		},
	})
	if err != nil {
		t.Fatalf("embed.New failed: %v", err)
	}

	if inst.Agent() == nil {
		t.Fatal("expected non-nil underlying agent")
	}
	if inst.Registry() == nil {
		t.Fatal("expected non-nil registry")
	}
	if inst.Memory() == nil {
		t.Fatal("expected non-nil memory")
	}

	res, err := inst.Run(context.Background(), "Deploy service")
	if err != nil {
		t.Fatalf("inst.Run failed: %v", err)
	}

	if res.Content != "Task executed successfully via embedded agent." {
		t.Errorf("unexpected content: %q", res.Content)
	}

	if len(events) == 0 {
		t.Errorf("expected recorded events, got 0")
	}
}

func TestEmbed_AutoDetectFailureWithoutEnv(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OLLAMA_HOST", "")

	_, err := embed.New(embed.Options{
		Workspace: t.TempDir(),
	})
	if err == nil {
		t.Fatal("expected error when no provider or env variables provided")
	}
}
