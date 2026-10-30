package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func applyPatch(t *testing.T, tool *ApplyPatchTool, patch string) string {
	t.Helper()
	args, _ := json.Marshal(map[string]string{"patch": patch})
	res, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	return res.(string)
}

func TestApplyPatch_SingleHunk(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "main.go")
	write(t, target, "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n")

	patch := `--- a/main.go
+++ b/main.go
@@ -1,4 +1,5 @@
 package main
 
 func main() {
-	fmt.Println("hello")
+	fmt.Println("hello krew")
+	os.Exit(0)
 }
`
	applyPatch(t, &ApplyPatchTool{Workspace: root}, patch)
	want := "package main\n\nfunc main() {\n\tfmt.Println(\"hello krew\")\n\tos.Exit(0)\n}\n"
	if got := read(t, target); got != want {
		t.Errorf("expected:\n%s\ngot:\n%s", want, got)
	}
}

func TestApplyPatch_MultipleHunksAndFiles(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "a.txt"), "one\ntwo\nthree\nfour\nfive\n")
	write(t, filepath.Join(root, "b.txt"), "alpha\nbeta\n")

	patch := `--- a/a.txt
+++ b/a.txt
@@ -1,3 +1,3 @@
 one
-two
+dos
 three
@@ -3,3 +3,3 @@
-three
+tres
 four
 five
--- a/b.txt
+++ b/b.txt
@@ -1,2 +1,3 @@
 alpha
 beta
+gamma
`
	applyPatch(t, &ApplyPatchTool{Workspace: root}, patch)

	if got := read(t, filepath.Join(root, "a.txt")); got != "one\ndos\ntres\nfour\nfive\n" {
		t.Errorf("unexpected a.txt:\n%q", got)
	}
	if got := read(t, filepath.Join(root, "b.txt")); got != "alpha\nbeta\ngamma\n" {
		t.Errorf("unexpected b.txt:\n%q", got)
	}
}

func TestApplyPatch_FuzzyDrift(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "f.txt"), strings.Repeat("nope\n", 5)+"target\nafter\n")

	// Header offset says line 1 but the real block sits at line 6; fuzzy
	// radius search must still locate the `target` anchor.
	patch := `--- a/f.txt
+++ b/f.txt
@@ -1,3 +1,3 @@
 nope
-target
+found
 after
`
	applyPatch(t, &ApplyPatchTool{Workspace: root}, patch)

	got := read(t, filepath.Join(root, "f.txt"))
	if strings.Contains(got, "target") || !strings.Contains(got, "found\nafter") {
		t.Errorf("expected fuzzy replacement applied at real position, got:\n%q", got)
	}
	if !strings.HasSuffix(got, "found\nafter\n") {
		t.Errorf("expected trailing content preserved:\n%q", got)
	}
}

func TestApplyPatch_CreationAndDeletion(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "gone.txt"), "bye\n")

	patch := `--- /dev/null
+++ b/new/dirs/added.txt
@@ -0,0 +1,2 @@
+fresh
+file
--- a/gone.txt
+++ /dev/null
@@ -1,1 +0,0 @@
-bye
`
	applyPatch(t, &ApplyPatchTool{Workspace: root}, patch)

	if got := read(t, filepath.Join(root, "new", "dirs", "added.txt")); got != "fresh\nfile\n" {
		t.Errorf("unexpected created content: %q", got)
	}
	if _, err := os.Stat(filepath.Join(root, "gone.txt")); !os.IsNotExist(err) {
		t.Errorf("expected gone.txt deleted, stat err=%v", err)
	}
}

func TestApplyPatch_RollbackOnConflictingHunk(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "code.go")
	original := "package main\n\nfunc main() {}\n"
	write(t, target, original)

	// Both hunks: first applies, second references absent context.
	patch := `--- a/code.go
+++ b/code.go
@@ -1,3 +1,3 @@
 package main
 
-func main() {}
+func main() { run() }
@@ -9,2 +10,2 @@
-missing context anchor
+still gone
`
	_, err := (&ApplyPatchTool{Workspace: root}).Execute(context.Background(), marshalPatch(patch))
	if err == nil {
		t.Fatal("expected failure for conflicting hunk")
	}
	if got := read(t, target); got != original {
		t.Errorf("expected rollback to original, got:\n%s", got)
	}
}

func TestApplyPatch_SandboxEscapesRejected(t *testing.T) {
	root := t.TempDir()
	patch := `--- ../../outside.txt
+++ b/../../outside.txt
@@ -1,1 +1,2 @@
+x
`
	if _, err := (&ApplyPatchTool{Workspace: root}).Execute(context.Background(), marshalPatch(patch)); err == nil {
		t.Fatal("expected sandbox escape rejection")
	}
}

func marshalPatch(patch string) json.RawMessage {
	b, _ := json.Marshal(map[string]string{"patch": patch})
	return b
}
