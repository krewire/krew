package commands

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
)

func TestWorkerCommand_KindMismatch(t *testing.T) {
	dir := t.TempDir()
	yaml := "project:\n  name: myapp\n  kind: app\n"
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	_ = os.Chdir(dir)

	fs := flag.NewFlagSet("worker", flag.ContinueOnError)
	RegisterWorker(fs)

	code := RunWorker(fs)
	if code != errs.ExitCodeUsage {
		t.Fatalf("expected exit code %v, got %v", errs.ExitCodeUsage, code)
	}
}

func TestWorkerCommand_DLQListEmpty(t *testing.T) {
	dir := t.TempDir()
	yaml := "project:\n  name: myworker\n  kind: worker\n"
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	_ = os.Chdir(dir)

	fs := flag.NewFlagSet("worker", flag.ContinueOnError)
	RegisterWorker(fs)
	_ = fs.Set("dlq", "list")
	_ = fs.Set("queue", "file")

	code := RunWorker(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected exit code success, got %v", code)
	}
}

func TestWorkerCommand_AppWithWorkerConfigAllowed(t *testing.T) {
	dir := t.TempDir()
	yaml := "project:\n  name: myapp\n  kind: app\nworker:\n  concurrency: 2\n"
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	_ = os.Chdir(dir)

	fs := flag.NewFlagSet("worker", flag.ContinueOnError)
	RegisterWorker(fs)
	_ = fs.Set("dlq", "list")

	code := RunWorker(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected exit code success, got %v", code)
	}
}

func TestWorkerCommand_DLQClear(t *testing.T) {
	dir := t.TempDir()
	yaml := "project:\n  name: myworker\n  kind: worker\n"
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	_ = os.Chdir(dir)

	fs := flag.NewFlagSet("worker", flag.ContinueOnError)
	RegisterWorker(fs)
	_ = fs.Set("dlq", "clear")
	_ = fs.Set("queue", "file")

	code := RunWorker(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("expected exit code success, got %v", code)
	}
}

func TestWorkerCommand_DLQUnknownSubcommand(t *testing.T) {
	dir := t.TempDir()
	yaml := "project:\n  name: myworker\n  kind: worker\n"
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	_ = os.Chdir(dir)

	fs := flag.NewFlagSet("worker", flag.ContinueOnError)
	RegisterWorker(fs)
	_ = fs.Set("dlq", "unknown-action")

	code := RunWorker(fs)
	if code != errs.ExitCodeUsage {
		t.Fatalf("expected exit code usage, got %v", code)
	}
}
