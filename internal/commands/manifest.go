package commands

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// manifestName is the build manifest written into the output directory; it
// lists every file the last successful build produced so the next run can
// prune stale artifacts (e.g. pages dropped by new exclude rules).
const manifestName = ".kiw-build-manifest"

// readManifest returns the relative paths recorded in outDir's manifest.
func readManifest(outDir string) []string {
	data, err := os.ReadFile(filepath.Join(outDir, manifestName))
	if err != nil {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			out = append(out, filepath.Clean(line))
		}
	}
	return out
}

// readManifestOutputs returns the runtime output directories recorded in the manifest.
func readManifestOutputs(outDir string) map[string]string {
	data, err := os.ReadFile(filepath.Join(outDir, manifestName))
	if err != nil {
		return nil
	}
	outputs := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# output:") {
			entry := strings.TrimPrefix(line, "# output:")
			parts := strings.SplitN(entry, "=", 2)
			if len(parts) == 2 {
				outputs[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}
	return outputs
}

// writeManifest records rel (slash-separated) paths and optional declared outputs
// as the current build output. Missing entries from a previous manifest were already pruned.
func writeManifest(outDir string, rel []string, outputs ...map[string]string) {
	sort.Strings(rel)
	var header []string
	if len(outputs) > 0 && len(outputs[0]) > 0 {
		header = append(header, "# multi-runtime outputs")
		var keys []string
		for k := range outputs[0] {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			header = append(header, fmt.Sprintf("# output:%s=%s", k, outputs[0][k]))
		}
	}

	body := ""
	if len(header) > 0 {
		body += strings.Join(header, "\n") + "\n"
	}
	body += strings.Join(rel, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(outDir, manifestName), []byte(body), 0o644); err != nil {
		slog.Warn("krew build: could not write build manifest", "err", err)
	}
}

// pruneStale deletes files recorded by the previous manifest that still
// exist, plus empty parent directories inside outDir. Only manifest-listed
// files are touched — anything else in the output directory is left alone.
func pruneStale(outDir string) {
	old := readManifest(outDir)
	if len(old) == 0 {
		return
	}
	for _, rel := range old {
		p := filepath.Join(outDir, filepath.FromSlash(rel))
		if err := os.Remove(p); err == nil {
			slog.Debug("pruned stale output", "file", rel)
		}
	}
	removeEmptyDirs(outDir)
	os.Remove(filepath.Join(outDir, manifestName))
}

// removeEmptyDirs prunes now-empty directories under root (deepest first).
func removeEmptyDirs(root string) {
	var dirs []string
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err == nil && d.IsDir() && p != root {
			dirs = append(dirs, p)
		}
		return nil
	})
	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))
	for _, d := range dirs {
		os.Remove(d)
	}
}

// collectCreated walks outDir and lists every regular file relative to it,
// excluding the manifest itself.
func collectCreated(outDir string) []string {
	var out []string
	filepath.WalkDir(outDir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(outDir, p)
		if rerr != nil || rel == manifestName {
			return nil
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	return out
}
