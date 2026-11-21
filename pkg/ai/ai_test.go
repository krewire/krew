package ai

import (
	"context"
	"testing"
)

type echoMockProvider struct{}

func (e *echoMockProvider) Name() string { return "echo" }
func (e *echoMockProvider) Complete(_ context.Context, req CompletionRequest) (*CompletionResponse, error) {
	content := ""
	if len(req.Messages) > 0 {
		content = "Echo: " + req.Messages[len(req.Messages)-1].Content
	}
	return &CompletionResponse{
		Content: content,
	}, nil
}
func (e *echoMockProvider) Stream(_ context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 2)
	ch <- StreamChunk{Delta: "Echo"}
	ch <- StreamChunk{Done: true}
	close(ch)
	return ch, nil
}

func TestClient_CompleteAndChat(t *testing.T) {
	client := NewClient(&echoMockProvider{}, "test-model").
		WithSystem("You are an assistant").
		WithTemperature(0.7)

	if client.DefaultModel() != "test-model" {
		t.Errorf("expected model 'test-model', got %q", client.DefaultModel())
	}
	if client.Provider().Name() != "echo" {
		t.Errorf("expected provider name 'echo', got %q", client.Provider().Name())
	}

	res, err := client.Complete(context.Background(), "Hello Krewire")
	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	if res != "Echo: Hello Krewire" {
		t.Errorf("expected 'Echo: Hello Krewire', got %q", res)
	}

	chatResp, err := client.Chat(context.Background(), []Message{
		SystemMessage("Be concise"),
		UserMessage("What is Go?"),
	})
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}
	if chatResp.Content != "Echo: What is Go?" {
		t.Errorf("expected 'Echo: What is Go?', got %q", chatResp.Content)
	}

	streamCh, err := client.Stream(context.Background(), []Message{UserMessage("Hi")})
	if err != nil {
		t.Fatalf("Stream error: %v", err)
	}
	firstChunk := <-streamCh
	if firstChunk.Delta != "Echo" {
		t.Errorf("expected chunk delta 'Echo', got %q", firstChunk.Delta)
	}
}

func TestMessageHelpers(t *testing.T) {
	sys := SystemMessage("sys")
	if sys.Role != RoleSystem || sys.Content != "sys" {
		t.Errorf("invalid system message: %+v", sys)
	}

	usr := UserMessage("usr")
	if usr.Role != RoleUser || usr.Content != "usr" {
		t.Errorf("invalid user message: %+v", usr)
	}

	asst := AssistantMessage("asst")
	if asst.Role != RoleAssistant || asst.Content != "asst" {
		t.Errorf("invalid assistant message: %+v", asst)
	}

	toolMsg := ToolMessage("id123", "result")
	if toolMsg.Role != RoleTool || toolMsg.ToolCallID != "id123" || toolMsg.Content != "result" {
		t.Errorf("invalid tool message: %+v", toolMsg)
	}
}
