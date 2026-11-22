package scaffold

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitExistingProjectPromptRefusal(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "krewire.yaml")
	if err := os.WriteFile(yamlPath, []byte("custom: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// User declines reset
	_, err := Init(InitOptions{
		Dir:          dir,
		Name:         "demo",
		Force:        false,
		ConfirmReset: func() bool { return false },
	})
	if err != nil {
		t.Fatal(err)
	}

	assertFileContains(t, yamlPath, "custom: true")
}

func TestInitExistingProjectResetWithConfirmation(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "krewire.yaml")
	if err := os.WriteFile(yamlPath, []byte("custom: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// User accepts reset
	_, err := Init(InitOptions{
		Dir:          dir,
		Name:         "demo",
		Force:        false,
		ConfirmReset: func() bool { return true },
	})
	if err != nil {
		t.Fatal(err)
	}

	assertFileContains(t, yamlPath, "name: demo")
	assertFileNotContains(t, yamlPath, "custom: true")
}

func TestInitExistingProjectResetWithForce(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "krewire.yaml")
	if err := os.WriteFile(yamlPath, []byte("custom: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Force reset
	_, err := Init(InitOptions{
		Dir:   dir,
		Name:  "demo",
		Force: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertFileContains(t, yamlPath, "name: demo")
	assertFileNotContains(t, yamlPath, "custom: true")
}

func TestNewCreatesBootstrap(t *testing.T) {
	parent := t.TempDir()
	created, err := New(Options{Name: "demo", Dir: parent})
	if err != nil {
		t.Fatal(err)
	}
	// Bootstrap creates 5 entries: .git, .gitignore, krewire.yaml, go.mod, main.go.
	if len(created) < 4 {
		t.Fatalf("created %d files, want at least 4: %v", len(created), created)
	}

	assertFileContains(t, filepath.Join(parent, "demo", "go.mod"), "module demo")
	assertFileContains(t, filepath.Join(parent, "demo", "go.mod"), "go "+GoVersion)
	// The scaffolded directive must never be older than the workspace modules
	// it requires, otherwise `go build` rejects the generated module.
	if lessThan(GoVersion, requiredGoBaseline(t)) {
		t.Errorf("GoVersion = %q is older than the krew module baseline %q", GoVersion, requiredGoBaseline(t))
	}
	assertFileNotContains(t, filepath.Join(parent, "demo", "go.mod"), "github.com/krewire/krewire/pkg/kern")
	assertFileContains(t, filepath.Join(parent, "demo", "krewire.yaml"), "name: demo")
	assertFileContains(t, filepath.Join(parent, "demo", "main.go"), "package main")
	assertFileNotContains(t, filepath.Join(parent, "demo", "main.go"), "github.com/krewire/krewire/pkg/kern")
	assertFileContains(t, filepath.Join(parent, "demo", ".gitignore"), "/demo")
}

func TestNewNestedPathResolution(t *testing.T) {
	parent := t.TempDir()
	created, err := New(Options{Name: "services/auth", Dir: parent})
	if err != nil {
		t.Fatal(err)
	}
	if len(created) < 4 {
		t.Fatalf("created %d files, want at least 4", len(created))
	}
	assertFileContains(t, filepath.Join(parent, "services", "auth", "krewire.yaml"), "name: auth")
	assertFileContains(t, filepath.Join(parent, "services", "auth", "go.mod"), "module auth")
	if _, err := os.Stat(filepath.Join(parent, "services", "auth", ".git")); err != nil {
		t.Errorf("expected .git in nested target: %v", err)
	}
}

func TestInitNestedPathResolution(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, "deep", "nested", "site")
	created, err := Init(InitOptions{Dir: target})
	if err != nil {
		t.Fatal(err)
	}
	if len(created) < 2 {
		t.Fatalf("created %d files, want at least 2", len(created))
	}
	assertFileContains(t, filepath.Join(target, "krewire.yaml"), "name: site")
	assertFileContains(t, filepath.Join(target, ".gitignore"), "/site")
}

func TestNewModuleOverride(t *testing.T) {
	parent := t.TempDir()
	if _, err := New(Options{Name: "demo", Dir: parent, Module: "github.com/acme/demo"}); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(parent, "demo", "go.mod"), "module github.com/acme/demo")
}

func TestNewRefusesNonEmptyDirectory(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "demo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := New(Options{Name: "demo", Dir: parent}); !errors.Is(err, ErrProjectExists) {
		t.Errorf("New() error = %v, want ErrProjectExists", err)
	}
}

func TestNewInvalidName(t *testing.T) {
	parent := t.TempDir()
	if _, err := New(Options{Name: "invalid:name", Dir: parent}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("New() error = %v, want ErrInvalidName", err)
	}
}

// newBootstrap creates a bootstrap project named "demo" inside parent and returns its path.
func newBootstrap(t *testing.T, parent string) string {
	t.Helper()
	if _, err := New(Options{Name: "demo", Dir: parent, Module: "example.com/demo"}); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(parent, "demo")
}

func TestEquipApp(t *testing.T) {
	dir := newBootstrap(t, t.TempDir())
	created, err := Equip(EquipOptions{
		Dir:     dir,
		Variant: VariantApp,
		Name:    "demo",
		Module:  "example.com/demo",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(dir, "go.mod"), "module example.com/demo")
	assertFileContains(t, filepath.Join(dir, "main.go"), "config.LoadMetadata")
	assertFileContains(t, filepath.Join(dir, "main.go"), "var Signature = \"build with krewire\"")
	assertFileNotContains(t, filepath.Join(dir, "main.go"), "cfg.yaml")
	assertFileContains(t, filepath.Join(dir, "internal/app/app.go"), "func New(meta *config.Metadata) (*rvweb.App, error)")
	for _, path := range []string{
		"internal/config/config.go",
		"internal/http/http.go",
		"web/layouts/shell.go",
		"web/pages/pages.go",
		"web/theme/theme.go",
		"assets/embed.go",
		"assets/public/app.css",
		"assets/public/app.js",
		"README.md",
		".gitignore",
		".env.example",
		".env",
	} {
		found := false
		for _, c := range created {
			if c == path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("created files missing %s: %v", path, created)
		}
	}
	assertFileContains(t, filepath.Join(dir, ".gitignore"), ".env")
	assertFileContains(t, filepath.Join(dir, ".env.example"), "APP_NAME=demo")
	assertFileContains(t, filepath.Join(dir, "internal/config/config.go"), "func Env(key, fallback string) string")
}

func TestEquipAppRefusesNonBootstrapOverwrite(t *testing.T) {
	dir := newBootstrap(t, t.TempDir())
	if err := os.MkdirAll(filepath.Join(dir, "internal/app"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "internal/app/app.go"), []byte("user code"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Equip(EquipOptions{Dir: dir, Variant: VariantApp, Name: "demo", Module: "example.com/demo"})
	if !errors.Is(err, ErrConflict) {
		t.Errorf("Equip() error = %v, want ErrConflict", err)
	}
}

func TestEquipAppMissingModule(t *testing.T) {
	dir := t.TempDir()
	if _, err := Equip(EquipOptions{Dir: dir, Variant: VariantApp}); err == nil {
		t.Error("Equip() = nil error, want module error")
	}
}

func TestEquipCLI(t *testing.T) {
	dir := newBootstrap(t, t.TempDir())
	created, err := Equip(EquipOptions{
		Dir:     dir,
		Variant: VariantCLI,
		Name:    "demo",
		Module:  "example.com/demo",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(dir, "go.mod"), "module example.com/demo")
	assertFileContains(t, filepath.Join(dir, "main.go"), "var Signature = \"build with krewire\"")
	assertFileContains(t, filepath.Join(dir, "main.go"), "tui.NewApp")
	assertFileContains(t, filepath.Join(dir, "main.go"), "internal/commands")
	assertFileContains(t, filepath.Join(dir, "internal/commands/commands.go"), "func Hello(_ *flag.FlagSet) errs.ExitCode")
	assertFileContains(t, filepath.Join(dir, "internal/commands/commands.go"), "hello, demo")
	for _, path := range []string{"README.md", ".gitignore"} {
		found := false
		for _, c := range created {
			if c == path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("created files missing %s: %v", path, created)
		}
	}
	if len(created) != 7 {
		t.Fatalf("equip cli created %d files, want 7: %v", len(created), created)
	}
}

func TestEquipCLIMissingModule(t *testing.T) {
	dir := t.TempDir()
	if _, err := Equip(EquipOptions{Dir: dir, Variant: VariantCLI, Name: "demo"}); err == nil {
		t.Error("Equip() = nil error, want module error")
	}
}

func TestEquipStatic(t *testing.T) {
	dir := newBootstrap(t, t.TempDir())
	created, err := Equip(EquipOptions{Dir: dir, Variant: VariantStatic, Name: "demo", Title: "My Site"})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(dir, "pages/index.kiw"), "My Site")
	assertFileContains(t, filepath.Join(dir, "layouts/Base.kiw"), "{{.Content}}")
	assertFileContains(t, filepath.Join(dir, "components/Hero.kiw"), "Getting Started")
	if _, err := os.Stat(filepath.Join(dir, "public/favicon.svg")); err != nil {
		t.Errorf("public/favicon.svg should exist: %v", err)
	}
	// No ssg.yaml is produced.
	if _, err := os.Stat(filepath.Join(dir, "ssg.yaml")); !os.IsNotExist(err) {
		t.Error("ssg.yaml should not exist")
	}
	// The bootstrap main.go is removed so static pages are previewed cleanly.
	if _, err := os.Stat(filepath.Join(dir, "main.go")); !os.IsNotExist(err) {
		t.Error("bootstrap main.go should be removed for a static project")
	}
	if len(created) != 6 {
		t.Fatalf("equip static created %d files, want 6: %v", len(created), created)
	}
}

func TestEquipBook(t *testing.T) {
	dir := newBootstrap(t, t.TempDir())
	created, err := Equip(EquipOptions{Dir: dir, Variant: VariantBook, Name: "demo", Title: "My Book"})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(dir, "krewire.yaml"), "title: My Book")
	assertFileContains(t, filepath.Join(dir, "content/docs/01-introduction.md"), "# Introduction")
	if _, err := os.Stat(filepath.Join(dir, "main.go")); !os.IsNotExist(err) {
		t.Error("bootstrap main.go should be removed for a book project")
	}
	if len(created) == 0 {
		t.Fatal("equip book created no files")
	}
}

func TestEquipStaticKeepsModifiedMain(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\n// user main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Equip(EquipOptions{Dir: dir, Variant: VariantStatic, Name: "demo"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); err != nil {
		t.Fatalf("user-modified main.go should be kept: %v", err)
	}
}

func TestEquipUnknownVariant(t *testing.T) {
	dir := t.TempDir()
	if _, err := Equip(EquipOptions{Dir: dir, Variant: Variant("nope")}); err == nil {
		t.Error("Equip() = nil error, want unknown variant error")
	}
}

func TestEquipTemplateRequiresEmptyDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Equip(EquipOptions{Dir: dir, Variant: VariantTemplate, TemplateURL: "https://example.com/x.git"})
	if !errors.Is(err, ErrNotEmpty) {
		t.Errorf("Equip() error = %v, want ErrNotEmpty", err)
	}
}

func TestEquipTemplateMissingURL(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "empty")
	if _, err := Equip(EquipOptions{Dir: dir, Variant: VariantTemplate}); err == nil {
		t.Error("Equip() = nil error, want missing URL error")
	}
}

func TestEquipTemplateClonesLocalRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	src := t.TempDir()
	if err := gitInitCommit(src); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "clone")
	created, err := Equip(EquipOptions{Dir: dir, Variant: VariantTemplate, TemplateURL: filepath.ToSlash(src)})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(dir, "hello.txt"), "from template")
	if len(created) == 0 {
		t.Fatal("clone created no files")
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), want) {
		t.Errorf("%s does not contain %q", filepath.Base(path), want)
	}
}

