package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/krewire/krew/internal/buildinfo"
	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krew/internal/shape"
	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/web/ssg/hmr"
)

// RegisterDev registers flags for the dev command.
func RegisterDev(fs *flag.FlagSet) {
	fs.String("addr", "", "listen address for the app (default :8080)")
	fs.Duration("interval", 500*time.Millisecond, "file-watch polling interval")
	fs.String("output", "", "output directory (default .krewire/build)")
	fs.String("o", "", "output directory (shorthand for --output)")
	fs.String("input", "", "content directory (default content)")
	fs.String("base", "", "URL base the site will be served under (default /)")
	registerRuntimeFlags(fs)
}

// RunDev runs the project in dev mode. For an app it rebuilds and restarts
// the child on change; for site/book projects it runs a live development server
// with file watching and auto-rebuild.
func RunDev(fs *flag.FlagSet) errs.ExitCode {
	rt, code := bootRuntime(fs)
	if code != errs.ExitCodeSuccess {
		return code
	}

	res, err := shape.Detect(rt.root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew: "+err.Error())
		return errs.ExitCodeUsage
	}

	if res.IsPreInit() {
		fmt.Fprintln(os.Stderr, "krew dev: project is not equipped yet — run 'krew init' to initialize")
		return errs.ExitCodeUsage
	}

	if res.Capabilities.HasInfra && !res.Capabilities.HasApp && !res.Capabilities.HasPages {
		fmt.Fprintln(os.Stderr, "krew dev: this is an infra configuration — use 'krew build --plan' or 'krew deploy' instead")
		return errs.ExitCodeUsage
	}

	if res.Capabilities.HasApp {
		return devApp(rt, fs)
	}
	if res.Capabilities.HasPages || res.Capabilities.HasBook {
		return devSite(rt, fs)
	}

	fmt.Fprintln(os.Stderr, "krew dev: no project found — run 'krew new <project>' first")
	return errs.ExitCodeUsage
}

// devSite runs a live development server for static sites and book projects,
// rebuilding into .krewire/build whenever source files change.
func devSite(rt *runtimeEnv, fs *flag.FlagSet) errs.ExitCode {
	root, cfg := rt.root, rt.cfg
	addr := resolvedAddr(flagValue(fs, "addr"), cfg)
	interval := fs.Lookup("interval").Value
	every := 500 * time.Millisecond
	if d, err := time.ParseDuration(interval.String()); err == nil && d > 0 {
		every = d
	}

	buildSite := func() errs.ExitCode {
		return RunBuild(fs)
	}

	slog.Info("building site for development")
	if code := buildSite(); code != errs.ExitCodeSuccess {
		slog.Error("initial site build failed")
		return code
	}

	outDir := joinRoot(root, firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output), config.DefaultOutput)

	hmrServer := hmr.NewServer()
	srv := &http.Server{
		Addr:    addr,
		Handler: hmrServer.Handler(extensionlessFS(outDir)),
	}

	serverErrCh := make(chan error, 1)
	go func() {
		displayAddr := addr
		if strings.HasPrefix(displayAddr, ":") {
			displayAddr = "http://localhost" + displayAddr
		} else {
			displayAddr = "http://" + displayAddr
		}
		slog.Info("dev server running", "url", displayAddr, "dir", outDir)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	watcher := newWatcher(root, every, cfg)
	slog.Info("watching for changes", "root", root, "interval", every)

	for {
		select {
		case err := <-serverErrCh:
			return fail(err)
		case sig := <-sigCh:
			slog.Info("dev received signal, stopping server", "signal", sig)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = srv.Shutdown(ctx)
			return errs.ExitCodeSuccess
		case <-watcher.Changed():
			watcher.Reset()
			slog.Info("change detected, rebuilding...")
			if code := buildSite(); code != errs.ExitCodeSuccess {
				slog.Error("site rebuild failed; keeping previous preview server running", "code", code)
				continue
			}
			slog.Info("site rebuilt successfully")
			hmrServer.Broadcast(hmr.Event{Type: hmr.EventReload})
		}
	}
}

// devApp runs the app, watching the module and declared asset/markup roots,
// rebuilding and restarting on change. A failed rebuild keeps the previous
// child running (RND-DEV-002).
func devApp(rt *runtimeEnv, fs *flag.FlagSet) errs.ExitCode {
	root, cfg := rt.root, rt.cfg
	env, debug := rt.env, rt.debug
	dir, err := os.MkdirTemp("", "krewire-dev-")
	if err != nil {
		return fail(err)
	}
	defer os.RemoveAll(dir)

	bin := filepath.Join(dir, "app")
	addr := resolvedAddr(flagValue(fs, "addr"), cfg)
	interval := fs.Lookup("interval").Value
	every := 500 * time.Millisecond
	if d, err := time.ParseDuration(interval.String()); err == nil && d > 0 {
		every = d
	}

	entry := resolveEntry(root)
	build := func() error {
		cmd := exec.Command("go", "build", "-ldflags="+buildinfo.LDFlags(false), "-o", bin, entry)
		cmd.Dir = root
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		slog.Info("dev build", "entry", entry)
		if err := cmd.Run(); err != nil {
			return err
		}
		if hasDir(root, "wasm") || cfg.Wasm.Entry != "" {
			_ = buildWASM(root, cfg, fs)
		}
		return nil
	}
	if err := build(); err != nil {
		return fail(err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	watcher := newWatcher(root, every, cfg)
	restart := func() { restartApp(bin, root, addr, env, debug) }

	restart()
	slog.Info("watching for changes", "root", root, "interval", every)
	for {
		select {
		case sig := <-sigCh:
			slog.Info("dev received signal, stopping app", "signal", sig)
			stopChild()
			return errs.ExitCodeSuccess
		case <-watcher.Changed():
			watcher.Reset()
			slog.Info("change detected")
			if err := build(); err != nil {
				slog.Error("dev build failed; keeping previous process running", "error", err)
				continue
			}
			stopChild()
			restart()
		}
	}
}
