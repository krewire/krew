package commands

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func TestRunRun_RejectsNonRunnableWorkloads(t *testing.T) {
	cases := []struct {
		kind    string
		content string
	}{
		{"site", "project:\n  kind: site\n"},
		{"book", "project:\n  kind: book\n"},
		{"infra", "project:\n  kind: infra\n"},
	}

	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(tc.content), 0o644); err != nil {
				t.Fatal(err)
			}
			chdir(t, dir)
			fs := flag.NewFlagSet("run", flag.ContinueOnError)
			RegisterRun(fs)
			code := RunRun(fs)
			if code != errs.ExitCodeUsage {
				t.Errorf("RunRun on %s = %v, want ExitCodeUsage (2)", tc.kind, code)
			}
		})
	}
}

func TestRunDev_RejectsInfraAndPreInit(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"infra", "project:\n  kind: infra\n"},
		{"pre-init", "title: PreInit\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.name == "pre-init" {
				for name, body := range map[string]string{
					"go.mod":       "module testkern\n\ngo 1.22\n",
					"main.go":      "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"hello, krewire\") }\n",
					"krewire.yaml": "title: PreInit\n",
					".gitignore":   ".krewire/\n",
				} {
					if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
						t.Fatal(err)
					}
				}
			} else {
				if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(tc.content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			chdir(t, dir)
			fs := flag.NewFlagSet("dev", flag.ContinueOnError)
			RegisterDev(fs)
			code := RunDev(fs)
			if code != errs.ExitCodeUsage {
				t.Errorf("RunDev on %s = %v, want ExitCodeUsage (2)", tc.name, code)
			}
		})
	}
}
