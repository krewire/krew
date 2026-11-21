package commands

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/kiw"
)

func setupTestRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// Create minimal krewire.yaml so findRoot() discovers it
	cfgPath := filepath.Join(dir, "krewire.yaml")
	if err := os.WriteFile(cfgPath, []byte("project:\n  kind: site\n"), 0o644); err != nil {
		t.Fatalf("failed to create krewire.yaml: %v", err)
	}
	return dir
}

func TestMake_Usage(t *testing.T) {
	fs := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fs)

	// No arguments
	code := RunMake(fs)
	if code != errs.ExitCodeUsage {
		t.Errorf("RunMake() with no args = %v, want ExitCodeUsage", code)
	}

	// Unknown kind
	fsUnknown := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fsUnknown)
	_ = fsUnknown.Parse([]string{"unknown_kind", "sample"})
	code = RunMake(fsUnknown)
	if code != errs.ExitCodeUsage {
		t.Errorf("RunMake() with unknown kind = %v, want ExitCodeUsage", code)
	}
}

func TestMake_Component(t *testing.T) {
	root := setupTestRoot(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	_ = os.Chdir(root)

	fs := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fs)
	_ = fs.Parse([]string{"component", "card-item"})

	code := RunMake(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunMake component failed with code %v", code)
	}

	target := filepath.Join(root, "components", "CardItem.kiw")
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected components/CardItem.kiw to exist: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, `{{class "card-item" .Variant "default" .Class}}`) || !strings.Contains(content, "{body}") {
		t.Errorf("expected component class card-item and a {body} slot, got: %s", content)
	}
	if diags := kiw.Lint("CardItem.kiw", content); len(diags) > 0 {
		t.Errorf("generated component does not lint clean: %v", diags)
	}

	// Duplicate component should fail
	fsDup := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fsDup)
	_ = fsDup.Parse([]string{"component", "CardItem"})
	if code := RunMake(fsDup); code != errs.ExitCodeFailure {
		t.Errorf("duplicate component generation should fail, got code %v", code)
	}

	// Colon syntax: make:component Button
	fsColon := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fsColon)
	_ = fsColon.Parse([]string{"make:component", "button"})
	if code := RunMake(fsColon); code != errs.ExitCodeSuccess {
		t.Errorf("make:component button failed, got code %v", code)
	}
	if _, err := os.Stat(filepath.Join(root, "components", "Button.kiw")); err != nil {
		t.Errorf("expected components/Button.kiw to exist: %v", err)
	}
}

func TestMake_Page(t *testing.T) {
	root := setupTestRoot(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	_ = os.Chdir(root)

	fs := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fs)
	_ = fs.Parse([]string{"page", "pricing"})

	code := RunMake(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunMake page failed with code %v", code)
	}

	target := filepath.Join(root, "pages", "pricing.kiw")
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected pages/pricing.kiw to exist: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, `title: "Pricing"`) || !strings.Contains(content, `<div class="pricing-page">`) {
		t.Errorf("unexpected page content: %s", content)
	}
}

func TestMake_Handler(t *testing.T) {
	root := setupTestRoot(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	_ = os.Chdir(root)

	fs := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fs)
	_ = fs.Parse([]string{"handler", "create-user"})

	code := RunMake(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunMake handler failed with code %v", code)
	}

	target := filepath.Join(root, "internal", "handlers", "create_user.go")
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected internal/handlers/create_user.go to exist: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, `func CreateUserHandler`) {
		t.Errorf("expected CreateUserHandler in content: %s", content)
	}
}

func TestMake_Migration(t *testing.T) {
	root := setupTestRoot(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	_ = os.Chdir(root)

	fs := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fs)
	_ = fs.Parse([]string{"--kind", "migration", "--name", "create_teams"})

	code := RunMake(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunMake migration failed with code %v", code)
	}

	entries, err := os.ReadDir(filepath.Join(root, "migrations"))
	if err != nil || len(entries) == 0 {
		t.Fatalf("expected migrations/ directory with created file: %v", err)
	}

	var foundGo string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), "_create_teams.go") {
			foundGo = e.Name()
			break
		}
	}
	if foundGo == "" {
		t.Fatalf("expected migration filename suffix _create_teams.go in %v", entries)
	}

	b, _ := os.ReadFile(filepath.Join(root, "migrations", foundGo))
	content := string(b)
	if !strings.Contains(content, "s.Create(\"teams\"") || !strings.Contains(content, "s.DropIfExists(\"teams\")") {
		t.Errorf("expected Laravel-style schema builder syntax in: %s", content)
	}
}

