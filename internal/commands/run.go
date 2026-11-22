package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"

	"github.com/krewire/krew/internal/buildinfo"
	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krew/internal/shape"
	"github.com/krewire/krewire/pkg/errs"
)

// RegisterRun registers flags for the run command.
func RegisterRun(fs *flag.FlagSet) {
	fs.String("addr", "", "listen address for the app (default "+defaultAddr+")")
	registerRuntimeFlags(fs)
}

// RunRun builds and runs a fullstack app in production mode: it compiles the
// project into a temp binary, executes it passing the listen address through
// APP_ADDR, and forwards SIGINT/SIGTERM so the app's graceful shutdown runs
// inside the app. Site and book projects are rejected with guidance to use
// build + serve.
//
// Extended by KWN-SCRIPT-9F3KQ: when the first positional argument is a Go
// file path it runs `go run <file> -- args` (KWN-SCR-001); when it matches
// `scripts.<task>` in krewire.yaml it runs the shell command (KWN-SCR-002).
// Precedence is file > script > project (KWN-SCR-004).
func RunRun(fs *flag.FlagSet) errs.ExitCode {
	rt, code := bootRuntime(fs)
	if code != errs.ExitCodeSuccess {
		return code
	}

	args := fs.Args()
	if len(args) > 0 {
		first := args[0]
		isFile := isGoFileArg(rt.root, first)
		_, isScript := rt.cfg.Scripts[first]
		if isFile && isScript {
			fmt.Fprintf(os.Stderr, "warning: both file and task %q exist; running file\n", first)
		}
		if isFile {
			return runGoFile(rt, first, args[1:])
		}
		if isScript {
			return runScript(rt, first, args[1:])
		}
		if strings.HasSuffix(first, ".go") {
			fmt.Fprintf(os.Stderr, "krew run: file %q not found\n", first)
			return errs.ExitCodeUsage
		}
		if len(rt.cfg.Scripts) > 0 && isTaskLike(first) {
			// Unknown task — only error if project is not a runnable app/cli,
			// otherwise treat as CLI args for backward compat (KWN-SCR-003).
			// Peek detection without fully booting: check for main.go.
			if !hasDir(rt.root, "cmd") && !hasFile(rt.root, "main.go") {
				fmt.Fprintf(os.Stderr, "krew run: unknown task %q\nAvailable tasks: %s\n", first, formatTasks(rt.cfg.Scripts))
				return errs.ExitCodeUsage
			}
		}
	}

	res, err := shape.Detect(rt.root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew: "+err.Error())
		return errs.ExitCodeUsage
	}

	if res.IsPreInit() {
		fmt.Fprintln(os.Stderr, "krew run: project is not equipped yet — run 'krew init' to initialize")
		return errs.ExitCodeUsage
	}

	if res.Capabilities.HasApp {
		return runApp(rt, fs)
	}
	if res.Capabilities.HasPages || res.Capabilities.HasBook {
		fmt.Fprintln(os.Stderr, "krew run: static sites and books — use 'krew dev' for a live dev server, or 'krew build' then 'krew serve'")
		return errs.ExitCodeUsage
	}
	if res.Capabilities.HasInfra {
		fmt.Fprintln(os.Stderr, "krew run: infra configuration — use 'krew build --plan' or 'krew deploy' instead")
		return errs.ExitCodeUsage
	}
	fmt.Fprintln(os.Stderr, "krew run: no executable main package found — add main.go or run 'krew new <project>'")
	return errs.ExitCodeUsage
}

// runApp builds the app to a temp binary and runs it, streaming output and
// forwarding signals.
func runApp(rt *runtimeEnv, fs *flag.FlagSet) errs.ExitCode {
	if rt.cfg.IsSSG() || hasDir(rt.root, "content") {
		if code := exportSiteAssets(rt.root, fs); code != errs.ExitCodeSuccess {
			return code
		}
	}

	dir, err := os.MkdirTemp("", "krewire-run-")
	if err != nil {
		return fail(err)
	}
	defer os.RemoveAll(dir)

	bin := filepath.Join(dir, "app")
	addr := resolvedAddr(flagValue(fs, "addr"), rt.cfg)

	entry := resolveEntry(rt.root)
	slog.Info("building app", "bin", bin, "entry", entry)
	build := exec.Command("go", "build", "-ldflags="+buildinfo.LDFlags(false), "-o", bin, entry)
	build.Dir = rt.root
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fail(fmt.Errorf("build app: %w", err))
	}

	slog.Info("running app", "addr", addr)
	var args []string
	if fs != nil {
		args = fs.Args()
	}
	cmd := exec.Command(bin, args...)
	cmd.Dir = rt.root
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = childEnviron(os.Environ(), addr, rt.env, rt.debug)
	if err := cmd.Start(); err != nil {
		return fail(err)
	}

	return waitChild(cmd)
}

