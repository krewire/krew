package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/krewire/krew/internal/buildinfo"
	"github.com/krewire/krew/internal/config"
	"github.com/krewire/krew/internal/mcp"
)

const (
	// DefaultDashboardPort is the default HTTP listen port for the developer dashboard.
	DefaultDashboardPort = "4000"

	// DefaultDashboardHost is the default host address binding for the developer dashboard.
	DefaultDashboardHost = "127.0.0.1"

	// DefaultDashboardShutdownTimeout is the maximum duration allocated for server drainage.
	DefaultDashboardShutdownTimeout = 3 * time.Second

	// DefaultDashboardPulseInterval is the SSE telemetry push cadence.
	DefaultDashboardPulseInterval = 2 * time.Second
)

type Server struct {
	root      string
	cfg       *config.Config
	env       string
	host      string
	port      string
	startTime time.Time
	listener  net.Listener
	server    *http.Server
}

func NewServer(root string, cfg *config.Config, env, host, port string) *Server {
	return &Server{
		root:      root,
		cfg:       cfg,
		env:       env,
		host:      host,
		port:      port,
		startTime: time.Now(),
	}
}

func (d *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", d.handleHome)
	mux.HandleFunc("/api/info", d.handleInfo)
	mux.HandleFunc("/api/metrics", d.handleMetrics)
	mux.HandleFunc("/api/health", d.handleHealth)
	mux.HandleFunc("/api/events", d.handleEvents)

	mcpSrv := mcp.NewServer(d.root, d.cfg)
	mcpTransport := mcp.NewHTTPTransport(mcpSrv, mcp.SecurityOptions{
		AllowedHosts: []string{d.host},
	})
	mcpTransport.Mount(mux, "/mcp")

	return mux
}

func (d *Server) Start() error {
	addr := net.JoinHostPort(d.host, d.port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("dashboard listen on %s: %w", addr, err)
	}
	d.listener = ln
	d.server = &http.Server{
		Handler: d.Handler(),
	}
	go func() {
		_ = d.server.Serve(ln)
	}()
	return nil
}

func (d *Server) Addr() string {
	if d.listener != nil {
		return d.listener.Addr().String()
	}
	return net.JoinHostPort(d.host, d.port)
}

func (d *Server) Stop(ctx context.Context) error {
	if d.server != nil {
		return d.server.Shutdown(ctx)
	}
	return nil
}

func (d *Server) CollectMetrics() RuntimeMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := int64(time.Since(d.startTime).Seconds())
	return RuntimeMetrics{
		Goroutines:    runtime.NumGoroutine(),
		AllocMB:       float64(m.Alloc) / 1024 / 1024,
		TotalAllocMB:  float64(m.TotalAlloc) / 1024 / 1024,
		SysMB:         float64(m.Sys) / 1024 / 1024,
		NumGC:         m.NumGC,
		UptimeSeconds: uptime,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
}

func (d *Server) CollectInfo() ProjectInfo {
	name := ""
	if d.cfg != nil {
		name = d.cfg.Title
	}
	if name == "" {
		name = filepath.Base(d.root)
	}
	routes := discoverRoutes(d.root)
	svcs := discoverServices(d.cfg)
	workers := discoverWorkers(d.cfg)

	cfgMap := map[string]any{
		"env": d.env,
	}
	if d.cfg != nil {
		cfgMap["title"] = d.cfg.Title
		cfgMap["base"] = d.cfg.Base
		cfgMap["output"] = d.cfg.Output
		cfgMap["version"] = d.cfg.Version
	}

	ver, fromSource := buildinfo.ResolveVersion(buildinfo.ModKrewire)
	if fromSource && ver != "" && ver != "dev" {
		ver = ver + " (dev)"
	}
	if ver == "" {
		ver = "dev"
	}

	return ProjectInfo{
		Name:      name,
		Env:       d.env,
		Root:      d.root,
		Version:   ver,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
		Port:      d.port,
		Services:  svcs,
		Workers:   workers,
		Routes:    routes,
		Config:    cfgMap,
	}
}

func (d *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	info := d.CollectInfo()
	initialJSON, err := json.Marshal(info)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := dashboardTemplate.Execute(w, map[string]any{
		"Title":       info.Name,
		"InitialJSON": template.JS(initialJSON),
	}); err != nil {
		http.Error(w, "Template Render Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (d *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	info := d.CollectInfo()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(info)
}

func (d *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := d.CollectMetrics()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(metrics)
}

func (d *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	metrics := d.CollectMetrics()
	res := HealthResponse{
		Status:        "ok",
		UptimeSeconds: metrics.UptimeSeconds,
		Goroutines:    metrics.Goroutines,
		MemoryMB:      metrics.AllocMB,
		Timestamp:     metrics.Timestamp,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(res)
}

func (d *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	initialMetrics := d.CollectMetrics()
	if b, err := json.Marshal(initialMetrics); err == nil {
		fmt.Fprintf(w, "event: pulse\ndata: %s\n\n", b)
		flusher.Flush()
	}

	ticker := time.NewTicker(DefaultDashboardPulseInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			m := d.CollectMetrics()
			b, err := json.Marshal(m)
			if err != nil {
				return
			}
			if _, err := fmt.Fprintf(w, "event: pulse\ndata: %s\n\n", b); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
