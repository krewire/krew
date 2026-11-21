package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/model"
	"github.com/krewire/krewire/pkg/sec"
)

// RegisterVerify registers flags for the krew verify command.
func RegisterVerify(fs *flag.FlagSet) {
	fs.Bool("json", false, "output verification record in machine-readable JSON format")
	fs.String("task", "active-task", "task identifier for evidence tracking")
	fs.String("agent", "krewire-agent", "agent identifier performing the verification")
	fs.String("ledger", "", "path to tamper-evident ledger JSON file to append evidence block")
	fs.Bool("verify-ledger", false, "verify cryptographic integrity of existing ledger file without running gates")
}

// RunVerify executes the canonical verification quality gates and constructs
// a tamper-evident, reproducible task verification evidence record.
func RunVerify(fs *flag.FlagSet) errs.ExitCode {
	dir, err := os.Getwd()
	if err != nil {
		return fail(err)
	}

	taskID := flagValue(fs, "task")
	if taskID == "" {
		taskID = "active-task"
	}
	agentID := flagValue(fs, "agent")
	if agentID == "" {
		agentID = "krewire-agent"
	}
	jsonOutput := flagBool(fs, "json")
	ledgerPath := flagValue(fs, "ledger")
	verifyLedgerOnly := flagBool(fs, "verify-ledger")

	if verifyLedgerOnly {
		if ledgerPath == "" {
			ledgerPath = filepath.Join(dir, ".krewire", "ledger.json")
		}
		l, err := model.LoadLedgerFile(ledgerPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: ledger verification failed: %v\n", err)
			return errs.ExitCodeFailure
		}
		fmt.Printf("✓ Verification ledger integrity verified (%d blocks, task: %s, file: %s)\n", len(l.Blocks), l.TaskID, ledgerPath)
		return errs.ExitCodeSuccess
	}

	// Determine git commit and working tree status
	gitCommit := "unversioned"
	workingTreeClean := false
	if out, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
		gitCommit = strings.TrimSpace(string(out))
	}
	if out, err := exec.Command("git", "status", "--porcelain").Output(); err == nil {
		workingTreeClean = len(strings.TrimSpace(string(out))) == 0
	}

	// Define canonical quality gates
	gateDefs := []struct {
		ID          string
		Type        model.QualityGateType
		Description string
		Command     string
		Args        []string
		Required    bool
	}{
		{
			ID:          "fmt",
			Type:        model.GateTypeFormat,
			Description: "Go code formatting check (gofmt -l .)",
			Command:     "gofmt",
			Args:        []string{"-l", "."},
			Required:    true,
		},
		{
			ID:          "vet",
			Type:        model.GateTypeStatic,
			Description: "Go static analysis and vet checks (go vet ./...)",
			Command:     "go",
			Args:        []string{"vet", "./..."},
			Required:    true,
		},
		{
			ID:          "test",
			Type:        model.GateTypeTest,
			Description: "Automated unit and integration test suite (go test ./...)",
			Command:     "go",
			Args:        []string{"test", "./..."},
			Required:    true,
		},
		{
			ID:          "build",
			Type:        model.GateTypeBuild,
			Description: "Pure-Go static compilation check (go build ./...)",
			Command:     "go",
			Args:        []string{"build", "./..."},
			Required:    true,
		},
	}

	// Check if layering_test.go exists in the directory
	if _, err := os.Stat(filepath.Join(dir, "layering_test.go")); err == nil {
		gateDefs = append(gateDefs, struct {
			ID          string
			Type        model.QualityGateType
			Description string
			Command     string
			Args        []string
			Required    bool
		}{
			ID:          "layering",
			Type:        model.GateTypeLayering,
			Description: "Downward-only layering architecture guard (layering_test.go)",
			Command:     "go",
			Args:        []string{"test", "-v", "./layering_test.go"},
			Required:    true,
		})
	}

	var evidences []model.VerificationEvidence

	for _, def := range gateDefs {
		cmdStr := def.Command + " " + strings.Join(def.Args, " ")
		ev := model.VerificationEvidence{
			GateID:      def.ID,
			Type:        def.Type,
			Description: def.Description,
			Command:     cmdStr,
			Required:    def.Required,
			Status:      model.GateAttempted,
		}

		start := time.Now()
		cmd := exec.Command(def.Command, def.Args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

		out, cmdErr := cmd.CombinedOutput()
		ev.DurationMs = time.Since(start).Milliseconds()
		rawOutput := strings.TrimSpace(string(out))

		// For gofmt, non-empty output indicates unformatted files
		if def.ID == "fmt" && len(rawOutput) > 0 {
			ev.Status = model.GateFailed
			ev.ExitCode = 1
			ev.Output = fmt.Sprintf("unformatted files detected:\n%s", rawOutput)
		} else if cmdErr != nil {
			ev.Status = model.GateFailed
			if exitErr, ok := cmdErr.(*exec.ExitError); ok {
				ev.ExitCode = exitErr.ExitCode()
			} else {
				ev.ExitCode = 1
			}
			ev.Output = sec.RedactSensitiveOutput(rawOutput)
		} else {
			ev.Status = model.GatePassed
			ev.ExitCode = 0
			ev.Output = sec.RedactSensitiveOutput(rawOutput)
		}

		ev.EvidenceHash = ev.ComputeHash(gitCommit)
		evidences = append(evidences, ev)
	}

	record := model.NewTaskVerificationRecord(taskID, agentID, gitCommit, workingTreeClean, evidences)

	if ledgerPath != "" {
		var l *model.VerificationLedger
		if _, err := os.Stat(ledgerPath); err == nil {
			if loaded, err := model.LoadLedgerFile(ledgerPath); err == nil {
				l = loaded
			}
		}
		if l == nil {
			l = model.NewVerificationLedger(taskID)
		}
		if _, err := l.AppendBlock(gitCommit, evidences); err == nil {
			_ = l.SaveFile(ledgerPath)
		}
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(record)
		if record.OverallVerified {
			return errs.ExitCodeSuccess
		}
		return errs.ExitCodeFailure
	}

	fmt.Printf("\n── Task Verification Evidence Record [%s] ─────────────\n", record.TaskID)
	fmt.Printf("  • Agent ID:           %s\n", record.AgentID)
	fmt.Printf("  • Git Commit:         %s\n", record.GitCommit)
	fmt.Printf("  • Working Tree Clean: %v\n", record.WorkingTreeClean)
	if ledgerPath != "" {
		fmt.Printf("  • Evidence Ledger:    Appended to %s\n", ledgerPath)
	}
	fmt.Printf("  • Verified Status:    %s\n\n", formatOverallVerified(record.OverallVerified))

	fmt.Printf("  [EVALUATED QUALITY GATES]\n")
	for _, g := range record.Gates {
		statusTag := formatGateStatus(g.Status)
		fmt.Printf("    %s %-10s (%dms) — %s\n", statusTag, g.GateID, g.DurationMs, g.Description)
		fmt.Printf("      • Command:  %s\n", g.Command)
		fmt.Printf("      • Exit:     %d\n", g.ExitCode)
		fmt.Printf("      • Hash:     %s\n", g.EvidenceHash)
		if g.Status == model.GateFailed && g.Output != "" {
			firstLine := strings.Split(g.Output, "\n")[0]
			fmt.Printf("      • Error:    %s\n", firstLine)
		}
	}

	fmt.Printf("\n  Summary: %s\n\n", record.Summary)

	if record.OverallVerified {
		return errs.ExitCodeSuccess
	}
	return errs.ExitCodeFailure
}

func formatOverallVerified(v bool) string {
	if v {
		return "✓ FULLY VERIFIED (All required quality gates passed)"
	}
	return "✗ NOT VERIFIED (One or more required quality gates failed or were skipped)"
}

func formatGateStatus(s model.VerificationStatus) string {
	switch s {
	case model.GatePassed:
		return "✓ [PASS]"
	case model.GateFailed:
		return "✗ [FAIL]"
	case model.GateSkipped:
		return "⚠ [SKIP]"
	case model.GateBlocked:
		return "⛔ [BLOCK]"
	default:
		return "ℹ [" + string(s) + "]"
	}
}
