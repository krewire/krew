package dsl

import "testing"

// The parser moved to github.com/krewire/krewire/pkg/kiw so that packages can parse a
// .kiw file without importing the devtool (the layering is strictly downward:
// boost → kiw → pkg/* → kern). This package survives as a
// compatibility alias, so its forwarding is the contract: a type alias, not a
// copy, means a caller cannot observe the difference.
func TestAliasForwardsToCanonicalParser(t *testing.T) {
	src := "---\ntitle: Landing\n---\n<h1>hi</h1>\n<style>a{}</style>"
	mod, err := ParseKiw(src)
	if err != nil {
		t.Fatal(err)
	}
	if mod.Frontmatter["title"] != "Landing" {
		t.Errorf("frontmatter title = %v, want Landing", mod.Frontmatter["title"])
	}
	if len(mod.Styles) != 1 {
		t.Errorf("styles = %v, want one entry", mod.Styles)
	}
	if _, err := DesugarTemplate(`<Card Title="x" />`); err != nil {
		t.Fatal(err)
	}
}
