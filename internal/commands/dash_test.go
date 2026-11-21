package commands

import (
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func runDashWithArgs(args ...string) (string, errs.ExitCode) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	fs := flag.NewFlagSet("dash", flag.ContinueOnError)
	RegisterDash(fs)
	_ = fs.Parse(args)

	code := RunDash(fs)

	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	out, _ := io.ReadAll(r)
	return string(out), code
}

func TestRunDash_Help(t *testing.T) {
	outHelp, codeHelp := runDashWithArgs("help")
	if codeHelp != errs.ExitCodeSuccess {
		t.Fatalf("expected ExitCodeSuccess for help, got %v", codeHelp)
	}
	if !strings.Contains(outHelp, "krew dash") {
		t.Errorf("expected usage output, got %q", outHelp)
	}
}

func TestRunDash_BareDryRun(t *testing.T) {
	out, code := runDashWithArgs("--dry-run")
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected ExitCodeSuccess for bare dry-run, got %v", code)
	}
	if !strings.Contains(out, "[dry-run] would start dev dashboard") {
		t.Errorf("expected server dry-run output, got %q", out)
	}
}

func TestRunDashboard_Alias(t *testing.T) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	fs := flag.NewFlagSet("dashboard", flag.ContinueOnError)
	RegisterDashboard(fs)
	_ = fs.Parse([]string{"--dry-run"})

	code := RunDashboard(fs)

	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	out, _ := io.ReadAll(r)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected ExitCodeSuccess for dashboard alias, got %v", code)
	}
	if !strings.Contains(string(out), "[dry-run] would start dev dashboard") {
		t.Errorf("expected server dry-run output, got %q", string(out))
	}
}

func TestRunDash_Start(t *testing.T) {
	out, code := runDashWithArgs("start", "--sandbox=go-1.27")
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected ExitCodeSuccess for start, got %v", code)
	}
	if !strings.Contains(out, "Isolated micro-container mounted") || !strings.Contains(out, "go-1.27") {
		t.Errorf("expected sandbox mount output, got %q", out)
	}
	if !strings.Contains(out, "PTY terminal attached") {
		t.Errorf("expected PTY terminal attached, got %q", out)
	}
	if !strings.Contains(out, "Autonomous ReAct Agent active") {
		t.Errorf("expected agent active output, got %q", out)
	}
}

func TestRunDash_StartDryRun(t *testing.T) {
	out, code := runDashWithArgs("start", "--sandbox=node-22", "--dry-run")
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected ExitCodeSuccess for dry-run, got %v", code)
	}
	if !strings.Contains(out, "[dry-run]") || !strings.Contains(out, "node-22") {
		t.Errorf("expected dry-run output, got %q", out)
	}
}

func TestRunDash_List(t *testing.T) {
	out, code := runDashWithArgs("list")
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected ExitCodeSuccess for list, got %v", code)
	}
	if !strings.Contains(out, "Active Krewire Dash Workspaces") {
		t.Errorf("expected workspace list output, got %q", out)
	}
}

func TestRunDash_Status(t *testing.T) {
	out, code := runDashWithArgs("status")
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected ExitCodeSuccess for status, got %v", code)
	}
	if !strings.Contains(out, "Krewire Dash Status") || !strings.Contains(out, "operational") {
		t.Errorf("expected status output, got %q", out)
	}
}

func TestRunDash_Console(t *testing.T) {
	out, code := runDashWithArgs("console", "--port=9000")
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected ExitCodeSuccess for console, got %v", code)
	}
	if !strings.Contains(out, "http://127.0.0.1:9000/console") {
		t.Errorf("expected console URL, got %q", out)
	}
}

func TestRunDash_Stop(t *testing.T) {
	out, code := runDashWithArgs("stop", "ws-sandbox-1")
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected ExitCodeSuccess for stop, got %v", code)
	}
	if !strings.Contains(out, "ws-sandbox-1") {
		t.Errorf("expected stop output, got %q", out)
	}
}

func TestRunDash_UnknownSubcommand(t *testing.T) {
	_, code := runDashWithArgs("invalid-subcommand")
	if code != errs.ExitCodeUsage {
		t.Fatalf("expected ExitCodeUsage for unknown subcommand, got %v", code)
	}
}
