package memory

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/krewire/krewire/pkg/ai"
)

func longMessage(n int) ai.Message {
	return ai.AssistantMessage(strings.Repeat("a", n*tokensPerRunes))
}

func TestTokenBuffer_BoundsAfterManyTurns(t *testing.T) {
	const maxTokens = 400
	b := NewTokenBuffer("system prompt", WithMaxTokens(maxTokens), WithKeepTurns(2))

	// 60+ turns of sizeable assistant/user payloads far exceed the budget.
	for i := 0; i < 60; i++ {
		b.Add(ai.UserMessage("user"), longMessage(50), longMessage(50))
	}

	if got := b.Tokens(); got > maxTokens {
		t.Fatalf("expected tokens within maxTokens=%d, got %d", maxTokens, got)
	}
	if len(b.messages) < 2 {
		t.Fatalf("expected recent turns preserved, got %d messages", len(b.messages))
	}
}

func TestTokenBuffer_PreservesSystemGoalAndRecentTurns(t *testing.T) {
	b := NewTokenBuffer("the system prompt", WithMaxTokens(300), WithKeepTurns(2))

	b.Add(ai.UserMessage("initial goal")) // captured as permanent goal
	b.Add(ai.AssistantMessage("ack"))     // turn 1
	for i := 0; i < 40; i++ {             // flood with filler
		b.Add(ai.UserMessage("filler"), ai.AssistantMessage("reply"), ai.AssistantMessage("tool"))
	}

	msgs := b.Messages()
	if len(msgs) == 0 || msgs[0].Content != "the system prompt" {
		t.Fatal("expected system prompt first")
	}
	if msgs[1].Content != "initial goal" || msgs[1].Role != ai.RoleUser {
		t.Fatalf("expected initial goal preserved as goal, got %+v", msgs[1])
	}
	for _, msg := range msgs[2:] {
		if msg.Content == "ack" {
			t.Error("expected old turns evicted")
			break
		}
	}
	// FIFO order of surviving turns.
	tail := msgs[len(msgs)-3:]
	if tail[0].Role != ai.RoleUser || tail[1].Role != ai.RoleAssistant {
		t.Fatalf("expected recent user/assistant FIFO order, got %+v", tail)
	}

	raw := b.RawMessages()
	if len(raw) == 0 || raw[0].Content != "initial goal" {
		t.Fatalf("expected goal lead RawMessages, got %+v", raw)
	}
}

func TestTokenBuffer_CompactionProducesCheckpoint(t *testing.T) {
	var mutex sync.Mutex
	var condensed [][]ai.Message

	b := NewTokenBuffer("system",
		WithMaxTokens(300),
		WithKeepTurns(2),
		WithCompactThreshold(0.85),
		WithCompactor(func(_ context.Context, evicted []ai.Message) (string, error) {
			mutex.Lock()
			defer mutex.Unlock()
			condensed = append(condensed, evicted)
			return "checkpoint facts", nil
		}),
	)

	b.Add(ai.UserMessage("goal"))
	for i := 0; i < 30; i++ {
		b.Add(ai.UserMessage("turn"), longMessage(30))
	}
	b.Wait()

	if len(condensed) == 0 {
		t.Fatal("expected the compactor to receive evicted turns")
	}
	msgs := b.Messages()
	found := false
	for _, msg := range msgs[:3] {
		if strings.Contains(msg.Content, "checkpoint facts") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected checkpoint summary near context head, got %+v", msgs[:3])
	}
	if got := b.Tokens(); got > 300 {
		t.Errorf("expected token growth bounded, got %d", got)
	}
}

func TestTokenBuffer_CompactorLeavesCheckpointOnError(t *testing.T) {
	b := NewTokenBuffer("system",
		WithMaxTokens(200),
		WithKeepTurns(2),
		WithCompactor(func(_ context.Context, _ []ai.Message) (string, error) {
			return "", errors.New("llm unavailable")
		}),
	)
	for i := 0; i < 30; i++ {
		b.Add(ai.UserMessage("turn"), longMessage(30))
	}
	b.Wait()

	for _, msg := range b.Messages() {
		if strings.Contains(msg.Content, checkpointPrefix) {
			t.Fatal("expected no checkpoint when the compactor fails")
		}
	}
}

