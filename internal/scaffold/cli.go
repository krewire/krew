// Package-level file for the command-line application (VariantCLI): kiw
// init --cli scaffolds a CLI built on pkg/tui + the foundation packages, mirroring
// the krew CLI's own layout (root main.go + internal/commands).
package scaffold

import (
	"fmt"
	"os"
)

// equipCLI shapes the bootstrap project into a command-line application. The root
// main.go becomes the tui.App harness, commands live under
// internal/commands, and go.mod is upgraded in place with pinned requires.
func equipCLI(opts EquipOptions) ([]string, error) {
	if opts.Module == "" {
		return nil, fmt.Errorf("equip cli: module path is required (read from go.mod)")
	}
	if opts.Name == "" {
		return nil, fmt.Errorf("equip cli: project name is required")
	}
	version := opts.Version
	if version == "" {
		version = "latest"
	}

	cwd, _ := os.Getwd()
	localPath := detectLocalModulePath(cwd)

	files := []file{
		{goModFile, goModTemplate(opts.Module, version, localPath)},
		{krewireYaml, cliKrewireYamlTemplate(opts.Name)},
		{mainGo, cliMainTemplate(opts.Module, opts.Name)},
		{"internal/commands/commands.go", cliCommandsTemplate(opts.Name)},
		{"internal/config/config.go", cliConfigTemplate()},
		{"README.md", cliReadmeTemplate(opts.Name)},
		{gitignoreFile, gitignoreTemplate(opts.Name)},
	}
	return writeVariant(opts.Dir, files)
}

func cliKrewireYamlTemplate(name string) string {
	return "project:\n  name: " + name + "\n  version: 0.1.0\n"
}

func cliMainTemplate(module, name string) string {
	return fmt.Sprintf(`// Entry point: build the CLI app and run it.
package main

import (
	"fmt"
	"os"

	"github.com/krewire/tui"
	"%s/internal/commands"
	"%s/internal/config"
)

const version = "0.1.0"

// Signature is the ecosystem watermark embedded into the binary at compile time via -ldflags.
var Signature = "build with krewire"

func main() {
	if _, err := config.LoadMetadata("krewire.yaml"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	app := tui.NewApp("%s", version).
		Command(tui.NewCommand("hello", "print a greeting", nil, commands.Hello))
	os.Exit(int(app.Run(os.Args[1:])))
}
`, module, module, name)
}

func cliCommandsTemplate(name string) string {
	return fmt.Sprintf(`// Package commands implements the %s CLI sub-commands.
package commands

import (
	"flag"
	"fmt"

	"github.com/krewire/krewire/pkg/errs"
)

// Hello prints a friendly greeting.
func Hello(_ *flag.FlagSet) errs.ExitCode {
	fmt.Println("hello, %s")
	return errs.ExitCodeSuccess
}
`, name, name)
}

func cliReadmeTemplate(name string) string {
	return fmt.Sprintf(`# %s

A Krewire CLI project scaffolded by the [krew CLI](https://github.com/krewire/krew).

Built on the [Krewire Ecosystem](https://github.com/krewire/krewire).

## Getting Started

	go run . hello

## Building

	go build -o ./%s .

## Testing

	go test ./...
`, name, name)
}

func cliConfigTemplate() string {
	return "// Package config defines the typed CLI configuration\n" +
		"// krewire.yaml: metadata and project-level config (metadata only)\n" +
		"package config\n\n" +
		"import (\n" +
		"\t\"fmt\"\n" +
		"\t\"os\"\n\n" +
		"\trconfig \"github.com/krewire/krewire/pkg/config\"\n" +
		"\tvalidation \"github.com/krewire/krewire/pkg/validation\"\n" +
		")\n\n" +
		"// Metadata mirrors krewire.yaml.\n" +
		"type Metadata struct {\n" +
		"\tProject Project `yaml:\"project\"`\n" +
		"}\n\n" +
		"// Project holds the project section of krewire.yaml.\n" +
		"type Project struct {\n" +
		"\tName    string `yaml:\"name\" validate:\"required\"`\n" +
		"\tKind    string `yaml:\"kind\" validate:\"required\"`\n" +
		"\tVersion string `yaml:\"version\"`\n" +
		"}\n\n" +
		"// LoadMetadata reads krewire.yaml from path, overlays the environment, and returns a\n" +
		"// validated Metadata.\n" +
		"func LoadMetadata(path string) (*Metadata, error) {\n" +
		"\tconfig := &Metadata{}\n" +
		"\tif err := rconfig.Load(path, config); err != nil {\n" +
		"\t\treturn nil, err\n" +
		"\t}\n" +
		"\tif err := rconfig.Override(config, os.LookupEnv); err != nil {\n" +
		"\t\treturn nil, err\n" +
		"\t}\n" +
		"\tif err := validation.Struct(config); err != nil {\n" +
		"\t\treturn nil, fmt.Errorf(\"config: %w\", err)\n" +
		"\t}\n" +
		"\treturn config, nil\n" +
		"}\n"
}
