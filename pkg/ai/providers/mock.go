package providers

import (
	"context"
	"strings"

	"github.com/krewire/krew/pkg/ai"
)

// MockProvider provides deterministic, offline responses for unit testing.
type MockProvider struct {
	ResponseContent string
	ResponseTool    *ai.ToolCall
	Err             error
}

// NewMock creates a MockProvider with a static response string.
func NewMock(content string) *MockProvider {
	return &MockProvider{
		ResponseContent: content,
	}
}

// Name returns the provider identifier.
func (m *MockProvider) Name() string {
	return "mock"
}

// Complete returns the mock response.
func (m *MockProvider) Complete(_ context.Context, _ ai.CompletionRequest) (*ai.CompletionResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	var toolCalls []ai.ToolCall
	if m.ResponseTool != nil {
		toolCalls = append(toolCalls, *m.ResponseTool)
	}

	return &ai.CompletionResponse{
		Content:      m.ResponseContent,
		ToolCalls:    toolCalls,
		FinishReason: "stop",
		Usage: ai.TokenUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}, nil
}

// Stream emits the mock response word by word over a channel.
func (m *MockProvider) Stream(_ context.Context, _ ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	ch := make(chan ai.StreamChunk, 10)
	go func() {
		defer close(ch)
		words := strings.Fields(m.ResponseContent)
		for i, w := range words {
			suffix := " "
			if i == len(words)-1 {
				suffix = ""
			}
			ch <- ai.StreamChunk{
				Delta: w + suffix,
			}
		}
		if m.ResponseTool != nil {
			ch <- ai.StreamChunk{
				ToolCall: m.ResponseTool,
			}
		}
		ch <- ai.StreamChunk{
			Done: true,
			Usage: ai.TokenUsage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		}
	}()

	return ch, nil
}
