// Package editors holds editor integrations for .kiw files; this test keeps
// the TextMate grammar loadable and its sample valid.
package editors

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/krewire/krewire/pkg/kiw"
)

// walk visits every regex-bearing key of a TextMate grammar node.
func walk(t *testing.T, path string, node any, fn func(path, key, re string)) {
	switch v := node.(type) {
	case map[string]any:
		for k, child := range v {
			if s, ok := child.(string); ok && (k == "match" || k == "begin" || k == "end") {
				fn(path, k, s)
				continue
			}
			walk(t, path+"/"+k, child, fn)
		}
	case []any:
		for _, child := range v {
			walk(t, path, child, fn)
		}
	}
}

func TestGrammar(t *testing.T) {
	data, err := os.ReadFile("vscode/syntaxes/kiw.tmLanguage.json")
	if err != nil {
		t.Fatal(err)
	}
	var g map[string]any
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("grammar is not valid JSON: %v", err)
	}
	if g["scopeName"] != "text.html.kiw" {
		t.Errorf("scopeName = %v", g["scopeName"])
	}
	// Oniguruma is a superset of RE2 for the constructs used here, except
	// lookahead, which the grammar uses only as "!(?!=)".
	n := 0
	walk(t, "", g, func(path, key, re string) {
		n++
		if _, err := regexp.Compile(strings.ReplaceAll(re, "(?!=)", "")); err != nil {
			t.Errorf("%s %s: %v", path, key, err)
		}
	})
	if n < 20 {
		t.Errorf("only %d patterns found", n)
	}
	repo, _ := g["repository"].(map[string]any)
	for _, rule := range []string{"frontmatter", "blocks", "interpolation", "template-action", "component-tag", "style", "script", "markdown"} {
		if _, ok := repo[rule]; !ok {
			t.Errorf("repository rule %q missing", rule)
		}
	}

	var pkg struct {
		Contributes struct {
			Grammars []struct{ Path, ScopeName string } `json:"grammars"`
		} `json:"contributes"`
	}
	data, err = os.ReadFile("vscode/package.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &pkg); err != nil || len(pkg.Contributes.Grammars) != 1 || pkg.Contributes.Grammars[0].ScopeName != "text.html.kiw" {
		t.Errorf("package.json grammar contribution invalid: %v %+v", err, pkg)
	}
	var lc map[string]any
	if data, err = os.ReadFile("vscode/language-configuration.json"); err != nil || json.Unmarshal(data, &lc) != nil {
		t.Errorf("language-configuration.json invalid: %v", err)
	}
}

func TestSampleLintsClean(t *testing.T) {
	src, err := os.ReadFile("vscode/samples/Sample.kiw")
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range kiw.Lint("Sample.kiw", string(src)) {
		t.Errorf("%v", &d)
	}
	if _, err := kiw.ParseKiw(string(src)); err != nil {
		t.Fatal(err)
	}
}
