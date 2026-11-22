package commands

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/krewire/krew/internal/scaffold"
	"github.com/krewire/krewire/pkg/errs"
)

func RegisterNew(fs *flag.FlagSet) {
	fs.String("module", "", "module path for the new project (defaults to the project name)")
	fs.String("dir", "", "directory to create the project in (defaults to the current directory)")
	fs.Bool("site", false, "equip a static site (pages/, layouts/, components/)")
	fs.Bool("book", false, "equip a content book (pkg/book)")
	fs.Bool("cli", false, "equip a command-line application (pkg/tui)")
	fs.Bool("app", false, "equip a fullstack monolith application")
	fs.Bool("worker", false, "equip a background job queue worker (cloud/worker)")
	fs.Bool("service", false, "equip a standalone microservice (cloud/service)")
	fs.Bool("gateway", false, "equip an API gateway (pkg/web/gateway)")
	fs.Bool("infra", false, "equip a cloud infrastructure blueprint (cloud/infra)")
	fs.String("template", "", "bootstrap from a remote git template (git URL)")
	fs.String("title", "", "site title for the site and book variants")
	fs.Bool("no-input", false, "never prompt; fail with usage when the project name is missing")
}

// RunNew creates a new Krewire project: initializing minimal ecosystem files and scaffolding.
func RunNew(fs *flag.FlagSet) errs.ExitCode {
	name := fs.Arg(0)
	if name == "" && shouldPrompt(fs) {
		r := bufio.NewReader(promptIn)
		name = ask(r, "Project name", "")
	}
	if name == "" {
		fmt.Fprintln(os.Stderr, "usage: krew new <project> [--template <git-url>]")
		fmt.Fprintln(os.Stderr, "  krew new creates a new directory; to set up the current directory use 'krew init'")
		fmt.Fprintln(os.Stderr, "  run it in a terminal without arguments to be asked for the name")
		return errs.ExitCodeUsage
	}

	site := flagBool(fs, "site")
	book := flagBool(fs, "book")
	cli := flagBool(fs, "cli")
	app := flagBool(fs, "app")
	worker := flagBool(fs, "worker")
	service := flagBool(fs, "service")
	gateway := flagBool(fs, "gateway")
	infra := flagBool(fs, "infra")
	templateURL := flagValue(fs, "template")

	if count := boolCount(site, book, cli, app, worker, service, gateway, infra, templateURL != ""); count > 1 {
		fmt.Fprintln(os.Stderr, "krew new: choose at most one variant: --site, --book, --cli, --app, --worker, --service, --gateway, --infra, or --template")
		return errs.ExitCodeUsage
	}

	parentDir := flagValue(fs, "dir")
	cleanedPath := filepath.Clean(name)
	projectName := filepath.Base(cleanedPath)
	if projectName == "." || projectName == "/" || projectName == "" {
		projectName = "krewire-project"
	}

	modulePath := flagValue(fs, "module")
	if modulePath == "" {
		modulePath = projectName
	}

	var targetDir string
	if filepath.IsAbs(cleanedPath) {
		targetDir = cleanedPath
	} else if parentDir != "" {
		targetDir = filepath.Join(parentDir, cleanedPath)
	} else {
		targetDir = cleanedPath
	}

	// 1. Run minimal Init (git init, .gitignore, krewire.yaml) in the target directory
	if _, err := scaffold.New(scaffold.Options{
		Name:   name,
		Dir:    parentDir,
		Module: modulePath,
	}); err != nil {
		return commandError(err)
	}

	// 2. Run scaffolding (default is fullstack monolith VariantApp)
	opts := scaffold.EquipOptions{
		Dir:         targetDir,
		Name:        projectName,
		Title:       flagValue(fs, "title"),
		TemplateURL: templateURL,
	}
	switch {
	case site:
		opts.Variant = scaffold.VariantStatic
		opts.Name = name
		opts.Title = firstNonEmpty(opts.Title, name)
	case book:
		opts.Variant = scaffold.VariantBook
		opts.Name = name
		opts.Title = firstNonEmpty(opts.Title, name)
	case cli:
		opts.Variant = scaffold.VariantCLI
		opts.Module = modulePath
		opts.Name = name
		opts.Version = resolveVersion()
	case worker:
		opts.Variant = scaffold.VariantWorker
		opts.Module = modulePath
		opts.Name = name
		opts.Version = resolveVersion()
	case service:
		opts.Variant = scaffold.VariantService
		opts.Module = modulePath
		opts.Name = name
		opts.Version = resolveVersion()
	case gateway:
		opts.Variant = scaffold.VariantGateway
		opts.Module = modulePath
		opts.Name = name
		opts.Version = resolveVersion()
	case infra:
		opts.Variant = scaffold.VariantInfra
		opts.Name = name
	case templateURL != "":
		opts.Variant = scaffold.VariantTemplate
	default: // app fullstack monolith scaffolding
		opts.Variant = scaffold.VariantApp
		opts.Module = modulePath
		opts.Name = name
		opts.Version = resolveVersion()
	}

	eqCreated, err := scaffold.Equip(opts)
	if err != nil {
		if isScaffoldUsage(err) {
			return commandError(err)
		}
		return fail(err)
	}

	slog.Info("scaffolded Krewire project", "name", name, "variant", opts.Variant, "files", len(eqCreated))
	for _, path := range eqCreated {
		fmt.Println("created " + filepath.Join(name, path))
	}
	fmt.Printf("\nNext steps:\n  cd %s\n  krew dev\n", name)
	return errs.ExitCodeSuccess
}
