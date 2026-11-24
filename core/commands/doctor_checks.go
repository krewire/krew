package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/krewire/krew/core/config"
	kiwver "github.com/krewire/krew/core/version"
)

// DiagnoseEnvironment runs all doctor checks and produces a DoctorReport.
func DiagnoseEnvironment(cwd string) DoctorReport {
	report := DoctorReport{
		Environment: map[string]string{
			"os":           runtime.GOOS,
			"arch":         runtime.GOARCH,
			"cpus":         fmt.Sprintf("%d", runtime.NumCPU()),
			"go_runtime":   runtime.Version(),
			"krew_version": kiwver.VersionString(),
			"kiw_version":  kiwver.VersionString(),
			"working_dir":  cwd,
		},
		Passed: true,
	}

	checkGoToolchain(&report)
	checkGitVCS(&report)
	checkGoWorkspace(cwd, &report)
	checkProjectConfig(cwd, &report)
	summarizeReport(&report)

	return report
}

func checkGoToolchain(report *DoctorReport) {
	goPath, err := exec.LookPath("go")
	if err != nil {
		report.Checks = append(report.Checks, DoctorCheck{
			Category:    "Toolchain",
			Name:        "Go Binary",
			Status:      StatusFail,
			Message:     "Go compiler ('go') not found in PATH",
			Remediation: "Install Go 1.27.1+ from https://go.dev/dl/ and add it to your PATH.",
		})
		report.Passed = false
		return
	}

	out, err := exec.Command(goPath, "version").CombinedOutput()
	verStr := strings.TrimSpace(string(out))
	if err != nil {
		report.Checks = append(report.Checks, DoctorCheck{
			Category:    "Toolchain",
			Name:        "Go Binary",
			Status:      StatusFail,
			Message:     fmt.Sprintf("Failed to query Go version: %v", err),
			Remediation: "Verify your Go installation.",
		})
		report.Passed = false
		return
	}

	report.Environment["go_compiler"] = verStr
	re := regexp.MustCompile(`go(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(verStr)
	if len(matches) >= 3 {
		major := matches[1]
		minor := matches[2]
		if major == "1" && (minor < "27") {
			report.Checks = append(report.Checks, DoctorCheck{
				Category:    "Toolchain",
				Name:        "Go Version",
				Status:      StatusWarn,
				Message:     fmt.Sprintf("Found %s; Krewire targets Go 1.27.1+", verStr),
				Remediation: "Upgrade Go to 1.27.1+ for full workspace and module compatibility.",
			})
			return
		}
		report.Checks = append(report.Checks, DoctorCheck{
			Category: "Toolchain",
			Name:     "Go Version",
			Status:   StatusPass,
			Message:  fmt.Sprintf("%s (valid toolchain)", verStr),
		})
		return
	}

	report.Checks = append(report.Checks, DoctorCheck{
		Category: "Toolchain",
		Name:     "Go Version",
		Status:   StatusPass,
		Message:  verStr,
	})
}

func checkGitVCS(report *DoctorReport) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		report.Checks = append(report.Checks, DoctorCheck{
			Category:    "Version Control",
			Name:        "Git Binary",
			Status:      StatusFail,
			Message:     "Git executable not found in PATH",
			Remediation: "Install Git from https://git-scm.com/.",
		})
		report.Passed = false
		return
	}

	out, err := exec.Command(gitPath, "version").CombinedOutput()
	gitVer := strings.TrimSpace(string(out))
	if err != nil {
		report.Checks = append(report.Checks, DoctorCheck{
			Category:    "Version Control",
			Name:        "Git Binary",
			Status:      StatusWarn,
			Message:     "Git binary found but failed to report version",
			Remediation: "Check Git executable permissions.",
		})
	} else {
		report.Checks = append(report.Checks, DoctorCheck{
			Category: "Version Control",
			Name:     "Git Binary",
			Status:   StatusPass,
			Message:  gitVer,
		})
	}

	// Git user identity check
	userName, _ := exec.Command(gitPath, "config", "user.name").CombinedOutput()
	userEmail, _ := exec.Command(gitPath, "config", "user.email").CombinedOutput()
	nameStr := strings.TrimSpace(string(userName))
	emailStr := strings.TrimSpace(string(userEmail))

	if nameStr == "" || emailStr == "" {
		report.Checks = append(report.Checks, DoctorCheck{
			Category:    "Version Control",
			Name:        "Git Identity",
			Status:      StatusWarn,
			Message:     "Git user.name or user.email is not set",
			Remediation: "Set global identity: git config --global user.name \"Name\" && git config --global user.email \"email@domain\"",
		})
	} else {
		report.Checks = append(report.Checks, DoctorCheck{
			Category: "Version Control",
			Name:     "Git Identity",
			Status:   StatusPass,
			Message:  fmt.Sprintf("%s <%s>", nameStr, emailStr),
		})
	}
}

func checkGoWorkspace(cwd string, report *DoctorReport) {
	goWorkPath := findGoWork(cwd)
	if goWorkPath == "" {
		report.Checks = append(report.Checks, DoctorCheck{
			Category: "Workspace",
			Name:     "Go Workspace",
			Status:   StatusPass,
			Message:  "Standalone module mode (no parent go.work)",
		})
		return
	}

	content, err := os.ReadFile(goWorkPath)
	if err != nil {
		report.Checks = append(report.Checks, DoctorCheck{
			Category:    "Workspace",
			Name:        "Go Workspace",
			Status:      StatusWarn,
			Message:     fmt.Sprintf("Found go.work at %s but unable to read: %v", goWorkPath, err),
			Remediation: "Verify file permissions on go.work.",
		})
		return
	}

	modCount := strings.Count(string(content), "./")
	report.Checks = append(report.Checks, DoctorCheck{
		Category: "Workspace",
		Name:     "Go Workspace",
		Status:   StatusPass,
		Message:  fmt.Sprintf("Active workspace at %s (~%d modules registered)", filepath.Dir(goWorkPath), modCount),
	})
}

func checkProjectConfig(cwd string, report *DoctorReport) {
	cfgPath := filepath.Join(cwd, "krewire.yaml")
	if _, err := os.Stat(cfgPath); err != nil {
		report.Checks = append(report.Checks, DoctorCheck{
			Category: "Project",
			Name:     "Configuration",
			Status:   StatusPass,
			Message:  "No krewire.yaml in current directory (run 'krew init' or 'krew new' to start one)",
		})
		return
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		report.Checks = append(report.Checks, DoctorCheck{
			Category:    "Project",
			Name:        "Configuration",
			Status:      StatusFail,
			Message:     fmt.Sprintf("krewire.yaml has syntax/schema errors: %v", err),
			Remediation: "Review and fix YAML syntax in krewire.yaml.",
		})
		report.Passed = false
		return
	}

	if _, issues, verr := validateProjectConfig(cwd); verr == nil && len(issues) > 0 {
		status, remediation := StatusWarn, "Fix or remove the reported keys; run 'krew config' for details."
		if config.HasErrors(issues) {
			status = StatusFail
			report.Passed = false
		}
		msgs := make([]string, 0, len(issues))
		for _, i := range issues {
			msgs = append(msgs, "krewire.yaml:"+i.String())
		}
		report.Checks = append(report.Checks, DoctorCheck{
			Category:    "Project",
			Name:        "Configuration",
			Status:      status,
			Message:     strings.Join(msgs, "; "),
			Remediation: remediation,
		})
		return
	}

	pName := cfg.Project.Name
	if pName == "" {
		pName = filepath.Base(cwd)
	}
	report.Checks = append(report.Checks, DoctorCheck{
		Category: "Project",
		Name:     "Configuration",
		Status:   StatusPass,
		Message:  fmt.Sprintf("Valid krewire.yaml (project: %q)", pName),
	})
}

func summarizeReport(report *DoctorReport) {
	passCount, warnCount, failCount := 0, 0, 0
	for _, c := range report.Checks {
		switch c.Status {
		case StatusPass:
			passCount++
		case StatusWarn:
			warnCount++
		case StatusFail:
			failCount++
		}
	}

	if failCount > 0 {
		report.Passed = false
		report.Summary = fmt.Sprintf("%d check(s) failed, %d warning(s), %d passed", failCount, warnCount, passCount)
	} else if warnCount > 0 {
		report.Summary = fmt.Sprintf("All required checks passed with %d warning(s), %d passed", warnCount, passCount)
	} else {
		report.Summary = fmt.Sprintf("All %d checks passed cleanly", passCount)
	}
}

func findGoWork(startDir string) string {
	dir := startDir
	for {
		candidate := filepath.Join(dir, "go.work")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
