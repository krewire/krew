package sdlc

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorkspaceSnapshot holds the contents of tracked files in the workspace.
type WorkspaceSnapshot struct {
	BaseDir string
	Files   map[string][]byte
}

// TakeSnapshot records the current file contents of the workspace matching scope.
func TakeSnapshot(baseDir string, scopes []string) (*WorkspaceSnapshot, error) {
	snap := &WorkspaceSnapshot{
		BaseDir: baseDir,
		Files:   make(map[string][]byte),
	}

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == ".krewire" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(baseDir, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		if !CheckScope(rel, scopes) {
			return nil
		}

		// Avoid large binaries (limit to 1MB per file for diff tracking)
		if info.Size() > 1024*1024 {
			return nil
		}

		data, err := os.ReadFile(path)
		if err == nil {
			snap.Files[rel] = data
		}
		return nil
	})

	return snap, err
}

// ComputeDiff compares the initial snapshot against the current workspace state.
func (snap *WorkspaceSnapshot) ComputeDiff() ([]FileDiff, error) {
	currentSnap, err := TakeSnapshot(snap.BaseDir, []string{"."})
	if err != nil {
		return nil, err
	}

	var diffs []FileDiff

	// 1. Check for modified and deleted files
	for path, oldData := range snap.Files {
		newData, exists := currentSnap.Files[path]
		if !exists {
			diffs = append(diffs, FileDiff{
				Path:         path,
				Action:       "deleted",
				LinesRemoved: countLines(oldData),
				DiffText:     fmt.Sprintf("--- a/%s\n+++ /dev/null\n", path),
				Status:       "pending",
			})
			continue
		}

		if !bytes.Equal(oldData, newData) {
			added, removed, diffText := simpleDiff(path, oldData, newData)
			diffs = append(diffs, FileDiff{
				Path:         path,
				Action:       "modified",
				LinesAdded:   added,
				LinesRemoved: removed,
				DiffText:     diffText,
				Status:       "pending",
			})
		}
	}

	// 2. Check for created files
	for path, newData := range currentSnap.Files {
		if _, exists := snap.Files[path]; !exists {
			diffs = append(diffs, FileDiff{
				Path:       path,
				Action:     "created",
				LinesAdded: countLines(newData),
				DiffText:   fmt.Sprintf("--- /dev/null\n+++ b/%s\n", path),
				Status:     "pending",
			})
		}
	}

	return diffs, nil
}

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	n := bytes.Count(data, []byte("\n"))
	if !bytes.HasSuffix(data, []byte("\n")) {
		n++
	}
	return n
}

func simpleDiff(path string, oldData, newData []byte) (int, int, string) {
	oldLines := strings.Split(string(oldData), "\n")
	newLines := strings.Split(string(newData), "\n")

	var diffBuilder strings.Builder
	diffBuilder.WriteString(fmt.Sprintf("--- a/%s\n+++ b/%s\n", path, path))

	added := 0
	removed := 0

	// Simple line-by-line diff display
	max := len(oldLines)
	if len(newLines) > max {
		max = len(newLines)
	}

	for i := 0; i < max; i++ {
		var o, n string
		hasOld := i < len(oldLines)
		hasNew := i < len(newLines)

		if hasOld {
			o = oldLines[i]
		}
		if hasNew {
			n = newLines[i]
		}

		if hasOld && hasNew && o == n {
			continue
		}

		if hasOld {
			diffBuilder.WriteString("-" + o + "\n")
			removed++
		}
		if hasNew {
			diffBuilder.WriteString("+" + n + "\n")
			added++
		}
	}

	return added, removed, diffBuilder.String()
}
