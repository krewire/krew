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
	// DefaultAnthropicBaseURL is the standard API base endpoint for Anthropic Claude.
	DefaultAnthropicBaseURL = "https://api.anthropic.com/v1"

	// DefaultAnthropicVersion is the Anthropic protocol version header.
	DefaultAnthropicVersion = "2023-06-01"
)

// AnthropicProvider implements the Anthropic Claude Messages API.
type AnthropicProvider struct {
	BaseURL    string
	APIKey     string
	Version    string
	HTTPClient *http.Client
}

// NewAnthropic creates a new Anthropic provider.
func NewAnthropic(apiKey string, opts ...func(*AnthropicProvider)) *AnthropicProvider {
	p := &AnthropicProvider{
		BaseURL:    DefaultAnthropicBaseURL,
		APIKey:     apiKey,
		Version:    DefaultAnthropicVersion,
		HTTPClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Name returns the provider identifier.
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

type anthropicContentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

type anthropicTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"input_schema"`
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature *float64           `json:"temperature,omitempty"`
	Tools       []anthropicTool    `json:"tools,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

type anthropicResponse struct {
	ID      string                  `json:"id"`
	Type    string                  `json:"type"`
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *AnthropicProvider) buildRequest(req ai.CompletionRequest, stream bool) anthropicRequest {
	var messages []anthropicMessage
	for _, m := range req.Messages {
		role := string(m.Role)
		if m.Role == ai.RoleTool {
			role = "user"
		}
		var blocks []anthropicContentBlock
		if m.Content != "" {
			blocks = append(blocks, anthropicContentBlock{
				Type: "text",
				Text: m.Content,
			})
		}
		for _, tc := range m.ToolCalls {
			blocks = append(blocks, anthropicContentBlock{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Name,
				Input: json.RawMessage(tc.Arguments),
			})
		}
		messages = append(messages, anthropicMessage{
			Role:    role,
			Content: blocks,
		})
	}

	var tools []anthropicTool
	for _, t := range req.Tools {
		tools = append(tools, anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Parameters,
		})
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	return anthropicRequest{
		Model:       req.Model,
		Messages:    messages,
		System:      req.System,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
		Tools:       tools,
		Stream:      stream,
	}
}

// Complete executes an Anthropic message completion.
func (p *AnthropicProvider) Complete(ctx context.Context, req ai.CompletionRequest) (*ai.CompletionResponse, error) {
	if err := ai.ValidateRequest(req); err != nil {
		return nil, err
	}

	payload := p.buildRequest(req, false)
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("anthropic: marshal request: %w", err)
	}

	url := p.BaseURL + "/messages"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("anthropic: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.APIKey)
	httpReq.Header.Set("anthropic-version", p.Version)

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("anthropic: read response: %w", err)
	}

	var res anthropicResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("anthropic: decode response: %w (status %d)", err, resp.StatusCode)
	}

	if res.Error != nil {
		return nil, fmt.Errorf("anthropic api error (%s): %s", res.Error.Type, res.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic api returned status %d: %s", resp.StatusCode, string(body))
	}

	var textBuilder strings.Builder
	var toolCalls []ai.ToolCall

	for _, block := range res.Content {
		switch block.Type {
		case "text":
			textBuilder.WriteString(block.Text)
		case "tool_use":
			toolCalls = append(toolCalls, ai.ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: string(block.Input),
			})
		}
	}

	return &ai.CompletionResponse{
		Content:   textBuilder.String(),
		ToolCalls: toolCalls,
		Usage: ai.TokenUsage{
			PromptTokens:     res.Usage.InputTokens,
			CompletionTokens: res.Usage.OutputTokens,
			TotalTokens:      res.Usage.InputTokens + res.Usage.OutputTokens,
		},
	}, nil
}

// Stream executes an Anthropic streaming completion.
func (p *AnthropicProvider) Stream(ctx context.Context, req ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	if err := ai.ValidateRequest(req); err != nil {
		return nil, err
	}

	payload := p.buildRequest(req, true)
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("anthropic: marshal request: %w", err)
	}

	url := p.BaseURL + "/messages"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("anthropic: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.APIKey)
	httpReq.Header.Set("anthropic-version", p.Version)

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: stream request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("anthropic stream status %d: %s", resp.StatusCode, string(body))
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
			var event struct {
				Type  string `json:"type"`
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(data), &event); err == nil {
				if event.Type == "content_block_delta" && event.Delta.Text != "" {
					if !sendChunk(ai.StreamChunk{Delta: event.Delta.Text}) {
						return
					}
				} else if event.Type == "message_stop" {
					sendChunk(ai.StreamChunk{Done: true})
					return
				}
			}
		}
		if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) && ctx.Err() == nil {
			sendChunk(ai.StreamChunk{Error: err})
		}
	}()

	return ch, nil
}
