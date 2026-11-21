package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/krewire/krew/pkg/ai"
)

func TestMockProvider(t *testing.T) {
	mock := NewMock("Test response")
	if mock.Name() != "mock" {
		t.Errorf("expected name mock, got %q", mock.Name())
	}

	resp, err := mock.Complete(context.Background(), ai.CompletionRequest{
		Model:    "mock-model",
		Messages: []ai.Message{ai.UserMessage("Hi")},
	})
	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	if resp.Content != "Test response" {
		t.Errorf("expected 'Test response', got %q", resp.Content)
	}

	streamCh, err := mock.Stream(context.Background(), ai.CompletionRequest{
		Model:    "mock-model",
		Messages: []ai.Message{ai.UserMessage("Hi")},
	})
	if err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	var collected string
	for chunk := range streamCh {
		collected += chunk.Delta
	}
	if collected != "Test response" {
		t.Errorf("expected collected 'Test response', got %q", collected)
	}
}

func TestOpenAIProvider_HTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.NotFound(w, r)
			return
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		resp := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "Hello from OpenAI mock",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     5,
				"completion_tokens": 10,
				"total_tokens":      15,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOpenAI("test-key", WithBaseURL(server.URL))
	resp, err := client.Complete(context.Background(), ai.CompletionRequest{
		Model:    "gpt-4o",
		Messages: []ai.Message{ai.UserMessage("Hello")},
	})
	if err != nil {
		t.Fatalf("OpenAI Complete error: %v", err)
	}
	if resp.Content != "Hello from OpenAI mock" {
		t.Errorf("expected 'Hello from OpenAI mock', got %q", resp.Content)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("expected TotalTokens 15, got %d", resp.Usage.TotalTokens)
	}
}

func TestAnthropicProvider_HTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/messages" {
			http.NotFound(w, r)
			return
		}
		if key := r.Header.Get("x-api-key"); key != "test-anthropic-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		resp := map[string]any{
			"id":   "msg_123",
			"type": "message",
			"role": "assistant",
			"content": []map[string]any{
				{
					"type": "text",
					"text": "Hello from Claude mock",
				},
			},
			"usage": map[string]any{
				"input_tokens":  8,
				"output_tokens": 12,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropic("test-anthropic-key")
	p.BaseURL = server.URL

	resp, err := p.Complete(context.Background(), ai.CompletionRequest{
		Model:    "claude-3-5-sonnet-20241022",
		Messages: []ai.Message{ai.UserMessage("Hi Claude")},
	})
	if err != nil {
		t.Fatalf("Anthropic Complete error: %v", err)
	}
	if resp.Content != "Hello from Claude mock" {
		t.Errorf("expected 'Hello from Claude mock', got %q", resp.Content)
	}
	if resp.Usage.TotalTokens != 20 {
		t.Errorf("expected TotalTokens 20, got %d", resp.Usage.TotalTokens)
	}
}

func TestGeminiProvider_HTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"role": "model",
						"parts": []map[string]any{
							{"text": "Hello from Gemini mock"},
						},
					},
					"finishReason": "STOP",
				},
			},
			"usageMetadata": map[string]any{
				"promptTokenCount":     4,
				"candidatesTokenCount": 6,
				"totalTokenCount":      10,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewGemini("test-gemini-key")
	p.BaseURL = server.URL

	resp, err := p.Complete(context.Background(), ai.CompletionRequest{
		Model:    "gemini-1.5-pro",
		Messages: []ai.Message{ai.UserMessage("Hi Gemini")},
	})
	if err != nil {
		t.Fatalf("Gemini Complete error: %v", err)
	}
	if resp.Content != "Hello from Gemini mock" {
		t.Errorf("expected 'Hello from Gemini mock', got %q", resp.Content)
	}
}

func TestOllamaProvider_HTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			http.NotFound(w, r)
			return
		}
		resp := map[string]any{
			"model": "llama3",
			"message": map[string]any{
				"role":    "assistant",
				"content": "Hello from Ollama mock",
			},
			"done":              true,
			"prompt_eval_count": 3,
			"eval_count":        7,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOllama()
	p.BaseURL = server.URL

	resp, err := p.Complete(context.Background(), ai.CompletionRequest{
		Model:    "llama3",
		Messages: []ai.Message{ai.UserMessage("Hi Ollama")},
	})
	if err != nil {
		t.Fatalf("Ollama Complete error: %v", err)
	}
	if resp.Content != "Hello from Ollama mock" {
		t.Errorf("expected 'Hello from Ollama mock', got %q", resp.Content)
	}
}

func TestStreamCancellation_OpenAI(t *testing.T) {
	streamStarted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		close(streamStarted)
		for i := 0; ; i++ {
			chunk := fmt.Sprintf("data: {\"choices\":[{\"delta\":{\"content\":\"chunk-%d\"}}]}\n\n", i)
			if _, err := w.Write([]byte(chunk)); err != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	client := NewOpenAI("test-key", WithBaseURL(server.URL))
	ch, err := client.Stream(ctx, ai.CompletionRequest{
		Model:    "gpt-4o",
		Messages: []ai.Message{ai.UserMessage("stream please")},
	})
	if err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	<-streamStarted
	<-ch
	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			for range ch {
			}
		}
	case <-time.After(1 * time.Second):
		t.Fatal("openai stream goroutine did not terminate on context cancellation")
	}
}

func TestStreamCancellation_Anthropic(t *testing.T) {
	streamStarted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		close(streamStarted)
		for i := 0; ; i++ {
			chunk := fmt.Sprintf("data: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text\",\"text\":\"chunk-%d\"}}\n\n", i)
			if _, err := w.Write([]byte(chunk)); err != nil {
				return
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	client := NewAnthropic("test-key")
	client.BaseURL = server.URL
	ch, err := client.Stream(ctx, ai.CompletionRequest{
		Model:    "claude-3-sonnet",
		Messages: []ai.Message{ai.UserMessage("stream please")},
	})
	if err != nil {
		t.Fatalf("Stream error: %v", err)
	}

	<-streamStarted
	<-ch
	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			for range ch {
			}
		}
	case <-time.After(1 * time.Second):
		t.Fatal("anthropic stream goroutine did not terminate on context cancellation")
	}
}
