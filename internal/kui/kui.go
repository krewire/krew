// Package kui copies Krewire UI (kui) components into a project's components/
// directory and tracks the copies, so customized components are never
// overwritten silently and upstream updates are visible.
//
// Every copy is recorded in components/.kui-lock.json with the kui version
// and the SHA-256 of the file as copied. Comparing that hash with the local
// file tells whether the developer modified it; comparing it with the current
// upstream source tells whether kui changed it.
package kui

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/krewire/krewire/pkg/ui/components"
)

// Dir is the project directory components are copied into.
const Dir = "components"

// LockFile is the lock file name inside Dir.
const LockFile = ".kui-lock.json"

// Lock records the components copied from kui.
type Lock struct {
	Version    int                  `json:"version"`
	Components map[string]LockEntry `json:"components"`
}

// LockEntry records one copied component.
type LockEntry struct {
	// Kui is the kui module version the file was copied from.
	Kui string `json:"kui"`
	// SHA256 is the hash of the file as copied.
	SHA256 string `json:"sha256"`
}

// Status describes a component in a project relative to kui.
type Status string

const (
	StatusMissing   Status = "not installed"
	StatusUpToDate  Status = "up to date"
	StatusModified  Status = "modified"
	StatusOutdated  Status = "update available"
	StatusDiverged  Status = "modified, update available"
	StatusUntracked Status = "untracked (differs from kui)"
)

// Action is what Install did for one component.
type Action string

const (
	ActionAdded     Action = "added"
	ActionUpdated   Action = "updated"
	ActionUnchanged Action = "unchanged"
	ActionReplaced  Action = "replaced (local changes overwritten)"
	ActionKept      Action = "kept (locally modified)"
)

// Result is the outcome for one component.
type Result struct {
	Name   string
	Action Action
}

// ErrModified is returned by Install when a locally modified component was
// kept instead of updated.
var ErrModified = errors.New("kui: locally modified components were kept")

func hash(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// Path returns the project path of a component file.
func Path(root, name string) string {
	return filepath.Join(root, Dir, name+".kiw")
}

// ReadLock reads root/components/.kui-lock.json; a missing file is empty.
func ReadLock(root string) (*Lock, error) {
	l := &Lock{Version: 1, Components: map[string]LockEntry{}}
	data, err := os.ReadFile(filepath.Join(root, Dir, LockFile))
	if errors.Is(err, os.ErrNotExist) {
		return l, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, l); err != nil {
		return nil, fmt.Errorf("kui: %s: %w", LockFile, err)
	}
	if l.Components == nil {
		l.Components = map[string]LockEntry{}
	}
	return l, nil
}

func writeLock(root string, l *Lock) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(l); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, Dir, LockFile), buf.Bytes(), 0o644)
}

// State returns the status of the named component in the project at root.
func State(root, name string, lock *Lock) (Status, error) {
	upstream, err := components.Get(name)
	if err != nil {
		return "", err
	}
	local, err := os.ReadFile(Path(root, name))
	if errors.Is(err, os.ErrNotExist) {
		return StatusMissing, nil
	}
	if err != nil {
		return "", err
	}
	lh, uh := hash(local), hash(upstream)
	entry, tracked := lock.Components[name]
	switch {
	case lh == uh:
		return StatusUpToDate, nil
	case !tracked:
		return StatusUntracked, nil
	case lh == entry.SHA256:
		return StatusOutdated, nil
	case entry.SHA256 == uh:
		return StatusModified, nil
	default:
		return StatusDiverged, nil
	}
}

// ParseRef turns a CLI reference into a Resolve reference: "ui/button" and
// "button" name a component, "ui:minimal" and "pack:minimal" name a pack.
func ParseRef(ref string) string {
	ref = strings.TrimPrefix(ref, "ui/")
	if p, ok := strings.CutPrefix(ref, "ui:"); ok {
		return "pack:" + p
	}
	return ref
}

// Install copies the referenced components and their dependencies into
// root/components. A component that is missing is added; one that is
// unmodified since it was copied is updated to the current kui version; one
// the developer modified (or that differs and was never tracked) is kept
// unless force is set, and ErrModified is returned after all others are done.
// With dryRun nothing is written.
func Install(root string, refs []string, force, dryRun bool) ([]Result, error) {
	for i, r := range refs {
		refs[i] = ParseRef(r)
	}
	names, err := components.Resolve(refs...)
	if err != nil {
		return nil, err
	}
	lock, err := ReadLock(root)
	if err != nil {
		return nil, err
	}
	if !dryRun {
		if err := os.MkdirAll(filepath.Join(root, Dir), 0o755); err != nil {
			return nil, err
		}
	}
	version := components.Version()
	var results []Result
	kept := false
	for _, name := range names {
		upstream, err := components.Get(name)
		if err != nil {
			return nil, err
		}
		st, err := State(root, name, lock)
		if err != nil {
			return nil, err
		}
		var act Action
		switch st {
		case StatusMissing:
			act = ActionAdded
		case StatusUpToDate:
			act = ActionUnchanged
		case StatusOutdated:
			act = ActionUpdated
		default:
			if force {
				act = ActionReplaced
			} else {
				act, kept = ActionKept, true
			}
		}
		results = append(results, Result{Name: name, Action: act})
		if dryRun || act == ActionKept {
			continue
		}
		if act != ActionUnchanged {
			if err := os.WriteFile(Path(root, name), upstream, 0o644); err != nil {
				return nil, err
			}
		}
		if e, ok := lock.Components[name]; !ok || act != ActionUnchanged || e.SHA256 != hash(upstream) {
			lock.Components[name] = LockEntry{Kui: version, SHA256: hash(upstream)}
		}
	}
	if !dryRun {
		if err := writeLock(root, lock); err != nil {
			return nil, err
		}
	}
	if kept {
		return results, ErrModified
	}
	return results, nil
}

// Installed returns the names of components present in root/components that
// are kui components, sorted.
func Installed(root string) []string {
	var out []string
	for _, name := range components.List() {
		if _, err := os.Stat(Path(root, name)); err == nil {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}
