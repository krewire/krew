package commands

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	kiwver "github.com/krewire/krew/core/version"
	"github.com/krewire/krewire/errs"
	"github.com/krewire/krewire/version"
	"github.com/krewire/tui/term"
)

// RegisterSelfUpdate registers flags for `krew self-update`.
func RegisterSelfUpdate(fs *flag.FlagSet) {
	fs.Bool("check", false, "check for available updates without installing")
	fs.Bool("force", false, "reinstall the latest version even if already up to date")
	fs.Bool("dry-run", false, "simulate update without replacing the binary")
	fs.String("version", "", "update to a specific version (default: latest stable)")
}

// RunSelfUpdate checks for and installs the latest stable version of the krew toolchain.
func RunSelfUpdate(fs *flag.FlagSet) errs.ExitCode {
	checkOnly := flagValue(fs, "check") == "true"
	force := flagValue(fs, "force") == "true"
	dryRun := flagValue(fs, "dry-run") == "true"
	targetVersionReq := flagValue(fs, "version")

	tm := term.NewTerminal()
	bold := func(s string) string { return tm.Paint(s, term.ColorCyan, []term.Style{term.StyleBold}) }
	green := func(s string) string { return tm.Paint(s, term.ColorGreen, nil) }
	yellow := func(s string) string { return tm.Paint(s, term.ColorYellow, nil) }

	curVerStr := kiwver.VersionString()
	cleanCur := strings.TrimPrefix(curVerStr, "v")
	curVer, err := version.ParseVersion(cleanCur)
	if err != nil {
		curVer = kiwver.Version
	}

	fmt.Printf("%s Checking for updates (current version: %s) ...\n", bold("→"), green("v"+curVer.String()))

	ctx, cancel := context.WithTimeout(context.Background(), defaultHTTPTimeout)
	defer cancel()

	rel, err := fetchLatestRelease(ctx, defaultReleaseURL)
	if err != nil {
		slog.Warn("failed to fetch release from GitHub API", "error", err)
		is404 := strings.Contains(err.Error(), "HTTP 404")
		if checkOnly {
			if is404 {
				fmt.Printf("%s No newer releases found on GitHub; krew is running current version (v%s)\n", green("✓"), curVer.String())
				return errs.ExitCodeSuccess
			}
			fmt.Fprintf(os.Stderr, "error checking for updates: %v\n", err)
			return errs.ExitCodeFailure
		}
		if dryRun {
			if is404 {
				fmt.Printf("%s [dry-run] No remote GitHub release found for krew (current version: v%s)\n", yellow("!"), curVer.String())
				return errs.ExitCodeSuccess
			}
			fmt.Fprintf(os.Stderr, "error checking for updates: %v\n", err)
			return errs.ExitCodeFailure
		}
		// If actual update requested, fallback to Go toolchain
		return fallbackGoInstallUpdate(curVer, "latest", dryRun)
	}

	targetTag := rel.TagName
	if targetVersionReq != "" {
		targetTag = targetVersionReq
	}
	cleanTarget := strings.TrimPrefix(targetTag, "v")
	targetVer, err := version.ParseVersion(cleanTarget)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid release tag %q: %v\n", targetTag, err)
		return errs.ExitCodeFailure
	}

	cmp := targetVer.Compare(curVer)
	if cmp <= 0 && !force {
		fmt.Printf("%s %s (v%s)\n", green("✓"), "krew is already up to date", curVer.String())
		return errs.ExitCodeSuccess
	}

	if checkOnly {
		if cmp > 0 {
			fmt.Printf("%s A new stable release of krew is available: %s -> %s\n", yellow("!"), yellow("v"+curVer.String()), green("v"+targetVer.String()))
			fmt.Println("  Run 'krew self-update' to update your installation.")
		} else {
			fmt.Printf("%s krew is up to date (v%s)\n", green("✓"), curVer.String())
		}
		return errs.ExitCodeSuccess
	}

	if dryRun {
		fmt.Printf("%s Would update krew from v%s to v%s\n", yellow("[dry-run]"), curVer.String(), targetVer.String())
		return errs.ExitCodeSuccess
	}

	fmt.Printf("%s Updating krew toolchain: %s -> %s ...\n", bold("→"), yellow("v"+curVer.String()), green("v"+targetVer.String()))

	// Locate current executable
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot determine krew executable path: %v\n", err)
		return errs.ExitCodeFailure
	}
	if resolved, err := filepath.EvalSymlinks(execPath); err == nil {
		execPath = resolved
	}

	// Try downloading binary release asset first
	asset := findMatchingAsset(rel.Assets, targetVer.String(), runtime.GOOS, runtime.GOARCH)
	if asset != nil {
		slog.Info("downloading release asset", "asset", asset.Name, "url", asset.BrowserDownloadURL)
		fmt.Printf("  Downloading %s (%d bytes) ...\n", asset.Name, asset.Size)
		if err := downloadAndInstallAsset(ctx, asset.BrowserDownloadURL, asset.Name, execPath); err == nil {
			fmt.Printf("%s krew successfully updated to %s\n", green("✓"), green("v"+targetVer.String()))
			fmt.Printf("  Binary installed at: %s\n", execPath)
			return errs.ExitCodeSuccess
		} else {
			slog.Warn("asset installation failed, falling back to go install", "error", err)
			fmt.Fprintf(os.Stderr, "  Asset installation encountered an issue: %v. Trying Go toolchain fallback...\n", err)
		}
	}

	// Fallback to Go toolchain
	return fallbackGoInstallUpdate(curVer, "v"+targetVer.String(), dryRun)
}
