// Package shape inspects a project directory to discover its capabilities.
package shape

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/krew/internal/gomod"
	"github.com/krewire/krewire/pkg/model"
)

// File and entry point constants for project shape detection (SSOT).
const (
	ConfigFileName = model.ConfigFileName
	GoModFileName  = gomod.GoModFileName
	EntryMainGo    = "main.go"
)

// MarkerPreInit is the marker reported for a project that only has the
// pre-scaffold bootstrap layout.
const MarkerPreInit = "pre-init"

// Capabilities captures every capability present in the project.
type Capabilities struct {
	HasApp     bool
	HasPages   bool
	HasSSG     bool
	HasBook    bool
	HasWorker  bool
	HasService bool
	HasGateway bool
	HasInfra   bool
}

// Result is the outcome of a detection.
type Result struct {
	Marker       string
	Capabilities Capabilities
}

// IsPreInit reports whether the project is a pre-scaffold bootstrap that must
// be equipped with source code before any command can run.
func (r Result) IsPreInit() bool {
	return r.Marker == MarkerPreInit
}

// String implements fmt.Stringer.
func (r Result) String() string {
	if r.Marker == "" {
		return "krewire"
	}
	return r.Marker
}

// IsApp reports whether dir builds an application package main directly from
// main.go or through cmd/ packages containing func main.
func IsApp(dir string) bool {
	return isApp(dir)
}

// Detect inspects the project rooted at dir.
func Detect(dir string, _ ...string) (Result, error) {
	preInit := isPreInit(dir)
	caps := Capabilities{
		HasApp:     !preInit && isApp(dir),
		HasPages:   isDir(dir, "pages"),
		HasSSG:     hasYamlKey(dir, "ssg"),
		HasBook:    isDir(dir, "content"),
		HasWorker:  hasYamlKey(dir, "worker"),
		HasService: hasYamlKey(dir, "service"),
		HasGateway: hasYamlKey(dir, "gateway"),
		HasInfra:   hasYamlKey(dir, "infra"),
	}

	if preInit {
		return Result{Marker: MarkerPreInit, Capabilities: caps}, nil
	}

	marker := "krewire"
	if isApp(dir) {
		marker = appMarker(dir)
	} else if isDir(dir, "pages") {
		marker = "pages/"
	} else if isDir(dir, "content") {
		marker = "content/"
	}

	return Result{Marker: marker, Capabilities: caps}, nil
}

// isPreInit reports whether dir is a fresh, pre-scaffold bootstrap project.
func isPreInit(dir string) bool {
	if !hasFile(dir, EntryMainGo) || !hasFile(dir, ConfigFileName) || !hasFile(dir, GoModFileName) {
		return false
	}
	data, err := os.ReadFile(filepath.Join(dir, EntryMainGo))
	if err != nil {
		return false
	}
	if !strings.Contains(string(data), "hello, krewire") {
		return false
	}
	for _, sub := range []string{"pages", "content", "cmd", "internal", "components"} {
		if isDir(dir, sub) {
			return false
		}
	}
	return true
}

// isApp reports whether dir builds an application package main directly from
// main.go or through cmd/ packages containing func main.
func isApp(dir string) bool {
	for _, p := range append([]string{dir}, cmdDirs(dir)...) {
		if packageMain(p) {
			return true
		}
	}
	return false
}

// cmdDirs returns every immediate subdirectory of dir/cmd.
func cmdDirs(dir string) []string {
	cmd := filepath.Join(dir, "cmd")
	entries, err := os.ReadDir(cmd)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, filepath.Join(cmd, e.Name()))
		}
	}
	return out
}

// appMarker names the strongest marker that flagged an app project.
func appMarker(dir string) string {
	if packageMain(dir) {
		return "main.go"
	}
	return "cmd/*"
}

// packageMain reports whether pkgDir is a package main whose files declare
// func main, excluding test files.
func packageMain(pkgDir string) bool {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return false
	}
	hasMain := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(pkgDir, e.Name()), nil, parser.SkipObjectResolution)
		if err != nil {
			continue
		}
		if f.Name.Name != "main" {
			return false
		}
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				hasMain = true
			}
		}
	}
	return hasMain
}

// hasYamlKey reports whether krewire.yaml declares a top-level key (e.g. `worker:`, `ssg:`).
func hasYamlKey(dir, key string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "krewire.yaml"))
	if err != nil {
		return false
	}
	prefix := key + ":"
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}

// HasFile reports whether file exists as a regular file under dir (SSOT).
func HasFile(dir, name string) bool {
	info, err := os.Stat(filepath.Join(dir, name))
	return err == nil && !info.IsDir()
}

// HasDir reports whether sub is a directory under dir (SSOT).
func HasDir(dir, sub string) bool {
	info, err := os.Stat(filepath.Join(dir, sub))
	return err == nil && info.IsDir()
}

func hasFile(dir, name string) bool { return HasFile(dir, name) }
func isDir(dir, sub string) bool    { return HasDir(dir, sub) }
