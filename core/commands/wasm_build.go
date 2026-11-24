package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/krewire/krew/core/config"
	"github.com/krewire/krewire/errs"
	"github.com/krewire/krewire/runtime/build"
)

// buildWASM compiles the project's WASM entry point (KWF-T4X9P FRK-WASM-002).
func buildWASM(root string, cfg *config.Config, fs *flag.FlagSet) errs.ExitCode {
	entry := firstNonEmpty(cfg.Wasm.Entry, DefaultWASMEntry)
	outDir := joinRoot(root, firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output), config.DefaultOutput)
	name := firstNonEmpty(cfg.Wasm.Name, DefaultWASMName)

	slog.Info("building WASM module", "entry", entry, "output", outDir)
	artifacts, err := build.BuildWASM(build.Config{
		Entry:  entry,
		OutDir: outDir,
		Name:   name,
	})
	if err != nil {
		return fail(err)
	}

	// Mirror artifacts into assets/ subdirectory for web client availability
	assetsDir := filepath.Join(outDir, "assets")
	_ = os.MkdirAll(assetsDir, 0o755)
	if data, err := os.ReadFile(artifacts.WASM); err == nil {
		_ = os.WriteFile(filepath.Join(assetsDir, filepath.Base(artifacts.WASM)), data, 0o644)
	}
	if data, err := os.ReadFile(artifacts.JS); err == nil {
		_ = os.WriteFile(filepath.Join(assetsDir, filepath.Base(artifacts.JS)), data, 0o644)
	}

	fp := build.Fingerprint(artifacts.Digest)
	slog.Info("WASM build complete", "wasm", filepath.Base(artifacts.WASM), "js", filepath.Base(artifacts.JS), "fingerprint", fp)
	fmt.Printf("created %s (fingerprint: %s)\n", artifacts.WASM, fp)
	return errs.ExitCodeSuccess
}
