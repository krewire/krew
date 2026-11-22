package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/krewire/krew/internal/buildinfo"
	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krewire/pkg/errs"
)

// buildAppBinary compiles the project Go binary into .krewire/build/bin/<name> (KWN-BUILD-001).
func buildAppBinary(root string, cfg *config.Config, fs *flag.FlagSet) errs.ExitCode {
	name := firstNonEmpty(cfg.Name, moduleName(root), DefaultBinaryName)
	outDir := joinRoot(root, firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output), config.DefaultOutput)
	binDir := filepath.Join(outDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fail(err)
	}
	binPath := filepath.Join(binDir, name)

	entry := resolveEntry(root)
	if entry == "." && !hasFile(root, "main.go") && !hasGoFiles(root) {
		return fail(fmt.Errorf("build binary: no main package found in root or cmd/"))
	}

	slog.Info("building binary", "entry", entry, "output", binPath)
	cmd := exec.Command("go", "build", "-trimpath", "-ldflags="+buildinfo.LDFlags(true), "-o", binPath, entry)
	cmd.Dir = root
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fail(fmt.Errorf("build binary: %w", err))
	}
	fmt.Println("created " + binPath)
	return errs.ExitCodeSuccess
}

// hasGoFiles reports whether dir contains non-test Go source files.
func hasGoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
			return true
		}
	}
	return false
}
