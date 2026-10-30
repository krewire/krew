package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/krewire/krewire/pkg/ai"
)

// DefaultTokenBudget is the heuristic context budget applied when no explicit
// limit is configured; it fits typical frontier model windows with headroom
// reserved for tool schemas and responses.
const DefaultTokenBudget = 100_000

// DefaultCompactThreshold is the fraction of the token budget at which
// automatic compaction triggers (85% capacity).
const DefaultCompactThreshold = 0.85

// checkpointPrefix labels the summary block injected after eviction.
const checkpointPrefix = "Prior conversation checkpoint (summary):"

// Compactor condenses evicted conversation turns into a single checkpoint
// summary. Implementations typically call an LLM; tests inject closures.
type Compactor func(ctx context.Context, evicted []ai.Message) (string, error)

// TokenBuffer implements Memory with token-accounted sliding-window eviction
// and automatic checkpoint compaction. It permanently preserves the system
// prompt, the initial user goal, and the most recent turns. Older turns are
// evicted and, when a Compactor is configured, compressed into one checkpoint
// summary asynchronously.
type TokenBuffer struct {
	mu         sync.Mutex
	system     string
	goal       *ai.Message  // initial user goal, never evicted
	checkpoint []ai.Message // summary blocks for evicted turns
	messages   []ai.Message
	speaker    int // heuristic token cost of messages
	maxTokens  int
	keepTurns  int
	threshold  float64
	compactor  Compactor

	gen        sync.WaitGroup
	generation int // bumped on Clear to invalidate in-flight writes
}

// TokenBufferOption customises a TokenBuffer at construction.
type TokenBufferOption func(*TokenBuffer)

// WithMaxTokens sets the total heuristic token budget. Non-positive values
// fall back to DefaultTokenBudget.
func WithMaxTokens(maxTokens int) TokenBufferOption {
	return func(b *TokenBuffer) {
		if maxTokens > 0 {
			b.maxTokens = maxTokens
		}
	}
}

// WithKeepTurns sets how many most-recent messages are always preserved
// during eviction. Non-positive values fall back to 12.
func WithKeepTurns(n int) TokenBufferOption {
	return func(b *TokenBuffer) {
		if n > 0 {
			b.keepTurns = n
		}
	}
}

// WithCompactor installs the summarizer invoked when eviction occurs.
func WithCompactor(c Compactor) TokenBufferOption {
	return func(b *TokenBuffer) {
		if c != nil {
			b.compactor = c
		}
	}
}

// WithCompactThreshold overrides the compaction trigger fraction.
func WithCompactThreshold(f float64) TokenBufferOption {
	return func(b *TokenBuffer) {
		if f > 0 && f < 1 {
			b.threshold = f
		}
	}
}

// NewTokenBuffer builds a token-aware sliding-window memory.
func NewTokenBuffer(systemPrompt string, opts ...TokenBufferOption) *TokenBuffer {
	b := &TokenBuffer{
		system:    systemPrompt,
		maxTokens: DefaultTokenBudget,
		keepTurns: 12,
		threshold: DefaultCompactThreshold,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// SetSystem updates or sets the system prompt.
func (b *TokenBuffer) SetSystem(sys string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.system = sys
}

// System returns the system prompt.
func (b *TokenBuffer) System() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.system
}

// Add appends messages, captures the initial goal, and evicts or compacts
// whenever the buffer reaches the configured capacity threshold.
func (b *TokenBuffer) Add(msgs ...ai.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, msg := range msgs {
		if msg.Role == ai.RoleUser && b.goal == nil {
			goal := msg
			b.goal = &goal
			continue
		}
		b.messages = append(b.messages, msg)
		b.speaker += MessageTokens(msg)
	}
	b.maybeEvictLocked()
}

// Messages returns a copy of the send-ready context: system prompt,
// checkpoint summaries, the initial goal, and surviving turns in FIFO order.
func (b *TokenBuffer) Messages() []ai.Message {
	b.mu.Lock()
	defer b.mu.Unlock()

	out := make([]ai.Message, 0, len(b.messages)+len(b.checkpoint)+1)
	if b.system != "" {
		out = append(out, ai.SystemMessage(b.system))
	}
	out = append(out, b.checkpoint...)
	if b.goal != nil {
		out = append(out, *b.goal)
	}
	out = append(out, b.messages...)
	return out
}

// RawMessages returns only the captured goal and surviving turns, without the
// system prompt or checkpoint blocks.
func (b *TokenBuffer) RawMessages() []ai.Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]ai.Message, 0, len(b.messages)+1)
	if b.goal != nil {
		out = append(out, *b.goal)
	}
	out = append(out, b.messages...)
	return out
}

