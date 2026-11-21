package boost

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/krewire/krewire/pkg/model"
)

var (
	// ErrInvalidCommitMessage is returned when a commit message violates Conventional Commits.
	ErrInvalidCommitMessage = errors.New("boost: commit message violates Conventional Commits standard")

	// ErrPillarViolation is returned when an action violates the Three Pillars (Secure, Sustainable, Scalable).
	ErrPillarViolation = errors.New("boost: action violates Three Pillars constitution")
)

var conventionalCommitRegex = regexp.MustCompile(`^(feat|fix|docs|test|refactor|perf|security|chore|style|ci|build|revert)(\([a-zA-Z0-9_\-\.\/]+\))?:\s+.+$`)

// ValidateCommitMessage validates that a commit message adheres to the ecosystem Conventional Commit standard.
func ValidateCommitMessage(msg string) error {
	trimmed := strings.TrimSpace(msg)
	if trimmed == "" {
		return fmt.Errorf("%w: commit message cannot be empty", ErrInvalidCommitMessage)
	}

	lines := strings.Split(trimmed, "\n")
	firstLine := strings.TrimSpace(lines[0])

	if !conventionalCommitRegex.MatchString(firstLine) {
		return fmt.Errorf("%w: format must be '<type>(<scope>): <description>', got %q", ErrInvalidCommitMessage, firstLine)
	}

	return nil
}

// ConstitutionFilter evaluates actions and context against Krewire's Three Pillars (3S).
type ConstitutionFilter struct {
	MaxCommitLOC int
}

// DefaultConstitutionFilter creates a standard constitution filter.
func DefaultConstitutionFilter() *ConstitutionFilter {
	return &ConstitutionFilter{
		MaxCommitLOC: 300,
	}
}

// ValidateFileMutations verifies that an agent's changes adhere to the atomic commit size invariant (<= 300 LOC).
func (f *ConstitutionFilter) ValidateFileMutations(mutations []model.FileMutation) error {
	totalChanged := 0
	for _, m := range mutations {
		totalChanged += m.LinesAdded + m.LinesRemoved
	}

	if f.MaxCommitLOC > 0 && totalChanged > f.MaxCommitLOC {
		return fmt.Errorf("%w: total changes (%d LOC) exceed atomic commit limit of %d LOC", ErrPillarViolation, totalChanged, f.MaxCommitLOC)
	}

	return nil
}

// EvaluatePillarHealth checks that project context properly respects the Three Pillars (Secure, Sustainable, Scalable).
func EvaluatePillarHealth(ctx model.ProjectContext) []model.ContextDiagnostic {
	var diags []model.ContextDiagnostic

	// 1. Check Secure Pillar
	if len(ctx.Pillars.Secure) == 0 {
		diags = append(diags, model.ContextDiagnostic{
			Severity: model.DiagMissing,
			Scope:    model.ScopeDomain,
			Message:  "Secure pillar has no declared context rules",
			Hint:     "Add baseline rules for SafePath, secret scanning, and CSP headers",
		})
	}

	// 2. Check Sustainable Pillar
	if len(ctx.Pillars.Sustainable) == 0 {
		diags = append(diags, model.ContextDiagnostic{
			Severity: model.DiagMissing,
			Scope:    model.ScopeDomain,
			Message:  "Sustainable pillar has no declared context rules",
			Hint:     "Add baseline rules for 10-30 MB RAM bounds and single static binary embeds",
		})
	}

	// 3. Check Scalable Pillar
	if len(ctx.Pillars.Scalable) == 0 {
		diags = append(diags, model.ContextDiagnostic{
			Severity: model.DiagMissing,
			Scope:    model.ScopeDomain,
			Message:  "Scalable pillar has no declared context rules",
			Hint:     "Add baseline rules for downward-only layering and progressive pipeline transitions",
		})
	}

	return diags
}
