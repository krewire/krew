# `ai`

Import: `github.com/krewire/krew/pkg/ai`

## Purpose

Native AI SDK for Go applications: multi-provider LLM integrations, streaming completions, structured schema generation, object extraction, and agentic tool invocation.

## Main API

- `ai.Provider`: Contract every driver implements — `Name()`, `Complete(ctx, req)`, and `Stream(ctx, req) (<-chan StreamChunk, error)`.
- `ai.NewClient(provider, defaultModel) *Client`: High-level wrapper with `Complete`, `Chat`, `Stream`, `WithSystem`, `WithTemperature`, and `WithTools`.
- `ai.GenerateObject[T](ctx, provider, model, prompt) (*T, error)`: Type-safe structured JSON object extraction.
- `ai.GenerateSchema(t reflect.Type) map[string]any`: Derives a JSON schema from a Go type.
- `ai.ServeSSE(w, stream)`: Streams chunks to a browser as Server-Sent Events.
- `ai.FormatMessages(messages)` / `ai.ValidateRequest(req)`: Prompt formatting and request validation.
- `ai.Tool` / `ai.ToolCall`: Tool definitions with input validation and execution handlers.

## Providers

`github.com/krewire/krew/pkg/ai/providers` ships five drivers:

- `providers.NewOpenAI(apiKey, opts...)` — also accepts `providers.WithBaseURL(url)`.
- `providers.NewAnthropic(apiKey, opts...)`
- `providers.NewGemini(apiKey, opts...)`
- `providers.NewOllama(opts...)` — local models.
- `providers.NewMock(content)` — deterministic provider for tests.

## Example

```go
package main

import (
	"context"
	"fmt"

	"github.com/krewire/krew/pkg/ai"
	"github.com/krewire/krew/pkg/ai/providers"
)

type Summary struct {
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

func main() {
	provider := providers.NewMock(`{"title":"Krewire","tags":["go","modular"]}`)

	obj, err := ai.GenerateObject[Summary](context.Background(), provider, "test-model", "Analyze Krewire")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", obj)
}
```
