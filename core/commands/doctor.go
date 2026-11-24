package commands

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/krewire/krewire/errs"
)

var (
	doctorFlagJSON bool
)

// RegisterDoctor configures flags for the `krew doctor` command.
func RegisterDoctor(fs *flag.FlagSet) {
	fs.BoolVar(&doctorFlagJSON, "json", false, "output diagnostic report in structured JSON format")
}

// CheckStatus represents the outcome of an individual diagnostic check.
type CheckStatus string

const (
	StatusPass CheckStatus = "pass"
	StatusWarn CheckStatus = "warn"
	StatusFail CheckStatus = "fail"
)

// DoctorCheck represents a single diagnostic evaluation.
type DoctorCheck struct {
	Category    string      `json:"category"`
	Name        string      `json:"name"`
	Status      CheckStatus `json:"status"`
	Message     string      `json:"message"`
	Remediation string      `json:"remediation,omitempty"`
}

// DoctorReport holds the complete diagnostic evaluation of the environment.
type DoctorReport struct {
	Timestamp   string            `json:"timestamp"`
	Environment map[string]string `json:"environment"`
	Checks      []DoctorCheck     `json:"checks"`
	Passed      bool              `json:"passed"`
	Summary     string            `json:"summary"`
}

// RunDoctor performs system, toolchain, and workspace diagnostics.
func RunDoctor(fs *flag.FlagSet) errs.ExitCode {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	report := DiagnoseEnvironment(cwd)

	if doctorFlagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return fail(err)
		}
		if !report.Passed {
			return errs.ExitCodeFailure
		}
		return errs.ExitCodeSuccess
	}

	renderDoctorConsole(report)

	if !report.Passed {
		return errs.ExitCodeFailure
	}
	return errs.ExitCodeSuccess
}
