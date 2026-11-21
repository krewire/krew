package gomod

import (
	"os"
	"path/filepath"
	"testing"
)

func writeMod(t *testing.T, dir, content string) string {
	t.Helper()
	p := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestReadParsesModuleAndRequireBlock(t *testing.T) {
	dir := t.TempDir()
	p := writeMod(t, dir, `module github.com/acme/demo

go 1.22

require (
	github.com/krewire/krewire/pkg/model v0.0.0-20260818-abcdef
	github.com/krewire/krewire/pkg v0.1.0
)
`)

	m, err := Read(p)
	if err != nil {
		t.Fatal(err)
	}
	if m.Path != "github.com/acme/demo" {
		t.Errorf("Path = %q, want github.com/acme/demo", m.Path)
	}
	if got := m.Requires["github.com/krewire/krewire/pkg/model"]; got != "v0.0.0-20260818-abcdef" {
		t.Errorf("framework version = %q", got)
	}
	if got := m.Requires["github.com/krewire/krewire/pkg"]; got != "v0.1.0" {
		t.Errorf("libs version = %q", got)
	}
	if !m.UsesKrewire() {
		t.Error("UsesKrewire() = false, want true")
	}
}

func TestReadParsesSingleLineRequire(t *testing.T) {
	dir := t.TempDir()
	p := writeMod(t, dir, `module demo

go 1.22

require github.com/krewire/krewire/pkg v0.1.0
`)

	m, err := Read(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Requires["github.com/krewire/krewire/pkg"]; got != "v0.1.0" {
		t.Errorf("libs version = %q, want v0.1.0", got)
	}
}

func TestReadRejectsMissingModule(t *testing.T) {
	dir := t.TempDir()
	p := writeMod(t, dir, "go 1.22\n")
	if _, err := Read(p); err == nil {
		t.Error("Read() error = nil, want a module directive error")
	}
}

func TestFindWalksUpDirectories(t *testing.T) {
	dir := t.TempDir()
	writeMod(t, dir, "module github.com/acme/demo\n\ngo 1.22\n")
	inner := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}

	m, err := Find(inner)
	if err != nil {
		t.Fatal(err)
	}
	if m.Path != "github.com/acme/demo" {
		t.Errorf("Path = %q", m.Path)
	}
}

func TestFindMissing(t *testing.T) {
	dir := t.TempDir()
	if _, err := Find(dir); err == nil {
		t.Error("Find() error = nil, want an error")
	}
}

func TestCheckGoVersion(t *testing.T) {
	tests := []struct {
		runtimeVer string
		targetVer  string
		wantErr    bool
	}{
		{"go1.27.1", "1.27.1", false},
		{"go1.27.2", "1.27.1", false},
		{"go1.28.0", "1.27.1", false},
		{"go1.26.9", "1.27.1", true},
		{"go1.20.0", "1.27.1", true},
		{"devel go1.28-abc", "1.27.1", false},
		{"go1.27.1 linux/amd64", "1.27.1", false},
	}

	for _, tt := range tests {
		err := CheckGoVersion(tt.runtimeVer, tt.targetVer)
		if (err != nil) != tt.wantErr {
			t.Errorf("CheckGoVersion(%q, %q) err = %v, wantErr = %v", tt.runtimeVer, tt.targetVer, err, tt.wantErr)
		}
	}
}
