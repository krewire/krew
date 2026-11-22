package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/orm/sqlite"
)

func runMigrate(t *testing.T, args ...string) (string, errs.ExitCode) {
	t.Helper()
	return captureStdout(t, func() errs.ExitCode {
		fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
		RegisterMigrate(fs)
		if err := fs.Parse(args); err != nil {
			t.Fatal(err)
		}
		return RunMigrate(fs)
	})
}

// fillMigration replaces the generated placeholder sections of the newest
// migration whose name ends in suffix.
func fillMigration(t *testing.T, dir, suffix, up, down string) {
	t.Helper()
	matches, _ := filepath.Glob(filepath.Join(dir, "*_"+suffix+".sql"))
	if len(matches) != 1 {
		t.Fatalf("expected one *_%s.sql in %s, got %v", suffix, dir, matches)
	}
	src, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "-- +up") || !strings.Contains(string(src), "-- +down") {
		t.Fatalf("generated migration lacks markers:\n%s", src)
	}
	text := strings.Replace(string(src), "-- Write your UP migration SQL here:", up, 1)
	text = strings.Replace(text, "-- Write your DOWN migration SQL here:", down, 1)
	writeFile(t, matches[0], text)
}

// TestMigrate_EndToEndInNewApp scaffolds a project with `krew new --app`,
// creates migrations with `krew migrate create`, and drives up, status, down,
// and redo against the SQLite file named in krewire.yaml.
func TestMigrate_EndToEndInNewApp(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	fsNew := flag.NewFlagSet("new", flag.ContinueOnError)
	RegisterNew(fsNew)
	_ = fsNew.Parse([]string{"--app", "--no-input", "shop"})
	if _, code := captureStdout(t, func() errs.ExitCode { return RunNew(fsNew) }); code != errs.ExitCodeSuccess {
		t.Fatalf("krew new --app = %v", code)
	}
	project := filepath.Join(dir, "shop")
	chdir(t, project)

	if _, code := runMigrate(t, "up"); code != errs.ExitCodeUsage {
		t.Fatalf("up without a configured database = %v, want usage", code)
	}

	yaml, err := os.ReadFile(filepath.Join(project, "krewire.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(project, "krewire.yaml"), string(yaml)+"\ndatabase:\n  driver: sqlite\n  dsn: data/shop.db\n")
	if err := os.MkdirAll(filepath.Join(project, "data"), 0o755); err != nil {
		t.Fatal(err)
	}

	if out, code := runMigrate(t, "create", "create_products.sql"); code != errs.ExitCodeSuccess || !strings.Contains(out, "created database/migrations/") {
		t.Fatalf("create = %v %q", code, out)
	}
	if _, code := runMigrate(t, "create", "add_price.sql"); code != errs.ExitCodeSuccess {
		t.Fatalf("second create = %v", code)
	}
	migrations := filepath.Join(project, "database", "migrations")
	fillMigration(t, migrations, "create_products", "CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT NOT NULL);", "DROP TABLE products;")
	fillMigration(t, migrations, "add_price", "ALTER TABLE products ADD COLUMN price INTEGER NOT NULL DEFAULT 0;", "ALTER TABLE products DROP COLUMN price;")

	out, code := runMigrate(t, "status")
	if code != errs.ExitCodeSuccess || strings.Count(out, "pending") != 2 {
		t.Fatalf("status before up = %v\n%s", code, out)
	}
	if out, code = runMigrate(t, "up"); code != errs.ExitCodeSuccess || strings.Count(out, "applied ") != 2 {
		t.Fatalf("up = %v\n%s", code, out)
	}
	if out, _ = runMigrate(t, "up"); !strings.Contains(out, "up to date") {
		t.Fatalf("second up = %q", out)
	}

	db, err := sqlite.Open(filepath.Join(project, "data", "shop.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(context.Background(), "INSERT INTO products (name, price) VALUES ('tea', 5)"); err != nil {
		t.Fatalf("schema from krew migrate up is not usable by the app: %v", err)
	}

	if out, code = runMigrate(t, "redo"); code != errs.ExitCodeSuccess || !strings.Contains(out, "rolled back") || !strings.Contains(out, "add_price") {
		t.Fatalf("redo = %v\n%s", code, out)
	}
	if out, code = runMigrate(t, "down", "2"); code != errs.ExitCodeSuccess || strings.Count(out, "rolled back") != 2 {
		t.Fatalf("down 2 = %v\n%s", code, out)
	}
	if out, _ = runMigrate(t, "down"); !strings.Contains(out, "nothing to roll back") {
		t.Fatalf("down on an empty history = %q", out)
	}

	if _, code = runMigrate(t, "up", "1"); code != errs.ExitCodeSuccess {
		t.Fatalf("up 1 = %v", code)
	}
	edited, _ := filepath.Glob(filepath.Join(migrations, "*_create_products.sql"))
	src, _ := os.ReadFile(edited[0])
	writeFile(t, edited[0], strings.Replace(string(src), "name TEXT NOT NULL", "name TEXT NOT NULL, sku TEXT", 1))
	if out, code = runMigrate(t, "status"); code != errs.ExitCodeFailure || !strings.Contains(out, "modified") {
		t.Fatalf("status over an edited migration = %v\n%s", code, out)
	}
	if _, code = runMigrate(t, "up"); code != errs.ExitCodeFailure {
		t.Fatalf("up over an edited migration = %v, want failure", code)
	}
}

func findLocalMonorepo() string {
	cur, _ := os.Getwd()
	for {
		if data, err := os.ReadFile(filepath.Join(cur, "go.mod")); err == nil {
			if strings.Contains(string(data), "module github.com/krewire/krewire") {
				return cur
			}
		}
		candidate := filepath.Join(cur, "krewire")
		if data, err := os.ReadFile(filepath.Join(candidate, "go.mod")); err == nil {
			if strings.Contains(string(data), "module github.com/krewire/krewire") {
				return candidate
			}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return ""
}

func TestMigrate_GoMigrations_EndToEnd(t *testing.T) {
	monorepo := findLocalMonorepo()
	dir := t.TempDir()
	chdir(t, dir)
	fsNew := flag.NewFlagSet("new", flag.ContinueOnError)
	RegisterNew(fsNew)
	_ = fsNew.Parse([]string{"--app", "--no-input", "blog"})
	if _, code := captureStdout(t, func() errs.ExitCode { return RunNew(fsNew) }); code != errs.ExitCodeSuccess {
		t.Fatalf("krew new --app = %v", code)
	}
	project := filepath.Join(dir, "blog")
	chdir(t, project)

	if monorepo != "" {
		goModPath := filepath.Join(project, "go.mod")
		data, _ := os.ReadFile(goModPath)
		writeFile(t, goModPath, string(data)+fmt.Sprintf("\nreplace github.com/krewire/krewire => %s\n", monorepo))

		// Copy monorepo go.sum so dependencies are resolved without network in tests
		if sumData, err := os.ReadFile(filepath.Join(monorepo, "go.sum")); err == nil {
			writeFile(t, filepath.Join(project, "go.sum"), string(sumData))
		}
	}

	yaml, err := os.ReadFile(filepath.Join(project, "krewire.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(project, "krewire.yaml"), string(yaml)+"\ndatabase:\n  driver: sqlite\n  dsn: data/blog.db\n")
	if err := os.MkdirAll(filepath.Join(project, "data"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Create Laravel-style Go migration by default (no .sql extension)
	if out, code := runMigrate(t, "create", "create_posts_table"); code != errs.ExitCodeSuccess || !strings.Contains(out, "created database/migrations/") {
		t.Fatalf("create = %v %q", code, out)
	}

	matches, _ := filepath.Glob(filepath.Join(project, "database", "migrations", "*_create_posts_table.go"))
	if len(matches) != 1 {
		t.Fatalf("expected one *_create_posts_table.go, got: %v", matches)
	}

	content, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "s.Create(\"posts\"") {
		t.Fatalf("generated Go migration lacks s.Create: %s", string(content))
	}

	// Status before up: pending
	out, code := runMigrate(t, "status")
	if code != errs.ExitCodeSuccess || !strings.Contains(out, "pending") {
		t.Fatalf("status before up = %v\n%s", code, out)
	}

	// Up: apply Go migration
	out, code = runMigrate(t, "up")
	if code != errs.ExitCodeSuccess || !strings.Contains(out, "applied ") {
		t.Fatalf("up = %v\n%s", code, out)
	}

	// Verify table exists and can be queried
	db, err := sqlite.Open(filepath.Join(project, "data", "blog.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(context.Background(), "INSERT INTO posts (created_at, updated_at) VALUES (datetime('now'), datetime('now'))"); err != nil {
		t.Fatalf("table created by Go migration is not usable: %v", err)
	}

	// Down: rollback Go migration
	out, code = runMigrate(t, "down")
	if code != errs.ExitCodeSuccess || !strings.Contains(out, "rolled back") {
		t.Fatalf("down = %v\n%s", code, out)
	}

	// Redo: rollback and re-apply
	out, code = runMigrate(t, "up")
	if code != errs.ExitCodeSuccess || !strings.Contains(out, "applied") {
		t.Fatalf("re-up = %v\n%s", code, out)
	}
	out, code = runMigrate(t, "redo")
	if code != errs.ExitCodeSuccess || !strings.Contains(out, "rolled back") || !strings.Contains(out, "applied") {
		t.Fatalf("redo = %v\n%s", code, out)
	}
}

func TestMigrate_ConfigPrecedenceAndErrors(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	writeFile(t, filepath.Join(dir, "krewire.yaml"), "project:\n  name: x\ndatabase:\n  dsn: yaml.db\n  migrations: db/migrations\n")
	writeFile(t, filepath.Join(dir, ".env"), "APP_DATABASE_DSN=env.db\n")
	t.Setenv("APP_DATABASE_DSN", "")
	_ = os.Unsetenv("APP_DATABASE_DSN")

	if out, code := runMigrate(t, "create", "init.sql"); code != errs.ExitCodeSuccess || !strings.Contains(out, "db/migrations/") {
		t.Fatalf("create must honor database.migrations: %v %q", code, out)
	}
	fillMigration(t, filepath.Join(dir, "db", "migrations"), "init", "CREATE TABLE t (id INTEGER);", "DROP TABLE t;")

	if _, code := runMigrate(t, "up"); code != errs.ExitCodeSuccess {
		t.Fatalf("up = %v", code)
	}
	if _, err := os.Stat(filepath.Join(dir, "env.db")); err != nil {
		t.Fatalf(".env APP_DATABASE_DSN must win over krewire.yaml: %v", err)
	}
	if _, code := runMigrate(t, "--dsn", "flag.db", "up"); code != errs.ExitCodeSuccess {
		t.Fatalf("up --dsn = %v", code)
	}
	if _, err := os.Stat(filepath.Join(dir, "flag.db")); err != nil {
		t.Fatalf("--dsn must win over the environment: %v", err)
	}

	for _, args := range [][]string{
		{},
		{"sideways"},
		{"down", "0"},
		{"down", "two"},
		{"status", "3"},
		{"create"},
		{"--driver", "postgres", "--dsn", "postgres://localhost/x", "up"},
		{"--dir", "nope", "up"},
	} {
		if _, code := runMigrate(t, args...); code != errs.ExitCodeUsage {
			t.Errorf("krew migrate %v = %v, want usage", args, code)
		}
	}
}

func TestRoundDuration(t *testing.T) {
	if got := roundDuration(412345 * time.Nanosecond); got != 412*time.Microsecond {
		t.Errorf("sub-millisecond = %s", got)
	}
	if got := roundDuration(1500 * time.Microsecond); got != 2*time.Millisecond {
		t.Errorf("millisecond rounding = %s", got)
	}
}

func TestMigrate_LegacyDirFallback(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	writeFile(t, filepath.Join(dir, "krewire.yaml"), "project:\n  name: legacy\ndatabase:\n  driver: sqlite\n  dsn: legacy.db\n")
	legacyDir := filepath.Join(dir, "migrations")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if out, code := runMigrate(t, "create", "legacy_init.sql"); code != errs.ExitCodeSuccess || !strings.Contains(out, "created migrations/") {
		t.Fatalf("create in legacy project must fallback to migrations/: %v %q", code, out)
	}
}
