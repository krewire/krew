package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// seedWorkspace builds a small tree used by the search-tool tests.
func seedWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(rel, content string) {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	write("pkg/api/handler.go", "package api\n\nfunc Handler() {}\nfunc helper() {}\n")
	write("pkg/api/handler_test.go", "package api\n\nfunc TestHandler() {}\n")
	write("cmd/main.go", "package main\n\nfunc main() { Handler() }\n")
	write("node_modules/pkg/index.go", "package ignore\nfunc fim() {}\n")
	write("api/handler.go", "package api\n\nfunc Handler() {}\nfunc helper() {}\n")
	write("api/handler_test.go", "package api\n\nfunc TestHandler() {}\n")
	write("main.go", "package main\n\nfunc main() { Handler() }\n")
	write("binary.dat", "\x00\x01\x02not_utf8")
	return root
}

func TestGrepSearch_MatchAndLineNumbers(t *testing.T) {
	root := seedWorkspace(t)
	tool := &GrepSearchTool{Workspace: root}
	args, _ := json.Marshal(map[string]any{"pattern": "func Handler"})

	raw, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	hits := raw.([]SearchHit)
	if len(hits) != 2 {
		t.Fatalf("expected 2 matches, got %+v", hits)
	}
	// `pkgs/api/handler.go` walks before `cmd`, lexical order asserts.
	first := hits[0]
	if first.Path != "api/handler.go" || first.Line != 3 {
		t.Errorf("unexpected first hit: %+v", first)
	}
	if first.Text != "func Handler() {}" {
		t.Errorf("unexpected matched text: %+v", first.Text)
	}

	// Include and exclude filters together.
	args, _ = json.Marshal(map[string]any{
		"pattern": "func",
		"include": "*.go",
		"exclude": "*_test.go",
	})
	raw, err = tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	for _, h := range raw.([]SearchHit) {
		if filepath.Ext(h.Path) == "" || h.Path == "api/handler_test.go" {
			t.Errorf("exclude filter leaked: %+v", h)
		}
	}
}

func TestGrepSearch_BinaryAndSkipDirs(t *testing.T) {
	root := seedWorkspace(t)
	tool := &GrepSearchTool{Workspace: root}

	args, _ := json.Marshal(map[string]any{"pattern": "not_utf8|ignore"})
	raw, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if hits := raw.([]SearchHit); len(hits) != 0 {
		t.Fatalf("expected binary and skipped matches suppressed, got %+v", hits)
	}
}

func TestFindFiles(t *testing.T) {
	root := seedWorkspace(t)
	tool := &FindFilesTool{Workspace: root}

	t.Run("name glob across depths", func(t *testing.T) {
		args, _ := json.Marshal(map[string]string{"pattern": "*.go"})
		raw, err := tool.Execute(context.Background(), args)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		found := raw.([]string)
		for _, f := range []string{"api/handler_test.go", "pkg/api/handler.go", "main.go", "pkg/api/handler.go"} {
			if !contains(found, f) {
				t.Errorf("expected %q in %v", f, found)
			}
		}
		// node_modules must never surface.
		for _, f := range found {
			if filepath.HasPrefix(f, "node_modules") {
				t.Errorf("skip dir leaked: %v", found)
			}
		}
	})

	t.Run("single hit star path", func(t *testing.T) {
		args, _ := json.Marshal(map[string]string{"pattern": "api/*.go"})
		raw, err := tool.Execute(context.Background(), args)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		found := raw.([]string)
		for _, f := range found {
			if f == "pkg/api/handler.go" || f == "pkg/api/handler_test.go" {
				t.Errorf("`*` must not cross directory boundaries: %v", found)
			}
		}
	})

	t.Run("double star recursion", func(t *testing.T) {
		args, _ := json.Marshal(map[string]string{"pattern": "**/*_test.go"})
		raw, err := tool.Execute(context.Background(), args)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if !contains(raw.([]string), "api/handler_test.go") {
			t.Errorf("expected recursive match, got %v", raw)
		}
	})
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}
