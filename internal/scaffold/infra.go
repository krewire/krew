// Package-level file for the declarative infrastructure project (VariantInfra).
package scaffold

import (
	"fmt"
)

// equipInfra shapes the bootstrap project into a declarative cloud infrastructure blueprint project.
func equipInfra(opts EquipOptions) ([]string, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("equip infra: project name is required")
	}

	files := []file{
		{krewireYaml, infraKrewireYamlTemplate(opts.Name)},
		{"infra/blueprint.go", infraBlueprintTemplate(opts.Name)},
		{"README.md", infraReadmeTemplate(opts.Name)},
		{gitignoreFile, gitignoreTemplate(opts.Name)},
	}

	created, err := writeVariant(opts.Dir, files)
	if err != nil {
		return nil, err
	}

	// Infra workload doesn't need root main.go
	if err := removeBootstrapMain(opts.Dir); err != nil {
		return nil, err
	}
	return created, nil
}

func infraKrewireYamlTemplate(name string) string {
	return fmt.Sprintf(`project:
  name: %s
  kind: infra
  version: 0.1.0

infra:
  provider: "krewire-cloud"
  region: "asia-southeast1"

server:
  port: 8080
`, name)
}

func infraBlueprintTemplate(name string) string {
	return fmt.Sprintf(`// Package infra defines cloud topology blueprints for %s.
package infra

import (
	"github.com/krewire/krewire/pkg/cloud/infra/provider"
	"github.com/krewire/krewire/pkg/cloud/infra/schema"
)

// DesiredResources returns the declared infrastructure resources.
func DesiredResources() []provider.Resource {
	return []provider.Resource{
		{
			Kind: schema.KindCompute,
			ID:   "%s-web",
			Properties: map[string]any{
				"runtime":  "fargate",
				"replicas": 2,
			},
		},
		{
			Kind: schema.KindDatabase,
			ID:   "%s-db",
			Properties: map[string]any{
				"engine":  "postgres",
				"storage": 50,
			},
			DependsOn: []string{"%s-web"},
		},
	}
}
`, name, name, name, name)
}

func infraReadmeTemplate(name string) string {
	return fmt.Sprintf(`# %s (Infrastructure)

Declarative cloud infrastructure project powered by Krewire.

## Commands

Evaluate and verify infrastructure blueprint plan:

`+"```bash"+`
krew build --plan
`+"```"+`

Stage deployment plan:

`+"```bash"+`
krew deploy --dry-run
`+"```"+`

Apply infrastructure deployment:

`+"```bash"+`
krew deploy --auto-approve
`+"```"+`
`, name)
}
