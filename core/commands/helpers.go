package commands

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/krewire/krew/core/config"
	"github.com/krewire/krew/core/gomod"
	"github.com/krewire/krew/core/scaffold"
	"github.com/krewire/krew/core/shape"
	"github.com/krewire/krewire/env"
	"github.com/krewire/krewire/errs"
)

func flagValue(fs *flag.FlagSet, name string) string {
	if fs == nil {
		return ""
	}
	if f := fs.Lookup(name); f != nil {
		return f.Value.String()
	}
	return ""
}

func flagInt(fs *flag.FlagSet, name string) int {
	v := flagValue(fs, name)
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}

func flagFloat(fs *flag.FlagSet, name string) float64 {
	v := flagValue(fs, name)
	if v == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func commandError(err error) errs.ExitCode {
	fmt.Fprint(os.Stderr, errs.FormatTree(err))
	switch {
	case errors.Is(err, scaffold.ErrInvalidName),
		errors.Is(err, scaffold.ErrProjectExists),
		errors.Is(err, scaffold.ErrNotEmpty),
		errors.Is(err, scaffold.ErrConflict):
		return errs.ExitCodeUsage
	default:
		return errs.ExitCodeFailure
	}
}

func fail(err error) errs.ExitCode {
	fmt.Fprint(os.Stderr, errs.FormatTree(err))
	printStackIfDebug(err)
	return errs.ExitCodeFailure
}

// printStackIfDebug renders an attached stack after the error line when the
// command resolved debug mode (KWL-P8W2N KWL-DIAGV-007).
func printStackIfDebug(err error) {
	if !debugEnabled {
		return
	}
	frames := errs.StackOf(err)
	if len(frames) == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, "stack trace (newest first):")
	fmt.Fprint(os.Stderr, errs.FormatStack(frames))
}

// usageOrFail prints err and maps core usage errors to exit code 2
// (KWL-K4T7W KWL-ENVV-007).
func usageOrFail(err error) errs.ExitCode {
	var ce interface{ ExitCode() errs.ExitCode }
	if errors.As(err, &ce) && ce.ExitCode() == errs.ExitCodeUsage {
		fmt.Fprint(os.Stderr, errs.FormatTree(err))
		printStackIfDebug(err)
		return errs.ExitCodeUsage
	}
	return fail(err)
}

// flagProvided reports whether the named flag was explicitly set on the
// command line, as opposed to holding its zero value.
func flagProvided(fs *flag.FlagSet, name string) bool {
	if fs == nil {
		return false
	}
	set := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			set = true
		}
	})
	return set
}

// truthStr renders a boolean as the KIW_DEBUG wire value.
func truthStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// defaultAddr is the fallback listen address when no address is specified.
const defaultAddr = ":8080"

// resolveAddrWithPort resolves the listening address, preferring an explicit command-line
// flag (addr or port) over a declared port in krewire.yaml, falling back to defaultAddr (:8080)
// (FRK-CONF-031).
func resolveAddrWithPort(flagAddr, flagPort string, cfg *config.Config) string {
	if flagAddr != "" {
		if !strings.Contains(flagAddr, ":") {
			return ":" + flagAddr
		}
		return flagAddr
	}
	if flagPort != "" {
		if strings.HasPrefix(flagPort, ":") {
			return flagPort
		}
		return ":" + flagPort
	}
	if cfg != nil && cfg.DeclaredPort() > 0 {
		return fmt.Sprintf(":%d", cfg.DeclaredPort())
	}
	return defaultAddr
}

// resolvedAddr resolves the listening address, preferring an explicit command-line
// flag over a declared port in krewire.yaml, falling back to defaultAddr (:8080)
// (FRK-CONF-031).
func resolvedAddr(flagAddr string, cfg *config.Config) string {
	return resolveAddrWithPort(flagAddr, "", cfg)
}

// childEnviron composes the child-process environment for run and dev,
// exporting APP_ADDR plus KIW_ENV/KIW_DEBUG (KWL-K4T7W KWL-ENVV-006).
// Vein is applied: env via env.Env.
func childEnviron(base []string, addr string, env env.Env, debug bool) []string {
	e := append([]string(nil), base...)
	if addr != "" {
		e = append(e, "APP_ADDR="+addr)
		parts := strings.Split(addr, ":")
		if len(parts) > 1 && parts[len(parts)-1] != "" {
			e = append(e, "PORT="+parts[len(parts)-1])
		}
	}
	return append(e, "KIW_ENV="+env.String(), "KIW_DEBUG="+truthStr(debug))
}

// firstNonEmpty returns the first non-empty string among the arguments (SSOT delegation).
func firstNonEmpty(vals ...string) string {
	return config.FirstNonEmpty(vals...)
}

// findRoot locates the project root walking up from the working directory.
// For app/cli/worker/service it is the Go module root (go.mod). For site/book
// which have no entry point, krewire.yaml or a declarative layout (pages/,
// content/) is sufficient (KWF-DF3PL FRK-FLS-001/002, KWL-K1N2Q).
func findRoot() (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cur, err = filepath.Abs(cur)
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(cur, gomod.GoModFileName)); err == nil && !info.IsDir() {
			return cur, nil
		}
		if info, err := os.Stat(filepath.Join(cur, config.ConfigFileName)); err == nil && !info.IsDir() {
			return cur, nil
		}
		if info, err := os.Stat(filepath.Join(cur, "pages")); err == nil && info.IsDir() {
			return cur, nil
		}
		if info, err := os.Stat(filepath.Join(cur, "content")); err == nil && info.IsDir() {
			return cur, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", fmt.Errorf("not inside a Go module — run 'krew new <project>' first")
		}
		cur = parent
	}
}

// joinRoot resolves a flag value relative to the project root, falling back to
// def when the flag is unset.
func joinRoot(root, value, def string) string {
	if value == "" {
		value = def
	}
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(root, value)
}

// hasDir reports whether path exists inside root as a directory (delegates to shape.HasDir).
func hasDir(root, path string) bool {
	return shape.HasDir(root, path)
}

// hasFile reports whether path exists inside root as a regular file (delegates to shape.HasFile).
func hasFile(root, path string) bool {
	return shape.HasFile(root, path)
}

// resolveEntry finds the main entry point package for go build / go run.
func resolveEntry(root string) string {
	if hasFile(root, "main.go") {
		return "."
	}
	if entries, err := os.ReadDir(filepath.Join(root, "cmd")); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				return "./cmd/" + e.Name()
			} else if e.Name() == "main.go" {
				return "./cmd"
			}
		}
	}
	return "."
}

// moduleName returns the last path element of the module declared in root's
// go.mod.
func moduleName(root string) string {
	mod, err := gomod.Read(filepath.Join(root, "go.mod"))
	if err != nil || mod.Path == "" {
		return filepath.Base(root)
	}
	parts := strings.Split(strings.TrimSuffix(mod.Path, "/"), "/")
	return parts[len(parts)-1]
}

// moduleDir walks upward from dir to find the directory containing go.mod.
func moduleDir(dir string) string {
	cur := dir
	for {
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			return cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return ""
		}
		cur = parent
	}
}

// containsGoPackageArg reports whether args contains a package-like path argument.
func containsGoPackageArg(args []string) bool {
	for _, a := range args {
		if strings.Contains(a, "/") || strings.HasPrefix(a, "./") || a == "." || a == "..." {
			return true
		}
	}
	return false
}