func TestTokenBuffer_ClearInvalidatesInFlightCompaction(t *testing.T) {
	release := make(chan struct{})
	b := NewTokenBuffer("system",
		WithMaxTokens(200),
		WithKeepTurns(2),
		WithCompactor(func(_ context.Context, _ []ai.Message) (string, error) {
			<-release // hold the summary until the test clears
			return "stale checkpoint", nil
		}),
	)
	b.Add(ai.UserMessage("turn"))
	for i := 0; i < 20; i++ {
		b.Add(ai.UserMessage("turn"), longMessage(30))
	}
	b.Clear()
	close(release)
	b.Wait()

	for _, msg := range b.Messages() {
		if strings.Contains(msg.Content, "stale checkpoint") {
			t.Fatal("expected stale compaction result to be discarded")
		}
	}
}

func TestTokenBuffer_CompactForce(t *testing.T) {
	b := NewTokenBuffer("system", WithMaxTokens(100_000), WithKeepTurns(0))
	for i := 0; i < 30; i++ {
		b.Add(ai.UserMessage("turn"), longMessage(30))
	}
	before := b.Len()
	b.Compact(50)
	after := b.Tokens()
	if after >= int(0.85*100_000) || before-b.Len() <= 0 {
		t.Fatalf("expected forced compaction, before=%d after=%d tokens=%d", before, b.Len(), after)
	}
}

func TestTokenEstimator(t *testing.T) {
	if got := EstimateTokens(""); got != 0 {
		t.Errorf("empty text: expected 0, got %d", got)
	}
	if got := EstimateTokens(strings.Repeat("x", tokensPerRunes)); got != 1 {
		t.Errorf("expected 1 token for %d chars, got %d", tokensPerRunes, got)
	}
	msg := ai.Message{Role: ai.RoleSystem}
	if got := MessageTokens(msg); got != 0 {
		t.Errorf("empty message: expected 0, got %d", got)
	}
	withCalls := ai.Message{
		Role: ai.RoleTool,
		ToolCalls: []ai.ToolCall{{
			ID:        "call_1",
			Name:      "write_file",
			Arguments: `{"path":"/x","content":"abc"}`,
		}},
	}
	if got := MessageTokens(withCalls); got <= toolCallOverhead {
		t.Errorf("expected tool call payload accounted, got %d", got)
	}
}

func TestLLMCompactor_BuildsRequest(t *testing.T) {
	var captured ai.CompletionRequest
	provider := fakeProvider{complete: func(req ai.CompletionRequest) (*ai.CompletionResponse, error) {
		captured = req
		return &ai.CompletionResponse{Content: "  compact!  "}, nil
	}}

	summary, err := LLMCompactor(provider, "test-model")(context.Background(), []ai.Message{
		ai.UserMessage("hello"), ai.AssistantMessage("world"),
	})
	if err != nil {
		t.Fatalf("compactor failed: %v", err)
	}
	if summary != "compact!" {
		t.Errorf("expected trimmed summary, got %q", summary)
	}
	if captured.Model != "test-model" || len(captured.Messages) != 2 {
		t.Errorf("unexpected completion request: %+v", captured)
	}
}

// fakeProvider satisfies ai.Provider for compactor tests.
type fakeProvider struct {
	complete func(req ai.CompletionRequest) (*ai.CompletionResponse, error)
}

func (f fakeProvider) Name() string { return "fake" }
func (f fakeProvider) Complete(_ context.Context, req ai.CompletionRequest) (*ai.CompletionResponse, error) {
	return f.complete(req)
}
func (f fakeProvider) Stream(_ context.Context, _ ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	return nil, errors.New("stream unsupported")
}