func TestMake_MigrationSQL(t *testing.T) {
	root := setupTestRoot(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	_ = os.Chdir(root)

	fs := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fs)
	_ = fs.Parse([]string{"--kind", "migration", "--name", "create_items.sql"})

	code := RunMake(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunMake migration failed with code %v", code)
	}

	entries, err := os.ReadDir(filepath.Join(root, "migrations"))
	if err != nil || len(entries) == 0 {
		t.Fatalf("expected migrations/ directory with created file: %v", err)
	}

	var foundSQL string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), "_create_items.sql") {
			foundSQL = e.Name()
			break
		}
	}
	if foundSQL == "" {
		t.Fatalf("expected migration filename suffix _create_items.sql in %v", entries)
	}

	b, _ := os.ReadFile(filepath.Join(root, "migrations", foundSQL))
	content := string(b)
	if !strings.Contains(content, "-- +up") || !strings.Contains(content, "-- +down") {
		t.Errorf("expected SQL migration markers in: %s", content)
	}
}

func TestMake_Model(t *testing.T) {
	root := setupTestRoot(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	_ = os.Chdir(root)

	fs := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fs)
	_ = fs.Parse([]string{"model", "account"})

	code := RunMake(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunMake model failed with code %v", code)
	}

	target := filepath.Join(root, "internal", "models", "account.go")
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected internal/models/account.go to exist: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "type Account struct") || (!strings.Contains(content, "orm.BaseModel") && !strings.Contains(content, `json:"created_at"`)) {
		t.Errorf("expected Account struct with BaseModel or timestamps, got: %s", content)
	}
}

func TestMake_Seeder(t *testing.T) {
	root := setupTestRoot(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	_ = os.Chdir(root)

	fs := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fs)
	_ = fs.Parse([]string{"seeder", "User"})

	code := RunMake(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunMake seeder failed with code %v", code)
	}

	target := filepath.Join(root, "seeds", "user_seeder.go")
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected seeds/user_seeder.go to exist: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "seeder.Register") || !strings.Contains(content, "UserSeeder") {
		t.Errorf("expected UserSeeder content, got: %s", content)
	}
}

func TestMake_Factory(t *testing.T) {
	root := setupTestRoot(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	_ = os.Chdir(root)

	fs := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fs)
	_ = fs.Parse([]string{"factory", "User"})

	code := RunMake(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunMake factory failed with code %v", code)
	}

	target := filepath.Join(root, "factories", "user_factory.go")
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected factories/user_factory.go to exist: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "UserFactory") || !strings.Contains(content, "seeder.NewFactory") {
		t.Errorf("expected UserFactory content, got: %s", content)
	}
}

func TestMake_Repository(t *testing.T) {
	root := setupTestRoot(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	_ = os.Chdir(root)

	fs := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fs)
	_ = fs.Parse([]string{"repository", "User"})

	code := RunMake(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunMake repository failed with code %v", code)
	}

	target := filepath.Join(root, "internal", "repositories", "user_repository.go")
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected internal/repositories/user_repository.go to exist: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "UserRepository") || !strings.Contains(content, "repository.Repository") {
		t.Errorf("expected UserRepository content, got: %s", content)
	}
}

func TestMake_Middleware(t *testing.T) {
	root := setupTestRoot(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	_ = os.Chdir(root)

	fs := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fs)
	_ = fs.Parse([]string{"middleware", "auth"})

	code := RunMake(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunMake middleware failed with code %v", code)
	}

	target := filepath.Join(root, "internal", "middleware", "auth.go")
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected internal/middleware/auth.go to exist: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "func Auth(next http.Handler) http.Handler") {
		t.Errorf("expected Auth middleware function, got: %s", content)
	}
}

func TestMake_Service(t *testing.T) {
	root := setupTestRoot(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	_ = os.Chdir(root)

	fs := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fs)
	_ = fs.Parse([]string{"service", "billing"})

	code := RunMake(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunMake service failed with code %v", code)
	}

	target := filepath.Join(root, "internal", "services", "billing.go")
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected internal/services/billing.go to exist: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "type BillingService interface") || !strings.Contains(content, "NewBillingService()") {
		t.Errorf("expected BillingService interface and constructor, got: %s", content)
	}
}

func TestMake_Chapter(t *testing.T) {
	root := setupTestRoot(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	_ = os.Chdir(root)

	fs := flag.NewFlagSet("make", flag.ContinueOnError)
	RegisterMake(fs)
	_ = fs.Parse([]string{"chapter", "intro"})

	code := RunMake(fs)
	if code != errs.ExitCodeSuccess {
		t.Fatalf("RunMake chapter failed with code %v", code)
	}

	target := filepath.Join(root, "content", "01-intro.md")
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected content/01-intro.md to exist: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "title: \"Intro\"") || !strings.Contains(content, "# Intro") {
		t.Errorf("expected frontmatter and heading in chapter, got: %s", content)
	}
}
