package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Search result caps so a broad pattern cannot flood the model context.
const (
	MaxGrepResults = 200
	MaxFindResults = 500
)

// defaultSkipDirs are workspace directories never traversed by search tools.
// They hold vendored, generated, or binary bulk that is never a search target.
var defaultSkipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".krewire":     true,
	"dist":         true,
	"build":        true,
	"bin":          true,
}

// skipDir reports whether a directory entry must not be descended into.
func skipDir(name string) bool {
	return defaultSkipDirs[name]
}

// SearchHit is one matched line reported by grep_search.
type SearchHit struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

// --- GrepSearchTool ---

// GrepSearchTool searches file contents with regular expressions.
type GrepSearchTool struct {
	Workspace string
}

func (t *GrepSearchTool) Name() string { return "grep_search" }

func (t *GrepSearchTool) Description() string {
	return "Search file contents with a regular expression. Supports case-insensitive matching, glob include/exclude filters, and built-in skip lists (.git, node_modules, vendor, build outputs). Returns file path, line number, and matched line."
}

func (t *GrepSearchTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": `Go regular expression to match (e.g. "func .*Handler")`,
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Optional directory to scope the search (defaults to workspace root)",
			},
			"include": map[string]any{
				"type":        "string",
				"description": `Optional glob filter for relative file paths (e.g. "*.go")`,
			},
			"exclude": map[string]any{
				"type":        "string",
				"description": `Optional glob of relative file paths to skip (e.g. "*_test.go")`,
			},
			"ignore_case": map[string]any{
				"type":        "boolean",
				"description": "Match case-insensitively (default false)",
			},
			"max_results": map[string]any{
				"type":        "integer",
				"description": fmt.Sprintf("Maximum matches returned (default and cap %d)", MaxGrepResults),
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GrepSearchTool) Execute(ctx context.Context, args json.RawMessage) (any, error) {
	var input struct {
		Pattern    string `json:"pattern"`
		Path       string `json:"path"`
		Include    string `json:"include"`
		Exclude    string `json:"exclude"`
		IgnoreCase bool   `json:"ignore_case"`
		MaxResults int    `json:"max_results"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}
	if strings.TrimSpace(input.Pattern) == "" {
		return nil, fmt.Errorf("pattern is required")
	}

	expr := input.Pattern
	if input.IgnoreCase {
		expr = "(?i)" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}
	if input.MaxResults <= 0 || input.MaxResults > MaxGrepResults {
		input.MaxResults = MaxGrepResults
	}
	include, err := compileOptionalGlob(input.Include)
	if err != nil {
		return nil, err
	}
	exclude, err := compileOptionalGlob(input.Exclude)
	if err != nil {
		return nil, err
	}

	root, err := SafePath(t.Workspace, orDot(input.Path))
	if err != nil {
		return nil, err
	}

	hits := []SearchHit{}
	err = walkWorkspace(root, func(rel string, info os.FileInfo) error {
		if !info.Mode().IsRegular() {
			return nil
		}
		if include != nil && !include.MatchString(rel) {
			return nil
		}
		if exclude != nil && exclude.MatchString(rel) {
			return nil
		}
		matchesInFile(filepath.Join(root, rel), rel, re, input.MaxResults, &hits)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return hits, nil
}

// matchesInFile greps a single file, accumulating hits until maxResults.
// Unreadable and binary files are skipped silently. NUL bytes, like git, mark
// a file binary.
func matchesInFile(path, display string, re *regexp.Regexp, maxResults int, hits *[]SearchHit) {
	data, err := os.ReadFile(path)
	if err != nil || isBinary(data) {
		return
	}
	const readLimit = 1 << 20 // avoid flooding context with enormous files
	if len(data) > readLimit {
		data = data[:readLimit]
	}
	for i, line := range strings.Split(string(data), "\n") {
		if re.MatchString(line) {
			*hits = append(*hits, SearchHit{Path: display, Line: i + 1, Text: line})
			if len(*hits) >= maxResults {
				return
			}
		}
	}
}

// isBinary reports whether data should be treated as binary: NUL bytes are
// invalid in every textual encoding tooling cares about.
func isBinary(data []byte) bool {
	const probe = 8_000
	if len(data) > probe {
		data = data[:probe]
	}
	return bytes.IndexByte(data, 0) >= 0 || !utf8.Valid(data)
}

// --- FindFilesTool ---

// FindFilesTool locates files and directories by glob pattern.
type FindFilesTool struct {
	Workspace string
}

func (t *FindFilesTool) Name() string { return "find_files" }

func (t *FindFilesTool) Description() string {
	return `Find files and directories by glob pattern relative to the workspace (e.g. "pkg/**/*.go", "README*", "Makefile"). "**" matches any number of directories.`
}

func (t *FindFilesTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "Glob pattern; without a slash it matches file names at any depth",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Optional directory to scope the search (defaults to workspace root)",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *FindFilesTool) Execute(ctx context.Context, args json.RawMessage) (any, error) {
	var input struct {
		Pattern string `json:"pattern"`
		Path    string `json:"path"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}
	pattern := strings.TrimSpace(input.Pattern)
	if pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}
	depthGlob := !strings.Contains(pattern, "/")
	matcher, err := compileOptionalGlob(pattern)
	if err != nil {
		return nil, err
	}

	root, err := SafePath(t.Workspace, orDot(input.Path))
	if err != nil {
		return nil, err
	}

	found := []string{}
	err = walkWorkspace(root, func(rel string, info os.FileInfo) error {
		candidate := rel
		if depthGlob {
			candidate = filepath.Base(rel)
		}
		if matcher != nil && matcher.MatchString(candidate) && len(found) < MaxFindResults {
			found = append(found, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}

// walkWorkspace descends root, invoking fn for every entry with a
// slash-separated path relative to root. Files make no distinction between
// root itself and "." and skip-listed directories are never descended.
func walkWorkspace(root string, fn func(rel string, info os.FileInfo) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}
		if d.IsDir() && skipDir(d.Name()) {
			return fs.SkipDir
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		return fn(filepath.ToSlash(rel), info)
	})
}

// compileOptionalGlob compiles a shell-like glob into an anchored regexp;
// an empty glob yields nil (matches everything).
func compileOptionalGlob(glob string) (*regexp.Regexp, error) {
	if strings.TrimSpace(glob) == "" {
		return nil, nil
	}
	expr, err := globToRegexp(glob)
	if err != nil {
		return nil, fmt.Errorf("invalid glob %q: %w", glob, err)
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid glob %q: %w", glob, err)
	}
	return re, nil
}

// globToRegexp converts a glob to an anchored regexp. "/" separates segments;
// "**" spans any number of segments; "*" matches within one segment.
func globToRegexp(glob string) (string, error) {
	if strings.Contains(glob, "**/**") {
		return "", fmt.Errorf("redundant **/** segment")
	}
	var b strings.Builder
	b.WriteString("^")
	for i, seg := range strings.Split(glob, "/") {
		if i > 0 {
			b.WriteString("/")
		}
		if seg == "**" {
			b.WriteString(".*")
			continue
		}
		writeGlobSegment(&b, seg)
	}
	b.WriteString("$")
	return b.String(), nil
}

// writeGlobSegment emits one glob segment: "*" becomes [^/]*, "?" becomes
// [^/], bracket expressions pass through verbatim, everything else is quoted.
func writeGlobSegment(b *strings.Builder, seg string) {
	for i := 0; i < len(seg); i++ {
		c := seg[i]
		switch {
		case c == '*':
			b.WriteString("[^/]*")
		case c == '?':
			b.WriteString("[^/]")
		case c == '[':
			if bracketEnd := strings.IndexByte(seg[i:], ']'); bracketEnd >= 0 {
				b.WriteString(seg[i : i+bracketEnd+1])
				i += bracketEnd
			} else {
				b.WriteString(regexp.QuoteMeta("["))
			}
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
}

// orDot maps an empty scope to the workspace root.
func orDot(path string) string {
	if strings.TrimSpace(path) == "" {
		return "."
	}
	return path
}
