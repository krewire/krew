package commands

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krewire/errs"
)

func TestRegisterNew(t *testing.T) {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	RegisterNew(fs)

	expectedFlags := []string{
		"module", "dir", "site", "book", "cli",
		"app", "worker", "service", "gateway", "infra",
		"template", "title", "no-input",
	}
	for _, name := range expectedFlags {
		if fs.Lookup(name) == nil {
			t.Errorf("flag %q was not registered by RegisterNew", name)
		}
	}
}

func TestRunNew_DefaultAndExplicitApp(t *testing.T) {
	for _, tc := range []struct {
		name        string
		flag        string
		projectName string
	}{
		{"default app", "", "default-app-proj"},
		{"explicit app", "app", "explicit-app-proj"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			chdir(t, dir)

			fs := flag.NewFlagSet("new", flag.ContinueOnError)
			RegisterNew(fs)
			if tc.flag != "" {
				_ = fs.Set(tc.flag, "true")
			}
			_ = fs.Parse([]string{tc.projectName})

			code := RunNew(fs)
			if code != errs.ExitCodeSuccess {
				t.Fatalf("RunNew failed with code %v", code)
			}

			projectDir := filepath.Join(dir, tc.projectName)
			for _, file := range []string{"krewire.yaml", "go.mod", "main.go"} {
				if _, err := os.Stat(filepath.Join(projectDir, file)); err != nil {
					t.Errorf("expected %s to exist: %v", file, err)
				}
			}
		})
	}
}

func TestRunNew_SiteAndBook(t *testing.T) {
	for _, tc := range []struct {
		variant   string
		flag      string
		title     string
		expect    string
		checkFile string
	}{
		{"site default title", "site", "", "my-site", "pages/index.kiw"},
		{"site custom title", "site", "My Beautiful Site", "My Beautiful Site", "pages/index.kiw"},
		{"book default title", "book", "", "my-book", "krewire.yaml"},
		{"book custom title", "book", "My Great Book", "My Great Book", "krewire.yaml"},
	} {
		t.Run(tc.variant, func(t *testing.T) {
			dir := t.TempDir()
			chdir(t, dir)

			projectName := "my-" + tc.flag
			if tc.title != "" {
				projectName += "-custom"
			}
			fs := flag.NewFlagSet("new", flag.ContinueOnError)
			RegisterNew(fs)
			_ = fs.Set(tc.flag, "true")
			if tc.title != "" {
				_ = fs.Set("title", tc.title)
			}
			_ = fs.Parse([]string{projectName})

			code := RunNew(fs)
			if code != errs.ExitCodeSuccess {
				t.Fatalf("RunNew with --%s failed: %v", tc.flag, code)
			}

			projectDir := filepath.Join(dir, projectName)
			targetFile := filepath.Join(projectDir, tc.checkFile)
			data, err := os.ReadFile(targetFile)
			if err != nil {
				t.Fatalf("failed to read %s: %v", tc.checkFile, err)
			}
			if !stringContains(string(data), tc.expect) {
				t.Errorf("%s expected to contain %q, got: %s", tc.checkFile, tc.expect, string(data))
			}
		})
	}
}

func TestRunNew_CLIAndGateway(t *testing.T) {
	for _, tc := range []struct {
		variant string
		flag    string
	}{
		{"cli", "cli"},
		{"gateway", "gateway"},
	} {
		t.Run(tc.variant, func(t *testing.T) {
			dir := t.TempDir()
			chdir(t, dir)

			projectName := "my-" + tc.flag
			fs := flag.NewFlagSet("new", flag.ContinueOnError)
			RegisterNew(fs)
			_ = fs.Set(tc.flag, "true")
			_ = fs.Parse([]string{projectName})

			code := RunNew(fs)
			if code != errs.ExitCodeSuccess {
				t.Fatalf("RunNew with --%s failed: %v", tc.flag, code)
			}

			projectDir := filepath.Join(dir, projectName)
			if _, err := os.Stat(filepath.Join(projectDir, "krewire.yaml")); err != nil {
				t.Fatalf("krewire.yaml missing: %v", err)
			}
			if _, err := os.Stat(filepath.Join(projectDir, "main.go")); err != nil {
				t.Fatalf("main.go missing: %v", err)
			}
		})
	}
}

