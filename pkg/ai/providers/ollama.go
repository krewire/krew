package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/krewire/krew/pkg/ai"
)

const (
	// DefaultOllamaBaseURL is the default local HTTP endpoint for Ollama daemon.
	DefaultOllamaBaseURL = "http://localhost:11434"
)

// OllamaProvider integrates local LLMs running via Ollama.
type OllamaProvider struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewOllama creates an Ollama provider pointing to DefaultOllamaBaseURL by default.
func NewOllama(opts ...func(*OllamaProvider)) *OllamaProvider {
	p := &OllamaProvider{
		BaseURL:    DefaultOllamaBaseURL,
		HTTPClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Name returns the provider identifier.
func (p *OllamaProvider) Name() string {
	return "ollama"
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Format   string          `json:"format,omitempty"`
}

type ollamaResponse struct {
	Model           string        `json:"model"`
	CreatedAt       string        `json:"created_at"`
	Message         ollamaMessage `json:"message"`
	Done            bool          `json:"done"`
	PromptEvalCount int           `json:"prompt_eval_count"`
	EvalCount       int           `json:"eval_count"`
	Error           string        `json:"error,omitempty"`
}

// Complete executes an Ollama chat completion.
func (p *OllamaProvider) Complete(ctx context.Context, req ai.CompletionRequest) (*ai.CompletionResponse, error) {
	if err := ai.ValidateRequest(req); err != nil {
		return nil, err
	}

	var messages []ollamaMessage
	if req.System != "" {
		messages = append(messages, ollamaMessage{
			Role:    "system",
			Content: req.System,
		})
	}
	for _, m := range req.Messages {
		messages = append(messages, ollamaMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}

	payload := ollamaRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   false,
	}
	if req.ResponseSchema != nil {
		payload.Format = "json"
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("ollama: marshal request: %w", err)
	}

	url := strings.TrimSuffix(p.BaseURL, "/") + "/api/chat"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("ollama: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama: execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama: read response: %w", err)
	}

	var res ollamaResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("ollama: decode response: %w", err)
	}

	if res.Error != "" {
		return nil, fmt.Errorf("ollama error: %s", res.Error)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	return &ai.CompletionResponse{
		Content:      res.Message.Content,
		FinishReason: "stop",
		Usage: ai.TokenUsage{
			PromptTokens:     res.PromptEvalCount,
			CompletionTokens: res.EvalCount,
			TotalTokens:      res.PromptEvalCount + res.EvalCount,
		},
	}, nil
}

// Stream executes a streaming chat completion on Ollama.
func (p *OllamaProvider) Stream(ctx context.Context, req ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	resp, err := p.Complete(ctx, req)
	if err != nil {
		return nil, err
	}
	ch := make(chan ai.StreamChunk, 2)
	go func() {
		defer close(ch)
		select {
		case ch <- ai.StreamChunk{Delta: resp.Content}:
		case <-ctx.Done():
			return
		}
		select {
		case ch <- ai.StreamChunk{Done: true}:
		case <-ctx.Done():
			return
		}
	}()
	return ch, nil
}