func assertFileNotContains(t *testing.T, path, notWant string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), notWant) {
		t.Errorf("%s should not contain %q", filepath.Base(path), notWant)
	}
}

// gitInitCommit creates a tiny git repository whose tracked files are listed
// by walkCreated.
func gitInitCommit(dir string) error {
	for _, f := range []struct{ name, body string }{
		{"hello.txt", "from template"},
	} {
		if err := os.WriteFile(filepath.Join(dir, f.name), []byte(f.body), 0o644); err != nil {
			return err
		}
	}
	cmd := exec.Command("git", "-C", dir, "init", "-q")
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "-C", dir, "add", ".")
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "-C", dir, "-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-q", "-m", "init")
	return cmd.Run()
}

// requiredGoBaseline is the `go` directive declared by the krew module itself,
// read from ../../go.mod. A generated project depends on the pkg/* packages,
// so the scaffolded directive must be at least this new. Reading the
// real file means raising the baseline elsewhere cannot silently desync.
func requiredGoBaseline(t *testing.T) string {
	t.Helper()
	dir := "."
	for i := 0; i < 6; i++ {
		path := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(path)
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "go "); ok {
					return strings.TrimSpace(rest)
				}
			}
			t.Fatalf("%s has no go directive", path)
		}
		dir = filepath.Join("..", dir)
	}
	t.Fatal("could not find go.mod in parent directories")
	return ""
}

