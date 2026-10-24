package agent

import "time"

// EventType identifies the lifecycle phase of an agent step.
type EventType string

const (
	EventStart      EventType = "start"
	EventThought    EventType = "thought"
	EventToolCall   EventType = "tool_call"
	EventToolResult EventType = "tool_result"
	EventMessage    EventType = "message"
	EventError      EventType = "error"
	EventDone       EventType = "done"
)

// Event is emitted in real time as the agent reasons and executes tools.
type Event struct {
	Type      EventType `json:"type"`
	Step      int       `json:"step"`
	Content   string    `json:"content,omitempty"`
	ToolID    string    `json:"tool_id,omitempty"`
	ToolName  string    `json:"tool_name,omitempty"`
	ToolArgs  string    `json:"tool_args,omitempty"`
	Result    any       `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// EventHandler receives streaming events during agent execution.
type EventHandler func(Event)
