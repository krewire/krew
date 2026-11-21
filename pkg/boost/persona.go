package boost

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"sync"

	"github.com/krewire/krewire/pkg/model"
)

var (
	// ErrPersonaNotFound is returned when a requested persona does not exist.
	ErrPersonaNotFound = errors.New("boost: persona not found")

	// ErrInvalidPersona is returned when a persona definition is missing mandatory fields.
	ErrInvalidPersona = errors.New("boost: invalid persona definition")
)

// Persona represents an autonomous agent persona profile with capabilities and limits.
type Persona struct {
	Name             string   `json:"name"`
	Title            string   `json:"title"`
	Role             string   `json:"role"`
	Description      string   `json:"description"`
	Mode             string   `json:"mode"` // "subagent", "primary"
	Capabilities     []string `json:"capabilities,omitempty"`
	AutonomousScope  string   `json:"autonomous_scope,omitempty"`
	RequiresApproval string   `json:"requires_approval,omitempty"`
	ForbiddenActions string   `json:"forbidden_actions,omitempty"`
	Tools            string   `json:"tools,omitempty"`
	Instructions     string   `json:"instructions,omitempty"`
	Aliases          []string `json:"aliases,omitempty"`
}

// CanonicalPersonaNames contains the 21 official Boost personas.
var CanonicalPersonaNames = []string{
	"auditor",
	"automator",
	"builder",
	"customizer",
	"decomposer",
	"deployer",
	"designer",
	"explorer",
	"housekeeper",
	"leader",
	"learner",
	"maintainer",
	"manager",
	"orchestrator",
	"planner",
	"researcher",
	"supervisor",
	"tester",
	"thinker",
	"tracker",
	"writer",
}

// defaultAliases maps intuitive role shorthands to canonical persona names.
var defaultAliases = map[string]string{
	"plan":      "planner",
	"build":     "builder",
	"coder":     "builder",
	"developer": "builder",
	"test":      "tester",
	"qa":        "tester",
	"design":    "designer",
	"architect": "designer",
	"deploy":    "deployer",
	"devops":    "deployer",
	"audit":     "auditor",
	"security":  "auditor",
	"sec":       "auditor",
	"maintain":  "maintainer",
	"refactor":  "maintainer",
	"track":     "tracker",
	"pmo":       "tracker",
	"write":     "writer",
	"docs":      "writer",
	"research":  "researcher",
	"automate":  "automator",
	"ci":        "automator",
}

