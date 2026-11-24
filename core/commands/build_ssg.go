package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"path"
	"strings"

	"github.com/krewire/hub/plugin"
	"github.com/krewire/krew/core/config"
	"github.com/krewire/krewire/errs"
	"github.com/krewire/krewire/web/ssg"
)

// buildSSGFromFile builds the project's SSG site from the file-based layout
// (pages/, components/, layouts/, content/, public/) using
// ssg.LoadFromDir. krewire.yaml supplies metadata (title, description, theme)
// and output dir.
//
// Detected plugins (e.g. Tailwind via tailwind.config.js) run first and their
// CSS/JS output paths are declared to the site, so `krew build` auto-injects the
// matching tags — the layout never names plugin assets by hand.
func buildSSGFromFile(root string, cfg *config.Config, fs *flag.FlagSet) (errs.ExitCode, []string, []string) {
	output := firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output, config.DefaultOutput)
	outDir := joinRoot(root, output, config.DefaultOutput)
	slog.Info("building SSG site from file layout", "root", root, "output", outDir)
	site, err := ssg.LoadFromDir(root)
	if err != nil {
		return fail(err), nil, nil
	}
	// The effective URL base honors krew --base first, then krewire.yaml
	// (build.base folded into cfg.Base at load). A file-based project has
	// no `ssg:` block, so thread it here; declarative sites receive it
	// through ToSSGConfig instead.
	if base := firstNonEmpty(flagValue(fs, "base"), cfg.Base, ""); base != "" {
		site.WithBasePath(base)
	}
	// Run detected plugins (Tailwind is the first; others follow the same pattern).
	for _, p := range plugin.Registry {
		if !p.Detect(root) {
			continue
		}
		slog.Info("plugin detected", "plugin", p.Name())
		if err := p.Build(root, outDir); err != nil {
			slog.Warn("plugin build failed", "plugin", p.Name(), "err", err)
			continue
		}
		slog.Info("plugin built", "plugin", p.Name())
		if ap, ok := p.(plugin.AssetProvider); ok {
			site.DeclareAsset(ap.Assets(root)...)
		}
	}
	if flagValue(fs, "bundle") != "false" {
		site.WithBundle(true)
	}
	if flagValue(fs, "minify") == "true" {
		site.WithMinify(true)
	}
	created, err := site.Build(outDir)
	if err != nil {
		return fail(err), nil, nil
	}
	warnAutoAssetErrors(site)
	// Post-render build pass: plugins that scan generated HTML artifacts
	// (such as Tailwind) re-run to capture classes emitted by component expansions.
	for _, p := range plugin.Registry {
		if p.Detect(root) {
			if err := p.Build(root, outDir); err != nil {
				slog.Warn("plugin post-build failed", "plugin", p.Name(), "err", err)
			}
		}
	}
	if flagValue(fs, "bundle") != "false" {
		site.WriteBundleAssets(outDir)
	}
	for _, p := range created {
		fmt.Println("created " + p)
	}
	css, js := site.InjectedAssets()
	return errs.ExitCodeSuccess, css, js
}

// buildSSGFromConfig builds the project's SSG site from the `ssg:` section
// of krewire.yaml. Top-level fields (title, output, theme) are merged into the
// ssg.Config so they don't need to be repeated under ssg:.
func buildSSGFromConfig(root string, cfg *config.Config, fs *flag.FlagSet) (errs.ExitCode, []string, []string) {
	ssgCfg := cfg.ToSSGConfig()
	if flagValue(fs, "bundle") != "false" {
		if ssgCfg.AutoAssets == nil {
			ssgCfg.AutoAssets = &ssg.AutoAssetConfig{}
		}
		trueVal := true
		ssgCfg.AutoAssets.Bundle = &trueVal
	}
	output := firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output, config.DefaultOutput)
	outDir := joinRoot(root, output, config.DefaultOutput)
	slog.Info("building SSG site from krewire.yaml", "output", outDir)
	site, created, err := ssg.BuildFromConfigSite(ssgCfg, outDir)
	if err != nil {
		return fail(err), nil, nil
	}
	warnAutoAssetErrors(site)
	for _, p := range created {
		fmt.Println("created " + p)
	}
	css, js := site.InjectedAssets()
	return errs.ExitCodeSuccess, css, js
}

// warnAutoAssetErrors surfaces asset-layer configuration problems (FRK-AS-074).
// The build keeps going with the default order, but a typo such as
// `auto_assets.order: {theme.css: {layer: last-milee}}` must not be silently
// ignored — it would reorder assets invisibly, which is exactly the class of
// drift the shared plan exists to prevent.
func warnAutoAssetErrors(site *ssg.Site) {
	for _, err := range site.AutoAssetErrors() {
		slog.Warn("krew build: auto_assets config ignored", "err", err)
	}
}

// versionedAssetURLs appends the product version as a cache-busting query,
// matching what the ssg half injects so both halves refresh together. The rule
// itself lives in pkg/web/ssg (VersionedURLs) — this only forwards to it,
// so the devtool and the generator cannot disagree on the query format.
func versionedAssetURLs(urls []string, version string) []string {
	return ssg.VersionedURLs(urls, version)
}

// withoutAsset drops the entry naming asset from an asset plan. Plan URLs are
// "/assets/<name>", so the base name is compared.
func withoutAsset(urls []string, asset string) []string {
	base := path.Base(strings.TrimPrefix(asset, "/"))
	out := make([]string, 0, len(urls))
	for _, u := range urls {
		if path.Base(u) == base {
			continue
		}
		out = append(out, u)
	}
	return out
}
