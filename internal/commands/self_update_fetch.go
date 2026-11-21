package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	kiwver "github.com/krewire/krew/internal/version"
	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/term"
	"github.com/krewire/krewire/pkg/version"
)

var (
	// defaultReleaseURL is the GitHub API endpoint for the latest release.
	// Overridden in tests to point to a mock HTTP server.
	defaultReleaseURL = "https://api.github.com/repos/krewire/krewire/releases/latest"
	// defaultHTTPTimeout is the network timeout for checking releases.
	defaultHTTPTimeout = 10 * time.Second
)

type githubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	PublishedAt string        `json:"published_at"`
	HTMLURL     string        `json:"html_url"`
	Assets      []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

func fetchLatestRelease(ctx context.Context, url string) (*githubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "kiw-self-update/"+kiwver.VersionString())

	client := &http.Client{Timeout: defaultHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("GitHub API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decoding release JSON: %w", err)
	}
	if rel.TagName == "" {
		return nil, fmt.Errorf("release has no tag_name")
	}
	return &rel, nil
}

func findMatchingAsset(assets []githubAsset, ver, targetOS, targetArch string) *githubAsset {
	for i := range assets {
		name := strings.ToLower(assets[i].Name)
		if strings.Contains(name, targetOS) && strings.Contains(name, targetArch) {
			if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip") {
				return &assets[i]
			}
		}
	}
	return nil
}

func fallbackGoInstallUpdate(curVer version.Version, targetTag string, dryRun bool) errs.ExitCode {
	if _, err := exec.LookPath("go"); err != nil {
		fmt.Fprintf(os.Stderr, "Go toolchain not found in PATH and GitHub releases unavailable.\n")
		fmt.Fprintf(os.Stderr, "Install or update manually: curl -fsSL https://krewire.com/scripts/install.sh | sh\n")
		return errs.ExitCodeFailure
	}

	pkg := "github.com/krewire/krew/cmd/krew@" + targetTag
	if dryRun {
		fmt.Printf("[dry-run] Would run: go install %s\n", pkg)
		return errs.ExitCodeSuccess
	}

	fmt.Printf("  Running: go install %s ...\n", pkg)
	cmd := exec.Command("go", "install", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "go install failed: %v\n", err)
		return errs.ExitCodeFailure
	}

	tm := term.NewTerminal()
	green := func(s string) string { return tm.Paint(s, term.ColorGreen, nil) }
	fmt.Printf("%s krew successfully updated via go install (%s)\n", green("✓"), targetTag)
	return errs.ExitCodeSuccess
}
