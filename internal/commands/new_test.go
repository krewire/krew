package commands

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func TestRunNew_WorkerServiceInfra(t *testing.T) {
	for _, tc := range []struct {
		kind     string
		flagName string
	}{
		{"worker", "worker"},
		{"service", "service"},
		{"infra", "infra"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			dir := t.TempDir()
			chdir(t, dir)

			projectName := "my-" + tc.kind
			fs := flag.NewFlagSet("new", flag.ContinueOnError)
			RegisterNew(fs)
			_ = fs.Set(tc.flagName, "true")
			_ = fs.Parse([]string{projectName})

			code := RunNew(fs)
			if code != errs.ExitCodeSuccess {
				t.Fatalf("RunNew with --%s failed: %v", tc.flagName, code)
			}

			projectDir := filepath.Join(dir, projectName)
			yamlFile := filepath.Join(projectDir, "krewire.yaml")
			data, err := os.ReadFile(yamlFile)
			if err != nil {
				t.Fatalf("krewire.yaml not created: %v", err)
			}
			if !stringContains(string(data), "kind: "+tc.kind) {
				t.Errorf("krewire.yaml missing kind: %s, got: %s", tc.kind, string(data))
			}
		})
	}
}

func TestRunInit_WorkerServiceInfra(t *testing.T) {
	for _, tc := range []struct {
		kind     string
		flagName string
	}{
		{"worker", "worker"},
		{"service", "service"},
		{"infra", "infra"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			dir := t.TempDir()
			chdir(t, dir)

			fs := flag.NewFlagSet("init", flag.ContinueOnError)
			RegisterInit(fs)
			_ = fs.Set(tc.flagName, "true")
			_ = fs.Parse(nil)

			code := RunInit(fs)
			if code != errs.ExitCodeSuccess {
				t.Fatalf("RunInit with --%s failed: %v", tc.flagName, code)
			}

			yamlFile := filepath.Join(dir, "krewire.yaml")
			data, err := os.ReadFile(yamlFile)
			if err != nil {
				t.Fatalf("krewire.yaml not created: %v", err)
			}
			if !stringContains(string(data), "kind: "+tc.kind) {
				t.Errorf("krewire.yaml missing kind: %s, got: %s", tc.kind, string(data))
			}
		})
	}
}

func stringContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || stringIndex(s, sub) >= 0)
}

func stringIndex(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
