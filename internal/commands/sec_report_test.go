package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderSecurityAssuranceReport_CleanProject(t *testing.T) {
	dir := t.TempDir()

	// Write clean krewire.yaml
	yaml := `project:
  name: demo-service
  kind: service
  version: "1.0.0"
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	regressions := renderSecurityAssuranceReport(&buf, dir)

	if len(regressions) != 0 {
		t.Fatalf("expected 0 regressions for clean project, got: %v", regressions)
	}

	out := buf.String()
	if !strings.Contains(out, "Workload Security Assurance Report [SERVICE]") {
		t.Errorf("output missing expected workload header: %s", out)
	}
	if !strings.Contains(out, "[VERIFIED CONTROLS — Active By Default]") {
		t.Errorf("output missing verified controls section: %s", out)
	}
	if !strings.Contains(out, "[CONFIGURATION-DEPENDENT CONTROLS") {
		t.Errorf("output missing config-dependent controls section: %s", out)
	}
	if !strings.Contains(out, "[RESIDUAL RISKS & DEVELOPER RESPONSIBILITIES]") {
		t.Errorf("output missing residual risks section: %s", out)
	}
	if !strings.Contains(out, "Zero hardcoded plaintext credentials detected") {
		t.Errorf("output missing secret hygiene confirmation: %s", out)
	}
}

func TestRenderSecurityAssuranceReport_DetectsRegressions(t *testing.T) {
	dir := t.TempDir()

	// Insecure krewire.yaml with plaintext password and insecure CORS
	insecureYaml := `project:
  name: insecure-app
  kind: app
database:
  password: "super_secret_hardcoded_pwd"
server:
  cors:
    origins:
      - "*"
    allow_credentials: true
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(insecureYaml), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	regressions := renderSecurityAssuranceReport(&buf, dir)

	if len(regressions) < 2 {
		t.Fatalf("expected at least 2 regressions (plaintext password + insecure cors), got: %v", regressions)
	}

	out := buf.String()
	if !strings.Contains(out, "[REGRESSION]") {
		t.Errorf("expected regression flag in output: %s", out)
	}
	if !strings.Contains(out, "plaintext secret detected") {
		t.Errorf("expected plaintext secret detection: %s", out)
	}
	if !strings.Contains(out, "insecure CORS configuration") {
		t.Errorf("expected CORS configuration error: %s", out)
	}
}
