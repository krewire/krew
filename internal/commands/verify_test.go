package commands

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func TestVerifyCommand_CleanProject(t *testing.T) {
	dir := t.TempDir()

	// Create valid clean Go module
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/clean\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	code := "package clean\n\n// Add returns sum.\nfunc Add(a, b int) int {\n\treturn a + b\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "clean.go"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	testCode := "package clean\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {\n\tif Add(1, 2) != 3 {\n\t\tt.Fail()\n\t}\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "clean_test.go"), []byte(testCode), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	_ = os.Chdir(dir)

	// Intercept stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	RegisterVerify(fs)
	_ = fs.Parse([]string{"--task", "test-task-1", "--agent", "test-agent"})

	exitCode := RunVerify(fs)

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()

	if exitCode != errs.ExitCodeSuccess {
		t.Fatalf("expected verify to succeed on clean project, got %v: %s", exitCode, out)
	}

	if !strings.Contains(out, "Task Verification Evidence Record [test-task-1]") {
		t.Errorf("missing task header: %s", out)
	}
	if !strings.Contains(out, "FULLY VERIFIED") {
		t.Errorf("expected FULLY VERIFIED in output: %s", out)
	}
	if !strings.Contains(out, "gofmt -l .") {
		t.Errorf("missing fmt gate: %s", out)
	}
}

func TestVerifyCommand_FailsOnUnformattedCode(t *testing.T) {
	dir := t.TempDir()

	// Create unformatted Go module
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/badfmt\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	badCode := "package badfmt\nfunc Sub(a, b int) int { return a-b }\n" // Not gofmt compliant
	if err := os.WriteFile(filepath.Join(dir, "bad.go"), []byte(badCode), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	_ = os.Chdir(dir)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	RegisterVerify(fs)

	exitCode := RunVerify(fs)

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()

	if exitCode != errs.ExitCodeFailure {
		t.Fatalf("expected verify to fail on unformatted code, got %v", exitCode)
	}
	if !strings.Contains(out, "NOT VERIFIED") {
		t.Errorf("expected NOT VERIFIED in output: %s", out)
	}
	if !strings.Contains(out, "[FAIL] fmt") {
		t.Errorf("expected fmt gate to fail: %s", out)
	}
}

func TestVerifyCommand_LedgerIntegration(t *testing.T) {
	dir := t.TempDir()

	// Clean module
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/ledgerproj\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	code := "package ledgerproj\n\n// Hello returns greeting.\nfunc Hello() string {\n\treturn \"hi\"\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "hello.go"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	testCode := "package ledgerproj\n\nimport \"testing\"\n\nfunc TestHello(t *testing.T) {\n\tif Hello() != \"hi\" {\n\t\tt.Fail()\n\t}\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "hello_test.go"), []byte(testCode), 0644); err != nil {
		t.Fatal(err)
	}

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	_ = os.Chdir(dir)

	ledgerFile := filepath.Join(dir, ".krewire", "ledger.json")

	// Run verify with --ledger
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	RegisterVerify(fs)
	_ = fs.Parse([]string{"--task", "task-ledger-1", "--ledger", ledgerFile})

	exitCode := RunVerify(fs)
	if exitCode != errs.ExitCodeSuccess {
		t.Fatalf("expected verify with --ledger to succeed, got %v", exitCode)
	}

	// Verify ledger file was written
	if _, err := os.Stat(ledgerFile); err != nil {
		t.Fatalf("expected ledger file at %s, got: %v", ledgerFile, err)
	}

	// Run verify --verify-ledger
	fsCheck := flag.NewFlagSet("verify", flag.ContinueOnError)
	RegisterVerify(fsCheck)
	_ = fsCheck.Parse([]string{"--verify-ledger", "--ledger", ledgerFile})

	checkCode := RunVerify(fsCheck)
	if checkCode != errs.ExitCodeSuccess {
		t.Fatalf("expected --verify-ledger to succeed, got %v", checkCode)
	}
}
