// Package shape classifies a project directory into one of the Krewire
// project kinds so commands can dispatch to the right pipeline.
package shape

import (
	"fmt"
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

// Kind identifies a Krewire project kind. Re-exported from model for SSOT.
type Kind = model.Kind

const (
	KindApp          = model.KindApp
	KindCLI          = model.KindCLI
	KindSite         = model.KindSite
	KindBook         = model.KindBook
	KindWorker       = model.KindWorker
	KindService      = model.KindService
	KindGateway      = model.KindGateway
	KindInfra        = model.KindInfra
	KindNone    Kind = ""
)

// MarkerPreInit is the marker reported for a project that only has the
// pre-scaffold bootstrap layout. It is not a kind: the project must be equipped
// with a variant before any workload command can run.
const MarkerPreInit = "pre-init"

// Capabilities captures every workload capability present in the project.
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
	// Kind is the primary detected project kind.
	Kind Kind
	// Marker names the file or config key that determined the kind, e.g.
	// "main.go" or "krewire.yaml#ssg".
	Marker string
	// Capabilities details the composable features discovered in the project.
	Capabilities Capabilities
}

// IsPreInit reports whether the project is a pre-scaffold bootstrap that must
// be equipped with a variant before any workload command can run.
func (r Result) IsPreInit() bool {
	return r.Kind == KindNone && r.Marker == MarkerPreInit
}

// String implements fmt.Stringer.
func (r Result) String() string {
	if r.Kind == KindNone {
		return "none"
	}
	return string(r.Kind)
}

// IsApp reports whether dir builds an application package main directly from
// main.go or through cmd/ packages containing func main.
func IsApp(dir string) bool {
	return isApp(dir)
}

// Detect classifies the project rooted at dir. explicitKind, when non-empty,
// comes from config.project.kind and wins over any marker detection.
func Detect(dir, explicitKind string) (Result, error) {
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

	if explicitKind != "" {
		parsed, err := model.ParseKind(explicitKind)
		if err != nil {
			return Result{Capabilities: caps}, fmt.Errorf("shape: unknown project kind %q", explicitKind)
		}
		return Result{Kind: parsed, Marker: "project.kind", Capabilities: caps}, nil
	}

	// 1. Explicit marker keys in krewire.yaml
	if hasYamlKey(dir, "worker") {
		return Result{Kind: KindWorker, Marker: "krewire.yaml#worker", Capabilities: caps}, nil
	}
	if hasYamlKey(dir, "service") {
		return Result{Kind: KindService, Marker: "krewire.yaml#service", Capabilities: caps}, nil
	}
	if hasYamlKey(dir, "gateway") {
		return Result{Kind: KindGateway, Marker: "krewire.yaml#gateway", Capabilities: caps}, nil
	}
	if hasYamlKey(dir, "infra") {
		return Result{Kind: KindInfra, Marker: "krewire.yaml#infra", Capabilities: caps}, nil
	}

	// 2. Pre-scaffold bootstrap (go.mod + krewire.yaml + starter main.go).
	// This is not a workload kind: the project must be equipped first.
	if preInit {
		return Result{Kind: KindNone, Marker: MarkerPreInit, Capabilities: caps}, nil
	}

	// 3. Application entry point
	if isApp(dir) {
		return Result{Kind: KindApp, Marker: appMarker(dir), Capabilities: caps}, nil
	}

	// 4. Declarative SSG site (ssg: key in krewire.yaml, or pages/ directory)
	if hasYamlKey(dir, "ssg") || isDir(dir, "pages") {
		marker := "pages/"
		if hasYamlKey(dir, "ssg") {
			marker = ConfigFileName + "#ssg"
		}
		return Result{Kind: KindSite, Marker: marker, Capabilities: caps}, nil
	}

	// 5. Content documentation (book)
	if isDir(dir, "content") {
		return Result{Kind: KindBook, Marker: "content/", Capabilities: caps}, nil
	}

	return Result{Kind: KindNone, Capabilities: caps}, nil
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
