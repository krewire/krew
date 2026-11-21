package commands

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func TestRunBundle_DirectAssets(t *testing.T) {
	tmp := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	// Create minimal krewire.yaml
	if err := os.WriteFile(filepath.Join(tmp, "krewire.yaml"), []byte("project:\n  name: test\n  kind: site\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create dummy public/assets
	pubAssets := filepath.Join(tmp, "public", "assets")
	if err := os.MkdirAll(pubAssets, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pubAssets, "custom.css"), []byte(".card{padding:10px;}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pubAssets, "custom.js"), []byte("console.log('hi');"), 0o644); err != nil {
		t.Fatal(err)
	}

	fs := flag.NewFlagSet("bundle", flag.ContinueOnError)
	RegisterBundle(fs)
	_ = fs.Parse([]string{"--output", ".krewire/build"})

	code := RunBundle(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected ExitCodeSuccess, got %v", code)
	}

	cssOut := filepath.Join(tmp, ".krewire", "build", "assets", "krewire.css")
	jsOut := filepath.Join(tmp, ".krewire", "build", "assets", "krewire.js")

	if _, err := os.Stat(cssOut); err != nil {
		t.Errorf("missing bundled css: %v", err)
	}
	if _, err := os.Stat(jsOut); err != nil {
		t.Errorf("missing bundled js: %v", err)
	}
}
