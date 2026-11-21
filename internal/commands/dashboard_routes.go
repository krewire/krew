package commands

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krewire/pkg/model"
)

type ProjectRoute struct {
	Path string `json:"path"`
	Kind string `json:"kind"` // "page", "content", "api"
	File string `json:"file"`
}

// ProjectService represents a configured or running microservice in the project.
type ProjectService struct {
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// ProjectInfo provides full metadata about the project, its shape, and environment.
type ProjectInfo struct {
	Name      string           `json:"name"`
	Kind      string           `json:"kind"`
	Env       string           `json:"env"`
	Root      string           `json:"root"`
	Version   string           `json:"version"`
	GoVersion string           `json:"go_version"`
	Platform  string           `json:"platform"`
	Port      string           `json:"port"`
	Services  []ProjectService `json:"services"`
	Workers   []string         `json:"workers"`
	Routes    []ProjectRoute   `json:"routes"`
	Config    map[string]any   `json:"config"`
}

// RuntimeMetrics encapsulates real-time memory and goroutine telemetry.
type RuntimeMetrics struct {
	Goroutines    int     `json:"goroutines"`
	AllocMB       float64 `json:"alloc_mb"`
	TotalAllocMB  float64 `json:"total_alloc_mb"`
	SysMB         float64 `json:"sys_mb"`
	NumGC         uint32  `json:"num_gc"`
	UptimeSeconds int64   `json:"uptime_seconds"`
	Timestamp     string  `json:"timestamp"`
}

// HealthResponse holds the standardized health check contract.
type HealthResponse struct {
	Status        string  `json:"status"`
	Kind          string  `json:"kind"`
	UptimeSeconds int64   `json:"uptime_seconds"`
	Goroutines    int     `json:"goroutines"`
	MemoryMB      float64 `json:"memory_mb"`
	Timestamp     string  `json:"timestamp"`
}

func discoverRoutes(root string) []ProjectRoute {
	var routes []ProjectRoute
	seen := make(map[string]bool)

	addRoute := func(p, kind, file string) {
		if !seen[p] {
			seen[p] = true
			routes = append(routes, ProjectRoute{Path: p, Kind: kind, File: file})
		}
	}

	// 1. Pages directory (.kiw, .html, .md)
	pagesDir := filepath.Join(root, "pages")
	if info, err := os.Stat(pagesDir); err == nil && info.IsDir() {
		_ = filepath.WalkDir(pagesDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if ext != ".kiw" && ext != ".html" && ext != ".md" {
				return nil
			}
			rel, err := filepath.Rel(pagesDir, path)
			if err != nil {
				return nil
			}
			routePath := "/" + filepath.ToSlash(rel)
			routePath = strings.TrimSuffix(routePath, ext)
			if strings.HasSuffix(routePath, "/index") {
				routePath = strings.TrimSuffix(routePath, "index")
			}
			if routePath == "" {
				routePath = "/"
			}
			addRoute(routePath, "page", filepath.Join("pages", rel))
			return nil
		})
	}

	// 2. Content directory (.md, .kiw)
	contentDir := "content"
	dir := filepath.Join(root, contentDir)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if ext != ".md" && ext != ".kiw" {
				return nil
			}
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return nil
			}
			routePath := "/" + filepath.ToSlash(rel)
			routePath = strings.TrimSuffix(routePath, ext)
			if strings.HasSuffix(routePath, "/index") {
				routePath = strings.TrimSuffix(routePath, "index")
			}
			if routePath == "" {
				routePath = "/"
			}
			addRoute(routePath, "content", filepath.Join(contentDir, rel))
			return nil
		})
	}

	// 3. Handlers directory (.go)
	for _, dirName := range []string{"handlers", filepath.Join("internal", "handlers"), "api"} {
		dir := filepath.Join(root, dirName)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
					return nil
				}
				rel, err := filepath.Rel(dir, path)
				if err != nil {
					return nil
				}
				base := strings.TrimSuffix(rel, ".go")
				routePath := "/api/" + filepath.ToSlash(base)
				addRoute(routePath, "api", filepath.Join(dirName, rel))
				return nil
			})
		}
	}

	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path < routes[j].Path
	})
	return routes
}

func discoverServices(cfg *config.Config) []ProjectService {
	var svcs []ProjectService
	if cfg == nil {
		return svcs
	}
	for _, s := range cfg.Services {
		svcs = append(svcs, ProjectService{
			Name:     s.Name,
			Port:     s.Port,
			Protocol: "http",
		})
	}
	if cfg.Runtime.Port > 0 {
		svcs = append(svcs, ProjectService{
			Name:     "runtime",
			Port:     cfg.Runtime.Port,
			Protocol: "http",
		})
	}
	if cfg.Dev.Port > 0 {
		svcs = append(svcs, ProjectService{
			Name:     "dev",
			Port:     cfg.Dev.Port,
			Protocol: "http",
		})
	}
	if cfg.Server.Port > 0 {
		svcs = append(svcs, ProjectService{
			Name:     "server",
			Port:     cfg.Server.Port,
			Protocol: "http",
		})
	}
	return svcs
}

func discoverWorkers(cfg *config.Config) []string {
	var workers []string
	if cfg == nil {
		return workers
	}
	if cfg.Kind() == string(model.KindWorker) {
		workers = append(workers, "default (worker kind)")
	}
	return workers
}
