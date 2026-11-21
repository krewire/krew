// Package-level file for the standalone microservice application (VariantService).
package scaffold

import (
	"fmt"
	"os"
	"os/exec"
)

// equipService shapes the bootstrap project into a standalone microservice application.
func equipService(opts EquipOptions) ([]string, error) {
	if opts.Module == "" {
		return nil, fmt.Errorf("equip service: module path is required (read from go.mod)")
	}
	if opts.Name == "" {
		return nil, fmt.Errorf("equip service: project name is required")
	}
	version := opts.Version
	if version == "" {
		version = "latest"
	}

	cwd, _ := os.Getwd()
	localPath := detectLocalModulePath(cwd)

	files := []file{
		{goModFile, goModTemplate(opts.Module, version, localPath)},
		{krewireYaml, serviceKrewireYamlTemplate(opts.Name)},
		{mainGo, serviceMainTemplate(opts.Module, opts.Name)},
		{"internal/handler/handler.go", serviceHandlerTemplate()},
		{"README.md", serviceReadmeTemplate(opts.Name)},
		{gitignoreFile, gitignoreTemplate(opts.Name)},
	}
	report, err := writeVariant(opts.Dir, files)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = opts.Dir
	_ = cmd.Run()
	return report, nil
}

func serviceKrewireYamlTemplate(name string) string {
	return fmt.Sprintf(`project:
  name: %s
  kind: service
  version: 0.1.0

service:
  name: %s
  port: 8081
`, name, name)
}

func serviceMainTemplate(module, name string) string {
	return fmt.Sprintf(`// Entry point: run the standalone microservice.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krewire/krewire/pkg/cloud/service/registry"
	"%s/internal/handler"
)

func main() {
	port := os.Getenv("APP_ADDR")
	if port == "" {
		port = ":8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/api/v1/status", handler.Status)

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	// In-memory service discovery registry registration
	reg := registry.NewMemoryRegistry()
	_ = reg.Register(context.Background(), registry.Service{
		ID:      "%s-1",
		Name:    "%s",
		Address: "http://localhost" + port,
	})

	slog.Info("service listening", "name", "%s", "addr", port)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "service error: %%v\n", err)
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	slog.Info("shutting down service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = server.Shutdown(ctx)
	_ = reg.Deregister(context.Background(), "%s-1")
	slog.Info("service stopped")
}
`, module, name, name, name, name)
}

func serviceHandlerTemplate() string {
	return `// Package handler implements HTTP endpoints for the service.
package handler

import (
	"encoding/json"
	"net/http"
)

// Health responds with a 200 OK health check status.
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// Status responds with service metadata.
func Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"service": "active",
		"version": "0.1.0",
	})
}
`
}

func serviceReadmeTemplate(name string) string {
	return fmt.Sprintf(`# %s (Microservice)

Microservice powered by Krewire.

## Development

Run the service locally:

`+"```bash"+`
krew run
`+"```"+`

Run with hot reload:

`+"```bash"+`
krew dev
`+"```"+`

## Building & Deploying

Compile into a production binary:

`+"```bash"+`
krew build
`+"```"+`

Stage deployable artifacts:

`+"```bash"+`
krew deploy
`+"```"+`
`, name)
}
