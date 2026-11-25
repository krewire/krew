package krew_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var compiledKrewBin string

func TestMain(m *testing.M) {
	// Compile krew binary once for the entire conformance package
	tmpDir, err := os.MkdirTemp("", "krew-conformance-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	binPath := filepath.Join(tmpDir, "krew")
	cmd := exec.Command("go", "build", "-trimpath", "-o", binPath, "../cli")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("failed to compile krew for conformance tests: " + err.Error() + "\n" + string(out))
	}
	compiledKrewBin = binPath

	os.Exit(m.Run())
}

func runKrew(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(compiledKrewBin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to execute krew: %v", err)
		}
	}
	return string(out), exitCode
}

func TestCommandContract_VersionAndHelp(t *testing.T) {
	t.Run("version command exits 0 and prints toolchain identity", func(t *testing.T) {
		out, code := runKrew(t, "", "version")
		if code != 0 {
			t.Errorf("expected exit code 0, got %d. Output: %s", code, out)
		}
		if !strings.Contains(out, "krew Unified Developer Toolchain") {
			t.Errorf("output missing toolchain name: %s", out)
		}
		if !strings.Contains(out, "Krewire Ecosystem") {
			t.Errorf("output missing ecosystem reference: %s", out)
		}
	})

	t.Run("general help exits 0 and lists core commands", func(t *testing.T) {
		out, code := runKrew(t, "", "--help")
		if code != 0 {
			t.Errorf("expected exit code 0, got %d. Output: %s", code, out)
		}
		if !strings.Contains(out, "CORE WORKFLOW") {
			t.Errorf("help missing CORE WORKFLOW section: %s", out)
		}
		if !strings.Contains(out, "new") || !strings.Contains(out, "dev") || !strings.Contains(out, "build") {
			t.Errorf("help missing core commands: %s", out)
		}
	})

	t.Run("unknown command exits with usage error and suggestion", func(t *testing.T) {
		out, code := runKrew(t, "", "nonexistent-cmd")
		if code != 2 {
			t.Errorf("expected usage exit code 2, got %d. Output: %s", code, out)
		}
		if !strings.Contains(out, "unknown command") && !strings.Contains(out, "USAGE") {
			t.Errorf("expected usage diagnostics in output: %s", out)
		}
	})
}

func TestCommandContract_NewAndScaffold_Hermetic(t *testing.T) {
	tempWorkspace := t.TempDir()

	t.Run("scaffolds clean site project non-interactively", func(t *testing.T) {
		out, code := runKrew(t, tempWorkspace, "new", "testsite", "--site", "--no-input")
		if code != 0 {
			t.Fatalf("krew new failed with code %d: %s", code, out)
		}

		projectDir := filepath.Join(tempWorkspace, "testsite")
		configPath := filepath.Join(projectDir, "krewire.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatalf("expected krewire.yaml to exist in %s", projectDir)
		}

		// Verify workspace inspection inside scaffolded project
		wsOut, wsCode := runKrew(t, projectDir, "ws", "info")
		if wsCode != 0 {
			t.Errorf("krew ws info failed with code %d: %s", wsCode, wsOut)
		}

		// Verify info inside scaffolded project
		infoOut, infoCode := runKrew(t, projectDir, "info")
		if infoCode != 0 {
			t.Errorf("krew info failed with code %d: %s", infoCode, infoOut)
		}
	})

	t.Run("fails closed when project directory already exists without force", func(t *testing.T) {
		// Attempt to scaffold into the same directory without force
		out, code := runKrew(t, tempWorkspace, "new", "testsite", "--site", "--no-input")
		if code == 0 {
			t.Errorf("expected non-zero exit code when destination exists, got 0. Output: %s", out)
		}
	})
}

func TestCommandContract_Doctor_Hermetic(t *testing.T) {
	out, code := runKrew(t, t.TempDir(), "doctor")
	// Doctor checks environment; must complete without panics or crashes
	if code != 0 && code != 1 {
		t.Errorf("doctor returned unexpected exit code %d: %s", code, out)
	}
}
