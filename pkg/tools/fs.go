package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SafePath ensures the path resolves within root unless root is empty.
func SafePath(root, rel string) (string, error) {
	if root == "" {
		return filepath.Clean(rel), nil
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	target := rel
	if !filepath.IsAbs(target) {
		target = filepath.Join(absRoot, target)
	} else {
		target = filepath.Clean(target)
	}

	if !strings.HasPrefix(target, absRoot) {
		return "", fmt.Errorf("security error: path %q outside sandbox %q", rel, root)
	}
	return target, nil
}

// --- ReadFileTool ---

type ReadFileTool struct {
	Workspace string
}

func (t *ReadFileTool) Name() string { return "read_file" }
func (t *ReadFileTool) Description() string {
	return "Read the contents of a text file from the workspace."
}
func (t *ReadFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Relative or absolute path to the file to read",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, args json.RawMessage) (any, error) {
	var input struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}
	target, err := SafePath(t.Workspace, input.Path)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(target)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}
	return string(b), nil
}

// --- WriteFileTool ---

type WriteFileTool struct {
	Workspace string
}

func (t *WriteFileTool) Name() string { return "write_file" }
func (t *WriteFileTool) Description() string {
	return "Create or overwrite a file with given content."
}
func (t *WriteFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the file to write",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Full file content",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteFileTool) Execute(ctx context.Context, args json.RawMessage) (any, error) {
	var input struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}
	target, err := SafePath(t.Workspace, input.Path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(target, []byte(input.Content), 0o644); err != nil {
		return nil, err
	}
	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(input.Content), input.Path), nil
}

// --- EditFileTool ---

type EditFileTool struct {
	Workspace string
}

func (t *EditFileTool) Name() string { return "edit_file" }
func (t *EditFileTool) Description() string {
	return "Perform exact target string replacement in an existing file."
}
func (t *EditFileTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path of the file to edit",
			},
			"target": map[string]any{
				"type":        "string",
				"description": "Exact text chunk to replace",
			},
			"replacement": map[string]any{
				"type":        "string",
				"description": "New replacement content",
			},
		},
		"required": []string{"path", "target", "replacement"},
	}
}

func (t *EditFileTool) Execute(ctx context.Context, args json.RawMessage) (any, error) {
	var input struct {
		Path        string `json:"path"`
		Target      string `json:"target"`
		Replacement string `json:"replacement"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}
	target, err := SafePath(t.Workspace, input.Path)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(target)
	if err != nil {
		return nil, fmt.Errorf("read file to edit: %w", err)
	}
	content := string(b)
	if !strings.Contains(content, input.Target) {
		return nil, errors.New("target text not found in file")
	}
	updated := strings.Replace(content, input.Target, input.Replacement, 1)
	if err := os.WriteFile(target, []byte(updated), 0o644); err != nil {
		return nil, err
	}
	return "File edited successfully", nil
}

// --- ListDirTool ---

type ListDirTool struct {
	Workspace string
}

func (t *ListDirTool) Name() string { return "list_dir" }
func (t *ListDirTool) Description() string {
	return "List directory contents with file names and directories."
}
func (t *ListDirTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Directory path (defaults to root workspace)",
			},
		},
	}
}

func (t *ListDirTool) Execute(ctx context.Context, args json.RawMessage) (any, error) {
	var input struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal(args, &input)
	if input.Path == "" {
		input.Path = "."
	}
	target, err := SafePath(t.Workspace, input.Path)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, e := range entries {
		suffix := ""
		if e.IsDir() {
			suffix = "/"
		}
		res = append(res, e.Name()+suffix)
	}
	return res, nil
}
