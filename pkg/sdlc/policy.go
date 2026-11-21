package sdlc

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/krewire/krew/pkg/agent"
)

// MutatingTools are tools that change workspace state.
var MutatingTools = map[string]bool{
	"write_file":  true,
	"edit_file":   true,
	"apply_patch": true,
}

// IsMutatingTool returns true if the tool alters repository or workspace state.
func IsMutatingTool(toolName string) bool {
	return MutatingTools[toolName]
}

// CheckScope verifies if a file path is contained within any of the allowed scope prefixes.
func CheckScope(path string, allowedScopes []string) bool {
	if len(allowedScopes) == 0 {
		return false
	}
	cleanPath := filepath.Clean(filepath.ToSlash(path))
	for _, scope := range allowedScopes {
		cleanScope := filepath.Clean(filepath.ToSlash(scope))
		if cleanScope == "." || cleanScope == "*" || cleanScope == "" {
			return true
		}
		if cleanPath == cleanScope {
			return true
		}
		if strings.HasPrefix(cleanPath, cleanScope+"/") {
			return true
		}
	}
	return false
}

// ExtractFilePath parses common file path arguments from tool payload JSON.
func ExtractFilePath(args json.RawMessage) string {
	var payload struct {
		Path     string `json:"path"`
		FilePath string `json:"file_path"`
		Target   string `json:"target"`
	}
	if err := json.Unmarshal(args, &payload); err == nil {
		if payload.Path != "" {
			return payload.Path
		}
		if payload.FilePath != "" {
			return payload.FilePath
		}
		if payload.Target != "" {
			return payload.Target
		}
	}
	return ""
}

// PolicyFilter encapsulates scope checking and interactive approval rules.
type PolicyFilter struct {
	Brief         TaskBrief
	CurrentStage  func() Stage
	HumanApprover agent.Approver
	HighRiskTools map[string]bool
}

// NewScopedApprover creates an agent.Approver that enforces:
// 1. Read-only tool access outside StageImplement.
// 2. Strict file path boundary confinement against TaskBrief.Scope.
// 3. Human approval for high-risk actions.
func NewScopedApprover(brief TaskBrief, currentStage func() Stage, humanApprover agent.Approver, highRiskTools ...string) agent.Approver {
	risks := make(map[string]bool)
	for _, t := range highRiskTools {
		risks[t] = true
	}
	// By default, run_command is considered high risk unless explicitly allowed.
	if len(highRiskTools) == 0 {
		risks["run_command"] = true
	}

	filter := &PolicyFilter{
		Brief:         brief,
		CurrentStage:  currentStage,
		HumanApprover: humanApprover,
		HighRiskTools: risks,
	}

	return filter.Approve
}

// Approve evaluates whether a tool call is permitted.
func (f *PolicyFilter) Approve(toolName string, args json.RawMessage) bool {
	stage := StageUnderstand
	if f.CurrentStage != nil {
		stage = f.CurrentStage()
	}

	// 1. In early stages (Understand, Plan), file modifications are strictly prohibited
	if IsMutatingTool(toolName) && stage != StageImplement {
		return false
	}

	// 2. In StageImplement, mutating tools must strictly stay within TaskBrief.Scope
	if IsMutatingTool(toolName) {
		path := ExtractFilePath(args)
		if path != "" && !CheckScope(path, f.Brief.Scope) {
			return false
		}
	}

	// 3. High-risk actions require explicit human approval if configured
	if f.HighRiskTools[toolName] {
		if f.HumanApprover != nil {
			return f.HumanApprover(toolName, args)
		}
		// If high-risk action has no human approver, reject to fail closed
		return false
	}

	// 4. If human approver is configured for other tools, consult it
	if f.HumanApprover != nil {
		return f.HumanApprover(toolName, args)
	}

	return true
}
