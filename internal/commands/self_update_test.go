package commands

import (
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/version"
)

func TestRegisterSelfUpdate(t *testing.T) {
	fs := flag.NewFlagSet("self-update", flag.ContinueOnError)
	RegisterSelfUpdate(fs)

	if fs.Lookup("check") == nil {
		t.Error("expected --check flag to be registered")
	}
	if fs.Lookup("force") == nil {
		t.Error("expected --force flag to be registered")
	}
	if fs.Lookup("dry-run") == nil {
		t.Error("expected --dry-run flag to be registered")
	}
	if fs.Lookup("version") == nil {
		t.Error("expected --version flag to be registered")
	}
}

func TestRunSelfUpdateCheckUpToDate(t *testing.T) {
	// Mock release equal to or older than current version
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := githubRelease{
			TagName: "v0.1.0",
			Name:    "v0.1.0",
		}
		_ = json.NewEncoder(w).Encode(rel)
	}))
	defer server.Close()

	origURL := defaultReleaseURL
	defaultReleaseURL = server.URL
	defer func() { defaultReleaseURL = origURL }()

	fs := flag.NewFlagSet("self-update", flag.ContinueOnError)
	RegisterSelfUpdate(fs)
	_ = fs.Parse([]string{"--check"})

	code := RunSelfUpdate(fs)
	if code != errs.ExitCodeSuccess {
		t.Errorf("got exit code %d, want %d", code, errs.ExitCodeSuccess)
	}
}

func TestRunSelfUpdateCheckNewerVersionAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := githubRelease{
			TagName: "v99.0.0",
			Name:    "v99.0.0",
		}
		_ = json.NewEncoder(w).Encode(rel)
	}))
	defer server.Close()

	origURL := defaultReleaseURL
	defaultReleaseURL = server.URL
	defer func() { defaultReleaseURL = origURL }()

	fs := flag.NewFlagSet("self-update", flag.ContinueOnError)
	RegisterSelfUpdate(fs)
	_ = fs.Parse([]string{"--check"})

	code := RunSelfUpdate(fs)
	if code != errs.ExitCodeSuccess {
		t.Errorf("got exit code %d, want %d", code, errs.ExitCodeSuccess)
	}
}

func TestRunSelfUpdateDryRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := githubRelease{
			TagName: "v99.0.0",
			Name:    "v99.0.0",
		}
		_ = json.NewEncoder(w).Encode(rel)
	}))
	defer server.Close()

	origURL := defaultReleaseURL
	defaultReleaseURL = server.URL
	defer func() { defaultReleaseURL = origURL }()

	fs := flag.NewFlagSet("self-update", flag.ContinueOnError)
	RegisterSelfUpdate(fs)
	_ = fs.Parse([]string{"--dry-run"})

	code := RunSelfUpdate(fs)
	if code != errs.ExitCodeSuccess {
		t.Errorf("got exit code %d, want %d", code, errs.ExitCodeSuccess)
	}
}

func TestFindMatchingAsset(t *testing.T) {
	assets := []githubAsset{
		{Name: "kiw_0.3.0_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux_amd64.tar.gz", Size: 1024},
		{Name: "kiw_0.3.0_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin_arm64.tar.gz", Size: 2048},
		{Name: "kiw_0.3.0_windows_amd64.zip", BrowserDownloadURL: "https://example.com/windows_amd64.zip", Size: 4096},
	}

	match := findMatchingAsset(assets, "0.3.0", "linux", "amd64")
	if match == nil || match.Name != "kiw_0.3.0_linux_amd64.tar.gz" {
		t.Errorf("expected linux_amd64 match, got %v", match)
	}

	matchDarwin := findMatchingAsset(assets, "0.3.0", "darwin", "arm64")
	if matchDarwin == nil || matchDarwin.Name != "kiw_0.3.0_darwin_arm64.tar.gz" {
		t.Errorf("expected darwin_arm64 match, got %v", matchDarwin)
	}

	matchNone := findMatchingAsset(assets, "0.3.0", "freebsd", "riscv64")
	if matchNone != nil {
		t.Errorf("expected nil for unsupported os/arch, got %v", matchNone)
	}
}

func TestVersionCompareLogic(t *testing.T) {
	v1 := version.MustParseVersion("0.2.0")
	v2 := version.MustParseVersion("0.2.1")
	v3 := version.MustParseVersion("0.3.0")

	if v2.Compare(v1) <= 0 {
		t.Error("v0.2.1 should be greater than v0.2.0")
	}
	if v3.Compare(v2) <= 0 {
		t.Error("v0.3.0 should be greater than v0.2.1")
	}
	if v1.Compare(v1) != 0 {
		t.Error("v0.2.0 should be equal to v0.2.0")
	}
}
