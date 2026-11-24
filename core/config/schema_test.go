package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSchemaIsDraft07WithCanonicalID(t *testing.T) {
	var raw map[string]any
	if err := json.Unmarshal(SchemaJSON, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["$schema"] != "http://json-schema.org/draft-07/schema#" {
		t.Errorf("$schema = %v, want Draft-07", raw["$schema"])
	}
	if raw["$id"] != SchemaURL {
		t.Errorf("$id = %v, want %s", raw["$id"], SchemaURL)
	}
}

// TestSchemaMatchesConfig keeps the schema and the Config struct in step:
// every yaml key krew reads is documented, and every documented key in a
// strict (additionalProperties: false) object is read by kiw.
func TestSchemaMatchesConfig(t *testing.T) {
	var walk func(rt reflect.Type, s *schemaNode, path string)
	walk = func(rt reflect.Type, s *schemaNode, path string) {
		s = s.resolve()
		for rt.Kind() == reflect.Pointer || rt.Kind() == reflect.Slice {
			if rt.Kind() == reflect.Slice {
				if s.Items == nil {
					return
				}
				s = s.Items.resolve()
			}
			rt = rt.Elem()
		}
		if rt.Kind() != reflect.Struct || s.Properties == nil {
			return
		}
		if _, allowExtra := s.additional(); allowExtra {
			return // loose object (pkg/ui layouts, infra resources)
		}
		seen := map[string]bool{}
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			name := strings.Split(f.Tag.Get("yaml"), ",")[0]
			if name == "" || name == "-" {
				continue
			}
			seen[name] = true
			ps, ok := s.Properties[name]
			if !ok {
				t.Errorf("schema is missing %s (field %s.%s)", joinPath(path, name), rt.Name(), f.Name)
				continue
			}
			walk(f.Type, ps, joinPath(path, name))
		}
		for name := range s.Properties {
			if !seen[name] && name != "$schema" {
				t.Errorf("schema documents %s, which %s does not read", joinPath(path, name), rt.Name())
			}
		}
	}
	walk(reflect.TypeOf(Config{}), rootSchema, "")
}

func TestValidate(t *testing.T) {
	src := `# yaml-language-server: $schema=https://krewire.com/schemas/krewire.json
project:
  name: demo
build:
  outdir: dist
  base: /
dev:
  port: 70000
port: 8080
base: /old/
theme:
  default: dark
  light:
    primry: "#000"
gateway:
  routes:
    - path: /api
      rate_limit: fast
`
	issues, err := Validate([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, i := range issues {
		got = append(got, i.String())
	}
	want := []string{
		`5:3: warning: unknown key "outdir" in build; did you mean "build.output"? (ignored)`,
		`8:9: error: dev.port: 70000 is out of range 1..65535`,
		`9:1: warning: unknown key "port"; did you mean "dev.port"? (ignored)`,
		`10:1: warning: base is deprecated. Use build.base instead.`,
		`14:5: warning: unknown key "primry" in theme.light; did you mean "theme.light.primary"? (ignored)`,
		`18:19: error: gateway.routes[0].rate_limit must be integer, got string`,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("issues:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if !HasErrors(issues) {
		t.Error("HasErrors = false")
	}
}

func TestValidateSyntaxError(t *testing.T) {
	if _, err := Validate([]byte("project:\n  name: [x\n")); err == nil || !strings.Contains(err.Error(), "line") {
		t.Errorf("err = %v, want YAML syntax error with line", err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// TestEcosystemConfigsValidate validates the krewire.yaml files of the sibling
// ecosystem repositories, when they are checked out next to this one, and the
// fixtures in this repository. KREWIRE_ECOSYSTEM_DIR overrides the directory
// holding the sibling checkouts (default: the parent of this repository).
func TestEcosystemConfigsValidate(t *testing.T) {
	repo := repoRoot(t)
	if repo == "" {
		t.Skip("not running in repo")
	}
	eco := filepath.Join(repo, "..")
	if dir := os.Getenv("KREWIRE_ECOSYSTEM_DIR"); dir != "" {
		eco = dir
	}
	var files []string
	for _, pattern := range []string{
		filepath.Join(repo, "testtw", "krewire.yaml"),
		filepath.Join(repo, "krewire.yaml"),
		filepath.Join(eco, "*", "krewire.yaml"),
		filepath.Join(eco, "*", "*", "krewire.yaml"),
		filepath.Join(eco, "*", "apps", "*", "krewire.yaml"),
	} {
		m, _ := filepath.Glob(pattern)
		files = append(files, m...)
	}
	checked := 0
	for _, f := range files {
		if strings.Contains(f, string(filepath.Separator)+"templates"+string(filepath.Separator)) {
			continue // Go text/template sources, not YAML
		}
		issues, err := ValidateFile(f)
		if err != nil {
			t.Errorf("%s: %v", f, err)
			continue
		}
		checked++
		for _, i := range issues {
			t.Errorf("%s:%s", f, i)
		}
	}
	if checked < 3 {
		t.Fatalf("validated only %d files", checked)
	}
	t.Logf("validated %d krewire.yaml files", checked)
}

func TestValidateFileMissing(t *testing.T) {
	if _, err := ValidateFile(filepath.Join(t.TempDir(), "krewire.yaml")); !os.IsNotExist(err) {
		t.Errorf("err = %v", err)
	}
}
