package commands

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func TestDiagnoseEnvironment(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	report := DiagnoseEnvironment(cwd)
	if report.Environment["os"] == "" {
		t.Errorf("expected OS in report environment, got empty")
	}
	if report.Environment["krew_version"] == "" {
		t.Errorf("expected kiw_version in report environment, got empty")
	}

	if len(report.Checks) == 0 {
		t.Errorf("expected at least one doctor check, got none")
	}

	foundGo := false
	for _, c := range report.Checks {
		if c.Category == "Toolchain" && c.Name == "Go Version" {
			foundGo = true
			if c.Status == StatusFail {
				t.Errorf("expected Go Version to pass, got fail: %s", c.Message)
			}
		}
	}

	if !foundGo {
		t.Errorf("expected Go Version check to be executed")
	}
}

func TestRunDoctorJSON(t *testing.T) {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	RegisterDoctor(fs)

	doctorFlagJSON = true
	defer func() { doctorFlagJSON = false }()

	code := RunDoctor(fs)
	if code != errs.ExitCodeSuccess {
		t.Errorf("expected exit code success (0), got %v", code)
	}
}

func TestFindGoWork(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	// We are inside krewire/tools/kiw/internal/commands, parent root has go.work
	found := findGoWork(cwd)
	if found == "" {
		t.Logf("no go.work found from %s (acceptable if standalone)", cwd)
	} else if filepath.Base(found) != "go.work" {
		t.Errorf("expected found file to be go.work, got %s", found)
	}
}
