package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/tui/term"
)

// AuthCredentials holds stored developer credentials for CLI operations.
type AuthCredentials struct {
	Endpoint  string    `json:"endpoint"`
	User      string    `json:"user"`
	Token     string    `json:"token"`
	TokenType string    `json:"token_type,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// DefaultAuthFilePath returns the canonical path to ~/.krew/auth.json,
// or respects the KREWIRE_AUTH_FILE environment variable override.
func DefaultAuthFilePath() (string, error) {
	if custom := os.Getenv("KREWIRE_AUTH_FILE"); custom != "" {
		return custom, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(home, ".krew", "auth.json"), nil
}

// LoadCredentials reads and decodes the stored credentials, or returns nil if absent.
func LoadCredentials() (*AuthCredentials, error) {
	path, err := DefaultAuthFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read credentials file %s: %w", path, err)
	}
	var creds AuthCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parse credentials file %s: %w", path, err)
	}
	return &creds, nil
}

// SaveCredentials writes credentials to ~/.krew/auth.json with 0600 permissions (S1: Secure).
func SaveCredentials(creds *AuthCredentials) error {
	path, err := DefaultAuthFilePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config directory %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("encode credentials: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

// RemoveCredentials deletes the stored credentials file.
func RemoveCredentials() error {
	path, err := DefaultAuthFilePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove credentials file %s: %w", path, err)
	}
	return nil
}

// RegisterAuth registers flags for the auth command group.
func RegisterAuth(fs *flag.FlagSet) {
	fs.String("token", "", "authentication API token or personal access token")
	fs.String("user", "", "user email or username for authentication")
	fs.String("email", "", "user email alias for --user")
	fs.String("password", "", "account password (non-interactive)")
	fs.String("endpoint", "", "custom identity/auth API endpoint URL")
	fs.Bool("browser", false, "launch browser for web-based interactive authentication")
}

// RunAuth dispatches the auth subcommands: login, logout, status, whoami.
func RunAuth(fs *flag.FlagSet) errs.ExitCode {
	sub := fs.Arg(0)
	switch sub {
	case "login":
		return runAuthLogin(fs)
	case "logout":
		return runAuthLogout(fs)
	case "status", "whoami":
		return runAuthStatus(fs)
	case "", "help", "-h", "--help":
		return authHelp()
	default:
		fmt.Fprintf(os.Stderr, "krew: unknown auth sub-command %q (supported: login, logout, status)\n", sub)
		return authHelp()
	}
}

func runAuthLogin(fs *flag.FlagSet) errs.ExitCode {
	var subFs *flag.FlagSet
	if len(fs.Args()) > 1 {
		subFs = flag.NewFlagSet("login", flag.ContinueOnError)
		RegisterAuth(subFs)
		_ = subFs.Parse(fs.Args()[1:])
	}

	token := authFlagValue(fs, subFs, "token")
	user := authFlagValue(fs, subFs, "user")
	if user == "" {
		user = authFlagValue(fs, subFs, "email")
	}
	password := authFlagValue(fs, subFs, "password")
	endpoint := authFlagValue(fs, subFs, "endpoint")
	if endpoint == "" {
		cfg, _ := config.Load(".")
		if cfg != nil {
			endpoint = cfg.AuthURL()
		} else {
			endpoint = config.DefaultAuthAPIURL
		}
	}
	endpoint = strings.TrimRight(endpoint, "/")

	// Direct token provided
	if token != "" {
		creds := &AuthCredentials{
			Endpoint:  endpoint,
			User:      user,
			Token:     token,
			TokenType: "Bearer",
			CreatedAt: time.Now().UTC(),
		}
		if creds.User == "" {
			creds.User = "token-user"
		}
		if err := SaveCredentials(creds); err != nil {
			return fail(err)
		}
		tm := term.NewTerminal()
		fmt.Printf("%s Successfully authenticated to %s\n", tm.Paint("✔", term.ColorGreen, nil), endpoint)
		fmt.Printf("  Identity: %s\n", creds.User)
		fmt.Printf("  Token stored in ~/.krew/auth.json (mode 0600)\n")
		return errs.ExitCodeSuccess
	}

	// Username/email and password provided
	if user != "" && password != "" {
		creds, err := loginWithPassword(endpoint, user, password)
		if err != nil {
			return fail(fmt.Errorf("authentication failed: %w", err))
		}
		if err := SaveCredentials(creds); err != nil {
			return fail(err)
		}
		tm := term.NewTerminal()
		fmt.Printf("%s Successfully authenticated as %s\n", tm.Paint("✔", term.ColorGreen, nil), user)
		fmt.Printf("  Endpoint: %s\n", endpoint)
		fmt.Printf("  Token stored in ~/.krew/auth.json (mode 0600)\n")
		return errs.ExitCodeSuccess
	}

	fmt.Fprintln(os.Stderr, "error: authentication credentials required")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  krew auth login --token <token>")
	fmt.Fprintln(os.Stderr, "  krew auth login --user <email> --password <password>")
	return errs.ExitCodeUsage
}

func loginWithPassword(endpoint, user, password string) (*AuthCredentials, error) {
	loginURL := endpoint + "/api/auth/login"
	reqBody, err := json.Marshal(map[string]string{
		"email":    user,
		"password": password,
	})
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(loginURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", loginURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("endpoint returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var res struct {
		Token string `json:"token"`
		User  struct {
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	email := res.User.Email
	if email == "" {
		email = user
	}

	return &AuthCredentials{
		Endpoint:  endpoint,
		User:      email,
		Token:     res.Token,
		TokenType: "Bearer",
		CreatedAt: time.Now().UTC(),
	}, nil
}

func runAuthLogout(fs *flag.FlagSet) errs.ExitCode {
	creds, err := LoadCredentials()
	if err != nil {
		return fail(err)
	}
	if creds == nil {
		fmt.Println("You are not currently logged in.")
		return errs.ExitCodeSuccess
	}

	endpoint := creds.Endpoint
	user := creds.User

	if err := RemoveCredentials(); err != nil {
		return fail(err)
	}

	tm := term.NewTerminal()
	fmt.Printf("%s Successfully logged out %s\n", tm.Paint("✔", term.ColorGreen, nil), user)
	fmt.Printf("  Removed credentials for %s\n", endpoint)
	return errs.ExitCodeSuccess
}

func runAuthStatus(fs *flag.FlagSet) errs.ExitCode {
	creds, err := LoadCredentials()
	if err != nil {
		return fail(err)
	}
	if creds == nil {
		fmt.Println("Not logged in. Run 'krew auth login' to authenticate.")
		return errs.ExitCodeSuccess
	}

	tm := term.NewTerminal()
	fmt.Printf("%s Authenticated to %s\n", tm.Paint("✔", term.ColorGreen, nil), creds.Endpoint)
	fmt.Printf("  Identity:     %s\n", creds.User)
	maskedToken := creds.Token
	if len(maskedToken) > 8 {
		maskedToken = maskedToken[:4] + strings.Repeat("*", len(maskedToken)-8) + maskedToken[len(maskedToken)-4:]
	} else if len(maskedToken) > 0 {
		maskedToken = strings.Repeat("*", len(maskedToken))
	}
	fmt.Printf("  Token:        %s\n", maskedToken)
	fmt.Printf("  Logged in at: %s\n", creds.CreatedAt.Format(time.RFC3339))
	return errs.ExitCodeSuccess
}

func authHelp() errs.ExitCode {
	fmt.Fprint(os.Stdout, `krew auth — manage developer authentication sessions

Usage:
  krew auth login [options]     authenticate with Krewire identity platform
  krew auth logout              remove active credentials and logout
  krew auth status              display current authentication status

Options:
  --token <token>               API token or Personal Access Token (PAT)
  --user, --email <email>       user email or username
  --password <password>         account password
  --endpoint <url>              custom auth service endpoint URL
  --browser                     open browser for web-based interactive authentication

Examples:
  krew auth login --token kiw_tok_abc123
  krew auth login --user dev@example.com --password secret
  krew auth logout
  krew auth status
`)
	return errs.ExitCodeSuccess
}

func authFlagValue(fs, subFs *flag.FlagSet, name string) string {
	if subFs != nil {
		if v := flagValue(subFs, name); v != "" {
			return v
		}
	}
	return flagValue(fs, name)
}