// Clear removes all conversation state except the system prompt and
// invalidates any in-flight compaction results; completed rewrites are
// discarded by their generation check, so Clear never blocks on a user
// supplied Compactor.
func (b *TokenBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.goal = nil
	b.checkpoint = nil
	b.messages = nil
	b.speaker = 0
	b.generation++
}

// Len returns the count of survived conversation messages (goal excluded).
func (b *TokenBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.messages)
}

// Tokens reports the current heuristic token estimate of all content the
// buffer keeps (system prompt, checkpoints, goal, and turns).
func (b *TokenBuffer) Tokens() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.totalTokensLocked()
}

// Compact forces token-based eviction to fit the given heuristic token
// budget, bypassing the capacity threshold. It satisfies the Memory interface.
func (b *TokenBuffer) Compact(budget int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if budget <= 0 {
		return
	}
	b.evictLocked(budget)
}

// Wait blocks until every in-flight background compaction has settled.
func (b *TokenBuffer) Wait() {
	b.gen.Wait()
}

func (b *TokenBuffer) totalTokensLocked() int {
	total := EstimateTokens(b.system) + tokensFor(b.checkpoint)
	if b.goal != nil {
		total += MessageTokens(*b.goal)
	}
	return total + b.speaker
}

// maybeEvictLocked runs automatic compaction when the buffer passes the
// configured threshold fraction of its budget.
func (b *TokenBuffer) maybeEvictLocked() {
	if len(b.messages) <= b.keepTurns {
		return
	}
	if float64(b.totalTokensLocked()) < b.threshold*float64(b.maxTokens) {
		return
	}
	b.evictLocked(b.maxTokens)
}

// evictLocked drops the oldest turns until the total fits the heuristic
// budget, never touching the keepTurns most recent messages and always
// keeping tool-result chains attached to their requests. When a Compactor is
// installed, the evicted block is summarised asynchronously into the
// checkpoint.
func (b *TokenBuffer) evictLocked(budget int) {
	soft := int(b.threshold * float64(budget))
	var evicted []ai.Message

	for b.totalTokensLocked() > soft && len(b.messages) > b.keepTurns {
		limit := len(b.messages) - b.keepTurns
		if limit <= 0 {
			break
		}
		end := firstTurnBoundary(b.messages[:limit])
		if end <= 0 {
			break
		}
		block := b.messages[:end]
		evicted = append(evicted, block...)
		b.speaker -= tokensFor(block)
		b.messages = append([]ai.Message(nil), b.messages[end:]...)
	}

	if len(evicted) == 0 {
		return
	}
	if b.compactor == nil {
		return
	}
	b.startCompactionLocked(evicted)
}

// firstTurnBoundary returns the index of the next user-role boundary at or
// after position 0 (exclusive), so a complete assistant/tool exchange is
// evicted with its triggering user message.
func firstTurnBoundary(msgs []ai.Message) int {
	for i := 1; i < len(msgs); i++ {
		if msgs[i].Role == ai.RoleUser {
			return i
		}
	}
	return len(msgs)
}

// startCompactionLocked summarises evicted turns in the background. Results
// from superseded generations are discarded.
func (b *TokenBuffer) startCompactionLocked(evicted []ai.Message) {
	prior := append([]ai.Message(nil), b.checkpoint...)
	generation := b.generation
	copies := make([]ai.Message, len(evicted))
	copy(copies, evicted)

	b.gen.Add(1)
	go func() {
		defer b.gen.Done()

		summary, err := b.compactor(context.Background(), copies)
		if err != nil || strings.TrimSpace(summary) == "" {
			return // keep the prior checkpoint intact
		}

		b.mu.Lock()
		defer b.mu.Unlock()
		if generation != b.generation {
			return
		}
		b.checkpoint = append(append([]ai.Message(nil), prior...),
			ai.SystemMessage(fmt.Sprintf("%s\n%s", checkpointPrefix, summary)))
	}()
}
