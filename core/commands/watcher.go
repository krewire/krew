package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/krewire/krew/core/config"
)

// watcher polls the watched roots for changes, emitting on every change of
// the signature. It is single-consumer: call Reset after each change.
type watcher struct {
	root      string
	every     time.Duration
	ch        chan struct{}
	sig       string
	done      chan struct{}
	watchDirs []string // additional directories to watch (from project.dirs)
}

// newWatcher returns a watcher polling root every d.
func newWatcher(root string, d time.Duration, cfg *config.Config) *watcher {
	watchDirs := defaultWatchDirs()
	if cfg != nil && cfg.Project.Dirs != (config.Dirs{}) {
		if cfg.Project.Dirs.Web != "" {
			watchDirs = append(watchDirs, cfg.Project.Dirs.Web)
		}
		if cfg.Project.Dirs.Public != "" {
			watchDirs = append(watchDirs, cfg.Project.Dirs.Public)
		}
		if cfg.Project.Dirs.Internal != "" {
			watchDirs = append(watchDirs, cfg.Project.Dirs.Internal)
		}
		if cfg.Project.Dirs.Cmd != "" {
			watchDirs = append(watchDirs, cfg.Project.Dirs.Cmd)
		}
	}
	w := &watcher{
		root:      root,
		every:     d,
		ch:        make(chan struct{}, 1),
		done:      make(chan struct{}),
		watchDirs: watchDirs,
	}
	go w.loop()
	return w
}

// defaultWatchDirs returns the canonical directories to watch (FRK-STR-031).
func defaultWatchDirs() []string {
	return []string{
		"web",
		"public",
		"assets",
		"layouts",
		"components",
		"pages",
		"content",
		"ui",
		"frontend",
		"static",
		"wasm",
	}
}

// Changed returns a channel signaled when a change is detected. Subsequent
// signals are suppressed until Reset is called.
func (w *watcher) Changed() <-chan struct{} { return w.ch }

// Reset re-arms the watcher after a change has been handled.
func (w *watcher) Reset() { w.sig = "" }

func (w *watcher) loop() {
	t := time.NewTicker(w.every)
	defer t.Stop()
	for {
		select {
		case <-w.done:
			return
		case <-t.C:
			now := w.signature()
			if w.sig != "" && now != w.sig {
				select {
				case w.ch <- struct{}{}:
				default:
				}
			}
			w.sig = now
		}
	}
}

func isIgnoredDir(name string) bool {
	if name == "" {
		return false
	}
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch name {
	case "node_modules", "vendor", ".git", ".krewire":
		return true
	}
	return false
}

// signature walks the watched roots and hashes path+modtime pairs, so
// content-safe rebuilds are triggered by edits within the watched set.
func (w *watcher) signature() string {
	var b strings.Builder
	_ = filepath.WalkDir(w.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != w.root && isIgnoredDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !w.watched(path) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		fmt.Fprintf(&b, "%s:%d:%d;", path, info.ModTime().UnixNano(), info.Size())
		return nil
	})
	return b.String()
}

// watched reports whether path belongs to the watched set:
// go sources anywhere plus files under configured asset/markup roots.
func (w *watcher) watched(path string) bool {
	if strings.HasPrefix(filepath.Base(path), ".") {
		return false
	}
	if strings.Contains(path, "/.git/") {
		return false
	}
	rel, err := filepath.Rel(w.root, path)
	if err != nil {
		return false
	}
	if strings.HasSuffix(rel, ".go") {
		return true
	}
	base := filepath.Base(rel)
	if base == "krewire.yaml" ||
		base == "package.json" ||
		strings.HasPrefix(base, "tailwind.config.") ||
		strings.HasPrefix(base, "postcss.config.") {
		return true
	}
	segs := strings.Split(filepath.ToSlash(rel), "/")
	if len(segs) > 1 {
		head := segs[0]
		for _, d := range w.watchDirs {
			if head == d {
				return true
			}
		}
	}
	return false
}
