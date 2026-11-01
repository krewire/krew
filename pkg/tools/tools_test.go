package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSafePath(t *testing.T) {
	tmp := t.TempDir()

	// Inside
	p, err := SafePath(tmp, "sub/file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != filepath.Join(tmp, "sub/file.txt") {
		t.Errorf("wrong path: %s", p)
	}

	// Outside escape attempt
	_, err = SafePath(tmp, "../escaped.txt")
	if err == nil {
		t.Fatal("expected escape error, got nil")
	}
}

func TestSafePath_SymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()

	// 1. Existing file outside root symlinked into root
	outsideFile := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("super secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	symlinkToFile := filepath.Join(root, "link_to_outside.txt")
	if err := os.Symlink(outsideFile, symlinkToFile); err != nil {
		t.Fatal(err)
	}

	_, err := SafePath(root, "link_to_outside.txt")
	if !errors.Is(err, ErrPathEscape) {
		t.Fatalf("expected ErrPathEscape for symlinked file pointing outside, got: %v", err)
	}

	// 2. Symlinked directory pointing outside root
	symlinkToDir := filepath.Join(root, "link_to_out_dir")
	if err := os.Symlink(outside, symlinkToDir); err != nil {
		t.Fatal(err)
	}

	_, err = SafePath(root, "link_to_out_dir/new_file.txt")
	if !errors.Is(err, ErrPathEscape) {
		t.Fatalf("expected ErrPathEscape for new file inside symlinked outside dir, got: %v", err)
	}

	// 3. ReadFileTool and WriteFileTool through symlinks return ErrPathEscape
	readTool := &ReadFileTool{Workspace: root}
	rArgs, _ := json.Marshal(map[string]string{"path": "link_to_outside.txt"})
	if _, err := readTool.Execute(context.Background(), rArgs); !errors.Is(err, ErrPathEscape) {
		t.Fatalf("expected ReadFileTool to return ErrPathEscape on symlink escape, got: %v", err)
	}

	writeTool := &WriteFileTool{Workspace: root}
	wArgs, _ := json.Marshal(map[string]string{"path": "link_to_out_dir/test.txt", "content": "bad"})
	if _, err := writeTool.Execute(context.Background(), wArgs); !errors.Is(err, ErrPathEscape) {
		t.Fatalf("expected WriteFileTool to return ErrPathEscape on symlink escape, got: %v", err)
	}
}

func TestFSTools_Lifecycle(t *testing.T) {
	tmp := t.TempDir()
	ctx := context.Background()

	writeTool := &WriteFileTool{Workspace: tmp}
	readTool := &ReadFileTool{Workspace: tmp}
	editTool := &EditFileTool{Workspace: tmp}
	listTool := &ListDirTool{Workspace: tmp}

	// 1. Write file
	wArgs, _ := json.Marshal(map[string]string{
		"path":    "hello.txt",
		"content": "Hello World",
	})
	if _, err := writeTool.Execute(ctx, wArgs); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// 2. Read file
	rArgs, _ := json.Marshal(map[string]string{
		"path": "hello.txt",
	})
	content, err := readTool.Execute(ctx, rArgs)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if content.(string) != "Hello World" {
		t.Errorf("expected 'Hello World', got %v", content)
	}

	// 3. Edit file
	eArgs, _ := json.Marshal(map[string]string{
		"path":        "hello.txt",
		"target":      "World",
		"replacement": "Krew",
	})
	if _, err := editTool.Execute(ctx, eArgs); err != nil {
		t.Fatalf("edit failed: %v", err)
	}

	content, _ = readTool.Execute(ctx, rArgs)
	if content.(string) != "Hello Krew" {
		t.Errorf("expected 'Hello Krew', got %v", content)
	}

	// 4. List dir
	lArgs, _ := json.Marshal(map[string]string{"path": "."})
	entries, err := listTool.Execute(ctx, lArgs)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	list := entries.([]string)
	if len(list) != 1 || list[0] != "hello.txt" {
		t.Errorf("expected [hello.txt], got %v", list)
	}
}

func TestRunCommandTool(t *testing.T) {
	tool := &RunCommandTool{}
	ctx := context.Background()

	args, _ := json.Marshal(map[string]string{
		"command": "echo 'krew agent test'",
	})
	res, err := tool.Execute(ctx, args)
	if err != nil {
		t.Fatalf("exec error: %v", err)
	}
	m := res.(map[string]any)
	if m["stdout"] != "krew agent test\n" {
		t.Errorf("expected stdout 'krew agent test\\n', got %q", m["stdout"])
	}
}

func TestRegistry(t *testing.T) {
	r := DefaultWorkspaceTools(os.TempDir())
	if _, ok := r.Get("read_file"); !ok {
		t.Error("missing read_file tool")
	}
	if _, ok := r.Get("run_command"); !ok {
		t.Error("missing run_command tool")
	}
	aiTools := r.ToAITools()
	if len(aiTools) != 6 {
		t.Errorf("expected 6 ai tools, got %d", len(aiTools))
	}
}
