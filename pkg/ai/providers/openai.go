package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/krewire/krew/pkg/ai"
)

const (
	// DefaultOpenAIBaseURL is the standard API base endpoint for OpenAI.
	DefaultOpenAIBaseURL = "https://api.openai.com/v1"
)

// OpenAIProvider interacts with OpenAI-compatible chat completion APIs
// (OpenAI, Groq, DeepSeek, Together, vLLM, local proxies).
type OpenAIProvider struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewOpenAI creates an OpenAI provider with optional custom base URL.
func NewOpenAI(apiKey string, opts ...func(*OpenAIProvider)) *OpenAIProvider {
	p := &OpenAIProvider{
		BaseURL:    DefaultOpenAIBaseURL,
		APIKey:     apiKey,
		HTTPClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// WithBaseURL configures a custom endpoint (e.g. for Groq, DeepSeek, or local vLLM).
func WithBaseURL(url string) func(*OpenAIProvider) {
	return func(p *OpenAIProvider) {
		p.BaseURL = strings.TrimSuffix(url, "/")
	}
}

// Name returns the provider identifier.
func (p *OpenAIProvider) Name() string {
	return "openai"
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	Name       string           `json:"name,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openAITool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string         `json:"name"`
		Description string         `json:"description,omitempty"`
		Parameters  map[string]any `json:"parameters,omitempty"`
	} `json:"function"`
}

type openAIRequest struct {
	Model          string          `json:"model"`
	Messages       []openAIMessage `json:"messages"`
	Temperature    *float64        `json:"temperature,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	Tools          []openAITool    `json:"tools,omitempty"`
	Stream         bool            `json:"stream,omitempty"`
	ResponseFormat map[string]any  `json:"response_format,omitempty"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Role      string           `json:"role"`
			Content   string           `json:"content"`
			ToolCalls []openAIToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (p *OpenAIProvider) buildRequest(req ai.CompletionRequest, stream bool) openAIRequest {
	var messages []openAIMessage
	if req.System != "" {
		messages = append(messages, openAIMessage{
			Role:    "system",
			Content: req.System,
		})
	}
	for _, m := range req.Messages {
		om := openAIMessage{
			Role:       string(m.Role),
			Content:    m.Content,
			Name:       m.Name,
			ToolCallID: m.ToolCallID,
		}
		for _, tc := range m.ToolCalls {
			otc := openAIToolCall{
				ID:   tc.ID,
				Type: "function",
			}
			otc.Function.Name = tc.Name
			otc.Function.Arguments = tc.Arguments
			om.ToolCalls = append(om.ToolCalls, otc)
		}
		messages = append(messages, om)
	}

	var tools []openAITool
	for _, t := range req.Tools {
		ot := openAITool{Type: "function"}
		ot.Function.Name = t.Name
		ot.Function.Description = t.Description
		ot.Function.Parameters = t.Parameters
		tools = append(tools, ot)
	}

	out := openAIRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Tools:       tools,
		Stream:      stream,
	}

	if req.ResponseSchema != nil {
		out.ResponseFormat = map[string]any{
			"type": "json_object",
		}
	}

	return out
}

// Complete executes an OpenAI chat completion.
func (p *OpenAIProvider) Complete(ctx context.Context, req ai.CompletionRequest) (*ai.CompletionResponse, error) {
	if err := ai.ValidateRequest(req); err != nil {
		return nil, err
	}

	payload := p.buildRequest(req, false)
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("openai: marshal request: %w", err)
	}

	url := p.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("openai: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai: execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openai: read response body: %w", err)
	}

	var res openAIResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("openai: decode response: %w (status %d)", err, resp.StatusCode)
	}

	if res.Error != nil {
		return nil, fmt.Errorf("openai api error (%s): %s", res.Error.Type, res.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api returned status %d: %s", resp.StatusCode, string(body))
	}
	if len(res.Choices) == 0 {
		return nil, errors.New("openai: no completion choices returned")
	}

	choice := res.Choices[0]
	var toolCalls []ai.ToolCall
	for _, tc := range choice.Message.ToolCalls {
		toolCalls = append(toolCalls, ai.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return &ai.CompletionResponse{
		Content:      choice.Message.Content,
		ToolCalls:    toolCalls,
		FinishReason: choice.FinishReason,
		Usage: ai.TokenUsage{
			PromptTokens:     res.Usage.PromptTokens,
			CompletionTokens: res.Usage.CompletionTokens,
			TotalTokens:      res.Usage.TotalTokens,
		},
	}, nil
}

// Stream executes an OpenAI streaming completion using SSE.
func (p *OpenAIProvider) Stream(ctx context.Context, req ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	if err := ai.ValidateRequest(req); err != nil {
		return nil, err
	}

	payload := p.buildRequest(req, true)
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("openai: marshal request: %w", err)
	}

	url := p.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("openai: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai: stream request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("openai stream returned status %d: %s", resp.StatusCode, string(body))
	}

	ch := make(chan ai.StreamChunk, 20)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		stop := context.AfterFunc(ctx, func() {
			resp.Body.Close()
		})
		defer stop()

		sendChunk := func(chunk ai.StreamChunk) bool {
			select {
			case ch <- chunk:
				return true
			case <-ctx.Done():
				return false
			}
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if strings.TrimSpace(data) == "[DONE]" {
				sendChunk(ai.StreamChunk{Done: true})
				return
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(data), &chunk); err == nil && len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta.Content
				if delta != "" {
					if !sendChunk(ai.StreamChunk{Delta: delta}) {
						return
					}
				}
				if chunk.Choices[0].FinishReason != "" {
					sendChunk(ai.StreamChunk{Done: true})
					return
				}
			}
		}
		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			sendChunk(ai.StreamChunk{Error: err})
		}
	}()

	return ch, nil
}
