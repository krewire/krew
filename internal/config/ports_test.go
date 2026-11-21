package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectRuntimePorts_NoConflict(t *testing.T) {
	dir := t.TempDir()
	config := &Config{
		Dev: DevConfig{Port: 8080},
	}

	pkgJSON := `{"scripts": {"dev": "vite --port 3000"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	detected, warnings := DetectRuntimePorts(dir, config)
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %v", warnings)
	}
	if len(detected) < 2 {
		t.Errorf("expected at least 2 detected ports, got %d", len(detected))
	}
}

func TestDetectRuntimePorts_WithConflict(t *testing.T) {
	dir := t.TempDir()
	config := &Config{
		Runtime: RuntimeConfig{Port: 3000},
	}

	pkgJSON := `{"scripts": {"dev": "vite --port 3000"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	_, warnings := DetectRuntimePorts(dir, config)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 conflict warning, got %d: %v", len(warnings), warnings)
	}
}
