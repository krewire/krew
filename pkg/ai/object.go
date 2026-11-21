package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// GenerateObject queries an LLM to generate a strictly typed struct according to
// JSON schema reflection.
func GenerateObject[T any](ctx context.Context, p Provider, model, prompt string) (*T, error) {
	var zero T
	typ := reflect.TypeOf(zero)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	schema := GenerateSchema(typ)

	req := CompletionRequest{
		Model:          model,
		Messages:       []Message{UserMessage(prompt)},
		ResponseSchema: schema,
	}

	resp, err := p.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ai: generate object: %w", err)
	}

	raw := cleanJSONOutput(resp.Content)
	var out T
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("ai: decode structured object: %w\nraw output: %s", err, raw)
	}

	return &out, nil
}

// cleanJSONOutput removes markdown code blocks and surrounding whitespace.
func cleanJSONOutput(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
	}
	return strings.TrimSpace(s)
}