// canonicalPersonaDefinitions provides baseline definitions for all 21 personas.
var canonicalPersonaDefinitions = map[string]*Persona{
	"planner": {
		Name:             "planner",
		Title:            "Planner Agent",
		Role:             "Analyze requirements, define scope, plan milestones, and manage dependencies/risks",
		Description:      "Planning subagent for requirements analysis, work breakdown, and verifiable project planning.",
		Mode:             "subagent",
		AutonomousScope:  "Choose elicitation technique, prioritization method, estimation approach",
		RequiresApproval: "Scope definition and final backlog sign-off",
		ForbiddenActions: "Write implementation code, alter system architecture, deploy live services",
		Aliases:          []string{"plan"},
	},
	"builder": {
		Name:             "builder",
		Title:            "Builder Agent",
		Role:             "Implement production-grade Go code, UI components, and unit tests adhering to strict layering",
		Description:      "Core engineering builder subagent for atomic feature implementation, bug fixes, and tests.",
		Mode:             "subagent",
		AutonomousScope:  "Implement code, run tests, format code, and stage atomic git commits within task scope",
		RequiresApproval: "File deletions, changing project dependencies, altering public API contracts",
		ForbiddenActions: "Bypass quality gates, commit hardcoded secrets, introduce circular package imports",
		Aliases:          []string{"build", "coder", "developer"},
	},
	"tester": {
		Name:             "tester",
		Title:            "Tester Agent",
		Role:             "Execute multi-gate qualification, unit tests, integration tests, and benchmark verification",
		Description:      "Quality assurance and test verification subagent.",
		Mode:             "subagent",
		AutonomousScope:  "Execute test runners, static analysis, benchmarks, and report gate pass/fail status",
		RequiresApproval: "Modifying test thresholds or skipping required quality gates",
		ForbiddenActions: "Fabricate test results, bypass failing assertions, suppress error output",
		Aliases:          []string{"test", "qa"},
	},
	"auditor": {
		Name:             "auditor",
		Title:            "Auditor Agent",
		Role:             "Conduct threat modeling, dependency hygiene auditing, secret scanning, and SafePath verification",
		Description:      "Security, compliance, and architectural governance auditor subagent.",
		Mode:             "subagent",
		AutonomousScope:  "Scan codebase for vulnerabilities, verify license adherence, audit dependencies",
		RequiresApproval: "Exempting security findings or overriding defense-in-depth rules",
		ForbiddenActions: "Commit unvetted secrets, weaken cryptographic primitives, disable CSRF/CORS checks",
		Aliases:          []string{"audit", "security", "sec"},
	},
	"designer": {
		Name:             "designer",
		Title:            "Designer Agent",
		Role:             "Design system architecture, domain boundary contracts, and Pop-Brutalist UI components",
		Description:      "System architecture and UI/UX design subagent.",
		Mode:             "subagent",
		AutonomousScope:  "Draft architectural decision records (ADRs), component layouts, and interface contracts",
		RequiresApproval: "Breaking changes to ecosystem API contracts, changing layering rules",
		ForbiddenActions: "Introduce peer coupling across Layer 4 platforms, introduce Node.js runtimes",
		Aliases:          []string{"design", "architect"},
	},
	"deployer": {
		Name:             "deployer",
		Title:            "Deployer Agent",
		Role:             "Build single static binaries, package Docker images, and verify production deployment health",
		Description:      "Deployment, staging, and operational release subagent.",
		Mode:             "subagent",
		AutonomousScope:  "Build release artifacts, run local staging containers, verify health probes",
		RequiresApproval: "Executing remote git push, deploying to live production, applying database migrations",
		ForbiddenActions: "Deploy unverified code, disable health checks, alter origin SSL posture",
		Aliases:          []string{"deploy", "devops"},
	},
	"maintainer": {
		Name:             "maintainer",
		Title:            "Maintainer Agent",
		Role:             "Refactor code, upgrade dependencies safely, prune dead code, and maintain package health",
		Description:      "Codebase maintenance, refactoring, and technical debt remediation subagent.",
		Mode:             "subagent",
		AutonomousScope:  "Clean up formatting, resolve deprecations, refactor internal package implementations",
		RequiresApproval: "Major semver bumps, changing public exported signatures, deleting packages",
		ForbiddenActions: "Break backward compatibility, introduce unvetted third-party libraries",
		Aliases:          []string{"maintain", "refactor"},
	},
	"automator": {
		Name:             "automator",
		Title:            "Automator Agent",
		Role:             "Maintain CI/CD workflows, automated build scripts, and reproducible dev environments",
		Description:      "Automation and workflow orchestration subagent.",
		Mode:             "subagent",
		AutonomousScope:  "Update GitHub Actions, refine test automation scripts, optimize build pipelines",
		RequiresApproval: "Modifying production deploy triggers, altering repository secret references",
		ForbiddenActions: "Add unpinned third-party actions, bypass status checks on protected branches",
		Aliases:          []string{"automate", "ci"},
	},
	"tracker": {
		Name:             "tracker",
		Title:            "Tracker Agent",
		Role:             "Track task progress against GitHub Issues, verify acceptance criteria, and update checklists",
		Description:      "Progress tracking, issue synchronization, and milestone accounting subagent.",
		Mode:             "subagent",
		AutonomousScope:  "Query issues, update checklist checkboxes upon passing tests, draft release notes",
		RequiresApproval: "Closing issues without passing quality gates, deleting milestones",
		ForbiddenActions: "Check off unverified task items, forge issue references in commit messages",
		Aliases:          []string{"track", "pmo"},
	},
	"writer": {
		Name:             "writer",
		Title:            "Writer Agent",
		Role:             "Author technical documentation, API manuals, architectural guides, and changelogs",
		Description:      "Technical writing, documentation portal maintenance, and user guide author subagent.",
		Mode:             "subagent",
		AutonomousScope:  "Write and update Markdown guides, correct typos, regenerate documentation indices",
		RequiresApproval: "Publishing changes to official public legal policies or licensing terms",
		ForbiddenActions: "Use Markdown tables in interactive chat responses, fabricate unverified statistics",
		Aliases:          []string{"write", "docs"},
	},
	"researcher": {
		Name:             "researcher",
		Title:            "Researcher Agent",
		Role:             "Investigate technical alternatives, benchmark algorithms, and review upstream specifications",
		Description:      "Codebase and technical domain research subagent.",
		Mode:             "subagent",
		AutonomousScope:  "Analyze codebases, run read-only benchmark comparisons, read RFCs and specifications",
		RequiresApproval: "Adopting new architectural patterns or changing foundational design baselines",
		ForbiddenActions: "Make mutating changes to source files, execute external unprompted network calls",
		Aliases:          []string{"research"},
	},
	"customizer": {
		Name:        "customizer",
		Title:       "Customizer Agent",
		Role:        "Tailor agent configurations, personas, skills, and slash commands for specific project domains",
		Description: "Project adaptation and agent customization subagent.",
		Mode:        "subagent",
	},
	"decomposer": {
		Name:        "decomposer",
		Title:       "Decomposer Agent",
		Role:        "Break down complex, ambiguous software objectives into granular, DAG-ordered subtasks",
		Description: "Task decomposition and dependency graphing subagent.",
		Mode:        "subagent",
	},
	"explorer": {
		Name:        "explorer",
		Title:       "Explorer Agent",
		Role:        "Explore unfamiliar codebases, map package dependency graphs, and identify dead paths",
		Description: "Repository discovery and structural exploration subagent.",
		Mode:        "subagent",
	},
	"housekeeper": {
		Name:        "housekeeper",
		Title:       "Housekeeper Agent",
		Role:        "Clean up temporary files, stale artifacts, redundant test fixtures, and orphaned branches",
		Description: "Workspace hygiene and temporary artifact pruning subagent.",
		Mode:        "subagent",
	},
	"leader": {
		Name:        "leader",
		Title:       "Leader Agent",
		Role:        "Coordinate multi-agent swarms, assign task shards to specialized agents, and aggregate results",
		Description: "Swarm coordination and multi-agent leadership subagent.",
		Mode:        "subagent",
	},
	"learner": {
		Name:        "learner",
		Title:       "Learner Agent",
		Role:        "Extract patterns from past developer corrections and codify learned project conventions into rules",
		Description: "Continuous learning and institutional knowledge codification subagent.",
		Mode:        "subagent",
	},
	"manager": {
		Name:        "manager",
		Title:       "Manager Agent",
		Role:        "Oversee project milestones, monitor agent execution budgets, and enforce governance rules",
		Description: "Engineering governance and lifecycle management subagent.",
		Mode:        "subagent",
	},
	"orchestrator": {
		Name:        "orchestrator",
		Title:       "Orchestrator Agent",
		Role:        "Execute complex multi-stage pipelines across build, test, package, and deployment systems",
		Description: "Pipeline orchestration and multi-system execution subagent.",
		Mode:        "subagent",
	},
	"supervisor": {
		Name:        "supervisor",
		Title:       "Supervisor Agent",
		Role:        "Supervise autonomous agent execution runs, enforce MaxSteps limits, and monitor process health",
		Description: "Agent execution supervision and safety circuit-breaker subagent.",
		Mode:        "subagent",
	},
	"thinker": {
		Name:        "thinker",
		Title:       "Thinker Agent",
		Role:        "Perform deep reasoning, invariant validation, and adversarial edge-case analysis",
		Description: "Deep algorithmic reasoning and mathematical/logical verification subagent.",
		Mode:        "subagent",
	},
}

