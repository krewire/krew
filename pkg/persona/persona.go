package persona

import (
	"errors"
	"strings"
	"sync"

	"github.com/krewire/krewire/pkg/boost"
)

// Persona represents an agent role with specialized instructions and behavior.
type Persona struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Instructions string   `json:"instructions"`
	DefaultTools []string `json:"default_tools,omitempty"`
}

var (
	// ErrInvalidPersonaName is returned when a persona has an empty name.
	ErrInvalidPersonaName = errors.New("persona: name cannot be empty")

	// ErrInvalidInstructions is returned when a persona has empty instructions.
	ErrInvalidInstructions = errors.New("persona: instructions cannot be empty")

	// ErrPersonaAlreadyExists is returned when registering a persona whose name is already taken.
	ErrPersonaAlreadyExists = errors.New("persona: persona with this name already exists")
)

// Validate ensures the persona has mandatory identifying and behavioral fields.
func (p Persona) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrInvalidPersonaName
	}
	if strings.TrimSpace(p.Instructions) == "" {
		return ErrInvalidInstructions
	}
	return nil
}

// Pre-defined built-in personas.
var (
	// CodingAssistant is tuned for software engineering, pair programming,
	// test-driven refactoring, and seamless integration with Krewire Boost.
	CodingAssistant = Persona{
		Name:        "coding-assistant",
		Description: "Expert software engineer for autonomous coding, refactoring, debugging, and review.",
		Instructions: `You are Krew, an elite AI coding assistant and agent.
You operate with surgical precision:
1. Understand the user goal and explore the codebase before making changes.
2. Maintain documentation integrity, existing style, and single source of truth (SSOT).
3. When modifying code, make atomic changes and verify with test execution.
4. Prefer editing existing files via exact string replacement over blind overwrites.
5. Provide concise, clear markdown explanations of what you accomplished.`,
		DefaultTools: []string{"read_file", "write_file", "edit_file", "grep_search", "find_files", "apply_patch", "list_dir", "run_command"},
	}

	// DevOpsRunner is tuned for server administration, infrastructure orchestration,
	// container management, and remote daemon tasks.
	DevOpsRunner = Persona{
		Name:        "devops-runner",
		Description: "Reliable DevOps and site reliability engineer for server tasks and deployments.",
		Instructions: `You are Krew DevOps Agent, an autonomous infrastructure and server automation runner.
1. Inspect system state, logs, network sockets, and environment variables carefully.
2. Execute shell commands with strict safety and explicit timeout boundaries.
3. Validate prerequisites before initiating deployments or service restarts.
4. Always report command output, status codes, and potential failure points clearly.`,
		DefaultTools: []string{"run_command", "read_file", "write_file", "grep_search", "find_files", "apply_patch", "fetch_url"},
	}

	// AutonomousWorker is tuned for headless background tasks, goal-directed multi-step
	// problem solving, and batch job processing.
	AutonomousWorker = Persona{
		Name:        "autonomous-worker",
		Description: "Goal-oriented autonomous agent for self-directed task completion.",
		Instructions: `You are Krew Autonomous Worker. You are given a high-level goal and work
systematically until the goal is fully accomplished:
1. Deconstruct the user prompt into clear milestones.
2. Use tools iteratively to inspect, modify, and verify your progress.
3. If an error occurs, analyze the root cause and adapt your approach without halting prematurely.
4. Conclude only when the objective is completely satisfied.`,
		DefaultTools: []string{"read_file", "write_file", "edit_file", "grep_search", "find_files", "apply_patch", "run_command", "fetch_url"},
	}

	// GeneralAssistant is a versatile conversational assistant.
	GeneralAssistant = Persona{
		Name:        "general-assistant",
		Description: "Helpful, versatile AI assistant for broad workflows and queries.",
		Instructions: `You are Krew, a helpful, fast, and capable AI assistant.
Answer clearly, accurately, and assist with any tasks or questions requested.`,
		DefaultTools: []string{"fetch_url"},
	}

	// ResearchAssistant is tuned for fast, methodical investigation of a
	// codebase or topic: locating facts, reading sources, and returning
	// findings without modifying the workspace.
	ResearchAssistant = Persona{
		Name:        "research-assistant",
		Description: "Methodical investigator for codebase exploration, fact finding, and analysis.",
		Instructions: `You are Krew Research, a precise codebase and information investigator.
1. Formulate the question you must answer before searching.
2. Search broadly, then read the specific files that matter.
3. Report findings with exact file paths, line references, and short evidence quotes.
4. Never modify files; your value is the quality and accuracy of the findings.`,
		DefaultTools: []string{"read_file", "grep_search", "find_files", "list_dir", "fetch_url"},
	}

	// SDLCPlanner is tuned for analyzing task briefs, scoping requirements,
	// and generating structured implementation plans before any file edits.
	SDLCPlanner = Persona{
		Name:        "sdlc-planner",
		Description: "Specialized SDLC planner for architectural breakdown, risk estimation, and concise planning.",
		Instructions: `You are Krew SDLC Planner. Your responsibility is to analyze the task brief and design a concise, actionable plan:
1. Examine the task objective, scope boundaries, and constraints.
2. Read the codebase to understand existing architecture and conventions.
3. Identify all target files and estimate risk levels (low, medium, high) for each step.
4. Never modify files during the planning stage; only inspect and plan.`,
		DefaultTools: []string{"read_file", "grep_search", "find_files", "list_dir"},
	}

	// SDLCImplementer is tuned for executing the approved plan with atomic edits
	// strictly bounded to granted scope.
	SDLCImplementer = Persona{
		Name:        "sdlc-implementer",
		Description: "Specialized SDLC implementer for atomic, test-driven code modifications strictly bounded to scope.",
		Instructions: `You are Krew SDLC Implementer. You implement changes according to the approved plan:
1. Confine all file creations and modifications strictly within the granted scope.
2. Apply atomic, precise changes without introducing unnecessary refactoring or style drift.
3. Keep changes clean, robust, and well-structured.
4. Prepare the code for deterministic verification checks.`,
		DefaultTools: []string{"read_file", "write_file", "edit_file", "apply_patch", "grep_search", "find_files", "list_dir"},
	}

	// SDLCVerifier is tuned for executing verification commands and analyzing qualification evidence.
	SDLCVerifier = Persona{
		Name:        "sdlc-verifier",
		Description: "Specialized SDLC verifier for deterministic testing and quality gate qualification.",
		Instructions: `You are Krew SDLC Verifier. Your responsibility is qualification and verification:
1. Execute the exact verification commands specified in the task brief.
2. Capture full execution outcomes, exit codes, and stdout/stderr evidence.
3. If checks fail, pinpoint the exact test failure or build error for immediate recovery.
4. Never claim completion or pass if any check exits with a non-zero code.`,
		DefaultTools: []string{"run_command", "read_file", "grep_search", "find_files"},
	}

	// SDLCReviewer is tuned for inspecting file diffs, verifying qualification evidence,
	// and assessing architectural or security risks.
	SDLCReviewer = Persona{
		Name:        "sdlc-reviewer",
		Description: "Specialized SDLC reviewer for diff review, security risk identification, and change summarization.",
		Instructions: `You are Krew SDLC Reviewer. Your responsibility is transparent human review preparation:
1. Review all file diffs produced during implementation.
2. Check that all verification checks and quality gates passed with deterministic evidence.
3. Identify potential security risks, edge cases, breaking changes, or unresolved questions.
4. Provide a clear, actionable change summary for human engineering approval.`,
		DefaultTools: []string{"read_file", "grep_search", "find_files"},
	}
)

