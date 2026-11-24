package config

import (
	"os"
	"path/filepath"
	"strings"

	pkgconfig "github.com/krewire/krewire/config"
)

// DefaultDatabaseDriver is the driver `krew migrate` assumes when none is set;
// it is the one driver krew links (pkg/orm/sqlite).
const DefaultDatabaseDriver = "sqlite"

// DefaultMigrationsDir is where `krew make migration` writes migration files.
const DefaultMigrationsDir = "database/migrations"

// DefaultSeedsDir is where `krew make seeder` writes seed files.
const DefaultSeedsDir = "seeds"

// DatabaseConfig is the `database:` section of krewire.yaml.
type DatabaseConfig struct {
	// Driver is the database/sql driver name, e.g. "sqlite" or "postgres".
	Driver string `yaml:"driver"`
	// DSN is the connection string; for SQLite, a file path relative to the
	// project root.
	DSN string `yaml:"dsn"`
	// Migrations is the migrations directory, relative to the project root.
	Migrations string `yaml:"migrations"`
	// Seeds is the seeds directory, relative to the project root.
	Seeds string `yaml:"seeds"`
}

// ResolveDatabase returns the effective database settings with the same
// precedence an app built on pkg/config sees: krewire.yaml < environment
// (APP_DATABASE_DRIVER, APP_DATABASE_DSN, APP_DATABASE_MIGRATIONS) < the
// explicit overrides (command-line flags; empty means unset). Defaults fill
// the driver and the migrations directory; the DSN has no default.
func (c *Config) ResolveDatabase(lookup func(string) (string, bool), override DatabaseConfig) (DatabaseConfig, error) {
	if lookup == nil {
		lookup = os.LookupEnv
	}
	var section struct {
		Database DatabaseConfig `yaml:"database"`
	}
	if c != nil {
		section.Database = c.Database
	}
	if err := pkgconfig.Override(&section, lookup); err != nil {
		return DatabaseConfig{}, err
	}
	db := section.Database
	db.Driver = firstNonEmpty(strings.TrimSpace(override.Driver), strings.TrimSpace(db.Driver), DefaultDatabaseDriver)
	db.DSN = firstNonEmpty(strings.TrimSpace(override.DSN), strings.TrimSpace(db.DSN))
	db.Migrations = firstNonEmpty(strings.TrimSpace(override.Migrations), strings.TrimSpace(db.Migrations), DefaultMigrationsDir)
	db.Seeds = firstNonEmpty(strings.TrimSpace(override.Seeds), strings.TrimSpace(db.Seeds), DefaultSeedsDir)
	return db, nil
}

// IsSQLite reports whether the driver names SQLite under any common spelling.
func (d DatabaseConfig) IsSQLite() bool {
	switch strings.ToLower(d.Driver) {
	case "sqlite", "sqlite3":
		return true
	}
	return false
}

// SQLitePath returns the DSN as a path anchored at root, so `krew migrate`
// opens the same file the app opens when it runs from the project root.
// ":memory:", "file:" URIs and absolute paths are returned unchanged.
func (d DatabaseConfig) SQLitePath(root string) string {
	dsn := d.DSN
	if dsn == "" || dsn == ":memory:" || strings.HasPrefix(dsn, "file:") || filepath.IsAbs(dsn) {
		return dsn
	}
	return filepath.Join(root, dsn)
}
