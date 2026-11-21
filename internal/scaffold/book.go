// Package-level file for the book variant (VariantBook): krew init
// --book scaffolds a content/ directory assembled by the book compiler.
package scaffold

import (
	"fmt"
)

// equipBook shapes the bootstrap project into a markdown content book. The sample
// chapters live under content/ (shared with the ssg layout) and the bootstrap
// main.go is removed, matching the book build path.
func equipBook(opts EquipOptions) ([]string, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("equip book: project name is required")
	}
	files := []file{
		{krewireYaml, bookKrewireYamlTemplate(opts.Name, opts.Title)},
		{"content/docs/01-introduction.md", introSample},
		{"content/docs/02-getting-started.md", gettingStartedSample},
	}
	report, err := writeVariant(opts.Dir, files)
	if err != nil {
		return nil, err
	}
	if err := removeBootstrapMain(opts.Dir); err != nil {
		return nil, err
	}
	return report, nil
}

func bookKrewireYamlTemplate(name, title string) string {
	return fmt.Sprintf(`project:
  name: %s
  kind: book
  version: 0.1.0

title: %s
author: ""
input: content

build:
  base: /
`, name, title)
}

const introSample = "# Introduction\n\nWelcome to your Krewire site.\n\nThis content directory is assembled into a book-shaped docs website.\n"

const gettingStartedSample = "# Getting Started\n\nWrite Markdown files in the `content/` directory — subdirectories become chapters with subchapters.\n\n```\n$ krew build\n$ krew serve\n```\n\nThen open `.krewire/build/docs/getting-started.html`.\n"
