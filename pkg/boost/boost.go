// Package boost ships Krewire Boost: a reusable AI agent setup
// (AGENTS.md, opencode.json, and a .agents/ preset) installable into any
// software project.
//
// The template files reside under templates/boost in the workspace and are
// exposed through Template. Install copies them into a target directory,
// refusing to overwrite managed files unless forced.
package boost

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type boostFS struct{}

func (b boostFS) Open(name string) (fs.File, error) {
	tfs, err := findTemplateFS()
	if err != nil {
		return nil, err
	}
	return tfs.Open(name)
}

func (b boostFS) ReadFile(name string) ([]byte, error) {
	tfs, err := findTemplateFS()
	if err != nil {
		return nil, err
	}
	return fs.ReadFile(tfs, name)
}

func (b boostFS) ReadDir(name string) ([]fs.DirEntry, error) {
	tfs, err := findTemplateFS()
	if err != nil {
		return nil, err
	}
	return fs.ReadDir(tfs, name)
}

func findTemplateFS() (fs.FS, error) {
	// 1. Check if environment variable is set
	if env := os.Getenv("KREWIRE_BOOST_TEMPLATES"); env != "" {
		if info, err := os.Stat(env); err == nil && info.IsDir() {
			return os.DirFS(env), nil
		}
	}

	// 2. Search upwards from cwd
	cwd, err := os.Getwd()
	if err == nil {
		dir := cwd
		for i := 0; i < 7; i++ {
			candidate := filepath.Join(dir, "templates", "boost")
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				return os.DirFS(candidate), nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	// 3. Search relative to executable
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for i := 0; i < 7; i++ {
			candidate := filepath.Join(dir, "templates", "boost")
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				return os.DirFS(candidate), nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	return nil, errors.New("boost: templates/boost directory not found")
}

// Template provides access to the Boost template tree under templates/boost.
var Template fs.FS = boostFS{}

// managed is the canonical set of top-level paths Install writes into a
// target directory. Anything copied lives under one of these prefixes.
func managed() []string {
	return []string{
		"AGENTS.md",
		"opencode.json",
		".cursor/mcp.json",
		".agents/agents",
		".agents/commands",
		".agents/skills",
		".agents/rules",
		".agents/templates",
		".agents/tools",
	}
}

var (
	// ErrTargetMissing is returned when the target directory does not exist
	// or is empty.
	ErrTargetMissing = errors.New("install: target directory does not exist")
	// ErrConflicts is returned when a managed file already exists in the
	// target and the install is not forced.
	ErrConflicts = errors.New("install: managed files already exist")
)

// Option configures an Install call.
type Option func(*options)

type options struct {
	force  bool
	dryRun bool
}

// WithForce permits overwriting existing managed files in the target.
func WithForce() Option {
	return func(o *options) { o.force = true }
}

// WithDryRun validates and reports the would-be writes without touching the
// target directory.
func WithDryRun() Option {
	return func(o *options) { o.dryRun = true }
}

// Install copies the Boost template into target. It returns the destination
// paths that were (or, with WithDryRun, would be) written, sorted, and a
// typed error when the target is unusable or existing files conflict.
func Install(target string, opts ...Option) ([]string, error) {
	target = filepath.Clean(target)

	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrTargetMissing, target)
		}
		return nil, fmt.Errorf("install: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrTargetMissing, target)
	}

	config := &options{}
	for _, apply := range opts {
		apply(config)
	}

	conflicts, err := detectConflicts(target)
	if err != nil {
		return nil, err
	}
	if len(conflicts) > 0 && !config.force && !config.dryRun {
		return nil, fmt.Errorf("%w: %s", ErrConflicts, listConflicts(target, conflicts))
	}

	relPaths, err := templatePaths()
	if err != nil {
		return nil, err
	}

	created := make([]string, 0, len(relPaths))
	for _, rel := range relPaths {
		dest := filepath.Join(target, rel)
		if config.dryRun {
			created = append(created, dest)
			continue
		}
		data, err := fs.ReadFile(Template, rel)
		if err != nil {
			return nil, fmt.Errorf("install: read %s: %w", rel, err)
		}
		if err := writeDest(dest, data); err != nil {
			return nil, err
		}
		created = append(created, dest)
	}
	sort.Strings(created)
	return created, nil
}

// Managed returns the canonical top-level paths the installer manages, as
// relative paths inside target directories.
func Managed() []string {
	return append([]string(nil), managed()...)
}

// detectConflicts returns managed paths that already exist under target.
func detectConflicts(target string) ([]string, error) {
	var conflicts []string
	for _, m := range managed() {
		exists, err := anyExist(filepath.Join(target, m))
		if err != nil {
			return nil, err
		}
		if exists {
			conflicts = append(conflicts, m)
		}
	}
	sort.Strings(conflicts)
	return conflicts, nil
}

// anyExist reports whether any file or directory exists at p.
func anyExist(p string) (bool, error) {
	if _, err := os.Lstat(p); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("install: stat %s: %w", p, err)
	}
	return true, nil
}

// templatePaths returns every file path inside the template tree,
// relative to the template root.
func templatePaths() ([]string, error) {
	var rels []string
	err := fs.WalkDir(Template, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rels = append(rels, p)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk template: %w", err)
	}
	return rels, nil
}

func listConflicts(target string, conflicts []string) string {
	parts := make([]string, len(conflicts))
	for i, c := range conflicts {
		parts[i] = filepath.Join(target, c)
	}
	return strings.Join(parts, ", ")
}

func writeDest(dest string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return os.WriteFile(dest, data, 0o644)
}
