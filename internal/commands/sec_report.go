package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/krewire/pkg/model"
)

// renderSecurityAssuranceReport prints an end-to-end user-facing security baseline report,
// distinguishing verified controls, configuration-dependent controls, and unsupported controls,
// along with residual risks and regression audits.
// Returns a slice of security violation strings.
func renderSecurityAssuranceReport(w io.Writer, dir string) []string {
	var regressions []string

	baseline := model.GetSecurityBaseline()

	fmt.Fprintf(w, "\n── Security Assurance Report ───────────────────────────\n")
	fmt.Fprintf(w, "  Threat Model: %s\n\n", baseline.ThreatModel)

	// Categorize controls
	var verified, configDep, unsupported []model.SecurityControl
	for _, ctrl := range baseline.Controls {
		switch ctrl.Status {
		case model.ControlVerified:
			verified = append(verified, ctrl)
		case model.ControlConfigDependent:
			configDep = append(configDep, ctrl)
		case model.ControlUnsupported:
			unsupported = append(unsupported, ctrl)
		}
	}

	// 2. Verified Controls
	fmt.Fprintf(w, "  [VERIFIED CONTROLS — Active By Default]\n")
	for _, c := range verified {
		fmt.Fprintf(w, "    ✓ %s: %s\n", c.Name, c.Description)
		fmt.Fprintf(w, "      • Failure Behavior: %s\n", c.FailureBehavior)
	}

	// 3. Configuration-Dependent Controls
	fmt.Fprintf(w, "\n  [CONFIGURATION-DEPENDENT CONTROLS — Operator / Developer Action Required]\n")
	for _, c := range configDep {
		fmt.Fprintf(w, "    ⚠ %s: %s\n", c.Name, c.Description)
		fmt.Fprintf(w, "      • Responsibility: %s\n", c.Responsibility)
		fmt.Fprintf(w, "      • Failure Behavior: %s\n", c.FailureBehavior)
	}

	// 4. Unsupported / Out-of-Scope Controls
	if len(unsupported) > 0 {
		fmt.Fprintf(w, "\n  [OUT-OF-SCOPE / UNSUPPORTED CONTROLS]\n")
		for _, c := range unsupported {
			fmt.Fprintf(w, "    ℹ %s: %s\n", c.Name, c.Description)
		}
	}

	// 5. Residual Risks
	if len(baseline.ResidualRisks) > 0 {
		fmt.Fprintf(w, "\n  [RESIDUAL RISKS & DEVELOPER RESPONSIBILITIES]\n")
		for _, r := range baseline.ResidualRisks {
			fmt.Fprintf(w, "    • %s\n", r)
		}
	}

	// 6. Security Regressions Static Audit
	fmt.Fprintf(w, "\n── Automated Security Regression Checks ────────────────────\n")
	postureIssues := auditSecurityPosture(dir)
	if len(postureIssues) > 0 {
		regressions = append(regressions, postureIssues...)
		for _, iss := range postureIssues {
			fmt.Fprintf(w, "    ✗ [REGRESSION] %s\n", iss)
		}
	} else {
		fmt.Fprintf(w, "    ✓ Zero hardcoded plaintext credentials detected\n")
		fmt.Fprintf(w, "    ✓ Dynamic environment references (${env:...}) contract upheld\n")
	}

	// Check for unsafe CORS configuration if present in krewire.yaml
	corsIssues := auditCorsConfiguration(dir)
	if len(corsIssues) > 0 {
		regressions = append(regressions, corsIssues...)
		for _, iss := range corsIssues {
			fmt.Fprintf(w, "    ✗ [REGRESSION] %s\n", iss)
		}
	}

	return regressions
}

func auditCorsConfiguration(dir string) []string {
	var issues []string
	cfgPath := filepath.Join(dir, "krewire.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	hasWildcardOrigin := false
	hasCredentialsTrue := false
	wildcardLine := 0
	credentialsLine := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "origins") && strings.Contains(lower, "*") {
			hasWildcardOrigin = true
			wildcardLine = i + 1
		} else if strings.HasPrefix(lower, "- \"*\"") || strings.HasPrefix(lower, "- '*'") || lower == "- *" {
			hasWildcardOrigin = true
			wildcardLine = i + 1
		}
		if strings.Contains(lower, "allow_credentials") && strings.Contains(lower, "true") {
			hasCredentialsTrue = true
			credentialsLine = i + 1
		}
	}

	if hasWildcardOrigin && hasCredentialsTrue {
		issues = append(issues, fmt.Sprintf("krewire.yaml:%d insecure CORS configuration: wildcard '*' origin cannot be combined with allow_credentials: true (line %d)", wildcardLine, credentialsLine))
	}

	return issues
}
