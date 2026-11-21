package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func TestLoadScriptsFromKrewireYaml(t *testing.T) {
	dir := t.TempDir()
	body := `title: Demo
scripts:
  lint: "go vet ./..."
  seed: "go run ./tools/seed.go"
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	RegisterRun(fs)
	if err := fs.Parse(nil); err != nil {
		t.Fatal(err)
	}
	rt, code := bootRuntime(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("bootRuntime = %d", code)
	}
	if len(rt.cfg.Scripts) != 2 {
		t.Fatalf("scripts = %v, want 2", rt.cfg.Scripts)
	}
	if rt.cfg.Scripts["lint"] != "go vet ./..." {
		t.Errorf("lint = %q", rt.cfg.Scripts["lint"])
	}
}

func TestRunGoFile_ExecutesAndForwardsExit(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"go.mod":       "module demo\n\ngo 1.22\n",
		"hello.go":     "package main\nimport \"fmt\"\nfunc main(){ fmt.Println(\"hello-script\") }",
		"fail.go":      "package main\nimport \"os\"\nfunc main(){ os.Exit(5) }",
		"args.go":      "package main\nimport (\"fmt\";\"os\")\nfunc main(){ fmt.Println(os.Args[1]) }",
		"krewire.yaml": "title: Demo\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	chdir(t, dir)

	run := func(args ...string) errs.ExitCode {
		fs := flag.NewFlagSet("run", flag.ContinueOnError)
		RegisterRun(fs)
		if err := fs.Parse(args); err != nil {
			t.Fatal(err)
		}
		return RunRun(fs)
	}
	if code := run("hello.go"); code != errs.ExitCodeSuccess {
		t.Errorf("run hello.go = %d want 0", code)
	}
	if code := run("fail.go"); code != errs.ExitCodeFailure {
		t.Errorf("run fail.go = %d want 1 (failure mapping)", code)
	}
	if code := run("hello.go", "--", "extra"); code != errs.ExitCodeSuccess {
		t.Errorf("run hello.go with extra args = %d want 0", code)
	}
	if code := run("args.go", "world"); code != errs.ExitCodeSuccess {
		t.Errorf("run args.go world = %d want 0", code)
	}
	if code := run("nonexist.go"); code != errs.ExitCodeUsage {
		t.Errorf("run nonexist.go = %d want 2 (usage)", code)
	}
}

func TestRunTask_ExecutesAndUnknownTask(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"go.mod":       "module demo\n\ngo 1.22\n",
		"krewire.yaml": "title: Demo\nscripts:\n  hello: \"echo hello-task\"\n  fail: \"false\"\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	chdir(t, dir)

	run := func(args ...string) errs.ExitCode {
		fs := flag.NewFlagSet("run", flag.ContinueOnError)
		RegisterRun(fs)
		if err := fs.Parse(args); err != nil {
			t.Fatal(err)
		}
		return RunRun(fs)
	}
	if code := run("hello"); code != errs.ExitCodeSuccess {
		t.Errorf("run hello task = %d want 0", code)
	}
	if code := run("fail"); code != errs.ExitCodeFailure {
		t.Errorf("run fail task = %d want 1", code)
	}
	if code := run("unknown-task"); code != errs.ExitCodeUsage {
		t.Errorf("run unknown-task = %d want 2", code)
	}
}

func TestPrecedence_FileOverScript(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"go.mod":       "module demo\n\ngo 1.22\n",
		"dup.go":       "package main\nfunc main(){}",
		"krewire.yaml": "title: Demo\nscripts:\n  dup.go: \"echo from-script\"\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	chdir(t, dir)
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	RegisterRun(fs)
	if err := fs.Parse([]string{"dup.go"}); err != nil {
		t.Fatal(err)
	}
	code := RunRun(fs)
	if code != errs.ExitCodeSuccess {
		t.Errorf("precedence dup.go = %d want 0 (file wins)", code)
	}
}

func TestBareRun_PreservesAppBehavior(t *testing.T) {
	dir := t.TempDir()
	main := `package main
import "fmt"
func main(){ fmt.Println("app-run") }
`
	for name, body := range map[string]string{
		"go.mod":       "module demo\n\ngo 1.22\n",
		"main.go":      main,
		"krewire.yaml": "title: Demo\nproject:\n  kind: app\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	chdir(t, dir)
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	RegisterRun(fs)
	if err := fs.Parse(nil); err != nil {
		t.Fatal(err)
	}
	if code := RunRun(fs); code != errs.ExitCodeSuccess {
		t.Errorf("bare run app = %d want 0", code)
	}
}

func TestIsGoFileArg(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "dir.go"), 0o755); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		arg  string
		want bool
	}{
		{"a.go", true},
		{"missing.go", false},
		{"dir.go", false},
		{"a.txt", false},
		{"a", false},
	}
	for _, tt := range tests {
		if got := isGoFileArg(dir, tt.arg); got != tt.want {
			t.Errorf("isGoFileArg(%q)=%v want %v", tt.arg, got, tt.want)
		}
	}
}

func TestInfoListsScripts(t *testing.T) {
	dir := t.TempDir()
	body := `title: Demo
scripts:
  build-all: "go build ./..."
  test-unit: "go test ./..."
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)
	fs := flag.NewFlagSet("info", flag.ContinueOnError)
	code := RunInfo(fs)
	if code != errs.ExitCodeSuccess {
		t.Errorf("RunInfo = %d want 0", code)
	}
}

func TestRunScript_SanitizesSpecialShellArgs(t *testing.T) {
	dir := t.TempDir()
	canaryFile := filepath.Join(dir, "pwned.txt")
	outputFile := filepath.Join(dir, "output.txt")

	body := fmt.Sprintf("title: Demo\nscripts:\n  record: \"echo > %s\"\n", outputFile)
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)

	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	RegisterRun(fs)
	maliciousArgs := []string{"record", "--", ";", "touch", canaryFile, "&", "|"}
	if err := fs.Parse(maliciousArgs); err != nil {
		t.Fatal(err)
	}

	code := RunRun(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunRun = %d, want 0", code)
	}

	// Verify canary file was NOT created (command injection prevented)
	if _, err := os.Stat(canaryFile); !os.IsNotExist(err) {
		t.Fatalf("vulnerability: canary file %s was created via command injection!", canaryFile)
	}

	// Verify special shell characters were passed literally in the command execution
	cmd := buildScriptCmd("echo", []string{";", "touch", "&", "|"})
	cmdStr := cmd.String()
	if !strings.Contains(cmdStr, "';'") || !strings.Contains(cmdStr, "'&'") || !strings.Contains(cmdStr, "'|'") {
		t.Fatalf("expected shell arguments to be safely quoted, got: %s", cmdStr)
	}
}
