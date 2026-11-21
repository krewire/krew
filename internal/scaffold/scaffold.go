// Package scaffold generates new Krewire projects.
package scaffold

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/krewire/krew/internal/config"
)

// Sentinels returned by the scaffold functions.
var (
	// ErrInvalidName is returned when the project name contains invalid
	// characters or path separators.
	ErrInvalidName = errors.New("invalid project name: use letters, digits, dashes, dots, or underscores")
	// ErrProjectExists is returned when the target directory exists and is
	// not empty.
	ErrProjectExists = errors.New("refusing to scaffold: directory already exists and is not empty")
	// ErrNotEmpty is returned when a variant target must be empty, such as a
	// --template clone target.
	ErrNotEmpty = errors.New("refusing to clone: target directory is not empty")
	// ErrConflict is returned when equipping a variant would overwrite a
	// file outside the bootstrap set (go.mod, krewire.yaml, .gitignore, main.go).
	ErrConflict = errors.New("refusing to equip: file already exists")
)

var namePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// Variant is the project shape equipped by Equip.
type Variant string

const (
	// VariantApp is a fullstack monolith with the entry point at the root
	// main.go (canonical layout KWF-CCI0N).
	VariantApp Variant = "app"
	// VariantStatic is a declarative static site configured solely through
	// the `ssg:` key in krewire.yaml (KWF-PT8OD). No ssg.yaml is produced.
	VariantStatic Variant = "static"
	// VariantBook is a book variant backed by pkg/book.
	VariantBook Variant = "book"
	// VariantCLI is a command-line application on pkg/tui + the foundation packages,
	// mirroring the krew CLI layout.
	VariantCLI Variant = "cli"
	// VariantWorker is a background job queue worker application (cloud/worker).
	VariantWorker Variant = "worker"
	// VariantService is a standalone microservice application (cloud/service).
	VariantService Variant = "service"
	// VariantGateway is an API gateway application proxying to upstream
	// services through the service registry (pkg/web/gateway).
	VariantGateway Variant = "gateway"
	// VariantInfra is a declarative cloud infrastructure blueprint project (cloud/infra).
	VariantInfra Variant = "infra"
	// VariantTemplate clones a remote starter repository into the target.
	VariantTemplate Variant = "template"
)

// Options configures bootstrap generation.
type Options struct {
	// Name is the project name and directory name. Required.
	Name string
	// Dir is the parent directory to create the project in. Defaults to the
	// current working directory.
	Dir string
	// Module is the Go module path. Defaults to the project name.
	Module string
}

// file is one output file of a scaffold.
type file struct {
	name string
	body string
}

// bootstrapBody is the stdlib-only main.go written by `krew new`. init
// upgrades it in place for the app variant and removes it for site/book
// variants.
const bootstrapBody = `package main

import "fmt"

func main() {
	fmt.Println("hello, krewire")
}
`

// InitOptions configures the minimal Krewire initialization in an existing or new directory.
type InitOptions struct {
	Dir          string
	Name         string
	Force        bool        // Overwrite krewire.yaml without prompting
	ConfirmReset func() bool // Prompt callback when krewire.yaml already exists
}

// Init initializes a minimal Krewire Ecosystem project:
// 1. Ensures directory exists.
// 2. Initializes git repository if .git does not exist.
// 3. Creates .gitignore if absent.
// 4. Creates or resets krewire.yaml to default settings (with confirmation if existing).
func Init(opts InitOptions) ([]string, error) {
	if opts.Dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		opts.Dir = wd
	}
	if err := os.MkdirAll(opts.Dir, 0o755); err != nil {
		return nil, err
	}

	if opts.Name == "" {
		opts.Name = filepath.Base(opts.Dir)
	}

	var created []string

	// 1. git init if .git is not present
	gitDir := filepath.Join(opts.Dir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		cmd := exec.Command("git", "init")
		cmd.Dir = opts.Dir
		if err := cmd.Run(); err == nil {
			created = append(created, ".git")
		}
	}

	// 2. .gitignore
	giPath := filepath.Join(opts.Dir, gitignoreFile)
	if _, err := os.Stat(giPath); os.IsNotExist(err) {
		if err := writeFile(giPath, []byte(gitignoreTemplate(opts.Name))); err != nil {
			return nil, err
		}
		created = append(created, gitignoreFile)
	}

	// 3. krewire.yaml
	yamlPath := filepath.Join(opts.Dir, krewireYaml)
	yamlContent := withSchemaHeader(fmt.Sprintf("project:\n  name: %s\n  version: 0.1.0\n", opts.Name))

	if _, err := os.Stat(yamlPath); err == nil {
		// Existing project: reset to default with confirmation
		shouldReset := opts.Force
		if !shouldReset && opts.ConfirmReset != nil {
			shouldReset = opts.ConfirmReset()
		}
		if shouldReset {
			if err := writeFile(yamlPath, []byte(yamlContent)); err != nil {
				return nil, err
			}
			created = append(created, krewireYaml+" (reset to default)")
		}
	} else {
		// New creation
		if err := writeFile(yamlPath, []byte(yamlContent)); err != nil {
			return nil, err
		}
		created = append(created, krewireYaml)
	}

	return created, nil
}

