package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krew/internal/scaffold"
	"github.com/krewire/krewire/pkg/errs"
	"github.com/krewire/krewire/pkg/orm/migrate"
)

const migrateUsage = `usage: krew migrate <command> [flags]

commands:
  up [n]          apply pending migrations (all, or the next n)
  down [n]        roll back the last n applied migrations (default 1)
  redo            roll back the last migration and apply it again
  status          list migrations and their state
  create <name>   create database/migrations/<timestamp>_<name>.go (use .sql for raw SQL)

The database comes from --dsn/--driver, else APP_DATABASE_DSN/APP_DATABASE_DRIVER
(.env is loaded), else database.dsn/database.driver in krewire.yaml.`

// RegisterMigrate registers flags for the migrate command.
func RegisterMigrate(fs *flag.FlagSet) {
	fs.String("dsn", "", "database DSN (overrides APP_DATABASE_DSN and database.dsn)")
	fs.String("driver", "", "database/sql driver (overrides APP_DATABASE_DRIVER and database.driver; default sqlite)")
	fs.String("dir", "", "migrations directory (overrides APP_DATABASE_MIGRATIONS and database.migrations; default database/migrations)")
	fs.Duration("lock-timeout", migrate.DefaultLockTimeout, "how long to wait for another migration runner")
}

// RunMigrate applies, rolls back, lists, or creates migrations.
func RunMigrate(fs *flag.FlagSet) errs.ExitCode {
	args := fs.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, migrateUsage)
		return errs.ExitCodeUsage
	}
	sub, args := args[0], args[1:]

	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew: "+err.Error())
		return errs.ExitCodeUsage
	}

	db, err := resolveMigrateDatabase(root, fs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew migrate: "+err.Error())
		return errs.ExitCodeUsage
	}

	switch sub {
	case "create", "new":
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "krew migrate: usage: krew migrate create <name>")
			return errs.ExitCodeUsage
		}
		if _, err := scaffold.MakeMigration(root, db.Migrations, args[0]); err != nil {
			fmt.Fprintln(os.Stderr, "krew migrate: "+err.Error())
			return errs.ExitCodeFailure
		}
		return errs.ExitCodeSuccess
	case "up", "down", "redo", "status":
	default:
		fmt.Fprintf(os.Stderr, "krew migrate: unknown command %q\n\n%s\n", sub, migrateUsage)
		return errs.ExitCodeUsage
	}

	steps, err := migrateSteps(sub, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew migrate: "+err.Error())
		return errs.ExitCodeUsage
	}
	timeout, _ := time.ParseDuration(flagValue(fs, "lock-timeout"))

	dir := db.Migrations
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}

	goFiles, _ := filepath.Glob(filepath.Join(dir, "*_*.go"))
	if len(goFiles) > 0 {
		return runGoMigrations(root, dir, db, sub, steps, timeout)
	}

	m, closeDB, err := openMigrator(root, db, timeout)
	if err != nil {
		printMigrateError(os.Stderr, err)
		return errs.ExitCodeUsage
	}
	defer closeDB()
	return runMigrateCommand(context.Background(), os.Stdout, m, sub, steps)
}

// migrateSteps parses the optional step count of up and down.
func migrateSteps(sub string, args []string) (int, error) {
	if len(args) == 0 {
		if sub == "down" {
			return 1, nil
		}
		return 0, nil
	}
	if (sub != "up" && sub != "down") || len(args) > 1 {
		return 0, fmt.Errorf("unexpected arguments %v for %q", args, sub)
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n < 1 {
		return 0, fmt.Errorf("step count must be a positive number, got %q", args[0])
	}
	return n, nil
}

// resolveMigrateDatabase merges flags, the environment (after .env), and
// krewire.yaml into the effective database settings.
func resolveMigrateDatabase(root string, fs *flag.FlagSet) (config.DatabaseConfig, error) {
	cfg, err := config.Load(root)
	if err != nil {
		return config.DatabaseConfig{}, err
	}
	if err := cfg.LoadDotEnv(root); err != nil {
		return config.DatabaseConfig{}, err
	}
	db, err := cfg.ResolveDatabase(os.LookupEnv, config.DatabaseConfig{
		Driver:     flagValue(fs, "driver"),
		DSN:        flagValue(fs, "dsn"),
		Migrations: flagValue(fs, "dir"),
	})
	if err != nil {
		return config.DatabaseConfig{}, err
	}
	if flagValue(fs, "dir") == "" && (cfg == nil || cfg.Database.Migrations == "") && os.Getenv("APP_DATABASE_MIGRATIONS") == "" {
		if !hasDir(root, "database/migrations") && hasDir(root, "migrations") {
			db.Migrations = "migrations"
		}
	}
	return db, nil
}
