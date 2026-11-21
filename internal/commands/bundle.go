package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krewire/pkg/assets"
	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/hub/plugin"
	"github.com/krewire/krewire/pkg/web/ssg"
)

// RegisterBundle registers flags for the `krew bundle` command.
func RegisterBundle(fs *flag.FlagSet) {
	fs.String("output", "", "target output directory (default: derived from krewire.yaml or .krewire/build)")
	fs.String("o", "", "alias for --output")
	fs.Bool("minify", false, "enable CSS and JS minification")
	fs.Bool("m", false, "alias for --minify")
}

// RunBundle bundles static CSS and JS assets into krewire.css and krewire.js.
func RunBundle(fs *flag.FlagSet) errs.ExitCode {
	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew: "+err.Error())
		return errs.ExitCodeUsage
	}
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}

	outFlag := firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output, config.DefaultOutput)
	outDir := joinRoot(root, outFlag, config.DefaultOutput)

	minify := flagValue(fs, "minify") == "true" || flagValue(fs, "m") == "true"
	slog.Info("bundling assets into krewire.css and krewire.js", "root", root, "output", outDir, "minify", minify)

	// 1. Detect if this is an SSG or file-based site layout
	if hasSiteLayout(root) {
		site, err := ssg.LoadFromDir(root)
		if err == nil {
			// Run detected plugins (e.g. Tailwind) first
			for _, p := range plugin.Registry {
				if p.Detect(root) {
					if berr := p.Build(root, outDir); berr == nil {
						if ap, ok := p.(plugin.AssetProvider); ok {
							site.DeclareAsset(ap.Assets(root)...)
						}
					}
				}
			}
			site.WithBundle(true)
			site.WithMinify(minify)
			if _, berr := site.Build(outDir); berr != nil {
				return fail(berr)
			}
			printBundleSuccess(outDir)
			return errs.ExitCodeSuccess
		}
	}

	// 2. Direct asset store bundling (scans public/assets and custom assets)
	store := assets.NewStore()
	publicDir := filepath.Join(root, "public")
	if info, err := os.Stat(publicDir); err == nil && info.IsDir() {
		store.MountDir(publicDir)
	}
	assetsDir := filepath.Join(root, "assets")
	if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
		store.MountDir(assetsDir)
	}

	names, err := store.Names()
	if err != nil || len(names) == 0 {
		// Also check outDir if assets were previously written there
		if info, err := os.Stat(filepath.Join(outDir, "assets")); err == nil && info.IsDir() {
			store.MountDir(outDir)
			names, _ = store.Names()
		}
	}

	var cssFiles []string
	var jsFiles []string
	for _, n := range names {
		lower := strings.ToLower(n)
		if strings.HasSuffix(lower, "krewire.css") || strings.HasSuffix(lower, "krewire.js") {
			continue // skip self
		}
		if strings.HasSuffix(lower, ".css") {
			cssFiles = append(cssFiles, n)
		} else if strings.HasSuffix(lower, ".js") {
			jsFiles = append(jsFiles, n)
		}
	}

	if len(cssFiles) == 0 && len(jsFiles) == 0 {
		fmt.Println("krew bundle: no CSS or JS assets found to bundle")
		return errs.ExitCodeSuccess
	}

	bundler := assets.NewBundler(assets.WithMinify(minify))
	res, err := bundler.Bundle(store, cssFiles, jsFiles)
	if err != nil {
		return fail(err)
	}

	created, err := bundler.WriteToDir(res, outDir)
	if err != nil {
		return fail(err)
	}

	for _, c := range created {
		fullPath := filepath.Join(outDir, filepath.FromSlash(c))
		size := int64(0)
		if fi, err := os.Stat(fullPath); err == nil {
			size = fi.Size()
		}
		fmt.Printf("bundled %s (%d bytes)\n", c, size)
	}

	return errs.ExitCodeSuccess
}

func printBundleSuccess(outDir string) {
	for _, rel := range []string{assets.DefaultCSSBundlePath, assets.DefaultJSBundlePath} {
		p := filepath.Join(outDir, filepath.FromSlash(rel))
		if fi, err := os.Stat(p); err == nil {
			fmt.Printf("bundled %s (%d bytes)\n", rel, fi.Size())
		}
	}
}

func hasSiteLayout(root string) bool {
	for _, sub := range []string{"pages", "layouts", "components", "public"} {
		if fi, err := os.Stat(filepath.Join(root, sub)); err == nil && fi.IsDir() {
			return true
		}
	}
	return false
}
