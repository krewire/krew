package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWatcherDetectsGoChange(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := newWatcher(dir, 20*time.Millisecond, nil)
	defer func() { close(w.done) }()

	select {
	case <-w.Changed():
		t.Fatal("must not signal on first scan")
	case <-time.After(60 * time.Millisecond):
	}

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-w.Changed():
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected change signal")
	}
}

func TestWatcherIgnoresNonWatched(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "node_modules", "pkg.js"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := newWatcher(dir, 20*time.Millisecond, nil)
	defer func() { close(w.done) }()

	if err := os.WriteFile(filepath.Join(dir, "node_modules", "pkg.js"), []byte("y\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-w.Changed():
		t.Fatal("node_modules must not be watched")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), "out")
	if err := copyDir(src, dst); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "sub", "b.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "b" {
		t.Errorf("copied content = %q, want b", got)
	}
}

func TestWatcherDetectsTailwindConfigChange(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "tailwind.config.js")
	if err := os.WriteFile(cfgPath, []byte("module.exports = {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := newWatcher(dir, 20*time.Millisecond, nil)
	defer func() { close(w.done) }()

	select {
	case <-w.Changed():
		t.Fatal("must not signal on initial scan")
	case <-time.After(60 * time.Millisecond):
	}

	if err := os.WriteFile(cfgPath, []byte("module.exports = { darkMode: 'class' }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-w.Changed():
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected change signal for tailwind.config.js edit")
	}
}

func TestWatcher_PrunesHeavyDirectoriesFromTraversal(t *testing.T) {
	dir := t.TempDir()

	heavyDirs := []string{
		"node_modules",
		"vendor",
		".git",
		".krewire",
		".hidden",
	}
	for _, h := range heavyDirs {
		sub := filepath.Join(dir, h, "sub")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(sub, "lib.go"), []byte("package sub\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	w := newWatcher(dir, 20*time.Millisecond, nil)
	defer func() { close(w.done) }()

	sig := w.signature()
	for _, h := range heavyDirs {
		if strings.Contains(sig, h) {
			t.Fatalf("expected ignored dir %q to be pruned from watcher signature, got: %s", h, sig)
		}
	}

	for _, h := range heavyDirs {
		target := filepath.Join(dir, h, "sub", "lib.go")
		if err := os.WriteFile(target, []byte("package sub\nfunc Edit(){}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	select {
	case <-w.Changed():
		t.Fatal("watcher must not signal for edits inside pruned directories")
	case <-time.After(100 * time.Millisecond):
	}
}
