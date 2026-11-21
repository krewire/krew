package config

import (
	"path/filepath"
	"testing"
)

func TestResolveDatabasePrecedence(t *testing.T) {
	config := &Config{Database: DatabaseConfig{Driver: "sqlite", DSN: "app.db"}}
	env := map[string]string{}
	lookup := func(k string) (string, bool) { v, ok := env[k]; return v, ok }

	got, err := config.ResolveDatabase(lookup, DatabaseConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if got != (DatabaseConfig{Driver: "sqlite", DSN: "app.db", Migrations: "migrations", Seeds: "seeds"}) {
		t.Fatalf("yaml + defaults = %+v", got)
	}

	env["APP_DATABASE_DSN"] = "from-env.db"
	env["APP_DATABASE_MIGRATIONS"] = "db/migrations"
	if got, _ = config.ResolveDatabase(lookup, DatabaseConfig{}); got.DSN != "from-env.db" || got.Migrations != "db/migrations" {
		t.Fatalf("env must override krewire.yaml: %+v", got)
	}

	if got, _ = config.ResolveDatabase(lookup, DatabaseConfig{DSN: "flag.db", Driver: "postgres"}); got.DSN != "flag.db" || got.Driver != "postgres" {
		t.Fatalf("flags must override env: %+v", got)
	}

	if got, _ = (*Config)(nil).ResolveDatabase(func(string) (string, bool) { return "", false }, DatabaseConfig{}); got.Driver != DefaultDatabaseDriver || got.DSN != "" {
		t.Fatalf("no config = %+v", got)
	}
}

func TestDatabaseSQLitePath(t *testing.T) {
	root := t.TempDir()
	cases := map[string]string{
		"todo.db":                     filepath.Join(root, "todo.db"),
		":memory:":                    ":memory:",
		"file:x.db?mode=rwc":          "file:x.db?mode=rwc",
		filepath.Join(root, "abs.db"): filepath.Join(root, "abs.db"),
	}
	for dsn, want := range cases {
		if got := (DatabaseConfig{DSN: dsn}).SQLitePath(root); got != want {
			t.Errorf("SQLitePath(%q) = %q, want %q", dsn, got, want)
		}
	}
	if !(DatabaseConfig{Driver: "SQLite3"}).IsSQLite() || (DatabaseConfig{Driver: "postgres"}).IsSQLite() {
		t.Error("IsSQLite misclassified a driver")
	}
}