// runCLI builds and runs a CLI project in production mode: it compiles the
// root main.go into a temp binary and executes it, forwarding SIGINT/SIGTERM
// so the CLI's own exit path resolves the process code. No listen address is
// injected for CLI projects.
func runCLI(rt *runtimeEnv, fs *flag.FlagSet) errs.ExitCode {
	dir, err := os.MkdirTemp("", "krewire-run-")
	if err != nil {
		return fail(err)
	}
	defer os.RemoveAll(dir)

	bin := filepath.Join(dir, "app")

	entry := resolveEntry(rt.root)
	slog.Info("building CLI", "bin", bin, "entry", entry)
	build := exec.Command("go", "build", "-ldflags="+buildinfo.LDFlags(false), "-o", bin, entry)
	build.Dir = rt.root
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fail(fmt.Errorf("build cli: %w", err))
	}

	cmd := exec.Command(bin, fs.Args()...)
	cmd.Dir = rt.root
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = childEnviron(os.Environ(), "", rt.env, rt.debug)
	if err := cmd.Start(); err != nil {
		return fail(err)
	}

	return waitChild(cmd)
}

// waitChild forwards SIGINT/SIGTERM to the child process and resolves the
// exit code from the child's status.
func waitChild(cmd *exec.Cmd) errs.ExitCode {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	for {
		select {
		case err := <-done:
			if err == nil {
				return errs.ExitCodeSuccess
			}
			return childExitCode(cmd, err)
		case sig := <-sigCh:
			slog.Info("forwarding signal to app", "signal", sig)
			if err := cmd.Process.Signal(sig); err != nil {
				slog.Warn("signal forward failed", "error", err)
			}
		}
	}
}

// childExitCode resolves a completed child command to a canonical exit code,
// preferring the child's own exit status when available.
func childExitCode(cmd *exec.Cmd, err error) errs.ExitCode {
	if cmd.Process != nil && cmd.ProcessState != nil {
		return errs.ExitCodeFromInt(cmd.ProcessState.ExitCode())
	}
	if err != nil {
		return errs.ExitCodeFailure
	}
	return errs.ExitCodeSuccess
}

// exportSiteAssets refreshes the site/book export so embedded assets are
// current before the app starts (RND-RUN-005). No-op when no site is declared.
func exportSiteAssets(root string, fs *flag.FlagSet) errs.ExitCode {
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}
	hasPages := hasDir(root, "pages")
	hasSSG := cfg.IsSSG()
	hasBook := hasDir(root, "content")
	if !hasPages && !hasSSG && !hasBook {
		return errs.ExitCodeSuccess
	}
	// Mirror RunBuild: SSG from file (pages/) takes precedence over ssg: config
	var planCSS, planJS []string
	if hasPages {
		code, css, js := buildSSGFromFile(root, cfg, fs)
		planCSS, planJS = css, js
		if code != errs.ExitCodeSuccess {
			return code
		}
	} else if hasSSG {
		code, css, js := buildSSGFromConfig(root, cfg, fs)
		planCSS, planJS = css, js
		if code != errs.ExitCodeSuccess {
			return code
		}
	}
	if hasBook {
		if code := buildManuscript(root, cfg, fs, hasPages || hasSSG, planCSS, planJS, ""); code != errs.ExitCodeSuccess {
			return code
		}
	}
	return errs.ExitCodeSuccess
}

func isGoFileArg(root, arg string) bool {
	if !strings.HasSuffix(arg, ".go") {
		return false
	}
	path := arg
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, arg)
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func runGoFile(rt *runtimeEnv, path string, extraArgs []string) errs.ExitCode {
	if len(extraArgs) > 0 && extraArgs[0] == "--" {
		extraArgs = extraArgs[1:]
	}
	args := []string{"run", "-ldflags=" + buildinfo.LDFlags(false), path}
	args = append(args, extraArgs...)
	slog.Info("running go file", "path", path)
	cmd := exec.Command("go", args...)
	cmd.Dir = rt.root
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = childEnviron(os.Environ(), "", rt.env, rt.debug)
	if err := cmd.Start(); err != nil {
		return fail(err)
	}
	return waitChild(cmd)
}

func quoteShellArg(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func quoteCmdArg(s string) string {
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

func buildScriptCmd(script string, extraArgs []string) *exec.Cmd {
	if len(extraArgs) > 0 {
		quoted := make([]string, len(extraArgs))
		if runtime.GOOS == "windows" {
			for i, a := range extraArgs {
				quoted[i] = quoteCmdArg(a)
			}
			script = script + " " + strings.Join(quoted, " ")
		} else {
			for i, a := range extraArgs {
				quoted[i] = quoteShellArg(a)
			}
			script = script + " " + strings.Join(quoted, " ")
		}
	}
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/C", script)
	}
	return exec.Command("sh", "-c", script)
}

func runScript(rt *runtimeEnv, task string, extraArgs []string) errs.ExitCode {
	script, ok := rt.cfg.Scripts[task]
	if !ok {
		fmt.Fprintf(os.Stderr, "krew run: unknown task %q\nAvailable tasks: %s\n", task, formatTasks(rt.cfg.Scripts))
		return errs.ExitCodeUsage
	}
	if len(extraArgs) > 0 && extraArgs[0] == "--" {
		extraArgs = extraArgs[1:]
	}
	cmd := buildScriptCmd(script, extraArgs)
	slog.Info("running script", "task", task, "command", cmd.String())
	cmd.Dir = rt.root
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = childEnviron(os.Environ(), "", rt.env, rt.debug)
	if err := cmd.Start(); err != nil {
		return fail(err)
	}
	return waitChild(cmd)
}

func isTaskLike(s string) bool {
	if s == "" {
		return false
	}
	if strings.Contains(s, "/") || strings.Contains(s, "\\") {
		return false
	}
	if strings.HasSuffix(s, ".go") {
		return false
	}
	return true
}

func formatTasks(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
