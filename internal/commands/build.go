package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krew/internal/shape"
	"github.com/krewire/krewire/pkg/errs"
)

const (
	// DefaultTarget is the default build target workload.
	DefaultTarget = "ssg"

	// DefaultThemeMode is the default theme selection mode.
	DefaultThemeMode = "auto"

	// DefaultWASMEntry is the default WASM source directory.
	DefaultWASMEntry = "./wasm"

	// DefaultWASMName is the default WASM module output identifier.
	DefaultWASMName = "runtime"

	// DefaultBinaryName is the default fallback binary name.
	DefaultBinaryName = "app"
)

// RegisterBuild registers flags for the build command.
func RegisterBuild(fs *flag.FlagSet) {
	fs.String("input", "", "content directory (default content)")
	fs.String("output", "", "output directory (default .krewire/build)")
	fs.String("o", "", "output directory (shorthand for --output)")
	fs.String("include", "", "comma-separated content glob patterns to build (default **/*.md)")
	fs.String("exclude", "", "comma-separated content globs to skip (default **/README.md,**/readme.md)")
	fs.String("base", "", "URL base the site will be served under (default /)")
	fs.String("title", "", "site title (defaults to the project name)")
	fs.String("author", "", "site author")
	fs.String("theme", "", "theme mode: auto, light, dark, or off (default auto)")
	fs.String("target", DefaultTarget, "build target: site, book, binary, wasm, or infra (default ssg)")
	fs.Bool("skip-deps", false, "skip executing build.depends_on commands")
	fs.Bool("deps-only", false, "run only build.depends_on commands and skip Krewire build")
	fs.Bool("bundle", true, "bundle all CSS and JS into assets/krewire.css and assets/krewire.js (default true)")
	fs.Bool("minify", false, "minify CSS and JS single-bundle assets")
	fs.Bool("plan", false, "infra: compile and show infrastructure plan without applying")
	fs.String("env", "", "infra: target environment (local, production, testing)")
}

// RunBuild builds the project for production. It dispatches according to the
// detected project capabilities (pages, book, app binary, worker, service, infra)
// or explicit --target flag (KWN-BUILD-001).
//
// Progressive enhancement is preserved: when web pages or content docs coexist
// in an app, worker, or service project, both the Go binary and web assets are built
// into the shared output (.krewire/build).
func RunBuild(fs *flag.FlagSet) errs.ExitCode {
	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew: "+err.Error())
		return errs.ExitCodeUsage
	}
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}

	skipDeps := flagValue(fs, "skip-deps") == "true"
	depsOnly := flagValue(fs, "deps-only") == "true"

	if depsOnly {
		if len(cfg.Build.DependsOn) == 0 {
			return errs.ExitCodeSuccess
		}
		_ = cfg.LoadDotEnv(root)
		return executeBuildDeps(root, cfg.Build.DependsOn, os.Environ())
	}

	if !skipDeps && len(cfg.Build.DependsOn) > 0 {
		_ = cfg.LoadDotEnv(root)
		if code := executeBuildDeps(root, cfg.Build.DependsOn, os.Environ()); code != errs.ExitCodeSuccess {
			return code
		}
	}

	outDir := joinRoot(root, firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output), config.DefaultOutput)
	pruneStale(outDir)

	target := ""
	if flagProvided(fs, "target") {
		target = strings.ToLower(strings.TrimSpace(flagValue(fs, "target")))
	}

	if handled, code := buildByTarget(root, cfg, fs, target, outDir); handled {
		return code
	}

	// Auto-detect project shape
	res, err := shape.Detect(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew: "+err.Error())
		return errs.ExitCodeUsage
	}

	if res.IsPreInit() {
		fmt.Fprintln(os.Stderr, "krew build: project is not equipped yet — run 'krew init' to initialize")
		return errs.ExitCodeUsage
	}

	built, firstErr := buildProject(root, cfg, fs, res)
	if !built {
		if firstErr != errs.ExitCodeSuccess {
			return firstErr
		}
		fmt.Fprintln(os.Stderr, "krew: no buildable workload found — add main.go, pages/*.kiw, an `ssg:`/`infra:` key to krewire.yaml, or content/")
		return errs.ExitCodeUsage
	}

	writeManifest(outDir, collectCreated(outDir), cfg.Build.Outputs)
	return errs.ExitCodeSuccess
}

func buildByTarget(root string, cfg *config.Config, fs *flag.FlagSet, target, outDir string) (bool, errs.ExitCode) {
	switch target {
	case "wasm":
		if code := buildWASM(root, cfg, fs); code != errs.ExitCodeSuccess {
			return true, code
		}
		writeManifest(outDir, collectCreated(outDir), cfg.Build.Outputs)
		return true, errs.ExitCodeSuccess
	case "infra":
		if code := buildInfra(root, cfg, fs); code != errs.ExitCodeSuccess {
			return true, code
		}
		writeManifest(outDir, collectCreated(outDir), cfg.Build.Outputs)
		return true, errs.ExitCodeSuccess
	case "binary":
		if code := buildAppBinary(root, cfg, fs); code != errs.ExitCodeSuccess {
			return true, code
		}
		writeManifest(outDir, collectCreated(outDir), cfg.Build.Outputs)
		return true, errs.ExitCodeSuccess
	case "ssg", "site":
		hasPages := hasDir(root, "pages")
		hasSSG := cfg.IsSSG()
		if !hasPages && !hasSSG {
			fmt.Fprintln(os.Stderr, "krew build: no SSG pages or ssg: config found")
			return true, errs.ExitCodeUsage
		}
		var code errs.ExitCode
		if hasPages {
			code, _, _ = buildSSGFromFile(root, cfg, fs)
		} else {
			code, _, _ = buildSSGFromConfig(root, cfg, fs)
		}
		if code != errs.ExitCodeSuccess {
			return true, code
		}
		writeManifest(outDir, collectCreated(outDir), cfg.Build.Outputs)
		return true, errs.ExitCodeSuccess
	case "book":
		if !hasDir(root, "content") {
			fmt.Fprintln(os.Stderr, "krew build: no content/ directory found for book")
			return true, errs.ExitCodeUsage
		}
		if code := buildManuscript(root, cfg, fs, false, nil, nil, ""); code != errs.ExitCodeSuccess {
			return true, code
		}
		writeManifest(outDir, collectCreated(outDir), cfg.Build.Outputs)
		return true, errs.ExitCodeSuccess
	case "":
		return false, errs.ExitCodeSuccess
	default:
		fmt.Fprintf(os.Stderr, "krew build: unknown target %q — supported: site, book, binary, wasm, infra\n", target)
		return true, errs.ExitCodeUsage
	}
}

