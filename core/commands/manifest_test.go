package commands

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestManifestLifecycle(t *testing.T) {
	dir := t.TempDir()

	// Initial write
	files := []string{"index.html", "about.html", "assets/style.css"}
	writeManifest(dir, files)

	// Read back
	got := readManifest(dir)
	if !slices.Equal(got, files) {
		t.Fatalf("readManifest() = %v, want %v", got, files)
	}

	// Create fake files
	for _, f := range files {
		p := filepath.Join(dir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Verify collectCreated sees all files except the manifest
	created := collectCreated(dir)
	slices.Sort(created)
	if !slices.Equal(created, files) {
		t.Fatalf("collectCreated() = %v, want %v", created, files)
	}

	// Prune stale: should delete all files listed in manifest and the manifest itself
	pruneStale(dir)
	for _, f := range files {
		p := filepath.Join(dir, filepath.FromSlash(f))
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("file %s should have been pruned", f)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, manifestName)); !os.IsNotExist(err) {
		t.Errorf("manifest file %s should have been removed", manifestName)
	}
}

func TestManifestWithMultiRuntimeOutputs(t *testing.T) {
	dir := t.TempDir()
	files := []string{"index.html"}
	outputs := map[string]string{
		"nodejs":  "dist",
		"laravel": "public/build",
	}

	writeManifest(dir, files, outputs)

	// readManifest should ignore comment headers and return only relative file paths
	gotFiles := readManifest(dir)
	if !slices.Equal(gotFiles, files) {
		t.Fatalf("readManifest() = %v, want %v", gotFiles, files)
	}

	// readManifestOutputs should retrieve the runtime mappings
	gotOutputs := readManifestOutputs(dir)
	if len(gotOutputs) != 2 || gotOutputs["nodejs"] != "dist" || gotOutputs["laravel"] != "public/build" {
		t.Fatalf("readManifestOutputs() = %v, want %v", gotOutputs, outputs)
	}
}
