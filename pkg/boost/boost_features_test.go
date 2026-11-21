package boost_test

import (
	"strings"
	"testing"

	"github.com/krewire/krew/pkg/boost"
	"github.com/krewire/krewire/pkg/model"
)

func TestPersonaRegistry_CanonicalPersonas(t *testing.T) {
	reg, err := boost.NewPersonaRegistry(boost.Template)
	if err != nil {
		t.Fatalf("failed to create persona registry: %v", err)
	}

	if reg.Count() < 21 {
		t.Errorf("expected at least 21 canonical personas, got %d", reg.Count())
	}

	// Verify all canonical names exist
	for _, name := range boost.CanonicalPersonaNames {
		p, ok := reg.Get(name)
		if !ok || p == nil {
			t.Errorf("expected persona %q to exist in registry", name)
		}
	}

	// Verify alias resolution
	aliasTests := []struct {
		alias        string
		expectedName string
	}{
		{"plan", "planner"},
		{"build", "builder"},
		{"coder", "builder"},
		{"qa", "tester"},
		{"test", "tester"},
		{"architect", "designer"},
		{"devops", "deployer"},
		{"security", "auditor"},
		{"refactor", "maintainer"},
		{"docs", "writer"},
	}

	for _, tt := range aliasTests {
		p, ok := reg.Get(tt.alias)
		if !ok || p.Name != tt.expectedName {
			t.Errorf("alias %q resolved to %v (ok=%v), want %q", tt.alias, p, ok, tt.expectedName)
		}
	}

	// Non-existent persona
	if _, ok := reg.Get("nonexistent_persona_xyz"); ok {
		t.Errorf("expected nonexistent persona to return false")
	}
}

func TestCommandRegistry_CanonicalCommands(t *testing.T) {
	reg, err := boost.NewCommandRegistry(boost.Template)
	if err != nil {
		t.Fatalf("failed to create command registry: %v", err)
	}

	// Test plan command retrieval
	planCmd, ok := reg.Get("/plan")
	if !ok || planCmd.Name != "plan" {
		t.Fatalf("failed to retrieve /plan command")
	}

	// Test command rendering with arguments
	rendered, err := reg.Render("/plan", "implement authentication with PKCE")
	if err != nil {
		t.Fatalf("failed to render /plan command: %v", err)
	}
	if !contains(rendered, "implement authentication with PKCE") {
		t.Errorf("expected rendered prompt to contain arguments: %s", rendered)
	}

	// Non-existent command
	if _, err := reg.Render("/unknown", "args"); err == nil {
		t.Errorf("expected error for nonexistent slash command")
	}
}

func TestValidateCommitMessage(t *testing.T) {
	validMessages := []string{
		"feat(auth): implement PKCE authorization code flow",
		"fix(sec): prevent path traversal in SafePath",
		"docs(boost): document 21 autonomous agent personas",
		"test(model): add unit tests for verification ledger",
		"chore(deps): bump dependencies",
		"refactor(ui): extract reusable component tokens",
	}

	for _, msg := range validMessages {
		if err := boost.ValidateCommitMessage(msg); err != nil {
			t.Errorf("expected valid commit message %q to pass, got: %v", msg, err)
		}
	}

	invalidMessages := []string{
		"",
		"updated files",
		"fix:something without space",
		"WIP",
		"feat : wrong space before colon",
	}

	for _, msg := range invalidMessages {
		if err := boost.ValidateCommitMessage(msg); err == nil {
			t.Errorf("expected invalid commit message %q to fail", msg)
		}
	}
}

func TestConstitutionFilter_ValidateFileMutations(t *testing.T) {
	filter := boost.DefaultConstitutionFilter()

	// Under 300 LOC limit
	smallMutations := []model.FileMutation{
		{Path: "pkg/auth/token.go", LinesAdded: 50, LinesRemoved: 10},
		{Path: "pkg/auth/token_test.go", LinesAdded: 80, LinesRemoved: 5},
	}
	if err := filter.ValidateFileMutations(smallMutations); err != nil {
		t.Errorf("expected small mutation to pass, got: %v", err)
	}

	// Over 300 LOC limit
	largeMutations := []model.FileMutation{
		{Path: "big.go", LinesAdded: 250, LinesRemoved: 100},
	}
	if err := filter.ValidateFileMutations(largeMutations); err == nil {
		t.Errorf("expected mutation over 300 LOC to fail")
	}
}

func TestEvaluatePillarHealth(t *testing.T) {
	// Empty pillars triggers diagnostics
	ctx := model.ProjectContext{
		Name: "test-app",
	}
	diags := boost.EvaluatePillarHealth(ctx)
	if len(diags) != 3 {
		t.Errorf("expected 3 diagnostics for 3 missing pillars, got %d", len(diags))
	}

	// Populated pillars
	ctx.Pillars.Secure = []model.ContextRule{{ID: "SEC-1"}}
	ctx.Pillars.Sustainable = []model.ContextRule{{ID: "SUSTAIN-1"}}
	ctx.Pillars.Scalable = []model.ContextRule{{ID: "SCALE-1"}}
	diags = boost.EvaluatePillarHealth(ctx)
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for complete pillars, got %d", len(diags))
	}
}

func TestTaskBrief_FormatAndParse(t *testing.T) {
	original := model.TaskBrief{
		ID:                   "issue-42",
		Title:                "Implement PKCE authentication",
		Objective:            "Add PKCE OAuth2 flow to auth package",
		Repository:           "auth",
		Scope:                []string{"pkg/auth/pkce.go", "pkg/auth/pkce_test.go"},
		Constraints:          []string{"pure-Go invariant", "zero third-party dependencies"},
		AcceptanceCriteria:   []string{"Unit tests pass", "CGO_ENABLED=0 builds successfully"},
		VerificationCommands: []string{"go test -v ./pkg/auth/...", "go vet ./..."},
	}

	formatted := boost.FormatTaskBrief(original)
	if !strings.Contains(formatted, "# Implement PKCE authentication") {
		t.Errorf("expected title in formatted markdown: %s", formatted)
	}
	if !strings.Contains(formatted, "## Objective") {
		t.Errorf("expected objective in formatted markdown: %s", formatted)
	}

	parsed, err := boost.ParseTaskBrief("issue-42", "Implement PKCE authentication", formatted)
	if err != nil {
		t.Fatalf("failed to parse task brief: %v", err)
	}

	if parsed.Objective != original.Objective {
		t.Errorf("expected objective %q, got %q", original.Objective, parsed.Objective)
	}
	if parsed.Repository != original.Repository {
		t.Errorf("expected repository %q, got %q", original.Repository, parsed.Repository)
	}
	if len(parsed.Scope) != len(original.Scope) {
		t.Errorf("expected %d allowed paths, got %d", len(original.Scope), len(parsed.Scope))
	}
	if len(parsed.Constraints) != len(original.Constraints) {
		t.Errorf("expected %d constraints, got %d", len(original.Constraints), len(parsed.Constraints))
	}
	if len(parsed.AcceptanceCriteria) != len(original.AcceptanceCriteria) {
		t.Errorf("expected %d acceptance criteria, got %d", len(original.AcceptanceCriteria), len(parsed.AcceptanceCriteria))
	}
	if len(parsed.VerificationCommands) != len(original.VerificationCommands) {
		t.Errorf("expected %d verification commands, got %d", len(original.VerificationCommands), len(parsed.VerificationCommands))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
