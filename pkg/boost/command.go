package boost

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

var (
	// ErrCommandNotFound is returned when a requested slash command does not exist.
	ErrCommandNotFound = errors.New("boost: slash command not found")

	// ErrInvalidCommand is returned when a slash command definition lacks mandatory fields.
	ErrInvalidCommand = errors.New("boost: invalid slash command definition")
)

// SlashCommand represents a governed workflow command (e.g. /plan, /build, /test).
type SlashCommand struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	ArgumentsDesc string `json:"arguments_desc,omitempty"`
	TargetPersona string `json:"target_persona,omitempty"`
	Template      string `json:"template"`
}

// CanonicalCommandNames lists the official Boost slash commands.
var CanonicalCommandNames = []string{
	"audit",
	"automate",
	"build",
	"customize",
	"decompose",
	"deploy",
	"design",
	"explore",
	"housekeep",
	"lead",
	"learn",
	"maintain",
	"manage",
	"orchestrate",
	"plan",
	"research",
	"supervise",
	"test",
	"think",
	"track",
	"verify",
	"write",
}

// defaultCommands provides baseline template specifications for canonical slash commands.
var defaultCommands = map[string]*SlashCommand{
	"plan": {
		Name:          "plan",
		Description:   "Turn ambiguous scope into a versioned SRS, prioritized backlog, and risk register.",
		ArgumentsDesc: "Scope description or task specification to plan.",
		TargetPersona: "planner",
		Template: `Analyze scope: $ARGUMENTS
Follow mandatory workflow: Understand -> Plan -> Build -> Verify -> Summarize.
Ground every requirement in verifiable sources and produce prioritized backlog with Gherkin acceptance criteria.`,
	},
	"build": {
		Name:          "build",
		Description:   "Implement feature or fix adhering strictly to downward layering and Three Pillars.",
		ArgumentsDesc: "Implementation task details, targeted packages, and requirements.",
		TargetPersona: "builder",
		Template: `Implement task: $ARGUMENTS
Ensure changes are <= 300 LOC, pass all 4 quality gates (gofmt, go vet, go test, CGO_ENABLED=0 go build), and avoid hardcoded secrets.`,
	},
	"test": {
		Name:          "test",
		Description:   "Run unit tests, integration tests, and benchmarks across target packages.",
		ArgumentsDesc: "Package paths or test flags.",
		TargetPersona: "tester",
		Template: `Execute test qualification for: $ARGUMENTS
Report concrete exit codes and assertion outcomes. Never fabricate test results.`,
	},
	"verify": {
		Name:          "verify",
		Description:   "Execute complete multi-gate qualification and record tamper-evident verification ledger.",
		ArgumentsDesc: "Commit SHA or scope to verify.",
		TargetPersona: "tester",
		Template: `Verify qualification gates for: $ARGUMENTS
Execute:
1. gofmt -l .
2. go vet ./...
3. go test -v ./...
4. CGO_ENABLED=0 go build ./...
Generate tamper-evident verification record.`,
	},
	"audit": {
		Name:          "audit",
		Description:   "Audit codebase for security vulnerabilities, SafePath traversal, secrets, and licensing.",
		ArgumentsDesc: "Target repository or package path.",
		TargetPersona: "auditor",
		Template: `Audit security and licensing compliance for: $ARGUMENTS
Scan for committed secrets, verify least-privilege permissions, and ensure zero unvetted dependencies.`,
	},
	"design": {
		Name:          "design",
		Description:   "Design system architecture, interface contracts, and Pop-Brutalist UI components.",
		ArgumentsDesc: "Architectural component or feature to design.",
		TargetPersona: "designer",
		Template: `Design architectural specification for: $ARGUMENTS
Ensure zero peer coupling across Layer 4 platforms and maintain pure-Go stdlib-first boundaries.`,
	},
	"deploy": {
		Name:          "deploy",
		Description:   "Build single static binaries or container images and verify health probes.",
		ArgumentsDesc: "Target workload or deployment environment.",
		TargetPersona: "deployer",
		Template: `Stage deployment for: $ARGUMENTS
Verify container health, listen ports, and HTTP 200 OK response on health endpoints.`,
	},
}

// CommandRegistry maintains registered slash commands and handles template rendering.
type CommandRegistry struct {
	commands map[string]*SlashCommand
}

// NewCommandRegistry creates a registry with canonical commands and scans fsys for overrides.
func NewCommandRegistry(fsys fs.FS) (*CommandRegistry, error) {
	r := &CommandRegistry{
		commands: make(map[string]*SlashCommand),
	}

	// 1. Seed defaults
	for name, cmd := range defaultCommands {
		c := *cmd
		r.commands[name] = &c
	}

	// 2. Scan filesystem if provided
	if fsys != nil {
		_ = r.loadFromFS(fsys)
	}

	return r, nil
}

// loadFromFS scans .agents/commands/*.md or commands/*.md from fsys.
func (r *CommandRegistry) loadFromFS(fsys fs.FS) error {
	candidateDirs := []string{".agents/commands", "commands"}
	for _, dir := range candidateDirs {
		entries, err := fs.ReadDir(fsys, dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			path := dir + "/" + e.Name()
			data, err := fs.ReadFile(fsys, path)
			if err != nil {
				continue
			}
			cmd, err := parseCommandMarkdown(data)
			if err == nil && cmd.Name != "" {
				r.Register(cmd)
			}
		}
	}
	return nil
}

// parseCommandMarkdown extracts command properties from markdown with YAML frontmatter.
func parseCommandMarkdown(data []byte) (*SlashCommand, error) {
	s := string(data)
	cmd := &SlashCommand{}

	if strings.HasPrefix(s, "---") {
		parts := strings.SplitN(s[3:], "---", 2)
		if len(parts) == 2 {
			fm := parts[0]
			body := parts[1]

			scanner := bufio.NewScanner(strings.NewReader(fm))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "name:") {
					cmd.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				} else if strings.HasPrefix(line, "description:") {
					cmd.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				}
			}
			cmd.Template = strings.TrimSpace(body)
		}
	}

	if cmd.Name == "" {
		return nil, ErrInvalidCommand
	}

	return cmd, nil
}

// Register registers or overrides a slash command.
func (r *CommandRegistry) Register(cmd *SlashCommand) {
	if cmd == nil || cmd.Name == "" {
		return
	}
	name := strings.TrimPrefix(strings.ToLower(cmd.Name), "/")
	cmd.Name = name
	r.commands[name] = cmd
}

// Get retrieves a slash command by name (supports both "plan" and "/plan").
func (r *CommandRegistry) Get(name string) (*SlashCommand, bool) {
	clean := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(name)), "/")
	cmd, ok := r.commands[clean]
	return cmd, ok
}

// List returns all registered slash commands sorted alphabetically.
func (r *CommandRegistry) List() []*SlashCommand {
	list := make([]*SlashCommand, 0, len(r.commands))
	for _, c := range r.commands {
		list = append(list, c)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})
	return list
}

// Render interpolates arguments into the command's prompt template.
func (r *CommandRegistry) Render(name, arguments string) (string, error) {
	cmd, ok := r.Get(name)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrCommandNotFound, name)
	}
	rendered := strings.ReplaceAll(cmd.Template, "$ARGUMENTS", strings.TrimSpace(arguments))
	return rendered, nil
}