// PersonaRegistry maintains loaded personas and alias indexes.
type PersonaRegistry struct {
	personas map[string]*Persona
	aliases  map[string]string
}

// NewPersonaRegistry creates a registry initialized with all 21 canonical personas
// and parses any overrides present in the provided filesystem.
func NewPersonaRegistry(fsys fs.FS) (*PersonaRegistry, error) {
	r := &PersonaRegistry{
		personas: make(map[string]*Persona),
		aliases:  make(map[string]string),
	}

	// 1. Seed canonical definitions
	for name, p := range canonicalPersonaDefinitions {
		// Clone
		cp := *p
		r.personas[name] = &cp
	}

	// 2. Register default aliases
	for alias, canonical := range defaultAliases {
		r.aliases[alias] = canonical
	}

	// 3. Scan filesystem if provided
	if fsys != nil {
		_ = r.loadFromFS(fsys)
	}

	return r, nil
}

var (
	defaultPersonaRegistry     *PersonaRegistry
	defaultPersonaRegistryOnce sync.Once
)

// DefaultRegistry returns the singleton canonical persona registry backed by Template.
func DefaultRegistry() *PersonaRegistry {
	defaultPersonaRegistryOnce.Do(func() {
		reg, err := NewPersonaRegistry(Template)
		if err != nil {
			reg = &PersonaRegistry{
				personas: make(map[string]*Persona),
				aliases:  make(map[string]string),
			}
			for name, p := range canonicalPersonaDefinitions {
				cp := *p
				reg.personas[name] = &cp
			}
			for alias, canonical := range defaultAliases {
				reg.aliases[alias] = canonical
			}
		}
		defaultPersonaRegistry = reg
	})
	return defaultPersonaRegistry
}

