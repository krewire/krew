package commands

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func TestRunFmt_CleanCode(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()

	goMod := "module example.com/testfmt\n\ngo 1.27.1\n"
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}

	mainCode := "package main\n\nfunc main() {}\n"
	if err := os.WriteFile(filepath.Join(tmp, "main.go"), []byte(mainCode), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	fs := flag.NewFlagSet("fmt", flag.ContinueOnError)
	RegisterFmt(fs)

	code := RunFmt(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("got exit code %v, want ExitCodeSuccess (0)", code)
	}
}

func TestRunFmt_UnformattedCodeCheckAndWrite(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()

	goMod := "module example.com/testfmt2\n\ngo 1.27.1\n"
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatal(err)
	}

	// Unformatted code (extra spaces, weird indentation)
	badCode := "package main\n\nfunc  main( )  {\n}\n"
	mainPath := filepath.Join(tmp, "main.go")
	if err := os.WriteFile(mainPath, []byte(badCode), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	fsCheck := flag.NewFlagSet("fmt", flag.ContinueOnError)
	RegisterFmt(fsCheck)

	// Check mode should detect failure
	code := RunFmt(fsCheck)
	if code != errs.ExitCodeFailure {
		t.Fatalf("RunFmt(check) exit code = %v, want ExitCodeFailure (1)", code)
	}

	// Write mode should format and succeed
	fsWrite := flag.NewFlagSet("fmt", flag.ContinueOnError)
	RegisterFmt(fsWrite)
	_ = fsWrite.Set("write", "true")

	code = RunFmt(fsWrite)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunFmt(write) exit code = %v, want ExitCodeSuccess (0)", code)
	}

	// Read formatted file
	data, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	want := "package main\n\nfunc main() {\n}\n"
	if string(data) != want {
		t.Fatalf("formatted code = %q, want %q", string(data), want)
	}
}
