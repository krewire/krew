package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FetchURLTool performs an HTTP GET request to fetch remote content or API data.
type FetchURLTool struct {
	Client *http.Client
}

func (t *FetchURLTool) Name() string { return "fetch_url" }
func (t *FetchURLTool) Description() string {
	return "Fetch the textual content or JSON of a given public URL over HTTP/HTTPS."
}
func (t *FetchURLTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The URL to fetch",
			},
		},
		"required": []string{"url"},
	}
}

func (t *FetchURLTool) Execute(ctx context.Context, args json.RawMessage) (any, error) {
	var input struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}

	client := t.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, input.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Krew-AI-Agent/0.1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024)) // limit to 512KB
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"status":      resp.StatusCode,
		"status_text": resp.Status,
		"content":     string(body),
	}, nil
}

// DefaultWorkspaceTools returns a ready-to-use registry equipped with file and execution tools.
func DefaultWorkspaceTools(workspace string) *Registry {
	r := NewRegistry()
	r.Register(
		&ReadFileTool{Workspace: workspace},
		&WriteFileTool{Workspace: workspace},
		&EditFileTool{Workspace: workspace},
		&ListDirTool{Workspace: workspace},
		&RunCommandTool{Workspace: workspace},
		&FetchURLTool{},
	)
	return r
}
