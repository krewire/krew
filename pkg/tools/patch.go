package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// hunkFuzzyRadius bounds the line-offset search when a hunk header's expected
// start position has drifted from the actual file content.
const hunkFuzzyRadius = 32

// patchFile is one patched file inside a unified diff; multi-file patches are
// supported.
type patchFile struct {
	relPath  string
	hunks    []patchHunk
	isDelete bool // +++ /dev/null: full-file deletion
	creation bool // --- /dev/null: new file
}

// patchHunk is one @@ block. lines carry raw prefixes:
// ' ' context, '-' removal, '+' addition.
type patchHunk struct {
	oldStart int // 1-based position in the original file
	lines    []patchLine
}

type patchLine struct {
	kind byte
	text string
}

// ApplyPatchTool applies standard unified diffs to workspace files.
type ApplyPatchTool struct {
	Workspace string
}

func (t *ApplyPatchTool) Name() string { return "apply_patch" }

func (t *ApplyPatchTool) Description() string {
	return `Apply a standard unified diff ("diff -u" format) to one or more workspace files. Hunks are located with tolerant context matching; a file is only rewritten when every hunk applies cleanly, so failures never damage existing content.`
}

func (t *ApplyPatchTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"patch": map[string]any{
				"type":        "string",
				"description": "The full unified diff text (--- a/..., +++ b/..., @@ hunks)",
			},
		},
		"required": []string{"patch"},
	}
}

