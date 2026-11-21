package commands

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krew/internal/gomod"
	"github.com/krewire/krewire/pkg/errs"
)

func newWsFlagSet(args ...string) *flag.FlagSet {
	fs := flag.NewFlagSet("ws", flag.ContinueOnError)
	_ = fs.Parse(args)
	return fs
}

func captureStdout(t *testing.T, fn func() errs.ExitCode) (string, errs.ExitCode) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := fn()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out), code
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunWs_DispatchesHelpAndRejectsUnknown(t *testing.T) {
	if _, code := captureStdout(t, func() errs.ExitCode { return RunWs(newWsFlagSet("help")) }); code != errs.ExitCodeUsage {
		t.Errorf("help exit code = %v, want Usage", code)
	}
	if _, code := captureStdout(t, func() errs.ExitCode { return RunWs(newWsFlagSet("bogus")) }); code != errs.ExitCodeUsage {
		t.Errorf("unknown sub-command exit code = %v, want Usage", code)
	}
}

func TestFindWorkspaceRoot_ClassifiesGoWorkAndMonorepo(t *testing.T) {
	hub := t.TempDir()
	writeFile(t, filepath.Join(hub, "go.work"), "go 1.26\n\nuse (\n\t./a\n\t./b\n)\n")
	nested := filepath.Join(hub, "a", "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	root, typ, members := findWorkspaceRoot(nested)
	if root != hub || !strings.Contains(typ, "go.work") {
		t.Errorf("hub detection root=%q typ=%q", root, typ)
	}
	if len(members) != 2 || members[0] != "./a" || members[1] != "./b" {
		t.Errorf("block-form members = %q", members)
	}

	solo := t.TempDir()
	writeFile(t, filepath.Join(solo, "go.mod"), "module example.com/solo\n")
	root2, typ2, members2 := findWorkspaceRoot(solo)
	if root2 != solo || !strings.Contains(typ2, "monorepo") || len(members2) != 1 || members2[0] != solo {
		t.Errorf("monorepo detection root=%q typ=%q members=%q", root2, typ2, members2)
	}
}

func TestRunWsInfo_PrintsRootTypeAndMembers(t *testing.T) {
	solo := t.TempDir()
	writeFile(t, filepath.Join(solo, "go.mod"), "module example.com/solo\n")
	chdir(t, solo)

	out, code := captureStdout(t, func() errs.ExitCode { return RunWs(newWsFlagSet("info")) })
	if code != errs.ExitCodeSuccess {
		t.Fatalf("info exit code = %v", code)
	}
	for _, want := range []string{"workspace:", "type:", "monorepo", "members:"} {
		if !strings.Contains(out, want) {
			t.Errorf("info output missing %q in %q", want, out)
		}
	}
}

func TestRunWsList_RendersProjectKindModuleTable(t *testing.T) {
	hub := t.TempDir()
	writeFile(t, filepath.Join(hub, "go.mod"), "module example.com/hub\n")
	writeFile(t, filepath.Join(hub, "go.work"), "go 1.26\n\nuse (\n\t.\n)\n")

	if _, err := gomod.Read(filepath.Join(hub, "go.mod")); err != nil {
		t.Fatalf("gomod fixture unreadable: %v", err)
	}
	chdir(t, hub)

	out, code := captureStdout(t, func() errs.ExitCode { return RunWs(newWsFlagSet("list")) })
	if code != errs.ExitCodeSuccess {
		t.Fatalf("list exit code = %v", code)
	}
	for _, want := range []string{"PROJECT", "example.com/hub"} {
		if !strings.Contains(out, want) {
			t.Errorf("list output missing %q in %q", want, out)
		}
	}
}

func TestRunWsAddRemove_UsageWithoutArgsOrGoWork(t *testing.T) {
	chdir(t, t.TempDir())

	if _, code := captureStdout(t, func() errs.ExitCode { return RunWs(newWsFlagSet("add")) }); code != errs.ExitCodeUsage {
		t.Errorf("add without path = %v, want Usage", code)
	}
	if _, code := captureStdout(t, func() errs.ExitCode { return RunWs(newWsFlagSet("remove")) }); code != errs.ExitCodeUsage {
		t.Errorf("remove without path = %v, want Usage", code)
	}
	if _, code := captureStdout(t, func() errs.ExitCode { return RunWs(newWsFlagSet("add", "./x")) }); code != errs.ExitCodeUsage {
		t.Errorf("add outside go.work layout = %v, want Usage", code)
	}
}

func TestRunWsSync_SucceedsInTrivialWorkspace(t *testing.T) {
	hub := t.TempDir()
	writeFile(t, filepath.Join(hub, "go.mod"), "module example.com/hub\n")
	writeFile(t, filepath.Join(hub, "go.work"), "go 1.26\n\nuse (\n\t.\n)\n")
	chdir(t, hub)

	if _, code := captureStdout(t, func() errs.ExitCode { return RunWs(newWsFlagSet("sync")) }); code != errs.ExitCodeSuccess {
		t.Fatalf("sync exit code = %v, want Success", code)
	}
}

func TestRunWsExec_AggregatesMemberFailures(t *testing.T) {
	hub := t.TempDir()
	writeFile(t, filepath.Join(hub, "good", "go.mod"), "module example.com/good\n")
	writeFile(t, filepath.Join(hub, "bad", "go.mod"), "module example.com/bad\n")
	writeFile(t, filepath.Join(hub, "go.work"), "go 1.26\n\nuse (\n\t./good\n\t./bad\n)\n")
	chdir(t, hub)

	_, ok := captureStdout(t, func() errs.ExitCode { return RunWs(newWsFlagSet("exec", "--", "true")) })
	if ok != errs.ExitCodeSuccess {
		t.Fatalf("exec true exit code = %v, want Success", ok)
	}

	_, failed := captureStdout(t, func() errs.ExitCode { return RunWs(newWsFlagSet("exec", "false")) })
	if failed != errs.ExitCodeFailure {
		t.Fatalf("exec false exit code = %v, want Failure", failed)
	}

	if _, code := captureStdout(t, func() errs.ExitCode { return RunWs(newWsFlagSet("exec")) }); code != errs.ExitCodeUsage {
		t.Errorf("exec without command = %v, want Usage", code)
	}
}

func TestHelpTextDocumentsSubcommandsExamplesAndTypes(t *testing.T) {
	out, code := captureStdout(t, func() errs.ExitCode { return RunWs(newWsFlagSet("help")) })
	if code != errs.ExitCodeUsage {
		t.Fatalf("help exit code = %v, want Usage", code)
	}

	// Subcommands
	for _, sub := range []string{"info", "list", "add", "remove", "sync", "exec"} {
		if !strings.Contains(out, sub) {
			t.Errorf("help text missing subcommand %q", sub)
		}
	}

	// Examples
	if !strings.Contains(out, "Examples:") || !strings.Contains(out, "krew ws info") {
		t.Errorf("help text missing examples section")
	}

	// Four workspace types
	for _, wsType := range []string{"monorepo", "multi-repo", "multi-project", "microservices"} {
		if !strings.Contains(out, wsType) {
			t.Errorf("help text missing workspace type %q", wsType)
		}
	}
}
