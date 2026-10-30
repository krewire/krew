package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/krewire/krewire/pkg/ai"
)

// compactorSystemPrompt directs the LLM to write one dense checkpoint out of
// evicted turns, prioritising durable facts over chatter.
const compactorSystemPrompt = `You compact finished conversation turns of an autonomous coding agent into a concise checkpoint summary.
Preserve: the user's intent, decisions made, files and commands touched, errors encountered, and any pending follow-ups.
Drop filler and repeated tool noise. Answer with the summary only.`

// LLMCompactor adapts an ai.Provider into a Compactor that compresses evicted
// turns into a model-generated checkpoint summary in one completion call.
func LLMCompactor(provider ai.Provider, model string) Compactor {
	return func(ctx context.Context, evicted []ai.Message) (string, error) {
		if provider == nil {
			return "", fmt.Errorf("compactor requires an ai provider")
		}
		transcript, err := json.Marshal(evicted)
		if err != nil {
			return "", fmt.Errorf("encode evicted turns: %w", err)
		}
		req := ai.CompletionRequest{
			Model: model,
			Messages: []ai.Message{
				ai.SystemMessage(compactorSystemPrompt),
				ai.UserMessage(string(transcript)),
			},
		}
		resp, err := provider.Complete(ctx, req)
		if err != nil {
			return "", fmt.Errorf("compaction completion: %w", err)
		}
		return strings.TrimSpace(resp.Content), nil
	}
}