// loadFromFS scans .agents/agents/*.md or agents/*.md from fsys.
func (r *PersonaRegistry) loadFromFS(fsys fs.FS) error {
	candidateDirs := []string{".agents/agents", "agents"}
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
			p, err := parsePersonaMarkdown(data)
			if err == nil && p.Name != "" {
				r.Register(p)
			}
		}
	}
	return nil
}

// parsePersonaMarkdown extracts persona attributes from a markdown file.
func parsePersonaMarkdown(data []byte) (*Persona, error) {
	s := string(data)
	p := &Persona{}

	// Extract Frontmatter
	if strings.HasPrefix(s, "---") {
		parts := strings.SplitN(s[3:], "---", 2)
		if len(parts) == 2 {
			fm := parts[0]
			body := parts[1]

			scanner := bufio.NewScanner(strings.NewReader(fm))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "name:") {
					p.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				} else if strings.HasPrefix(line, "description:") {
					p.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				} else if strings.HasPrefix(line, "mode:") {
					p.Mode = strings.TrimSpace(strings.TrimPrefix(line, "mode:"))
				}
			}
			p.Instructions = strings.TrimSpace(body)
		}
	}

	if p.Name == "" {
		return nil, ErrInvalidPersona
	}

	return p, nil
}

// Register registers or overrides a persona in the registry.
func (r *PersonaRegistry) Register(p *Persona) {
	if p == nil || p.Name == "" {
		return
	}
	r.personas[p.Name] = p
	for _, alias := range p.Aliases {
		r.aliases[strings.ToLower(alias)] = p.Name
	}
}

// Get resolves a persona by exact name or alias.
func (r *PersonaRegistry) Get(nameOrAlias string) (*Persona, bool) {
	key := strings.ToLower(strings.TrimSpace(nameOrAlias))
	if canonical, ok := r.aliases[key]; ok {
		key = canonical
	}
	p, ok := r.personas[key]
	return p, ok
}

// List returns all registered personas sorted alphabetically by name.
func (r *PersonaRegistry) List() []*Persona {
	list := make([]*Persona, 0, len(r.personas))
	for _, p := range r.personas {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})
	return list
}