func (t *ApplyPatchTool) Execute(ctx context.Context, args json.RawMessage) (any, error) {
	var input struct {
		Patch string `json:"patch"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}
	files, err := parseUnifiedDiff(input.Patch)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no file hunks found in patch")
	}

	// Rollback is structural: every file is rebuilt in memory and only written
	// once all of its hunks have matched, so a rejected hunk leaves the file
	// byte-for-byte untouched.
	var summary []string
	for _, pf := range files {
		if err := t.applyFile(pf); err != nil {
			return nil, fmt.Errorf("patch aborted before %q (files left untouched): %w", pf.relPath, err)
		}
		action := "patched"
		switch {
		case pf.isDelete:
			action = "deleted"
		case pf.creation:
			action = "created"
		}
		summary = append(summary, fmt.Sprintf("%s %s (%d hunks)", action, pf.relPath, len(pf.hunks)))
	}
	return strings.Join(summary, "; "), nil
}

// applyFile materialises one file's post-patch content in memory and writes
// it only when all hunks have applied.
func (t *ApplyPatchTool) applyFile(pf patchFile) error {
	target, err := SafePath(t.Workspace, pf.relPath)
	if err != nil {
		return err
	}

	if pf.isDelete {
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("delete file: %w", err)
		}
		return nil
	}

	var lines []string
	if !pf.creation {
		raw, err := os.ReadFile(target)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file not found: %v", err)
			}
			return fmt.Errorf("read file: %w", err)
		}
		lines = splitKeepNewlines(string(raw))
	}
	for _, h := range pf.hunks {
		if lines, err = applyHunk(h, lines); err != nil {
			return err
		}
	}
	return writeTarget(target, strings.Join(lines, ""))
}

func writeTarget(target, content string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write patched file: %w", err)
	}
	return nil
}

// splitKeepNewlines splits content into per-line segments, each carrying its
// trailing newline. The trailing empty element from SplitAfter is dropped.
func splitKeepNewlines(content string) []string {
	if content == "" {
		return nil
	}
	segments := strings.SplitAfter(content, "\n")
	if segments[len(segments)-1] == "" {
		segments = segments[:len(segments)-1]
	}
	return segments
}

// lineText extracts the bare text (no newline, no CR) of a file segment.
func lineText(segment string) string {
	return strings.TrimSuffix(strings.TrimSuffix(segment, "\n"), "\r")
}

// fuzzyEqual compares expected and actual text tolerating trailing whitespace
// and line-ending drift.
func fuzzyEqual(expected, actual string) bool {
	expected = strings.TrimRight(expected, " \t\r\n")
	actual = strings.TrimRight(actual, " \t\r\n")
	return expected == actual || strings.TrimSpace(expected) == strings.TrimSpace(actual)
}

// applyHunk locates hunk's old-side block in lines and returns the spliced
// result. Context/removal lines reduce to fuzzy equality; the search scans up
// to hunkFuzzyRadius lines around the header's stated position.
func applyHunk(h patchHunk, lines []string) ([]string, error) {
	oldCount := 0
	for _, l := range h.lines {
		if l.kind == ' ' || l.kind == '-' {
			oldCount++
		}
	}

	matchAt := func(cand int) bool {
		if cand < 0 || cand+oldCount > len(lines) {
			return false
		}
		j := 0
		for _, l := range h.lines {
			if l.kind == '+' {
				continue
			}
			if !fuzzyEqual(l.text, lineText(lines[cand+j])) {
				return false
			}
			j++
		}
		return true
	}

	expected := h.oldStart - 1
	if expected > len(lines)-oldCount {
		expected = len(lines) - oldCount
	}
	if expected < 0 {
		expected = 0
	}

	cand := -1
	for off := 0; off <= hunkFuzzyRadius && cand == -1; off++ {
		for _, want := range []int{expected + off, expected - off} {
			if matchAt(want) {
				cand = want
				break
			}
		}
	}
	if cand == -1 {
		return nil, fmt.Errorf("hunk does not apply cleanly: no context match within %d lines of position %d", hunkFuzzyRadius, h.oldStart)
	}

	res := make([]string, 0, len(lines)-oldCount+len(h.lines))
	res = append(res, lines[:cand]...)
	for _, l := range h.lines {
		switch l.kind {
		case '+':
			res = append(res, l.text+"\n")
		case ' ':
			res = append(res, lines[cand])
			cand++
		case '-':
			cand++ // removed; skipped in output
		}
	}
	return append(res, lines[cand:]...), nil
}

// parseHunkHeader parses a "@@ -a,b +c,d @@ [section]" line.
func parseHunkHeader(header string) (patchHunk, int, int, error) {
	var h patchHunk
	body := strings.TrimPrefix(header, "@@")
	if idx := strings.Index(body, "@@"); idx >= 0 {
		body = body[:idx]
	}
	oldCount, newCount := 0, 0
	for _, field := range strings.Fields(body) {
		old := strings.HasPrefix(field, "-")
		num := strings.TrimLeft(field, "-+")
		if !strings.HasPrefix(field, "-") && !strings.HasPrefix(field, "+") {
			continue
		}

		startStr, countStr := num, "1"
		if idx := strings.IndexByte(num, ','); idx >= 0 {
			startStr, countStr = num[:idx], num[idx+1:]
		}
		start, err := strconv.Atoi(startStr)
		if err != nil {
			return h, 0, 0, fmt.Errorf("hunk header %q: %w", header, err)
		}
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return h, 0, 0, fmt.Errorf("hunk header %q: %w", header, err)
		}
		if old {
			h.oldStart = start
			oldCount = count
		} else {
			newCount = count
		}
	}
	return h, oldCount, newCount, nil
}

// parseUnifiedDiff parses a full unified diff into per-file records. Git-style
// path prefixes (a/, b/) and /dev/null creations/deletions are handled.
func parseUnifiedDiff(patch string) ([]patchFile, error) {
	var files []patchFile
	var oldPath string

	rawLines := strings.SplitAfter(patch, "\n")
	trim := func(segment string) string { return strings.TrimSuffix(strings.TrimSuffix(segment, "\n"), "\r") }

	i := 0
	for i < len(rawLines) {
		line := trim(rawLines[i])

		switch {
		case strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "Index: ") ||
			strings.HasPrefix(line, "new file mode") || strings.HasPrefix(line, "deleted file mode") ||
			strings.HasPrefix(line, "similarity index") || strings.HasPrefix(line, "rename "):
			i++ // git diff boilerplate
			continue

		case strings.HasPrefix(line, "--- "):
			oldPath = stripAB(strings.TrimSpace(strings.TrimPrefix(line, "--- ")))
			i++
			continue

		case !strings.HasPrefix(line, "+++ "):
			i++ // inter-file whitespace, comments, anything else
			continue
		}

		target := strings.TrimSpace(strings.TrimPrefix(line, "+++ "))
		if idx := strings.IndexAny(target, " \t"); idx >= 0 {
			target = target[:idx]
		}
		target = stripAB(target)

		pf := patchFile{}
		if target == "/dev/null" {
			if oldPath == "" || oldPath == "/dev/null" {
				return nil, fmt.Errorf("+++ /dev/null without a --- source path")
			}
			pf.isDelete = true
			pf.relPath = oldPath
		} else {
			pf.relPath = target
			pf.creation = oldPath == "" || oldPath == "/dev/null"
		}
		i++

		// Consume this file's hunks until the next file header or EOF.
		for i < len(rawLines) && strings.HasPrefix(trim(rawLines[i]), "@@") {
			h, oldCount, newCount, err := parseHunkHeader(trim(rawLines[i]))
			if err != nil {
				return nil, err
			}
			i++

			// Body lines are consumed until the header's promised old/new line
			// counts are satisfied; this also keeps a following "--- " file
			// header from being mistaken for removal text.
			for i < len(rawLines) && (oldCount > 0 || newCount > 0) {
				body := trim(rawLines[i])

				if strings.HasPrefix(body, "\\") { // "\ No newline at end of file"
					i++
					continue
				}
				if body == "" {
					if oldCount > 0 && newCount > 0 { // blank context line
						h.lines = append(h.lines, patchLine{kind: ' ', text: ""})
						oldCount--
						newCount--
						i++
						continue
					}
					break
				}

				kind := body[0]
				text := body[1:]
				if kind != ' ' && kind != '-' && kind != '+' {
					break // not body content; let the outer loop reprocess
				}
				h.lines = append(h.lines, patchLine{kind: kind, text: text})
				if kind == ' ' || kind == '-' {
					oldCount--
				}
				if kind == ' ' || kind == '+' {
					newCount--
				}
				i++
			}
			pf.hunks = append(pf.hunks, h)
		}

		if len(pf.hunks) == 0 {
			return nil, fmt.Errorf("file %q has no @@ hunks", pf.relPath)
		}
		files = append(files, pf)
		oldPath = ""
	}
	return files, nil
}

// stripAB removes the conventional "a/" / "b/" diff path prefixes.
func stripAB(path string) string {
	if strings.HasPrefix(path, "a/") || strings.HasPrefix(path, "b/") {
		return path[2:]
	}
	return path
}
