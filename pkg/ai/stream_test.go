package ai

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServeSSE(t *testing.T) {
	ch := make(chan StreamChunk, 4)
	ch <- StreamChunk{Delta: "Hello"}
	ch <- StreamChunk{Delta: " world"}
	ch <- StreamChunk{ToolCall: &ToolCall{ID: "call_1", Name: "search", Arguments: `{"q":"go"}`}}
	ch <- StreamChunk{Done: true}
	close(ch)

	rec := httptest.NewRecorder()
	err := ServeSSE(rec, ch)
	if err != nil {
		t.Fatalf("ServeSSE error: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "event: delta\ndata: {\"delta\":\"Hello\"}") {
		t.Errorf("missing delta Hello in body: %s", body)
	}
	if !strings.Contains(body, "event: delta\ndata: {\"delta\":\" world\"}") {
		t.Errorf("missing delta world in body: %s", body)
	}
	if !strings.Contains(body, "event: tool_call") {
		t.Errorf("missing tool_call in body: %s", body)
	}
	if !strings.Contains(body, "event: done") {
		t.Errorf("missing done event in body: %s", body)
	}
}

func TestServeSSE_Error(t *testing.T) {
	ch := make(chan StreamChunk, 2)
	ch <- StreamChunk{Error: errors.New("rate limited")}
	close(ch)

	rec := httptest.NewRecorder()
	err := ServeSSE(rec, ch)
	if err == nil || !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("expected rate limited error, got %v", err)
	}
	if !strings.Contains(rec.Body.String(), "event: error") {
		t.Errorf("expected event: error in body: %s", rec.Body.String())
	}
}
