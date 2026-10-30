package memory

import (
	"strings"
	"unicode/utf8"

	"github.com/krewire/krewire/pkg/ai"
)

// Token estimation constants for the fast heuristic estimator. All LLM
// tokenizers differ, so the engine deals in conservative approximations:
// roughly one token per four characters of plain text, plus fixed overhead
// per structural element.
const (
	tokensPerRunes   = 4 // one token per this many runes
	messageOverhead  = 4 // role framing overhead for every message
	toolCallOverhead = 8 // framing overhead for every tool call entry
)

// EstimateTokens returns a fast heuristic token count for a text segment.
// It approximates BPE counts without paying a tokenizer dependency.
func EstimateTokens(text string) int {
	runes := utf8.RuneCountInString(text)
	if runes == 0 {
		return 0
	}
	tokens := (runes + tokensPerRunes - 1) / tokensPerRunes
	if tokens == 0 {
		tokens = 1
	}
	return tokens
}

// MessageTokens estimates the token cost of one conversation message,
// including its tool call payloads.
func MessageTokens(msg ai.Message) int {
	tokens := messageOverhead + EstimateTokens(msg.Content)
	for _, call := range msg.ToolCalls {
		tokens += toolCallOverhead + EstimateTokens(call.Name) + EstimateTokens(call.Arguments)
	}
	if msg.ToolCallID != "" {
		tokens += toolCallOverhead
	}
	if strings.TrimSpace(msg.Content) == "" && len(msg.ToolCalls) == 0 && msg.ToolCallID == "" {
		return 0
	}
	return tokens
}

// tokensFor sums heuristic token counts over a message slice.
func tokensFor(msgs []ai.Message) int {
	total := 0
	for _, msg := range msgs {
		total += MessageTokens(msg)
	}
	return total
}
