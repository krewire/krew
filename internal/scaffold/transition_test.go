package scaffold

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krewire/pkg/model"
)

// TestWorkloadTransitions verifies that projects can progressively evolve
// between workload kinds without broken configuration, data loss, or syntax errors.
func TestWorkloadTransitions(t *testing.T) {
	// ── 1. Transition 1: App to App + Worker ─────────────────────────────────
	t.Run("App_To_App_Plus_Worker", func(t *testing.T) {
		dir := t.TempDir()

		// Start with an App
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/myapp\ngo 1.27.1\n"), 0644); err != nil {
			t.Fatal(err)
		}
		appYaml := `project:
  name: myapp
  kind: app
  version: "0.1.0"
server:
  port: 8080
`
		if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(appYaml), 0644); err != nil {
			t.Fatal(err)
		}

		// Equip worker into sub-package internal/worker
		workerDir := filepath.Join(dir, "worker")
		if err := os.MkdirAll(workerDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(workerDir, "main.go"), []byte(bootstrapBody), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(workerDir, "krewire.yaml"), []byte("project:\n  name: myworker\n"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := Equip(EquipOptions{
			Dir:     workerDir,
			Name:    "myworker",
			Module:  "example.com/myapp/worker",
			Variant: VariantWorker,
		})
		if err != nil {
			t.Fatalf("Equip(VariantWorker) failed: %v", err)
		}

		// Verify both configurations load and validate cleanly
		appCfg, err := config.Load(dir)
		if err != nil || model.Kind(appCfg.Project.Kind) != model.KindApp {
			t.Errorf("expected App configuration intact, got: %+v, err: %v", appCfg, err)
		}

		workerCfg, err := config.Load(workerDir)
		if err != nil || model.Kind(workerCfg.Project.Kind) != model.KindWorker {
			t.Errorf("expected Worker configuration intact, got: %+v, err: %v", workerCfg, err)
		}
	})

	// ── 2. Transition 2: Incompatible Variant Refusal ────────────────────────
	t.Run("Refuses_Conflicting_Overwrites", func(t *testing.T) {
		dir := newBootstrap(t, t.TempDir())
		if err := os.MkdirAll(filepath.Join(dir, "internal/app"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "internal/app/app.go"), []byte("custom user app code"), 0o644); err != nil {
			t.Fatal(err)
		}

		_, err := Equip(EquipOptions{
			Dir:     dir,
			Name:    "demo",
			Module:  "example.com/demo",
			Variant: VariantApp,
		})
		if !errors.Is(err, ErrConflict) {
			t.Errorf("expected ErrConflict when equipping variant over existing code, got: %v", err)
		}
	})
}
