package shape

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krewire/pkg/model"
)

func writeFile(t *testing.T, dir, name, body string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectAppByRootMainGo(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/app\n")
	writeFile(t, dir, "main.go", `package main

func main() {}

`)
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindApp || got.Marker != "main.go" {
		t.Errorf("got %v (%s), want app (main.go)", got.Kind, got.Marker)
	}
}

func TestDetectAppByCmdPackage(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/app\n")
	writeFile(t, dir, "cmd/server/main.go", `package main

func main() {}

`)
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindApp || got.Marker != "cmd/*" {
		t.Errorf("got %v (%s), want app (cmd/*)", got.Kind, got.Marker)
	}
}

func TestDetectNonMainGoIsNotApp(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/lib\n")
	writeFile(t, dir, "main.go", `package lib

func Main() {}

`)
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindNone {
		t.Errorf("got %v, want none", got.Kind)
	}
}

func TestDetectSite(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "krewire.yaml", "ssg:\n  title: Hi\n")
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindSite {
		t.Errorf("got %v, want site", got.Kind)
	}
}

func TestDetectSiteByPagesDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pages/index.kiw", "---\ntitle: Hi\n---\n<h1>Hi</h1>\n")
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindSite || got.Marker != "pages/" {
		t.Errorf("got %v (%s), want site (pages/)", got.Kind, got.Marker)
	}
}

func TestDetectSiteByConfigButAppPrecedes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "krewire.yaml", "ssg:\n")
	writeFile(t, dir, "main.go", `package main

func main() {}

`)
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindApp {
		t.Errorf("got %v, want app (precedence)", got.Kind)
	}
}

func TestDetectBook(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "content/index.md", "# Hi\n")
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindBook {
		t.Errorf("got %v, want book", got.Kind)
	}
}

func TestDetectWorkerByYamlKey(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "krewire.yaml", "worker:\n  concurrency: 4\n")
	writeFile(t, dir, "main.go", `package main

func main() {}
`)
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindWorker || got.Marker != "krewire.yaml#worker" {
		t.Errorf("got %v (%s), want worker (krewire.yaml#worker)", got.Kind, got.Marker)
	}
}

func TestDetectServiceByYamlKey(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "krewire.yaml", "service:\n  name: auth\n  port: 8081\n")
	writeFile(t, dir, "main.go", `package main

func main() {}
`)
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindService || got.Marker != "krewire.yaml#service" {
		t.Errorf("got %v (%s), want service (krewire.yaml#service)", got.Kind, got.Marker)
	}
}

func TestDetectInfraByYamlKey(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "krewire.yaml", "infra:\n  provider: aws\n")
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindInfra || got.Marker != "krewire.yaml#infra" {
		t.Errorf("got %v (%s), want infra (krewire.yaml#infra)", got.Kind, got.Marker)
	}
}

func TestDetectPreInitBootstrap(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/starter\n\ngo 1.27.1\n")
	writeFile(t, dir, "krewire.yaml", "project:\n  name: starter\n")
	writeFile(t, dir, ".gitignore", ".krewire/\n")
	writeFile(t, dir, "main.go", `package main

import "fmt"

func main() {
	fmt.Println("hello, krewire")
}
`)
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindNone || got.Marker != MarkerPreInit {
		t.Errorf("got %v (%s), want none (pre-init)", got.Kind, got.Marker)
	}
}

func TestDetectGatewayKey(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "krewire.yaml", "gateway:\n  port: 8080\n")
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindGateway || got.Marker != "krewire.yaml#gateway" {
		t.Errorf("got %v (%s), want gateway (krewire.yaml#gateway)", got.Kind, got.Marker)
	}
}

func TestDetectAllEightExplicitKinds(t *testing.T) {
	for _, kind := range model.AllKinds {
		t.Run(string(kind), func(t *testing.T) {
			dir := t.TempDir()
			got, err := Detect(dir, string(kind))
			if err != nil {
				t.Fatalf("Detect failed for explicit kind %s: %v", kind, err)
			}
			if got.Kind != kind || got.Marker != "project.kind" {
				t.Errorf("got %v (%s), want %s (project.kind)", got.Kind, got.Marker, kind)
			}
		})
	}
}

func TestDetectCustomExplicitKind(t *testing.T) {
	got, err := Detect(t.TempDir(), "lambda")
	if err != nil {
		t.Fatalf("unexpected error for custom explicit kind: %v", err)
	}
	if got.Kind != Kind("lambda") {
		t.Errorf("got kind %v, want lambda", got.Kind)
	}
}

func TestDetectNone(t *testing.T) {
	got, err := Detect(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindNone {
		t.Errorf("got %v, want none", got.Kind)
	}
}
