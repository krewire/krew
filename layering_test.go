package krew_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func repoDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "krew", "go.mod")); err == nil {
		return filepath.Join(dir, "krew")
	}
	return dir
}

// TestDevtoolIsolation verifies that krew functions purely as a developer toolchain CLI
// and contains no internal/pkg AI engine packages, which have been promoted to core (krewire/pkg/ai).
func TestDevtoolIsolation(t *testing.T) {
	root := repoDir(t)
	pkgDir := filepath.Join(root, "pkg")
	if info, err := os.Stat(pkgDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(pkgDir)
		if err == nil && len(entries) > 0 {
			t.Errorf("krew must only be a devtool CLI: unexpected packages in %s", pkgDir)
		}
	}
}

// TestPeerPlatformDecoupling enforces Zero Peer Coupling (Lateral Independence).
// Sibling platforms at Layer 4 (cloud, auth, krew) and Layer 5 (dash)
// must never import source code or private packages of peer platforms directly.
func TestPeerPlatformDecoupling(t *testing.T) {
	root := repoDir(t)
	fset := token.NewFileSet()

	bannedPeerPrefixes := []string{
		"github.com/krewire/cloud",
		"github.com/krewire/auth",
		"github.com/krewire/dash",
	}

	for _, scanDir := range []string{"pkg", "cmd", "internal"} {
		dirPath := filepath.Join(root, scanDir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}
		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
				return err
			}
			rel, _ := filepath.Rel(root, path)

			node, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if parseErr != nil {
				t.Fatalf("failed to parse %s: %v", path, parseErr)
			}

			for _, imp := range node.Imports {
				val, _ := strconv.Unquote(imp.Path.Value)
				for _, banned := range bannedPeerPrefixes {
					if strings.HasPrefix(val, banned) {
						t.Errorf("zero peer coupling violation in %s: krew must not import sibling platform %q", rel, val)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("failed walking %s directory: %v", scanDir, err)
		}
	}
}
