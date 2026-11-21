// Package-level file for the git template variant (VariantTemplate): krewire
// init --template shallow-clones a starter repository into an empty target.
package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// equipTemplate bootstraps from a remote (or local) git starter repository.
// The target directory must be empty; the clone is shallow (--depth 1).
func equipTemplate(opts EquipOptions) ([]string, error) {
	url := strings.TrimSpace(opts.TemplateURL)
	if url == "" {
		return nil, fmt.Errorf("equip template: a git URL is required")
	}
	dir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return nil, err
	}
	if empty, err := targetEmpty(dir); err != nil {
		return nil, err
	} else if !empty {
		return nil, fmt.Errorf("%w: %s", ErrNotEmpty, dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	cmd := exec.Command("git", "clone", "--depth", "1", url, dir) // #nosec G204 — devtool, user-supplied URL
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("equip template: git clone: %w", err)
	}
	return walkCreated(dir)
}

// walkCreated lists the files created by a git clone, relative to dir,
// skipping the .git metadata directory.
func walkCreated(dir string) ([]string, error) {
	var created []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			if path == dir {
				return nil
			}
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		created = append(created, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(created)
	return created, nil
}
