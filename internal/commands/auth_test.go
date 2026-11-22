package commands_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krew/internal/commands"
	"github.com/krewire/krewire/pkg/errs"
)

func TestAuthLoginWithTokenAndLogout(t *testing.T) {
	tmpDir := t.TempDir()
	authFile := filepath.Join(tmpDir, ".krew", "auth.json")
	t.Setenv("KREWIRE_AUTH_FILE", authFile)

	// 1. Initial status: not logged in
	fsStatus := flag.NewFlagSet("auth", flag.ContinueOnError)
	commands.RegisterAuth(fsStatus)
	_ = fsStatus.Parse([]string{"status"})
	code := commands.RunAuth(fsStatus)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected status to exit 0 when not logged in, got %v", code)
	}

	// 2. Login with token
	fsLogin := flag.NewFlagSet("auth", flag.ContinueOnError)
	commands.RegisterAuth(fsLogin)
	_ = fsLogin.Parse([]string{"login", "--token", "kiw_tok_test999", "--user", "alice@example.com", "--endpoint", "https://auth.example.com"})
	code = commands.RunAuth(fsLogin)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected login to exit 0, got %v", code)
	}

	// Verify file was written with 0600 permissions
	fi, err := os.Stat(authFile)
	if err != nil {
		t.Fatalf("expected auth file to exist: %v", err)
	}
	perm := fi.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected auth file perm 0600, got %o", perm)
	}

	// Load credentials and assert contents
	creds, err := commands.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials failed: %v", err)
	}
	if creds == nil {
		t.Fatal("expected creds not nil")
	}
	if creds.Token != "kiw_tok_test999" {
		t.Errorf("expected token kiw_tok_test999, got %s", creds.Token)
	}
	if creds.User != "alice@example.com" {
		t.Errorf("expected user alice@example.com, got %s", creds.User)
	}
	if creds.Endpoint != "https://auth.example.com" {
		t.Errorf("expected endpoint https://auth.example.com, got %s", creds.Endpoint)
	}

	// 3. Status when logged in
	fsStatus2 := flag.NewFlagSet("auth", flag.ContinueOnError)
	commands.RegisterAuth(fsStatus2)
	_ = fsStatus2.Parse([]string{"status"})
	code = commands.RunAuth(fsStatus2)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected status to exit 0 when logged in, got %v", code)
	}

	// 4. Logout
	fsLogout := flag.NewFlagSet("auth", flag.ContinueOnError)
	commands.RegisterAuth(fsLogout)
	_ = fsLogout.Parse([]string{"logout"})
	code = commands.RunAuth(fsLogout)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected logout to exit 0, got %v", code)
	}

	// Verify file was removed
	if _, err := os.Stat(authFile); !os.IsNotExist(err) {
		t.Errorf("expected auth file to be removed, but still exists: %v", err)
	}

	// 5. Logout again (should succeed gracefully)
	fsLogout2 := flag.NewFlagSet("auth", flag.ContinueOnError)
	commands.RegisterAuth(fsLogout2)
	_ = fsLogout2.Parse([]string{"logout"})
	code = commands.RunAuth(fsLogout2)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected redundant logout to exit 0, got %v", code)
	}
}

func TestAuthLoginMissingCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	authFile := filepath.Join(tmpDir, "auth.json")
	t.Setenv("KREWIRE_AUTH_FILE", authFile)

	fs := flag.NewFlagSet("auth", flag.ContinueOnError)
	commands.RegisterAuth(fs)
	_ = fs.Parse([]string{"login"})
	code := commands.RunAuth(fs)
	if code != errs.ExitCodeUsage {
		t.Fatalf("expected login without credentials to exit with ExitCodeUsage (2), got %v", code)
	}
}

func TestAuthHelp(t *testing.T) {
	fs := flag.NewFlagSet("auth", flag.ContinueOnError)
	commands.RegisterAuth(fs)
	_ = fs.Parse([]string{"help"})
	code := commands.RunAuth(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected help to exit 0, got %v", code)
	}
}
