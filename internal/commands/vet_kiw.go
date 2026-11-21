package commands

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/krewire/pkg/kiw"
)

// lintKiwFiles runs kiw.Lint on every .kiw file under root, skipping hidden,
// vendor, node_modules and build directories, and returns the diagnostics
// with paths relative to root.
func lintKiwFiles(root string) []string {
	var out []string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if path != root && (strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" || name == "testdata") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".kiw") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		for _, d := range kiw.Lint(filepath.ToSlash(rel), string(src)) {
			out = append(out, d.Error())
		}
		return nil
	})
	return out
}