// lessThan reports whether Go version a is older than b, comparing dotted
// numeric components and ignoring any patch suffix. Missing components count
// as zero, so "1.27" < "1.27.1".
func lessThan(a, b string) bool {
	fa, fb := goComponents(a), goComponents(b)
	for i := 0; i < 3; i++ {
		if fa[i] != fb[i] {
			return fa[i] < fb[i]
		}
	}
	return false
}

// goComponents splits a dotted Go version into three numeric components,
// returning zeros for absent parts. A non-numeric part becomes zero.
func goComponents(v string) [3]int {
	var out [3]int
	for i, part := range strings.SplitN(strings.TrimSpace(v), ".", 3) {
		n := 0
		for _, r := range part {
			if r < '0' || r > '9' {
				n = 0
				break
			}
			n = n*10 + int(r-'0')
		}
		out[i] = n
	}
	return out
}

func TestNew_SideBySideInNodeJSProject(t *testing.T) {
	nodeDir := t.TempDir()

	// 1. Setup Node.js project
	pkgJSON := `{"name": "my-node-app", "version": "1.0.0"}`
	if err := os.WriteFile(filepath.Join(nodeDir, "package.json"), []byte(pkgJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	hostGitignore := "node_modules/\ndist/\n"
	if err := os.WriteFile(filepath.Join(nodeDir, ".gitignore"), []byte(hostGitignore), 0o644); err != nil {
		t.Fatal(err)
	}

	// 2. Scaffold Krewire service side-by-side in api/
	created, err := New(Options{
		Name:   "api",
		Dir:    nodeDir,
		Module: "my-node-app/api",
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if len(created) == 0 {
		t.Fatal("expected files to be created")
	}

	// 3. Verify Node.js files remain untouched
	assertFileContains(t, filepath.Join(nodeDir, "package.json"), "my-node-app")
	assertFileContains(t, filepath.Join(nodeDir, ".gitignore"), "node_modules/")

	// 4. Verify Krewire files created inside api/ subdirectory
	assertFileContains(t, filepath.Join(nodeDir, "api", "go.mod"), "module my-node-app/api")
	assertFileContains(t, filepath.Join(nodeDir, "api", "krewire.yaml"), "name: api")
	assertFileContains(t, filepath.Join(nodeDir, "api", ".gitignore"), ".krewire")
}

func TestNew_SideBySideInLaravelProject(t *testing.T) {
	laraDir := t.TempDir()

	// 1. Setup Laravel project
	compJSON := `{"name": "laravel/laravel", "require": {"php": "^8.2"}}`
	if err := os.WriteFile(filepath.Join(laraDir, "composer.json"), []byte(compJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(laraDir, "artisan"), []byte("#!/usr/bin/env php"), 0o755); err != nil {
		t.Fatal(err)
	}
	hostGitignore := "/vendor\n.env\n"
	if err := os.WriteFile(filepath.Join(laraDir, ".gitignore"), []byte(hostGitignore), 0o644); err != nil {
		t.Fatal(err)
	}

	// 2. Scaffold Krewire worker side-by-side in worker/
	created, err := New(Options{
		Name:   "worker",
		Dir:    laraDir,
		Module: "laravel-app/worker",
	})
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if len(created) == 0 {
		t.Fatal("expected files to be created")
	}

	// 3. Verify Laravel files remain untouched
	assertFileContains(t, filepath.Join(laraDir, "composer.json"), "laravel/laravel")
	assertFileContains(t, filepath.Join(laraDir, ".gitignore"), "/vendor")

	// 4. Verify Krewire files created inside worker/ subdirectory
	assertFileContains(t, filepath.Join(laraDir, "worker", "go.mod"), "module laravel-app/worker")
	assertFileContains(t, filepath.Join(laraDir, "worker", "krewire.yaml"), "name: worker")
	assertFileContains(t, filepath.Join(laraDir, "worker", ".gitignore"), ".krewire")
}

func TestEquip_Worker(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":       "module testworker\n\ngo 1.22\n",
		"main.go":      bootstrapBody,
		"krewire.yaml": "project:\n  name: testworker\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	created, err := Equip(EquipOptions{
		Dir:     dir,
		Name:    "testworker",
		Module:  "testworker",
		Variant: VariantWorker,
	})
	if err != nil {
		t.Fatalf("Equip(VariantWorker) failed: %v", err)
	}
	if len(created) == 0 {
		t.Fatal("expected files created")
	}

	assertFileContains(t, filepath.Join(dir, "main.go"), "cloud/worker")
	assertFileContains(t, filepath.Join(dir, "internal", "jobs", "jobs.go"), "ExampleJob")
}

func TestEquip_Service(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":       "module testsvc\n\ngo 1.22\n",
		"main.go":      bootstrapBody,
		"krewire.yaml": "project:\n  name: testsvc\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	created, err := Equip(EquipOptions{
		Dir:     dir,
		Name:    "testsvc",
		Module:  "testsvc",
		Variant: VariantService,
	})
	if err != nil {
		t.Fatalf("Equip(VariantService) failed: %v", err)
	}
	if len(created) == 0 {
		t.Fatal("expected files created")
	}

	assertFileContains(t, filepath.Join(dir, "main.go"), "cloud/service/registry")
	assertFileContains(t, filepath.Join(dir, "internal", "handler", "handler.go"), "func Health")
}

func TestEquip_Gateway(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":       "module testgw\n\ngo 1.22\n",
		"main.go":      bootstrapBody,
		"krewire.yaml": "project:\n  name: testgw\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	created, err := Equip(EquipOptions{
		Dir:     dir,
		Name:    "testgw",
		Module:  "testgw",
		Variant: VariantGateway,
	})
	if err != nil {
		t.Fatalf("Equip(VariantGateway) failed: %v", err)
	}
	if len(created) == 0 {
		t.Fatal("expected files created")
	}

	assertFileContains(t, filepath.Join(dir, "krewire.yaml"), "gateway:")
	assertFileContains(t, filepath.Join(dir, "main.go"), "pkg/web/gateway")
	assertFileContains(t, filepath.Join(dir, "main.go"), "registry.NewMemoryRegistry()")
	assertFileContains(t, filepath.Join(dir, "internal", "config", "config.go"), "func (c *Config) Routes()")
}

func TestEquip_Infra(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"krewire.yaml": "project:\n  name: testinfra\n",
		"main.go":      bootstrapBody,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	created, err := Equip(EquipOptions{
		Dir:     dir,
		Name:    "testinfra",
		Variant: VariantInfra,
	})
	if err != nil {
		t.Fatalf("Equip(VariantInfra) failed: %v", err)
	}
	if len(created) == 0 {
		t.Fatal("expected files created")
	}

	assertFileContains(t, filepath.Join(dir, "infra", "blueprint.go"), "DesiredResources")
	// Verify placeholder main.go was removed for infra
	if _, err := os.Stat(filepath.Join(dir, "main.go")); !os.IsNotExist(err) {
		t.Errorf("expected placeholder main.go to be removed for infra variant")
	}
}