// Names returns all canonical persona names in sorted order.
func (r *PersonaRegistry) Names() []string {
	names := make([]string, 0, len(r.personas))
	for name := range r.personas {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Count returns the number of registered personas.
func (r *PersonaRegistry) Count() int {
	return len(r.personas)
}

// Lookup resolves a persona by exact name or alias (alias for Get).
func (r *PersonaRegistry) Lookup(nameOrAlias string) (*Persona, bool) {
	return r.Get(nameOrAlias)
}

// FormatTaskBrief formats a model.TaskBrief into canonical GitHub Issue markdown.
func FormatTaskBrief(brief model.TaskBrief) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", brief.Title))
	sb.WriteString("## Objective\n")
	sb.WriteString(brief.Objective)
	sb.WriteString("\n\n")

	if brief.Repository != "" {
		sb.WriteString("## Repository\n")
		sb.WriteString(brief.Repository)
		sb.WriteString("\n\n")
	}

	if len(brief.Scope) > 0 {
		sb.WriteString("## Scope\n")
		for _, p := range brief.Scope {
			sb.WriteString(fmt.Sprintf("- `%s`\n", p))
		}
		sb.WriteString("\n")
	}

	if len(brief.Constraints) > 0 {
		sb.WriteString("## Constraints\n")
		for _, c := range brief.Constraints {
			sb.WriteString(fmt.Sprintf("- %s\n", c))
		}
		sb.WriteString("\n")
	}

	if len(brief.AcceptanceCriteria) > 0 {
		sb.WriteString("## Acceptance Criteria\n")
		for _, a := range brief.AcceptanceCriteria {
			sb.WriteString(fmt.Sprintf("- [ ] %s\n", a))
		}
		sb.WriteString("\n")
	}

	if len(brief.VerificationCommands) > 0 {
		sb.WriteString("## Verification Commands\n")
		for _, v := range brief.VerificationCommands {
			sb.WriteString(fmt.Sprintf("- `%s`\n", v))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ParseTaskBrief extracts a structured model.TaskBrief from a markdown issue description.
func ParseTaskBrief(id, title, body string) (model.TaskBrief, error) {
	brief := model.TaskBrief{
		ID:            strings.TrimSpace(id),
		Title:         strings.TrimSpace(title),
		MaxAutonomy:   model.AutonomyAutonomous,
		RequirePlan:   true,
		RequireReview: true,
		TimeoutSec:    600,
		MaxRetries:    2,
	}

	scanner := bufio.NewScanner(strings.NewReader(body))
	var currentSection string
	var sectionLines []string

	commitSection := func() {
		if currentSection == "" {
			return
		}
		text := strings.TrimSpace(strings.Join(sectionLines, "\n"))
		switch currentSection {
		case "objective":
			brief.Objective = text
		case "repository":
			brief.Repository = text
		case "scope":
			for _, line := range sectionLines {
				cleaned := cleanListEntry(line)
				if cleaned != "" {
					brief.Scope = append(brief.Scope, cleaned)
				}
			}
		case "constraints":
			for _, line := range sectionLines {
				cleaned := cleanListEntry(line)
				if cleaned != "" {
					brief.Constraints = append(brief.Constraints, cleaned)
				}
			}
		case "acceptance criteria", "acceptance":
			for _, line := range sectionLines {
				cleaned := cleanListEntry(line)
				if cleaned != "" {
					brief.AcceptanceCriteria = append(brief.AcceptanceCriteria, cleaned)
				}
			}
		case "verification commands", "verification":
			for _, line := range sectionLines {
				cleaned := cleanListEntry(line)
				if cleaned != "" {
					brief.VerificationCommands = append(brief.VerificationCommands, cleaned)
				}
			}
		}
		sectionLines = nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "#") {
			commitSection()
			heading := strings.TrimLeft(trimmed, "# ")
			currentSection = strings.ToLower(strings.TrimSpace(heading))
			continue
		}

		if currentSection != "" {
			sectionLines = append(sectionLines, line)
		}
	}
	commitSection()

	return brief, nil
}

func cleanListEntry(line string) string {
	trimmed := strings.TrimSpace(line)
	for _, prefix := range []string{"- [ ] ", "- [x] ", "- ", "* ", "• "} {
		if strings.HasPrefix(trimmed, prefix) {
			trimmed = strings.TrimPrefix(trimmed, prefix)
			break
		}
	}
	trimmed = strings.Trim(trimmed, "`")
	return strings.TrimSpace(trimmed)
}
