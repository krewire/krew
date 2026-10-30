package memory

import (
	"fmt"
	"strings"

	"github.com/krewire/krewire/pkg/ai"
)

// RenderTranscriptMarkdown loads a stored transcript and renders it as a
// readable Markdown document with per-role sections. Tool calls appear as
// fenced request blocks and tool results as fenced output blocks.
func RenderTranscriptMarkdown(dir, sessionID string) (string, error) {
	msgs, err := LoadSession(dir, sessionID)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Session %s\n", sessionID))
	b.WriteString(fmt.Sprintf("\n%d messages\n", len(msgs)))

	for _, msg := range msgs {
		b.WriteString("\n---\n")
		switch msg.Role {
		case ai.RoleUser:
			writeSection(&b, "## You", msg.Content)
		case ai.RoleAssistant:
			for _, call := range msg.ToolCalls {
				title := fmt.Sprintf("### Tool call: %s (`%s`)", call.Name, call.ID)
				writeSection(&b, title, call.Arguments)
			}
			if msg.Content != "" {
				writeSection(&b, "## Krew", msg.Content)
			}
		case ai.RoleTool:
			title := fmt.Sprintf("### Tool result (`%s`)", msg.ToolCallID)
			writeSection(&b, title, msg.Content)
		case ai.RoleSystem:
			// Sequences replayed from memory carry the system prompt; render
			// it once as context metadata.
			writeSection(&b, "### System prompt", msg.Content)
		default:
			writeSection(&b, "### "+string(msg.Role), msg.Content)
		}
	}
	return b.String(), nil
}

// writeSection appends a heading and its fenced body so multi-line tool
// output and assistant prose stay delimited in the exported document.
func writeSection(b *strings.Builder, heading, content string) {
	content = strings.TrimRight(content, "\n")
	b.WriteString("\n" + heading + "\n")
	if content != "" {
		b.WriteString("\n```\n" + content + "\n```\n")
	}
}
