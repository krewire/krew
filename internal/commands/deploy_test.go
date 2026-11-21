package commands

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func TestRunDeploy_BinaryWorkloads(t *testing.T) {
	for _, kind := range []string{"app", "cli", "worker", "service"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range map[string]string{
				"go.mod":       "module test" + kind + "\n\ngo 1.22\n",
				"main.go":      "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"ok\") }\n",
				"krewire.yaml": "project:\n  name: " + kind + "-bin\n  kind: " + kind + "\n",
			} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			chdir(t, dir)
			fs := flag.NewFlagSet("deploy", flag.ContinueOnError)
			RegisterDeploy(fs)
			_ = fs.Set("dry-run", "true")

			code := RunDeploy(fs)
			if code != errs.ExitCodeSuccess {
				t.Fatalf("RunDeploy on %s failed: %v", kind, code)
			}

			distBin := filepath.Join(dir, ".krewire", "dist", "test"+kind)
			if _, err := os.Stat(distBin); err != nil {
				t.Fatalf("expected binary at %s: %v", distBin, err)
			}

			// Verify SHA256SUMS was generated
			sumsFile := filepath.Join(dir, ".krewire", "dist", "SHA256SUMS")
			if _, err := os.Stat(sumsFile); err != nil {
				t.Fatalf("expected SHA256SUMS at %s: %v", sumsFile, err)
			}
		})
	}
}

func TestRunDeploy_SystemdTarget(t *testing.T) {
	for _, kind := range []string{"worker", "service", "app"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range map[string]string{
				"go.mod":       "module test" + kind + "\n\ngo 1.22\n",
				"main.go":      "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"ok\") }\n",
				"krewire.yaml": "project:\n  name: " + kind + "-sys\n  kind: " + kind + "\n",
			} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			chdir(t, dir)
			fs := flag.NewFlagSet("deploy", flag.ContinueOnError)
			RegisterDeploy(fs)
			_ = fs.Set("target", "systemd")
			_ = fs.Set("user", "myuser")
			_ = fs.Set("group", "mygroup")
			_ = fs.Set("dry-run", "true")

			code := RunDeploy(fs)
			if code != errs.ExitCodeSuccess {
				t.Fatalf("RunDeploy on %s systemd target failed: %v", kind, code)
			}

			distBin := filepath.Join(dir, ".krewire", "dist", "test"+kind)
			if _, err := os.Stat(distBin); err != nil {
				t.Fatalf("expected binary at %s: %v", distBin, err)
			}

			unitFile := filepath.Join(dir, ".krewire", "dist", "test"+kind+".service")
			unitData, err := os.ReadFile(unitFile)
			if err != nil {
				t.Fatalf("expected systemd service at %s: %v", unitFile, err)
			}
			unit := string(unitData)
			if !strings.Contains(unit, "User=myuser") {
				t.Errorf("systemd unit missing user: %s", unit)
			}
			if !strings.Contains(unit, "Group=mygroup") {
				t.Errorf("systemd unit missing group: %s", unit)
			}

			installScript := filepath.Join(dir, ".krewire", "dist", "install.sh")
			if _, err := os.Stat(installScript); err != nil {
				t.Fatalf("expected install.sh at %s: %v", installScript, err)
			}
		})
	}
}

func TestRunDeploy_InfraTarget(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"krewire.yaml": "project:\n  name: my-infra\n  kind: infra\ninfra:\n  provider: krewire-cloud\n  region: asia-southeast1\nserver:\n  port: 8080\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	chdir(t, dir)
	fs := flag.NewFlagSet("deploy", flag.ContinueOnError)
	RegisterDeploy(fs)
	_ = fs.Set("target", "infra")
	_ = fs.Set("plan", "true")

	code := RunDeploy(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunDeploy on infra target failed: %v", code)
	}

	planFile := filepath.Join(dir, ".krewire", "dist", "infra", "plan.json")
	data, err := os.ReadFile(planFile)
	if err != nil {
		t.Fatalf("expected plan in dist at %s: %v", planFile, err)
	}
	if !strings.Contains(string(data), "my-infra") {
		t.Errorf("plan missing my-infra resource: %s", string(data))
	}
}