var (
	builtins = []Persona{
		CodingAssistant,
		DevOpsRunner,
		AutonomousWorker,
		GeneralAssistant,
		ResearchAssistant,
		SDLCPlanner,
		SDLCImplementer,
		SDLCVerifier,
		SDLCReviewer,
	}

	mu       sync.RWMutex
	registry = make(map[string]Persona)
)

func boostToPersona(bp *boost.Persona) Persona {
	var tools []string
	if bp.Tools != "" {
		for _, t := range strings.Split(bp.Tools, ",") {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				tools = append(tools, trimmed)
			}
		}
	}
	instr := bp.Instructions
	if instr == "" {
		instr = bp.Description
	}
	return Persona{
		Name:         bp.Name,
		Description:  bp.Description,
		Instructions: instr,
		DefaultTools: tools,
	}
}

// All returns a slice of all built-in, boost, and registered personas.
func All() []Persona {
	mu.RLock()
	defer mu.RUnlock()

	seen := make(map[string]bool)
	result := make([]Persona, 0, len(builtins)+len(registry)+25)
	for _, p := range builtins {
		result = append(result, p)
		seen[p.Name] = true
	}
	for _, p := range registry {
		if !seen[p.Name] {
			result = append(result, p)
			seen[p.Name] = true
		}
	}
	for _, bp := range boost.DefaultRegistry().List() {
		if !seen[bp.Name] {
			result = append(result, boostToPersona(bp))
			seen[bp.Name] = true
		}
	}
	return result
}

// Names returns a slice of all available persona names.
func Names() []string {
	personas := All()
	names := make([]string, len(personas))
	for i, p := range personas {
		names[i] = p.Name
	}
	return names
}

// Lookup finds a persona by exact name, normalized name, or common aliases.
func Lookup(name string) (Persona, bool) {
	norm := strings.ToLower(strings.TrimSpace(name))
	if norm == "" {
		return Persona{}, false
	}

	// 1. Direct match with built-ins or canonical aliases
	switch norm {
	case "coding-assistant", "coding", "code", "dev":
		return CodingAssistant, true
	case "devops-runner", "devops", "ops", "sre":
		return DevOpsRunner, true
	case "autonomous-worker", "autonomous", "worker":
		return AutonomousWorker, true
	case "general-assistant", "general", "assistant":
		return GeneralAssistant, true
	case "research-assistant", "research", "investigate":
		return ResearchAssistant, true
	case "sdlc-planner", "planner", "plan":
		return SDLCPlanner, true
	case "sdlc-implementer", "implementer", "impl":
		return SDLCImplementer, true
	case "sdlc-verifier", "verifier", "verify":
		return SDLCVerifier, true
	case "sdlc-reviewer", "reviewer", "review":
		return SDLCReviewer, true
	}

	// 2. Custom registry lookup
	mu.RLock()
	p, ok := registry[norm]
	mu.RUnlock()
	if ok {
		return p, true
	}

	// 3. Canonical Boost personas lookup
	if bp, found := boost.DefaultRegistry().Lookup(norm); found {
		return boostToPersona(bp), true
	}

	return Persona{}, false
}

// Register registers a custom persona into the runtime registry.
func Register(p Persona) error {
	if err := p.Validate(); err != nil {
		return err
	}

	norm := strings.ToLower(strings.TrimSpace(p.Name))
	if _, exists := Lookup(norm); exists {
		return ErrPersonaAlreadyExists
	}

	mu.Lock()
	defer mu.Unlock()
	registry[norm] = p
	return nil
}
