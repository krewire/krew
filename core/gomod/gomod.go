// Package gomod reads minimal information out of go.mod files.
package gomod

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GoModFileName is the canonical file name for Go module declarations (SSOT).
const GoModFileName = "go.mod"

// Module describes the module path of a go.mod file, required Go version, and its direct requires.
type Module struct {
	// Path is the module path declared in the go.mod file.
	Path string
	// GoVersion is the minimum Go version declared in the go.mod file (e.g. "1.27.1").
	GoVersion string
	// Requires maps the path of each direct dependency to its version.
	Requires map[string]string
}

// Read parses the go.mod file at path.
func Read(path string) (*Module, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m := &Module{Requires: map[string]string{}}
	inRequireBlock := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "module "):
			m.Path = strings.TrimSpace(strings.TrimPrefix(line, "module "))
		case strings.HasPrefix(line, "go "):
			m.GoVersion = strings.TrimSpace(strings.TrimPrefix(line, "go "))
		case line == "require (":
			inRequireBlock = true
		case line == ")" && inRequireBlock:
			inRequireBlock = false
		case inRequireBlock:
			m.recordRequire(line)
		case strings.HasPrefix(line, "require "):
			m.recordRequire(strings.TrimSpace(strings.TrimPrefix(line, "require ")))
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if m.Path == "" {
		return nil, fmt.Errorf("no module directive found in %s", path)
	}
	return m, nil
}

// Find locates the nearest go.mod walking up from dir.
func Find(dir string) (*Module, error) {
	cur, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	for {
		p := filepath.Join(cur, GoModFileName)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return Read(p)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return nil, fmt.Errorf("%s not found (from %s)", GoModFileName, dir)
		}
		cur = parent
	}
}

// UsesKrewire reports whether the module depends on the Krewire framework or
// libraries.
func (m *Module) UsesKrewire() bool {
	if m == nil {
		return false
	}
	for path := range m.Requires {
		if strings.HasPrefix(path, "github.com/krewire/") {
			return true
		}
	}
	return false
}

func (m *Module) recordRequire(line string) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return
	}
	m.Requires[fields[0]] = fields[1]
}

// CheckGoVersion compares a runtime Go version string (e.g. "go1.27.1")
// against a target minimum version (e.g. "1.27.1" or "1.27").
// It returns an error if the runtime version is strictly older than targetVer.
func CheckGoVersion(runtimeVer, targetVer string) error {
	if targetVer == "" {
		return nil
	}
	cleanRuntime := strings.TrimPrefix(runtimeVer, "go")
	if i := strings.Index(cleanRuntime, " "); i >= 0 {
		cleanRuntime = cleanRuntime[:i]
	}
	if i := strings.Index(cleanRuntime, "-"); i >= 0 {
		cleanRuntime = cleanRuntime[:i]
	}
	if strings.HasPrefix(cleanRuntime, "devel") {
		return nil
	}

	cleanTarget := strings.TrimPrefix(targetVer, "go")
	if i := strings.Index(cleanTarget, " "); i >= 0 {
		cleanTarget = cleanTarget[:i]
	}

	rParts := strings.Split(cleanRuntime, ".")
	tParts := strings.Split(cleanTarget, ".")

	maxLen := len(rParts)
	if len(tParts) > maxLen {
		maxLen = len(tParts)
	}

	for i := 0; i < maxLen; i++ {
		rNum := 0
		if i < len(rParts) {
			rNum, _ = strconv.Atoi(rParts[i])
		}
		tNum := 0
		if i < len(tParts) {
			tNum, _ = strconv.Atoi(tParts[i])
		}
		if rNum < tNum {
			return fmt.Errorf("runtime Go version %s is older than required minimum %s", runtimeVer, targetVer)
		}
		if rNum > tNum {
			return nil
		}
	}
	return nil
}
