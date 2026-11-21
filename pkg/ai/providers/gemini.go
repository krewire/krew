package providers

import (
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
	// DefaultGeminiBaseURL is the standard API base endpoint for Google Gemini.
	DefaultGeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"
)

// GeminiProvider implements the Google Gemini / Vertex AI REST API.
type GeminiProvider struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewGemini creates a Google Gemini provider.
func NewGemini(apiKey string, opts ...func(*GeminiProvider)) *GeminiProvider {
	p := &GeminiProvider{
		BaseURL:    DefaultGeminiBaseURL,
		APIKey:     apiKey,
		HTTPClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Name returns the provider identifier.
func (p *GeminiProvider) Name() string {
	return "gemini"
}

type geminiPart struct {
	Text         string `json:"text,omitempty"`
	FunctionCall *struct {
		Name string         `json:"name"`
		Args map[string]any `json:"args"`
	} `json:"functionCall,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiFunctionDeclaration struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type geminiRequest struct {
	Contents          []geminiContent `json:"contents"`
	SystemInstruction *struct {
		Parts []geminiPart `json:"parts"`
	} `json:"systemInstruction,omitempty"`
	Tools []struct {
		FunctionDeclarations []geminiFunctionDeclaration `json:"functionDeclarations"`
	} `json:"tools,omitempty"`
	GenerationConfig *struct {
		Temperature      *float64 `json:"temperature,omitempty"`
		MaxOutputTokens  int      `json:"maxOutputTokens,omitempty"`
		ResponseMimeType string   `json:"responseMimeType,omitempty"`
	} `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
			Role  string       `json:"role"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func (p *GeminiProvider) buildRequest(req ai.CompletionRequest) geminiRequest {
	var contents []geminiContent
	for _, m := range req.Messages {
		role := "user"
		if m.Role == ai.RoleAssistant {
			role = "model"
		}
		var parts []geminiPart
		if m.Content != "" {
			parts = append(parts, geminiPart{Text: m.Content})
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: parts,
		})
	}

	out := geminiRequest{
		Contents: contents,
	}

	if req.System != "" {
		out.SystemInstruction = &struct {
			Parts []geminiPart `json:"parts"`
		}{
			Parts: []geminiPart{{Text: req.System}},
		}
	}

	if len(req.Tools) > 0 {
		var decls []geminiFunctionDeclaration
		for _, t := range req.Tools {
			decls = append(decls, geminiFunctionDeclaration{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			})
		}
		out.Tools = []struct {
			FunctionDeclarations []geminiFunctionDeclaration `json:"functionDeclarations"`
		}{
			{FunctionDeclarations: decls},
		}
	}

	if req.Temperature != nil || req.MaxTokens > 0 || req.ResponseSchema != nil {
		out.GenerationConfig = &struct {
			Temperature      *float64 `json:"temperature,omitempty"`
			MaxOutputTokens  int      `json:"maxOutputTokens,omitempty"`
			ResponseMimeType string   `json:"responseMimeType,omitempty"`
		}{
			Temperature:     req.Temperature,
			MaxOutputTokens: req.MaxTokens,
		}
		if req.ResponseSchema != nil {
			out.GenerationConfig.ResponseMimeType = "application/json"
		}
	}

	return out
}

// Complete executes a Gemini completion.
func (p *GeminiProvider) Complete(ctx context.Context, req ai.CompletionRequest) (*ai.CompletionResponse, error) {
	if err := ai.ValidateRequest(req); err != nil {
		return nil, err
	}

	payload := p.buildRequest(req)
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("gemini: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.BaseURL, req.Model, p.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("gemini: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini: execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gemini: read response: %w", err)
	}

	var res geminiResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("gemini: decode response: %w (status %d)", err, resp.StatusCode)
	}

	if res.Error != nil {
		return nil, fmt.Errorf("gemini api error (%s): %s", res.Error.Status, res.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini api returned status %d: %s", resp.StatusCode, string(body))
	}
	if len(res.Candidates) == 0 {
		return nil, errors.New("gemini: no candidates returned")
	}

	cand := res.Candidates[0]
	var textBuilder strings.Builder
	var toolCalls []ai.ToolCall

	for _, part := range cand.Content.Parts {
		if part.Text != "" {
			textBuilder.WriteString(part.Text)
		}
		if part.FunctionCall != nil {
			argsBytes, _ := json.Marshal(part.FunctionCall.Args)
			toolCalls = append(toolCalls, ai.ToolCall{
				ID:        part.FunctionCall.Name,
				Name:      part.FunctionCall.Name,
				Arguments: string(argsBytes),
			})
		}
	}

	return &ai.CompletionResponse{
		Content:      textBuilder.String(),
		ToolCalls:    toolCalls,
		FinishReason: cand.FinishReason,
		Usage: ai.TokenUsage{
			PromptTokens:     res.UsageMetadata.PromptTokenCount,
			CompletionTokens: res.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      res.UsageMetadata.TotalTokenCount,
		},
	}, nil
}

// Stream executes a Gemini stream request.
func (p *GeminiProvider) Stream(ctx context.Context, req ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	// For standard stream requests, fall back to complete or stream
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
