package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"strings"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krewire/pkg/book"
	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/ui"
)

// buildManuscript renders the project's content/ directory with the book engine.
// Settings come from krewire.yaml in the project root, overridden by flags.
// In a hybrid project (ssg pages present) the book suppresses its generated
// root TOC so the ssg landing page owns "/". Include/exclude path globs
// resolve flag > krewire.yaml `build:` > book defaults (README skipped).
// localeBasePath resolves the URL base a non-default locale is served under.
// The mount path must be part of it: book.Build exports pages under the mount
// directory, so a base without the mount emits links that point at pages the
// export never wrote (e.g. /id/workloads/app instead of /id/docs/workloads/app).
func localeBasePath(base, locale, mount string) string {
	path := strings.TrimRight(base, "/") + "/" + locale + "/"
	if m := strings.Trim(mount, "/"); m != "" {
		path += m + "/"
	}
	return path
}

// buildManuscript renders the book workload. lang is the BCP-47 tag of the
// content being built ("" for a single-locale book) and becomes <html lang> on
// every generated page, so a per-locale i18n build never mislabels its output.
func buildManuscript(root string, cfg *config.Config, fs *flag.FlagSet, hybrid bool, planCSS, planJS []string, lang string) errs.ExitCode {
	title := firstNonEmpty(flagValue(fs, "title"), cfg.Title, moduleName(root))
	input := firstNonEmpty(flagValue(fs, "input"), cfg.Input)
	if input == "" {
		input = config.DefaultInput
	}
	include := globList(fs, "include", cfg.Build.Include)
	exclude := globList(fs, "exclude", cfg.Build.Exclude)
	bcfg := book.Config{
		Input:      joinRoot(root, input, config.DefaultInput),
		Output:     joinRoot(root, firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output), config.DefaultOutput),
		Title:      title,
		Author:     firstNonEmpty(flagValue(fs, "author"), cfg.Author),
		BasePath:   firstNonEmpty(flagValue(fs, "base"), cfg.Base, "/"),
		Version:    cfg.Version,
		NavLinks:   navFromConfig(cfg.Nav),
		FooterText: cfg.Footer,
		Theme:      bookThemeFrom(fs, cfg.Theme),
		MountPath:  cfg.Book.Mount,
		NoRootTOC:  hybrid && cfg.Book.TOC == nil,
		Include:    include,
		Exclude:    exclude,
		Lang:       lang,
	}
	slog.Info("building book workload", "input", bcfg.Input, "output", bcfg.Output)
	if hybrid {
		// The ssg half owns the site's asset order. Reuse the plan it reported
		// instead of repeating the list here, so the book pages and the landing
		// page can never load stylesheets in different orders
		// (AGENTS.md § No Hardcoding).
		bcfg.ExtraCSS = versionedAssetURLs(planCSS, cfg.Version)
		bcfg.ExtraJS = versionedAssetURLs(planJS, cfg.Version)
	}
	created, err := book.Build(bcfg)
	if err != nil {
		return fail(err)
	}
	for _, path := range created {
		fmt.Println("created " + path)
	}
	return errs.ExitCodeSuccess
}

