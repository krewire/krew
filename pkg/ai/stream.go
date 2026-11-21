package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ServeSSE streams chunks from an AI completion stream to an HTTP client using
// standard Server-Sent Events (SSE).
func ServeSSE(w http.ResponseWriter, stream <-chan StreamChunk) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return fmt.Errorf("ai: http.ResponseWriter does not support http.Flusher")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	for chunk := range stream {
		if chunk.Error != nil {
			errPayload, _ := json.Marshal(map[string]string{"error": chunk.Error.Error()})
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", errPayload)
			flusher.Flush()
			return chunk.Error
		}

		if chunk.ToolCall != nil {
			tcPayload, _ := json.Marshal(chunk.ToolCall)
			fmt.Fprintf(w, "event: tool_call\ndata: %s\n\n", tcPayload)
			flusher.Flush()
			continue
		}

		if chunk.Delta != "" {
			deltaPayload, _ := json.Marshal(map[string]string{"delta": chunk.Delta})
			fmt.Fprintf(w, "event: delta\ndata: %s\n\n", deltaPayload)
			flusher.Flush()
		}

		if chunk.Done {
			fmt.Fprintf(w, "event: done\ndata: {}\n\n")
			flusher.Flush()
			return nil
		}
	}

	return nil
}
