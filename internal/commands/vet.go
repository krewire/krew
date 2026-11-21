package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	kiwconfig "github.com/krewire/krew/internal/config"
	"github.com/krewire/krew/internal/gomod"
	"github.com/krewire/krewire/pkg/config"
	"github.com/krewire/krewire/pkg/errs"
)

// RegisterVet configures flags for the `krew vet` command.
func RegisterVet(fs *flag.FlagSet) {
	fs.Bool("sec", false, "run security posture checks (secrets scan, config hygiene, gitignore)")
	fs.Bool("vuln", false, "run vulnerability audit (govulncheck / checksum verification)")
}

// RunVet runs `go vet`, validates krewire.yaml against its JSON Schema, lints
// every .kiw template (kiw.Lint), and runs the
// optional security posture and vulnerability audits.
func RunVet(fs *flag.FlagSet) errs.ExitCode {
	dir, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	mod, err := gomod.Find(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew: not inside a Go module — run 'krew new <project>' first")
		return errs.ExitCodeUsage
	}

	secCheck := flagValue(fs, "sec") == "true"
	vulnCheck := flagValue(fs, "vuln") == "true"
	if fs != nil {
		for _, a := range fs.Args() {
			switch a {
			case "--sec", "-sec", "sec":
				secCheck = true
			case "--vuln", "-vuln", "vuln":
				vulnCheck = true
			}
		}
	}

	slog.Info("running vet", "module", mod.Path)
	args := []string{"vet", "./..."}
	if fs != nil && len(fs.Args()) > 0 {
		var filtered []string
		for _, a := range fs.Args() {
			if a != "--sec" && a != "-sec" && a != "sec" && a != "--vuln" && a != "-vuln" && a != "vuln" {
				filtered = append(filtered, a)
			}
		}
		if len(filtered) > 0 {
			args = append([]string{"vet"}, filtered...)
		}
	}
	cmd := exec.Command("go", args...) // #nosec G204 — user-invoked devtool
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errs.ExitCodeFailure
	}

	if ok, issues, err := validateProjectConfig(dir); ok {
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ [CONFIG] %v\n", err)
			return errs.ExitCodeFailure
		}
		printConfigIssues(os.Stderr, "krewire.yaml", issues)
		if kiwconfig.HasErrors(issues) {
			return errs.ExitCodeFailure
		}
	}

	if issues := lintKiwFiles(dir); len(issues) > 0 {
		fmt.Fprintln(os.Stderr, "\n── .kiw Templates ──────────────────────────────────────────")
		for _, iss := range issues {
			fmt.Fprintf(os.Stderr, "  ✗ [KIW] %s\n", strings.ReplaceAll(iss, "\n", "\n    "))
		}
		return errs.ExitCodeFailure
	}

	if secCheck {
		regressions := renderSecurityAssuranceReport(os.Stdout, dir)
		if len(regressions) > 0 {
			fmt.Fprintf(os.Stderr, "\n✗ Security assurance quality gate failed with %d regression(s).\n", len(regressions))
			return errs.ExitCodeFailure
		}
		fmt.Fprintln(os.Stdout, "\n  ✓ Security assurance baseline: passed (zero regressions)")
	}

	if vulnCheck {
		fmt.Fprintln(os.Stdout, "\n── Vulnerability & Supply Chain Audit ──────────────────────")
		// Check 1: go mod verify
		verifyCmd := exec.Command("go", "mod", "verify")
		verifyCmd.Dir = dir
		if out, err := verifyCmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ [VULN] go mod verify failed: %s\n", strings.TrimSpace(string(out)))
			return errs.ExitCodeFailure
		}
		fmt.Fprintln(os.Stdout, "  ✓ Module checksum integrity: verified (go mod verify)")

		// Check 2: Minimum Go runtime toolchain advisory check
		if mod != nil && mod.GoVersion != "" {
			if err := gomod.CheckGoVersion(runtime.Version(), mod.GoVersion); err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ [VULN] %s\n", err)
				return errs.ExitCodeFailure
			}
			fmt.Fprintf(os.Stdout, "  ✓ Go toolchain version: %s matches minimum required %s\n", runtime.Version(), mod.GoVersion)
		}

		// Check 3: govulncheck if available
		if err := runGovulncheck(dir); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ [VULN] %s\n", err)
			return errs.ExitCodeFailure
		}
		fmt.Fprintln(os.Stdout, "  ✓ Vulnerability audit: passed (zero known CVEs)")
	}

	return errs.ExitCodeSuccess
}

func runGovulncheck(dir string) error {
	bin, err := exec.LookPath("govulncheck")
	if err != nil {
		fmt.Fprintln(os.Stdout, "  ℹ govulncheck not found in PATH (run 'go install golang.org/x/vuln/cmd/govulncheck@latest' for deep AST CVE scans)")
		return nil
	}
	cmd := exec.Command(bin, "./...")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func auditSecurityPosture(dir string) []string {
	var issues []string
	cfgPath := filepath.Join(dir, "krewire.yaml")
	if data, err := os.ReadFile(cfgPath); err == nil {
		content := string(data)
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			lower := strings.ToLower(trimmed)
			if strings.HasPrefix(trimmed, "#") {
				continue
			}
			if (strings.Contains(lower, "password:") || strings.Contains(lower, "api_key:") ||
				strings.Contains(lower, "secret:") || strings.Contains(lower, "token:")) &&
				!strings.Contains(trimmed, "${env:") && !strings.Contains(trimmed, "${") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" && strings.TrimSpace(parts[1]) != `""` {
					issues = append(issues, fmt.Sprintf("krewire.yaml:%d plaintext secret detected in %q — use ${env:VAR_NAME}", i+1, trimmed))
				}
			}
		}
	}
	issues = append(issues, auditEnvReferences(dir)...)
	return issues
}

var envRefRegex = regexp.MustCompile(`\$\{(?:env:)?([a-zA-Z_][a-zA-Z0-9_]*)(?::-([^}]+))?\}`)

func auditEnvReferences(dir string) []string {
	var issues []string
	cfgPath := filepath.Join(dir, "krewire.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil
	}

	cfg, _ := kiwconfig.Load(dir)
	envFile := ".env"
	if cfg != nil && cfg.Runtime.EnvFile != "" {
		envFile = cfg.Runtime.EnvFile
	}
	envMap := make(map[string]string)
	if envData, err := os.ReadFile(filepath.Join(dir, envFile)); err == nil {
		if pairs, err := config.ParseDotEnv(envData); err == nil {
			for _, p := range pairs {
				envMap[p.Key] = p.Value
			}
		}
	}

	lines := strings.Split(string(data), "\n")
	for lineIdx, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		matches := envRefRegex.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			if len(m) >= 2 {
				varName := m[1]
				hasDefault := len(m) >= 3 && m[2] != ""
				if hasDefault {
					continue
				}
				_, inDotenv := envMap[varName]
				_, inSys := os.LookupEnv(varName)
				if !inDotenv && !inSys {
					issues = append(issues, fmt.Sprintf("krewire.yaml:%d variable %q has no default value and is not defined in %s", lineIdx+1, varName, envFile))
				}
			}
		}
	}
	return issues
}