// globList resolves a repeatable/comma-separated glob flag over its
// krewire.yaml fallback. Unset everywhere returns nil (book defaults);
// an explicitly empty value disables that filter side.
func globList(fs *flag.FlagSet, name string, fallback []string) []string {
	raw := flagValue(fs, name)
	if raw == "" {
		return fallback
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	if out == nil {
		out = []string{}
	}
	return out
}

// themeFrom resolves the theme from the --theme flag and krewire.yaml,
// defaulting to auto. A mode of off disables the switcher.
func themeFrom(fs *flag.FlagSet, cfg *config.Theme) *ui.Theme {
	mode := strings.ToLower(strings.TrimSpace(flagValue(fs, "theme")))
	if mode == "" && cfg != nil {
		mode = strings.ToLower(strings.TrimSpace(cfg.Default))
	}
	if mode == "" {
		mode = "auto"
	}
	switch mode {
	case "off", "none", "disabled":
		return nil
	}
	t := &ui.Theme{Default: mode}
	if cfg != nil {
		t.Light = cfg.Light.UI(ui.DefaultLightPalette)
		t.Dark = cfg.Dark.UI(ui.DefaultDarkPalette)
	}
	return t
}

// bookThemeFrom resolves the book theme from the same flag/config as
// themeFrom but returns the book's local Theme type.
func bookThemeFrom(fs *flag.FlagSet, cfg *config.Theme) *book.Theme {
	mode := strings.ToLower(strings.TrimSpace(flagValue(fs, "theme")))
	if mode == "" && cfg != nil {
		mode = strings.ToLower(strings.TrimSpace(cfg.Default))
	}
	if mode == "" {
		mode = "auto"
	}
	switch mode {
	case "off", "none", "disabled":
		return nil
	}
	t := &book.Theme{Default: mode}
	if cfg != nil {
		t.Light = bookPalette(cfg.Light, book.DefaultLightPalette)
		t.Dark = bookPalette(cfg.Dark, book.DefaultDarkPalette)
	}
	return t
}

func bookPalette(p config.Palette, defaults book.Palette) book.Palette {
	out := defaults
	if p.Base1 != "" {
		out.Base1 = book.Color(p.Base1)
	}
	if p.Base1Content != "" {
		out.Base1Content = book.Color(p.Base1Content)
	}
	if p.Base2 != "" {
		out.Base2 = book.Color(p.Base2)
	}
	if p.Base2Content != "" {
		out.Base2Content = book.Color(p.Base2Content)
	}
	if p.Base3 != "" {
		out.Base3 = book.Color(p.Base3)
	}
	if p.Base3Content != "" {
		out.Base3Content = book.Color(p.Base3Content)
	}
	if p.Primary != "" {
		out.Primary = book.Color(p.Primary)
	}
	if p.PrimaryContent != "" {
		out.PrimaryContent = book.Color(p.PrimaryContent)
	}
	if p.Secondary != "" {
		out.Secondary = book.Color(p.Secondary)
	}
	if p.SecondaryContent != "" {
		out.SecondaryContent = book.Color(p.SecondaryContent)
	}
	if p.Accent != "" {
		out.Accent = book.Color(p.Accent)
	}
	if p.AccentContent != "" {
		out.AccentContent = book.Color(p.AccentContent)
	}
	if p.Ghost != "" {
		out.Ghost = book.Color(p.Ghost)
	}
	if p.GhostContent != "" {
		out.GhostContent = book.Color(p.GhostContent)
	}
	if p.Neutral != "" {
		out.Neutral = book.Color(p.Neutral)
	}
	if p.NeutralContent != "" {
		out.NeutralContent = book.Color(p.NeutralContent)
	}
	if p.Success != "" {
		out.Success = book.Color(p.Success)
	}
	if p.SuccessContent != "" {
		out.SuccessContent = book.Color(p.SuccessContent)
	}
	if p.Info != "" {
		out.Info = book.Color(p.Info)
	}
	if p.InfoContent != "" {
		out.InfoContent = book.Color(p.InfoContent)
	}
	if p.Warning != "" {
		out.Warning = book.Color(p.Warning)
	}
	if p.WarningContent != "" {
		out.WarningContent = book.Color(p.WarningContent)
	}
	if p.Error != "" {
		out.Error = book.Color(p.Error)
	}
	if p.ErrorContent != "" {
		out.ErrorContent = book.Color(p.ErrorContent)
	}
	return out
}

// navFromConfig converts krewire.yaml nav links into book links.
func navFromConfig(links []config.Link) []book.Link {
	if len(links) == 0 {
		return nil
	}
	out := make([]book.Link, 0, len(links))
	for _, l := range links {
		out = append(out, book.Link{Text: l.Text, URL: l.URL})
	}
	return out
}
