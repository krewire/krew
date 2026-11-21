// Package ai provides unified, provider-agnostic LLM primitives for the
// Krewire ecosystem (KWN-AI-001). It powers intelligent web components,
// background worker processing, and autonomous AI coding agent workflows.
package ai

import (
	"context"
	"errors"
	"fmt"
)

// Role represents the role of a message author in a conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message is an individual message in a chat completion conversation.
type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// SystemMessage creates a new system instruction message.
func SystemMessage(content string) Message {
	return Message{Role: RoleSystem, Content: content}
}

// UserMessage creates a new user message.
func UserMessage(content string) Message {
	return Message{Role: RoleUser, Content: content}
}

// AssistantMessage creates a new assistant response message.
func AssistantMessage(content string) Message {
	return Message{Role: RoleAssistant, Content: content}
}

// ToolMessage creates a tool execution result message responding to a specific ToolCallID.
func ToolMessage(toolCallID, content string) Message {
	return Message{Role: RoleTool, ToolCallID: toolCallID, Content: content}
}

// ToolCall represents an LLM request to invoke an external function/tool.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool defines an invokable function registered with the model.
type Tool struct {
	Name        string                                              `json:"name"`
	Description string                                              `json:"description"`
	Parameters  map[string]any                                      `json:"parameters"`
	Handler     func(ctx context.Context, args []byte) (any, error) `json:"-"`
}

// TokenUsage records prompt and completion token statistics.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CompletionRequest is the input payload sent to an AI provider.
type CompletionRequest struct {
	Model          string         `json:"model"`
	Messages       []Message      `json:"messages"`
	System         string         `json:"system,omitempty"`
	Temperature    *float64       `json:"temperature,omitempty"`
	MaxTokens      int            `json:"max_tokens,omitempty"`
	Tools          []Tool         `json:"tools,omitempty"`
	ResponseSchema map[string]any `json:"response_schema,omitempty"`
}

// CompletionResponse is the result returned by an AI provider.
type CompletionResponse struct {
	Content      string     `json:"content"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason,omitempty"`
	Usage        TokenUsage `json:"usage"`
}

// StreamChunk represents a streamed fragment of text or tool execution.
// Usage, when set, carries the token statistics reported with the terminal
// chunk of a stream.
type StreamChunk struct {
	Delta    string
	ToolCall *ToolCall
	Done     bool
	Error    error
	Usage    TokenUsage
}

// Provider defines the standard contract required by all LLM drivers.
type Provider interface {
	Name() string
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
}

// Client wraps an AI provider with sensible defaults and high-level methods.
type Client struct {
	provider     Provider
	defaultModel string
	system       string
	temperature  *float64
	tools        []Tool
}

// NewClient initializes a new high-level AI client.
func NewClient(provider Provider, defaultModel string) *Client {
	return &Client{
		provider:     provider,
		defaultModel: defaultModel,
	}
}

// WithSystem sets a default system prompt for the client.
func (c *Client) WithSystem(sys string) *Client {
	c.system = sys
	return c
}

// WithTemperature sets the sampling temperature.
func (c *Client) WithTemperature(temp float64) *Client {
	c.temperature = &temp
	return c
}

// WithTools registers default tools available to this client.
func (c *Client) WithTools(tools ...Tool) *Client {
	c.tools = append(c.tools, tools...)
	return c
}

// Complete executes a simple single-turn prompt and returns the text response.
func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
	if c.provider == nil {
		return "", errors.New("ai: provider is required")
	}
	req := CompletionRequest{
		Model:       c.defaultModel,
		System:      c.system,
		Temperature: c.temperature,
		Messages:    []Message{UserMessage(prompt)},
		Tools:       c.tools,
	}
	resp, err := c.provider.Complete(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// Chat executes a multi-turn conversation with tool-calling capabilities.
func (c *Client) Chat(ctx context.Context, messages []Message) (*CompletionResponse, error) {
	if c.provider == nil {
		return nil, errors.New("ai: provider is required")
	}
	req := CompletionRequest{
		Model:       c.defaultModel,
		System:      c.system,
		Temperature: c.temperature,
		Messages:    messages,
		Tools:       c.tools,
	}
	return c.provider.Complete(ctx, req)
}

// Stream initiates a streaming completion returning a channel of incremental chunks.
func (c *Client) Stream(ctx context.Context, messages []Message) (<-chan StreamChunk, error) {
	if c.provider == nil {
		return nil, errors.New("ai: provider is required")
	}
	req := CompletionRequest{
		Model:       c.defaultModel,
		System:      c.system,
		Temperature: c.temperature,
		Messages:    messages,
		Tools:       c.tools,
	}
	return c.provider.Stream(ctx, req)
}

// Provider returns the underlying Provider implementation.
func (c *Client) Provider() Provider {
	return c.provider
}

// DefaultModel returns the configured default model name.
func (c *Client) DefaultModel() string {
	return c.defaultModel
}

// ValidateRequest performs basic sanity checks on a completion request.
func ValidateRequest(req CompletionRequest) error {
	if req.Model == "" {
		return errors.New("ai: model name is required")
	}
	if len(req.Messages) == 0 && req.System == "" {
		return errors.New("ai: at least one message or system prompt is required")
	}
	return nil
}

// FormatMessages turns a slice of messages into a debuggable transcript.
func FormatMessages(messages []Message) string {
	var b string
	for _, m := range messages {
		b += fmt.Sprintf("[%s]: %s\n", m.Role, m.Content)
	}
	return b
}
