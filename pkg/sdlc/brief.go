package sdlc

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

// Validate ensures the TaskBrief has all required fields populated.
func (b *TaskBrief) Validate() error {
	if strings.TrimSpace(b.ID) == "" {
		return fmt.Errorf("%w: ID cannot be empty", ErrInvalidBrief)
	}
	if strings.TrimSpace(b.Title) == "" {
		return fmt.Errorf("%w: Title cannot be empty", ErrInvalidBrief)
	}
	if strings.TrimSpace(b.Objective) == "" {
		return fmt.Errorf("%w: Objective cannot be empty", ErrInvalidBrief)
	}
	if len(b.Scope) == 0 {
		return fmt.Errorf("%w: Scope must declare at least one allowed directory or file pattern", ErrInvalidBrief)
	}
	return nil
}

// ParseTaskBriefFromMarkdown extracts structured fields from a markdown GitHub issue description.
// It parses standard headings: Objective, Scope, Constraints, Acceptance Criteria, and Verification Commands.
func ParseTaskBriefFromMarkdown(id, title, body string) (TaskBrief, error) {
	brief := TaskBrief{
		ID:            strings.TrimSpace(id),
		Title:         strings.TrimSpace(title),
		RequirePlan:   true,
		RequireReview: true,
		Timeout:       10 * time.Minute,
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

	if brief.Objective == "" {
		// Fallback: if no ## Objective heading was parsed, use whole body
		brief.Objective = strings.TrimSpace(body)
	}

	if len(brief.Scope) == 0 {
		brief.Scope = []string{"."}
	}

	return brief, brief.Validate()
}

func cleanListEntry(line string) string {
	s := strings.TrimSpace(line)
	s = strings.TrimPrefix(s, "- [ ]")
	s = strings.TrimPrefix(s, "- [x]")
	s = strings.TrimPrefix(s, "- [X]")
	s = strings.TrimPrefix(s, "* [ ]")
	s = strings.TrimPrefix(s, "* [x]")
	s = strings.TrimPrefix(s, "* [X]")
	s = strings.TrimPrefix(s, "-")
	s = strings.TrimPrefix(s, "*")
	s = strings.TrimPrefix(s, "•")
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "`")
	return strings.TrimSpace(s)
}
