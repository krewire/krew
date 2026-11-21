package ai

import (
	"context"
	"testing"
)

type mockSummaryProvider struct {
	jsonPayload string
}

func (m *mockSummaryProvider) Name() string { return "mock" }
func (m *mockSummaryProvider) Complete(_ context.Context, _ CompletionRequest) (*CompletionResponse, error) {
	return &CompletionResponse{
		Content: m.jsonPayload,
	}, nil
}
func (m *mockSummaryProvider) Stream(_ context.Context, _ CompletionRequest) (<-chan StreamChunk, error) {
	return nil, nil
}

type ArticleSummary struct {
	Title      string   `json:"title"`
	WordCount  int      `json:"word_count"`
	KeyPoints  []string `json:"key_points"`
	IsComplete bool     `json:"is_complete"`
}

func TestGenerateObject(t *testing.T) {
	mockJSON := "```json\n{\n  \"title\": \"Krewire Architecture\",\n  \"word_count\": 1200,\n  \"key_points\": [\"modular\", \"downward deps\", \"zero js\"],\n  \"is_complete\": true\n}\n```"
	p := &mockSummaryProvider{jsonPayload: mockJSON}

	summary, err := GenerateObject[ArticleSummary](context.Background(), p, "mock-model", "Summarize this article")
	if err != nil {
		t.Fatalf("GenerateObject error: %v", err)
	}

	if summary.Title != "Krewire Architecture" {
		t.Errorf("expected Title 'Krewire Architecture', got %q", summary.Title)
	}
	if summary.WordCount != 1200 {
		t.Errorf("expected WordCount 1200, got %d", summary.WordCount)
	}
	if len(summary.KeyPoints) != 3 {
		t.Errorf("expected 3 key points, got %d", len(summary.KeyPoints))
	}
	if !summary.IsComplete {
		t.Errorf("expected IsComplete true")
	}
}
