package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krew/internal/shape"
	"github.com/krewire/krewire/pkg/book"
	"github.com/krewire/krewire/pkg/errs"
)

// serveDefaultTimeouts mirrors web.App.Serve to prevent Slowloris attacks.
const (
	serveReadHeaderTimeout = 5 * time.Second
	serveReadTimeout       = 15 * time.Second
	serveWriteTimeout      = 15 * time.Second
	serveIdleTimeout       = 60 * time.Second
	serveShutdownTimeout   = 10 * time.Second
)

// listenAndServe starts an http.Server with hardened timeouts on addr and
// blocks until the process receives SIGINT/SIGTERM, then shuts down
// gracefully within serveShutdownTimeout.
func listenAndServe(addr string, h http.Handler) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: serveReadHeaderTimeout,
		ReadTimeout:       serveReadTimeout,
		WriteTimeout:      serveWriteTimeout,
		IdleTimeout:       serveIdleTimeout,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), serveShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		_ = srv.Close()
		return err
	}
	return nil
}

// RegisterServe registers flags for the serve command.
func RegisterServe(fs *flag.FlagSet) {
	fs.String("input", "", "content directory (default content)")
	fs.String("addr", "", "listen address (default :8080)")
	registerRuntimeFlags(fs)
}

// RunServe starts the current project locally for every kind
// (KWN-6K41E RND-SRV-001): an app is compiled and listened on, a cli
// executes with argument passthrough, and site/book projects preview their
// static output over HTTP. A hybrid project (ssg + book) serves the merged
// .krewire/build statically. It shares the runtime bootstrap with run and
// dev (RND-SRV-002).
func RunServe(fs *flag.FlagSet) errs.ExitCode {
	rt, code := bootRuntime(fs)
	if code != errs.ExitCodeSuccess {
		return code
	}
	res, err := shape.Detect(rt.root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew: "+err.Error())
		return errs.ExitCodeUsage
	}

	if res.Capabilities.HasApp {
		return runApp(rt, fs)
	}
	if res.Capabilities.HasPages || rt.cfg.IsSSG() || res.Capabilities.HasBook {
		return serveStaticBuild(rt.root, fs)
	}
	fmt.Fprintln(os.Stderr, "krew serve: no project found — add a root main.go, an `ssg:` key in krewire.yaml, or a content/ directory")
	return errs.ExitCodeUsage
}

// serveBook previews the content directory with the book handler. The
// input resolves like every other setting: --input flag > krewire.yaml
// `input:` > config.DefaultInput ("content"). Include/exclude globs follow
// the same precedence as krew build so preview matches the export.
func serveBook(root string, fs *flag.FlagSet) errs.ExitCode {
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}
	input := joinRoot(root, flagValue(fs, "input"), firstNonEmpty(cfg.Input, config.DefaultInput))
	addr := resolvedAddr(flagValue(fs, "addr"), cfg)
	b, err := book.LoadWithRules(
		input, "", "", "/",
		globList(fs, "include", cfg.Build.Include),
		globList(fs, "exclude", cfg.Build.Exclude),
	)
	if err != nil {
		return fail(err)
	}
	slog.Info("serving book", "addr", addr)
	if err := book.Serve(b, addr); err != nil {
		return fail(err)
	}
	return errs.ExitCodeSuccess
}

// serveStaticBuild previews the built .krewire/build output over HTTP — the
// shared preview for hybrid (ssg + book) projects and file-mode sites.
func serveStaticBuild(root string, fs *flag.FlagSet) errs.ExitCode {
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}
	outDir := joinRoot(root, firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output), config.DefaultOutput)
	if info, err := os.Stat(filepath.Join(outDir, "index.html")); err != nil || info.IsDir() {
		hasIndex := false
		for _, o := range cfg.Build.Outputs {
			p := joinRoot(root, o, o)
			if fi, err := os.Stat(filepath.Join(p, "index.html")); err == nil && !fi.IsDir() {
				hasIndex = true
				break
			}
		}
		if !hasIndex {
			fmt.Fprintf(os.Stderr, "krew serve: no build output at %s — run 'krew build' first\n", outDir)
			return errs.ExitCodeUsage
		}
	}
	addr := resolvedAddr(flagValue(fs, "addr"), cfg)
	slog.Info("serving static build", "dir", outDir, "addr", addr)
	roots := []string{outDir}
	for _, o := range cfg.Build.Outputs {
		p := joinRoot(root, o, o)
		if hasDir(root, o) {
			roots = append(roots, p)
		}
	}
	var handler http.Handler
	if len(roots) > 1 {
		handler = virtualRootFS(roots...)
	} else {
		handler = extensionlessFS(outDir)
	}
	if err := listenAndServe(addr, handler); err != nil {
		return fail(err)
	}
	return errs.ExitCodeSuccess
}

// extensionlessFS serves the build output with extensionless URL resolution,
// mirroring the emitted sibling .html files: /docs → docs.html and
// /docs/x → docs/x.html. Directory listings are never shown.
func extensionlessFS(root string) http.Handler {
	fs := http.Dir(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := path.Clean("/" + r.URL.Path)
		if p == "/" {
			p = "/index.html"
		}
		for _, candidate := range []string{p, p + ".html"} {
			f, err := fs.Open(candidate)
			if err != nil {
				continue
			}
			st, statErr := f.Stat()
			closeErr := f.Close()
			if statErr != nil || closeErr != nil || st.IsDir() {
				continue
			}
			http.ServeFile(w, r, filepath.Join(root, filepath.FromSlash(candidate)))
			return
		}
		http.NotFound(w, r)
	})
}

// serveSSG previews the project's declarative ssg site from krewire.yaml.
func serveSSG(fs *flag.FlagSet, cfg *config.Config) errs.ExitCode {
	ssgCfg := cfg.ToSSGConfig()
	addr := resolvedAddr(flagValue(fs, "addr"), cfg)
	slog.Info("serving site", "addr", addr)
	site, err := ssgCfg.Site()
	if err != nil {
		return fail(err)
	}
	if err := listenAndServe(addr, site.Handler()); err != nil {
		return fail(err)
	}
	return errs.ExitCodeSuccess
}
