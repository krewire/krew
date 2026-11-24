package commands

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/krewire/krew/core/config"
	"github.com/krewire/krew/core/scaffold"
	"github.com/krewire/krewire/errs"
)

// RegisterMake registers flags for the make command.
func RegisterMake(fs *flag.FlagSet) {
	fs.String("kind", "", "generator kind: component, page, handler, migration, model, seeder, factory, repository, middleware, service, job, chapter, doc")
	fs.String("name", "", "name for the generated artifact")
}

// RunMake scaffolds new components, pages, handlers, migrations, models, middleware, services, jobs, or chapters.
func RunMake(fs *flag.FlagSet) errs.ExitCode {
	kind := flagValue(fs, "kind")
	name := flagValue(fs, "name")

	// Support positional arguments: krew make <kind> <name>
	args := fs.Args()
	if kind == "" && len(args) > 0 {
		kind = args[0]
		args = args[1:]
	}
	// Support colon syntax e.g. "make:component" or "component:Card"
	if strings.Contains(kind, ":") {
		parts := strings.SplitN(kind, ":", 2)
		if strings.ToLower(parts[0]) == "make" {
			kind = parts[1]
		} else if name == "" && parts[1] != "" {
			kind = parts[0]
			name = parts[1]
		}
	}
	if name == "" && len(args) > 0 {
		name = args[0]
	}

	if kind == "" || name == "" {
		fmt.Fprintln(os.Stderr, "krew make: usage: krew make <kind> <name>")
		fmt.Fprintln(os.Stderr, "supported kinds: component, page, handler, migration, model, seeder, factory, repository, middleware, service, job, chapter, doc")
		fmt.Fprintln(os.Stderr, "example: krew make component Card")
		fmt.Fprintln(os.Stderr, "hint: to start from a ready-made kui component instead, run 'krew ui list' and 'krew add ui/<name>'")
		return errs.ExitCodeUsage
	}

	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "krew: "+err.Error())
		return errs.ExitCodeUsage
	}

	if err := Make(root, kind, name); err != nil {
		fmt.Fprintln(os.Stderr, "krew make: "+err.Error())
		if errors.Is(err, scaffold.ErrUnknownKind) {
			return errs.ExitCodeUsage
		}
		return errs.ExitCodeFailure
	}
	return errs.ExitCodeSuccess
}

// Make executes the scaffolding generator for a given kind and artifact name.
// Migrations go to the directory `krew migrate` reads (database.migrations).
func Make(root, kind, name string) error {
	if strings.EqualFold(kind, "migration") {
		dir := config.DefaultMigrationsDir
		if cfg, err := config.Load(root); err == nil {
			if db, err := cfg.ResolveDatabase(os.LookupEnv, config.DatabaseConfig{}); err == nil {
				dir = db.Migrations
			}
		}
		if dir == config.DefaultMigrationsDir && !hasDir(root, "database/migrations") && hasDir(root, "migrations") {
			dir = "migrations"
		}
		_, err := scaffold.MakeMigration(root, dir, name)
		return err
	}
	return scaffold.MakeArtifact(root, kind, name)
}
