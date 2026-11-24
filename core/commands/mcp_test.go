package commands

import (
	"flag"
	"testing"
)

func TestIsLoopback(t *testing.T) {
	testCases := []struct {
		host     string
		loopback bool
	}{
		{"127.0.0.1", true},
		{"127.0.0.1:4000", true},
		{"localhost", true},
		{"localhost:8080", true},
		{"::1", true},
		{"[::1]:4000", true},
		{"127.0.0.2", true},
		{"", true}, // default is loopback
		{"0.0.0.0", false},
		{"0.0.0.0:4000", false},
		{"::", false},
		{"192.168.1.100", false},
		{"10.0.0.1", false},
		{"attacker.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.host, func(t *testing.T) {
			got := isLoopback(tc.host)
			if got != tc.loopback {
				t.Errorf("isLoopback(%q) = %v, expected %v", tc.host, got, tc.loopback)
			}
		})
	}
}

func TestRegisterMCP_Flags(t *testing.T) {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	RegisterMCP(fs)

	expectedFlags := []string{
		"stdio",
		"port",
		"host",
		"auth-token",
		"token",
		"allowed-origins",
		"allowed-hosts",
	}

	for _, name := range expectedFlags {
		if fs.Lookup(name) == nil {
			t.Errorf("expected flag %q to be registered", name)
		}
	}
}

func TestRunMCP_NonLoopback_RefusedWithoutAuth(t *testing.T) {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	RegisterMCP(fs)
	_ = fs.Set("port", "4999")
	_ = fs.Set("host", "0.0.0.0")

	// Running with non-loopback and no token should fail immediately
	code := RunMCP(fs)
	if code == 0 {
		t.Errorf("expected RunMCP to fail when binding to 0.0.0.0 without auth token, got exit code %d", code)
	}
}

func TestRunMCP_ConflictingFlags_Refused(t *testing.T) {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	RegisterMCP(fs)
	_ = fs.Set("stdio", "true")
	_ = fs.Set("port", "4000")

	code := RunMCP(fs)
	if code == 0 {
		t.Errorf("expected RunMCP to fail when both --stdio and --port are set, got exit code %d", code)
	}
}
