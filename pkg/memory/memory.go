package memory

import (
	"sync"

	"github.com/krewire/krewire/pkg/ai"
)

// Memory manages conversation messages and context compaction.
type Memory interface {
	Add(msgs ...ai.Message)
	Messages() []ai.Message
	Clear()
	Len() int
	Compact(maxTokens int)
}

// Buffer implements an in-memory sliding-window conversation memory.
type Buffer struct {
	mu       sync.RWMutex
	system   string
	messages []ai.Message
	maxMsgs  int
}

// NewBuffer creates a conversation memory buffer with an optional message limit.
func NewBuffer(systemPrompt string, maxMessages int) *Buffer {
	if maxMessages <= 0 {
		maxMessages = 100
	}
	return &Buffer{
		system:  systemPrompt,
		maxMsgs: maxMessages,
	}
}

// SetSystem updates or sets the system prompt.
func (b *Buffer) SetSystem(sys string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.system = sys
}

// System returns the system prompt.
func (b *Buffer) System() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.system
}

// Add appends one or more messages to the buffer.
func (b *Buffer) Add(msgs ...ai.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = append(b.messages, msgs...)
	if len(b.messages) > b.maxMsgs {
		excess := len(b.messages) - b.maxMsgs
		b.messages = b.messages[excess:]
	}
}

// Messages returns a copy of all conversation messages including system message.
func (b *Buffer) Messages() []ai.Message {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]ai.Message, 0, len(b.messages)+1)
	if b.system != "" {
		out = append(out, ai.SystemMessage(b.system))
	}
	out = append(out, b.messages...)
	return out
}

// RawMessages returns only the conversation messages without the system prompt.
func (b *Buffer) RawMessages() []ai.Message {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]ai.Message, len(b.messages))
	copy(out, b.messages)
	return out
}

// Clear removes all conversation messages.
func (b *Buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = nil
}

// Len returns the count of conversation messages.
func (b *Buffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.messages)
}

// Compact trims old messages if estimate exceeds limit, preserving the most recent turns.
func (b *Buffer) Compact(maxMessages int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if maxMessages > 0 && len(b.messages) > maxMessages {
		excess := len(b.messages) - maxMessages
		b.messages = b.messages[excess:]
	}
}