func TestRunNew_VariantConflict(t *testing.T) {
	conflicts := [][]string{
		{"--site", "--book"},
		{"--app", "--worker"},
		{"--cli", "--service"},
		{"--gateway", "--infra"},
		{"--site", "--template=https://github.com/example/template"},
	}
	for _, flags := range conflicts {
		t.Run(strings.Join(flags, "_"), func(t *testing.T) {
			dir := t.TempDir()
			chdir(t, dir)

			fs := flag.NewFlagSet("new", flag.ContinueOnError)
			RegisterNew(fs)
			args := append(flags, "conflict-proj")
			_ = fs.Parse(args)

			code := RunNew(fs)
			if code != errs.ExitCodeUsage {
				t.Fatalf("expected ExitCodeUsage (%v), got %v", errs.ExitCodeUsage, code)
			}
		})
	}
}

func TestRunNew_CustomDirAndModule(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	customParent := filepath.Join(dir, "sub", "parent")
	customModule := "github.com/example/custommod"
	projectName := "custom-proj"

	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	RegisterNew(fs)
	_ = fs.Set("dir", customParent)
	_ = fs.Set("module", customModule)
	_ = fs.Parse([]string{projectName})

	code := RunNew(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunNew with custom dir and module failed: %v", code)
	}

	targetDir := filepath.Join(customParent, projectName)
	goModPath := filepath.Join(targetDir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("go.mod missing at %s: %v", goModPath, err)
	}
	if !stringContains(string(data), "module "+customModule) {
		t.Errorf("go.mod missing module declaration: %s", string(data))
	}
}

func TestRunNew_AbsolutePath(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	absPath := filepath.Join(dir, "abs-proj")

	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	RegisterNew(fs)
	_ = fs.Parse([]string{absPath})

	code := RunNew(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunNew with absolute path failed: %v", code)
	}

	if _, err := os.Stat(filepath.Join(absPath, "krewire.yaml")); err != nil {
		t.Fatalf("krewire.yaml missing at absolute path: %v", err)
	}
}

func TestRunNew_DotName(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	RegisterNew(fs)
	_ = fs.Parse([]string{"."})

	code := RunNew(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunNew with '.' failed: %v", code)
	}

	yamlFile := filepath.Join(dir, "krewire.yaml")
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		t.Fatalf("krewire.yaml missing: %v", err)
	}
	if !stringContains(string(data), "project:") {
		t.Errorf("expected project block in yaml, got %s", string(data))
	}
}

func TestRunNew_EmptyNameNoPrompt(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	oldIn, oldOut, oldInteractive := promptIn, promptOut, interactive
	interactive = func() bool { return false }
	t.Cleanup(func() { promptIn, promptOut, interactive = oldIn, oldOut, oldInteractive })

	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	RegisterNew(fs)
	_ = fs.Parse([]string{})

	code := RunNew(fs)
	if code != errs.ExitCodeUsage {
		t.Fatalf("expected ExitCodeUsage, got %v", code)
	}
}

func TestRunNew_DirectoryExistsError(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	projectName := "existing-proj"
	targetDir := filepath.Join(dir, projectName)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "existing.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	RegisterNew(fs)
	_ = fs.Parse([]string{projectName})

	code := RunNew(fs)
	if code == errs.ExitCodeSuccess {
		t.Fatalf("expected RunNew to fail when target directory exists and is not empty, got %v", code)
	}
}

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
			if !stringContains(string(data), "name: ") {
				t.Errorf("krewire.yaml missing project name, got: %s", string(data))
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
			if !stringContains(string(data), "name: ") {
				t.Errorf("krewire.yaml missing project name, got: %s", string(data))
			}
		})
	}
}

func TestRunNew_TemplateError(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	RegisterNew(fs)
	_ = fs.Set("template", "file:///nonexistent/repo.git")
	_ = fs.Parse([]string{"tpl-proj"})

	code := RunNew(fs)
	if code == errs.ExitCodeSuccess {
		t.Fatalf("expected RunNew with invalid template to fail, got %v", code)
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
