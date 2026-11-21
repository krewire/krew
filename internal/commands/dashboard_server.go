package commands

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

type dashboardServer struct {
	root      string
	cfg       *config.Config
	kind      string
	env       string
	host      string
	port      string
	startTime time.Time
	listener  net.Listener
	server    *http.Server
}

func newDashboardServer(root string, cfg *config.Config, kind, env, host, port string) *dashboardServer {
	return &dashboardServer{
		root:      root,
		cfg:       cfg,
		kind:      kind,
		env:       env,
		host:      host,
		port:      port,
		startTime: time.Now(),
	}
}

func (d *dashboardServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", d.handleHome)
	mux.HandleFunc("/api/info", d.handleInfo)
	mux.HandleFunc("/api/metrics", d.handleMetrics)
	mux.HandleFunc("/api/health", d.handleHealth)
	mux.HandleFunc("/api/events", d.handleEvents)

	mcpSrv := mcp.NewServer(d.root, d.cfg)
	mcpTransport := mcp.NewHTTPTransport(mcpSrv)
	mcpTransport.Mount(mux, "/mcp")

	return mux
}

func (d *dashboardServer) Start() error {
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

func (d *dashboardServer) Addr() string {
	if d.listener != nil {
		return d.listener.Addr().String()
	}
	return net.JoinHostPort(d.host, d.port)
}

func (d *dashboardServer) Stop(ctx context.Context) error {
	if d.server != nil {
		return d.server.Shutdown(ctx)
	}
	return nil
}

func (d *dashboardServer) collectMetrics() RuntimeMetrics {
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

func (d *dashboardServer) collectInfo() ProjectInfo {
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
		"kind": d.kind,
		"env":  d.env,
	}
	if d.cfg != nil {
		cfgMap["title"] = d.cfg.Title
		cfgMap["base"] = d.cfg.Base
		cfgMap["output"] = d.cfg.Output
		cfgMap["version"] = d.cfg.Version
	}

	return ProjectInfo{
		Name:      name,
		Kind:      d.kind,
		Env:       d.env,
		Root:      d.root,
		Version:   qualifiedVersion(buildinfo.ModKrewire),
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
		Port:      d.port,
		Services:  svcs,
		Workers:   workers,
		Routes:    routes,
		Config:    cfgMap,
	}
}

func (d *dashboardServer) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	info := d.collectInfo()
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

func (d *dashboardServer) handleInfo(w http.ResponseWriter, r *http.Request) {
	info := d.collectInfo()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(info)
}

func (d *dashboardServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := d.collectMetrics()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(metrics)
}

func (d *dashboardServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	metrics := d.collectMetrics()
	res := HealthResponse{
		Status:        "ok",
		Kind:          d.kind,
		UptimeSeconds: metrics.UptimeSeconds,
		Goroutines:    metrics.Goroutines,
		MemoryMB:      metrics.AllocMB,
		Timestamp:     metrics.Timestamp,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(res)
}

func (d *dashboardServer) handleEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Send initial pulse immediately
	initialMetrics := d.collectMetrics()
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
			m := d.collectMetrics()
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
