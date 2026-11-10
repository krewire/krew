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

// TestLayering verifies downward-only dependencies across krew monorepo subsystems.
func TestLayering(t *testing.T) {
	root := repoDir(t)
	fset := token.NewFileSet()

	// Leaf subsystems must never import high-level orchestrators
	bannedInLeaf := map[string][]string{
		"tools":   {"github.com/krewire/krew/pkg/agent", "github.com/krewire/krew/cmd"},
		"memory":  {"github.com/krewire/krew/pkg/agent", "github.com/krewire/krew/cmd"},
		"persona": {"github.com/krewire/krew/pkg/agent", "github.com/krewire/krew/cmd"},
	}

	err := filepath.Walk(filepath.Join(root, "pkg"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}

		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) < 2 {
			return nil
		}
		subsystem := parts[1]

		banned, exists := bannedInLeaf[subsystem]
		if !exists {
			return nil
		}

		node, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			t.Fatalf("failed to parse %s: %v", path, parseErr)
		}

		for _, imp := range node.Imports {
			val, _ := strconv.Unquote(imp.Path.Value)
			for _, b := range banned {
				if strings.HasPrefix(val, b) {
					t.Errorf("layering violation in %s: subsystem %q must not import %q", rel, subsystem, val)
				}
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("failed walking pkg directory: %v", err)
	}
}

// TestPeerPlatformDecoupling enforces Zero Peer Coupling (Lateral Independence).
// Sibling platforms at Layer 4 (cloud, forge, auth, krew) and Layer 5 (dash)
// must never import source code or private packages of peer platforms directly.
func TestPeerPlatformDecoupling(t *testing.T) {
	root := repoDir(t)
	fset := token.NewFileSet()

	bannedPeerPrefixes := []string{
		"github.com/krewire/cloud",
		"github.com/krewire/forge",
		"github.com/krewire/auth",
		"github.com/krewire/dash",
	}

	for _, scanDir := range []string{"pkg", "cmd"} {
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
