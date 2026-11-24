package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"text/tabwriter"
	"time"

	"github.com/krewire/hub/sqlite"
	"github.com/krewire/krew/core/config"
	"github.com/krewire/krewire/errs"
	"github.com/krewire/krewire/orm"
	"github.com/krewire/krewire/orm/migrate"
)

// openMigrator opens the database and loads the migrations directory. kiw
// links only the SQLite driver; other drivers must be registered by the
// process, which for krew means they are not available.
func openMigrator(root string, db config.DatabaseConfig, timeout time.Duration) (*migrate.Migrator, func(), error) {
	if db.DSN == "" {
		return nil, nil, errs.WithHint(errors.New("no database configured"),
			"set database.dsn in krewire.yaml, APP_DATABASE_DSN in the environment or .env, or pass --dsn")
	}
	dir := db.Migrations
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return nil, nil, errs.WithHint(fmt.Errorf("migrations directory %s not found", dir),
			"create a migration first: krew migrate create <name>")
	}

	var conn *orm.DB
	var err error
	switch {
	case db.IsSQLite():
		conn, err = sqlite.Open(db.SQLitePath(root))
	case slices.Contains(sql.Drivers(), db.Driver):
		conn, err = orm.Open(db.Driver, db.DSN)
	default:
		return nil, nil, errs.WithHint(fmt.Errorf("database driver %q is not built into kiw", db.Driver),
			"krew bundles only the SQLite driver; for "+db.Driver+", import your driver and call migrate.Up from pkg/orm/migrate in your app")
	}
	if err != nil {
		return nil, nil, err
	}
	m, err := migrate.New(conn, os.DirFS(dir), migrate.Options{LockTimeout: timeout})
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	return m, func() { _ = conn.Close() }, nil
}

// runMigrateCommand runs one migrate sub-command and prints its outcome.
func runMigrateCommand(ctx context.Context, w io.Writer, m *migrate.Migrator, sub string, steps int) errs.ExitCode {
	var results []migrate.Result
	var err error
	switch sub {
	case "status":
		return printMigrateStatus(ctx, w, m)
	case "up":
		results, err = m.Up(ctx, steps)
	case "down":
		results, err = m.Down(ctx, steps)
	case "redo":
		results, err = m.Redo(ctx)
	}
	for _, r := range results {
		verb := "applied"
		if r.Direction == migrate.DirectionDown {
			verb = "rolled back"
		}
		fmt.Fprintf(w, "%s %s (%s)\n", verb, r.Migration, roundDuration(r.Duration))
	}
	if err != nil {
		printMigrateError(os.Stderr, err)
		return errs.ExitCodeFailure
	}
	if len(results) == 0 {
		if sub == "up" {
			fmt.Fprintln(w, "nothing to migrate: the database is up to date")
		} else {
			fmt.Fprintln(w, "nothing to roll back: no migration is applied")
		}
	}
	return errs.ExitCodeSuccess
}

// printMigrateStatus prints one line per migration. It fails when an applied
// migration was edited or deleted, so CI can gate on it.
func printMigrateStatus(ctx context.Context, w io.Writer, m *migrate.Migrator) errs.ExitCode {
	st, err := m.Status(ctx)
	if err != nil {
		printMigrateError(os.Stderr, err)
		return errs.ExitCodeFailure
	}
	if len(st) == 0 {
		fmt.Fprintln(w, "no migrations yet: krew migrate create <name>")
		return errs.ExitCodeSuccess
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "STATE\tMIGRATION\tAPPLIED AT")
	code := errs.ExitCodeSuccess
	for _, s := range st {
		at := "-"
		if !s.AppliedAt.IsZero() {
			at = s.AppliedAt.UTC().Format(time.RFC3339)
		}
		fmt.Fprintf(tw, "%s\t%s_%s\t%s\n", s.State, s.Version, s.Name, at)
		if s.State == migrate.StateModified || s.State == migrate.StateMissing {
			code = errs.ExitCodeFailure
		}
	}
	_ = tw.Flush()
	return code
}

func printMigrateError(w io.Writer, err error) {
	fmt.Fprintln(w, "krew migrate: "+err.Error())
	if hint := errs.HintOf(err); hint != "" {
		fmt.Fprintln(w, "hint: "+hint)
	}
}

// roundDuration keeps sub-millisecond timings readable instead of "0s".
func roundDuration(d time.Duration) time.Duration {
	if d < time.Millisecond {
		return d.Round(time.Microsecond)
	}
	return d.Round(time.Millisecond)
}
