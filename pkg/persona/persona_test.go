package persona

import (
	"testing"
)

func TestBuiltinPersonas_Validation(t *testing.T) {
	personas := All()
	if len(personas) < 9 {
		t.Fatalf("expected at least 9 built-in personas, got %d", len(personas))
	}

	for _, p := range personas {
		if err := p.Validate(); err != nil {
			t.Errorf("persona %q failed validation: %v", p.Name, err)
		}
		if p.Description == "" {
			t.Errorf("persona %q has empty description", p.Name)
		}
		if len(p.Instructions) == 0 {
			t.Errorf("persona %q has empty instructions", p.Name)
		}
	}
}

func TestLookup_BuiltinsAndAliases(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		expectFound bool
	}{
		{"coding-assistant", "coding-assistant", true},
		{"coding", "coding-assistant", true},
		{"code", "coding-assistant", true},
		{"dev", "coding-assistant", true},
		{"devops-runner", "devops-runner", true},
		{"devops", "devops-runner", true},
		{"ops", "devops-runner", true},
		{"sre", "devops-runner", true},
		{"autonomous-worker", "autonomous-worker", true},
		{"worker", "autonomous-worker", true},
		{"autonomous", "autonomous-worker", true},
		{"general-assistant", "general-assistant", true},
		{"general", "general-assistant", true},
		{"assistant", "general-assistant", true},
		{"research-assistant", "research-assistant", true},
		{"research", "research-assistant", true},
		{"investigate", "research-assistant", true},
		{"sdlc-planner", "sdlc-planner", true},
		{"planner", "sdlc-planner", true},
		{"sdlc-implementer", "sdlc-implementer", true},
		{"implementer", "sdlc-implementer", true},
		{"sdlc-verifier", "sdlc-verifier", true},
		{"verifier", "sdlc-verifier", true},
		{"sdlc-reviewer", "sdlc-reviewer", true},
		{"reviewer", "sdlc-reviewer", true},
		// Boost canonical personas and aliases
		{"auditor", "auditor", true},
		{"builder", "builder", true},
		{"thinker", "thinker", true},
		{"tester", "tester", true},
		{"architect", "designer", true},
		{"non-existent", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		p, found := Lookup(tt.input)
		if found != tt.expectFound {
			t.Errorf("Lookup(%q): expected found=%v, got %v", tt.input, tt.expectFound, found)
			continue
		}
		if found && p.Name != tt.expected {
			t.Errorf("Lookup(%q): expected persona name %q, got %q", tt.input, tt.expected, p.Name)
		}
	}
}

func TestRegister_CustomPersona(t *testing.T) {
	custom := Persona{
		Name:         "custom-qa-agent",
		Description:  "Specialized QA agent for end-to-end regression validation",
		Instructions: "Execute end-to-end regression test suites and analyze test reports.",
		DefaultTools: []string{"run_command", "read_file"},
	}

	// Initial registration should succeed
	if err := Register(custom); err != nil {
		t.Fatalf("failed to register custom persona: %v", err)
	}

	// Lookup should find it
	found, ok := Lookup("custom-qa-agent")
	if !ok || found.Name != custom.Name {
		t.Fatalf("expected custom persona to be found via Lookup, got ok=%v", ok)
	}

	// Duplicate registration should fail
	if err := Register(custom); err != ErrPersonaAlreadyExists {
		t.Fatalf("expected ErrPersonaAlreadyExists on duplicate registration, got %v", err)
	}

	// Registration of invalid persona should fail
	if err := Register(Persona{}); err != ErrInvalidPersonaName {
		t.Fatalf("expected ErrInvalidPersonaName, got %v", err)
	}

	if err := Register(Persona{Name: "valid-name"}); err != ErrInvalidInstructions {
		t.Fatalf("expected ErrInvalidInstructions, got %v", err)
	}
}

func TestNames_ContainsAllPersonas(t *testing.T) {
	names := Names()
	if len(names) < 9 {
		t.Fatalf("expected at least 9 names, got %d", len(names))
	}

	expectedNames := []string{
		"coding-assistant",
		"devops-runner",
		"autonomous-worker",
		"general-assistant",
		"research-assistant",
		"sdlc-planner",
		"sdlc-implementer",
		"sdlc-verifier",
		"sdlc-reviewer",
	}

	nameMap := make(map[string]bool)
	for _, n := range names {
		nameMap[n] = true
	}

	for _, expected := range expectedNames {
		if !nameMap[expected] {
			t.Errorf("expected names to contain %q", expected)
		}
	}
}