// New generates a new minimal Krewire bootstrap project and returns the created paths relative to Dir.
func New(opts Options) ([]string, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	cleanedPath := filepath.Clean(opts.Name)
	projectName := filepath.Base(cleanedPath)
	if projectName == "." || projectName == "/" || projectName == "" {
		projectName = "krewire-project"
	}

	if !namePattern.MatchString(projectName) {
		return nil, ErrInvalidName
	}

	var target string
	if filepath.IsAbs(cleanedPath) {
		target = cleanedPath
	} else if opts.Dir != "" {
		target = filepath.Join(opts.Dir, cleanedPath)
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		target = filepath.Join(wd, cleanedPath)
	}

	module := opts.Module
	if module == "" {
		module = projectName
	}

	empty, err := targetEmpty(target)
	if err != nil {
		return nil, err
	}
	if !empty {
		return nil, ErrProjectExists
	}

	initCreated, err := Init(InitOptions{
		Dir:  target,
		Name: projectName,
	})
	if err != nil {
		return nil, err
	}

	var created []string
	for _, f := range initCreated {
		created = append(created, filepath.Join(opts.Name, f))
	}

	// Add go.mod and placeholder main.go if not present
	goModPath := filepath.Join(target, goModFile)
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		if err := writeFile(goModPath, []byte(fmt.Sprintf("module %s\n\ngo %s\n", module, GoVersion))); err == nil {
			created = append(created, filepath.Join(opts.Name, goModFile))
		}
	}
	mainPath := filepath.Join(target, mainGo)
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		if err := writeFile(mainPath, []byte(bootstrapBody)); err == nil {
			created = append(created, filepath.Join(opts.Name, mainGo))
		}
	}

	return created, nil
}

// GoVersion is the Go language version emitted into scaffolded go.mod files.
// It is the single source of truth for the Krewire Go baseline: a generated
// project must declare a version at least as new as the go directive of the
// ecosystem module it requires, or `go build` rejects the module.
//
// Keep in sync with the `go` directive in the workspace go.work and in every
// module go.mod (see internal/docs/version.md §Go Baseline).
const GoVersion = "1.27.3"

// EquipOptions configures Equip. Dir is the project root to shape in place.
type EquipOptions struct {
	// Dir is the project root directory to equip. Required.
	Dir string
	// Name is the resolved project name, provided by the commands layer:
	// the module base for the app variant, the target directory base for
	// static and book. Required.
	Name string
	// Variant selects the shape to equip.
	Variant Variant
	// Module is the existing module path (from go.mod). Required by the app
	// variant to rewrite go.mod with pinned requires.
	Module string
	// Title is used by the static and book variants as the site title.
	Title string
	// Version pins the github.com/krewire/krewire require. Defaults to latest.
	Version string
	// TemplateURL is the git URL cloned by the template variant.
	TemplateURL string
}

// Equip shapes an existing project root into the requested variant and
// returns the created paths, relative to Dir. The bootstrap files (go.mod,
// krewire.yaml, .gitignore, main.go) may be upgraded in place; any other
// pre-existing file is refused.
func Equip(opts EquipOptions) ([]string, error) {
	if opts.Dir == "" {
		return nil, fmt.Errorf("equip: directory is required")
	}
	switch opts.Variant {
	case VariantApp:
		return equipApp(opts)
	case VariantStatic:
		return equipStatic(opts)
	case VariantBook:
		return equipBook(opts)
	case VariantCLI:
		return equipCLI(opts)
	case VariantWorker:
		return equipWorker(opts)
	case VariantService:
		return equipService(opts)
	case VariantGateway:
		return equipGateway(opts)
	case VariantInfra:
		return equipInfra(opts)
	case VariantTemplate:
		return equipTemplate(opts)
	default:
		return nil, fmt.Errorf("unknown variant %q", opts.Variant)
	}
}

const (
	goModFile     = "go.mod"
	krewireYaml   = config.ConfigFileName
	gitignoreFile = ".gitignore"
	mainGo        = "main.go"
)

// bootstrapUpgrade names the files init may overwrite in place.
var bootstrapUpgrade = map[string]bool{
	goModFile:     true,
	krewireYaml:   true,
	gitignoreFile: true,
	mainGo:        true,
}

// withSchemaHeader prefixes a krewire.yaml body with the yaml-language-server
// modeline that points editors at the canonical JSON Schema, unless the body
// already declares a schema.
func withSchemaHeader(body string) string {
	if strings.Contains(body, "$schema") {
		return body
	}
	return config.SchemaHeader + "\n" + body
}

// writeVariant writes files into Dir, refusing to overwrite anything outside
// the bootstrap set, and returns the created paths relative to Dir.
func writeVariant(dir string, files []file) ([]string, error) {
	var created []string
	for _, f := range files {
		path := filepath.Join(dir, f.name)
		if _, err := os.Stat(path); err == nil && !bootstrapUpgrade[f.name] {
			return nil, fmt.Errorf("%w: %s", ErrConflict, f.name)
		}
		body := f.body
		if f.name == krewireYaml {
			body = withSchemaHeader(body)
		}
		if err := writeFile(path, []byte(body)); err != nil {
			return nil, err
		}
		created = append(created, f.name)
	}
	return created, nil
}

// removeBootstrapMain removes the bootstrap's placeholder main.go (written by
// `krew new`) when equipping a site or book variant, so shape detection is
// pinned by project.kind rather than a stray app marker. It is a no-op when
// the file is absent or was modified by the user.
func removeBootstrapMain(dir string) error {
	path := filepath.Join(dir, mainGo)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(string(data)) == strings.TrimSpace(bootstrapBody) {
		return os.Remove(path)
	}
	return nil
}

func gitignoreTemplate(name string) string {
	return fmt.Sprintf(`# Compiled binary
/%s

# Environment files
.env
.env.local
.env.*.local
!.env.example

# Krewire build outputs
site/
.krewire/

# Editor and OS noise
.idea/
.vscode/
*.swp
.DS_Store
`, name)
}

// targetEmpty reports whether target is an absent path or an empty directory.
func targetEmpty(target string) (bool, error) {
	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	if !info.IsDir() {
		return false, fmt.Errorf("cannot scaffold: %s exists and is not a directory", target)
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func writeFile(path string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}