func buildProject(root string, cfg *config.Config, fs *flag.FlagSet, res shape.Result) (bool, errs.ExitCode) {
	hasPages := hasDir(root, "pages")
	hasSSG := cfg.IsSSG()
	hasBook := hasDir(root, "content")

	built := false
	var firstErr errs.ExitCode = errs.ExitCodeSuccess

	// 1. Infra dispatch
	if res.Capabilities.HasInfra {
		if code := buildInfra(root, cfg, fs); code != errs.ExitCodeSuccess {
			return false, code
		}
		built = true
	}

	// 2. Binary dispatch
	if res.Capabilities.HasApp {
		if code := buildAppBinary(root, cfg, fs); code != errs.ExitCodeSuccess {
			return false, code
		}
		built = true
	}

	// 3. Static web assets / Book documentation
	var planCSS, planJS []string
	if hasPages {
		code, css, js := buildSSGFromFile(root, cfg, fs)
		planCSS, planJS = css, js
		if code != errs.ExitCodeSuccess {
			if firstErr == errs.ExitCodeSuccess {
				firstErr = code
			}
		} else {
			built = true
		}
	} else if hasSSG {
		code, css, js := buildSSGFromConfig(root, cfg, fs)
		planCSS, planJS = css, js
		if code != errs.ExitCodeSuccess {
			if firstErr == errs.ExitCodeSuccess {
				firstErr = code
			}
		} else {
			built = true
		}
	}

	if hasBook {
		hybrid := hasPages || hasSSG
		bookBuilt, code := buildBookWorkload(root, cfg, fs, hybrid, planCSS, planJS)
		if bookBuilt {
			built = true
		} else if code != errs.ExitCodeSuccess && firstErr == errs.ExitCodeSuccess {
			firstErr = code
		}
	}

	// 4. WASM entry point if configured or wasm directory present
	if cfg.Wasm.Entry != "" || hasDir(root, "wasm") {
		if code := buildWASM(root, cfg, fs); code != errs.ExitCodeSuccess {
			return false, code
		}
		built = true
	}

	return built, firstErr
}

func buildBookWorkload(root string, cfg *config.Config, fs *flag.FlagSet, hybrid bool, planCSS, planJS []string) (bool, errs.ExitCode) {
	if cfg.I18n == nil || len(cfg.I18n.Locales) == 0 {
		if code := buildManuscript(root, cfg, fs, hybrid, planCSS, planJS, ""); code != errs.ExitCodeSuccess {
			return false, code
		}
		return true, errs.ExitCodeSuccess
	}

	built := false
	var firstErr errs.ExitCode = errs.ExitCodeSuccess

	for _, locale := range cfg.I18n.Locales {
		localeInput := filepath.Join("content", locale, "docs")
		if !hasDir(root, localeInput) {
			localeInput = filepath.Join("content", "docs", locale)
		}
		if !hasDir(root, localeInput) && cfg.Input != "" {
			localeInput = filepath.Join(cfg.Input, locale)
			if !hasDir(root, localeInput) {
				localeInput = filepath.Join(cfg.Input, locale, "docs")
			}
		}
		if !hasDir(root, localeInput) {
			slog.Warn("i18n: skipping locale, input dir missing", "locale", locale, "input", localeInput)
			continue
		}
		origInput := cfg.Input
		origOutput := cfg.Output
		origBase := cfg.Base
		cfg.Input = localeInput
		if locale == cfg.I18n.DefaultLocale {
			cfg.Output = origOutput
		} else {
			cfg.Output = filepath.Join(origOutput, locale)
			cfg.Base = localeBasePath(origBase, locale, cfg.Book.Mount)
		}
		slog.Info("i18n: building book for locale", "locale", locale, "input", cfg.Input, "output", cfg.Output, "base", cfg.Base)
		if code := buildManuscript(root, cfg, fs, hybrid, planCSS, planJS, locale); code != errs.ExitCodeSuccess {
			if firstErr == errs.ExitCodeSuccess {
				firstErr = code
			}
		} else {
			built = true
			if locale == cfg.I18n.DefaultLocale {
				_ = copyDir(filepath.Join(root, origOutput, "docs"), filepath.Join(root, origOutput, locale, "docs"))
			}
		}
		cfg.Input = origInput
		cfg.Output = origOutput
		cfg.Base = origBase
	}
	return built, firstErr
}
