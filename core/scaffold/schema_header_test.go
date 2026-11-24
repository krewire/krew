package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/krew/core/config"
)

// Every scaffolded krewire.yaml points editors at the JSON Schema and
// validates against it without issues.
func TestScaffoldedConfigsCarrySchemaAndValidate(t *testing.T) {
	check := func(name, dir string) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(dir, "krewire.yaml"))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !strings.HasPrefix(string(data), config.SchemaHeader+"\n") {
			t.Errorf("%s: krewire.yaml lacks the $schema modeline:\n%s", name, data)
		}
		issues, err := config.Validate(data)
		if err != nil || len(issues) > 0 {
			t.Errorf("%s: krewire.yaml does not validate: %v %v", name, err, issues)
		}
	}

	dir := t.TempDir()
	if _, err := Init(InitOptions{Dir: dir, Name: "demo"}); err != nil {
		t.Fatal(err)
	}
	check("init", dir)

	for _, v := range []Variant{VariantApp, VariantStatic, VariantBook, VariantCLI, VariantWorker, VariantService, VariantGateway, VariantInfra} {
		dir := t.TempDir()
		if _, err := Equip(EquipOptions{Dir: dir, Variant: v, Name: "demo", Title: "Demo", Module: "example.com/demo", Version: "v0.0.0"}); err != nil {
			t.Fatalf("%s: %v", v, err)
		}
		check(string(v), dir)
	}
}
