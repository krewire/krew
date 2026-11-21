// Package mcp implements a native Model Context Protocol (MCP) server for
// Krewire projects (KWN-MCP-001). It enables AI coding agents (Antigravity,
// Claude, Cursor, OpenCode) to inspect routes, stream dev telemetry, and
// invoke scaffolding tools.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/krewire/krew/internal/buildinfo"
	kiwconfig "github.com/krewire/krew/internal/config"
	kiwcontext "github.com/krewire/krew/internal/context"
	"github.com/krewire/krew/internal/scaffold"
	"github.com/krewire/krew/internal/shape"
	"github.com/krewire/krewire/pkg/model"
	"github.com/krewire/krewire/pkg/sec"
)

// ProtocolVersion is the MCP specification version supported.
const ProtocolVersion = "2024-11-05"

// Request represents a JSON-RPC 2.0 request or notification.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

// Error represents a JSON-RPC 2.0 error object.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Standard JSON-RPC error codes.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// Server coordinates MCP protocol interactions for a Krewire project.
type Server struct {
	root      string
	config    *kiwconfig.Config
	startTime time.Time
}

// NewServer initializes a new MCP Server for the given project root and config.
func NewServer(root string, config *kiwconfig.Config) *Server {
	return &Server{
		root:      root,
		config:    config,
		startTime: time.Now(),
	}
}

// HandleRequest processes an incoming JSON-RPC request and returns a response,
// or nil if the request was a notification.
func (s *Server) HandleRequest(ctx context.Context, req Request) *Response {
	if req.ID == nil && strings.HasPrefix(req.Method, "notifications/") {
		return nil // Notifications do not receive a response
	}

	res, err := s.dispatch(ctx, req.Method, req.Params)
	if err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   err,
		}
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  res,
	}
}

func (s *Server) dispatch(ctx context.Context, method string, paramsRaw json.RawMessage) (any, *Error) {
	switch method {
	case "initialize":
		return map[string]any{
			"protocolVersion": ProtocolVersion,
			"capabilities": map[string]any{
				"tools":     map[string]any{},
				"resources": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "krewire-mcp",
				"version": "0.1.0",
			},
		}, nil

	case "ping":
		return map[string]any{}, nil

	case "resources/list":
		return map[string]any{
			"resources": []map[string]any{
				{
					"uri":         "krewire://project/info",
					"name":        "Project Information",
					"description": "Project kind, configuration, Go toolchain, and services manifest",
					"mimeType":    "application/json",
				},
				{
					"uri":         "krewire://project/routes",
					"name":        "Project Routes & Endpoints",
					"description": "List of discovered pages (.kiw), markdown documents, and Go HTTP handlers",
					"mimeType":    "application/json",
				},
				{
					"uri":         "krewire://runtime/metrics",
					"name":        "Runtime Telemetry & Memory",
					"description": "Live goroutine count, memory allocations (MB), and GC statistics",
					"mimeType":    "application/json",
				},
				{
					"uri":         "krewire://runtime/health",
					"name":        "Health Probe Status",
					"description": "Project uptime and health check status",
					"mimeType":    "application/json",
				},
			},
		}, nil

	case "resources/read":
		var p struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(paramsRaw, &p); err != nil {
			return nil, &Error{Code: CodeInvalidParams, Message: "invalid resource uri parameter"}
		}
		text, err := s.readResource(p.URI)
		if err != nil {
			return nil, &Error{Code: CodeInvalidParams, Message: err.Error()}
		}
		return map[string]any{
			"contents": []map[string]any{
				{
					"uri":      p.URI,
					"mimeType": "application/json",
					"text":     text,
				},
			},
		}, nil

	case "tools/list":
		return map[string]any{
			"tools": []map[string]any{
				{
					"name":        "kiw_inspect_project",
					"description": "Inspect Krewire project metadata, workload kind, environment, Go toolchain, and services.",
					"inputSchema": map[string]any{
						"type":       "object",
						"properties": map[string]any{},
					},
				},
				{
					"name":        "kiw_list_routes",
					"description": "List all discovered pages (.kiw), markdown documents, and Go HTTP handlers in the project.",
					"inputSchema": map[string]any{
						"type":       "object",
						"properties": map[string]any{},
					},
				},
				{
					"name":        "kiw_make_scaffold",
					"description": "Scaffold a new component, page, handler, model, middleware, or service using Krewire generators.",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"kind": map[string]any{
								"type":        "string",
								"description": "Generator kind: component, page, handler, migration, model, middleware, service",
								"enum":        []string{"component", "page", "handler", "migration", "model", "middleware", "service"},
							},
							"name": map[string]any{
								"type":        "string",
								"description": "Name for the artifact (e.g. Card, UserHandler, AuthMiddleware)",
							},
						},
						"required": []string{"kind", "name"},
					},
				},
				{
					"name":        "kiw_get_metrics",
					"description": "Retrieve live goroutine count, memory allocations, total sys memory, and GC statistics.",
					"inputSchema": map[string]any{
						"type":       "object",
						"properties": map[string]any{},
					},
				},
				{
					"name":        "kiw_run_tests",
					"description": "Run the project test suite (go test ./...) and return structured test outcome.",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"pkg": map[string]any{
								"type":        "string",
								"description": "Target package path (default: ./...)",
							},
						},
					},
				},
				{
					"name":        "kiw_context",
					"description": "Inspect authoritative project context (metadata, workload kind, Go version, tier, Three Pillars invariants) as machine-readable JSON per ENG-AI-001.",
					"inputSchema": map[string]any{
						"type":       "object",
						"properties": map[string]any{},
					},
				},
				{
					"name":        "kiw_verify",
					"description": "Execute canonical quality gates (fmt, vet, test, build, layering) and generate a tamper-evident task verification record, optionally appending to the verification ledger.",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"task_id": map[string]any{
								"type":        "string",
								"description": "Task or issue identifier (e.g. issue-42)",
							},
							"agent_id": map[string]any{
								"type":        "string",
								"description": "Agent identifier performing verification",
							},
							"save_ledger": map[string]any{
								"type":        "boolean",
								"description": "Whether to append evidence block to .krewire/ledger.json",
							},
						},
					},
				},
			},
		}, nil

	case "tools/call":
		var p struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(paramsRaw, &p); err != nil {
			return nil, &Error{Code: CodeInvalidParams, Message: "invalid tool call parameters"}
		}
		resultText, isErr := s.callTool(ctx, p.Name, p.Arguments)
		return map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": resultText,
				},
			},
			"isError": isErr,
		}, nil

	default:
		return nil, &Error{Code: CodeMethodNotFound, Message: fmt.Sprintf("method not found: %s", method)}
	}
}

func (s *Server) readResource(uri string) (string, error) {
	switch uri {
	case "krewire://project/info":
		info := s.collectInfo()
		b, _ := json.MarshalIndent(info, "", "  ")
		return string(b), nil
	case "krewire://project/routes":
		routes := s.collectRoutes()
		b, _ := json.MarshalIndent(routes, "", "  ")
		return string(b), nil
	case "krewire://runtime/metrics":
		metrics := s.collectMetrics()
		b, _ := json.MarshalIndent(metrics, "", "  ")
		return string(b), nil
	case "krewire://runtime/health":
		m := s.collectMetrics()
		health := map[string]any{
			"status":         "ok",
			"uptime_seconds": m["uptime_seconds"],
			"goroutines":     m["goroutines"],
			"timestamp":      m["timestamp"],
		}
		b, _ := json.MarshalIndent(health, "", "  ")
		return string(b), nil
	default:
		return "", fmt.Errorf("unknown resource uri: %s", uri)
	}
}

func (s *Server) callTool(ctx context.Context, name string, argsRaw json.RawMessage) (string, bool) {
	switch name {
	case "kiw_inspect_project":
		info := s.collectInfo()
		b, _ := json.MarshalIndent(info, "", "  ")
		return string(b), false

	case "kiw_list_routes":
		routes := s.collectRoutes()
		b, _ := json.MarshalIndent(routes, "", "  ")
		return string(b), false

	case "kiw_get_metrics":
		metrics := s.collectMetrics()
		b, _ := json.MarshalIndent(metrics, "", "  ")
		return string(b), false

	case "kiw_make_scaffold":
		var p struct {
			Kind string `json:"kind"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(argsRaw, &p); err != nil {
			return fmt.Sprintf("invalid arguments: %v", err), true
		}
		if err := scaffold.MakeArtifact(s.root, p.Kind, p.Name); err != nil {
			return fmt.Sprintf("scaffolding failed: %v", err), true
		}
		return fmt.Sprintf("Successfully scaffolded %s %q in project", p.Kind, p.Name), false

	case "kiw_run_tests":
		pkg := "./..."
		var p struct {
			Pkg string `json:"pkg"`
		}
		if len(argsRaw) > 0 {
			_ = json.Unmarshal(argsRaw, &p)
		}
		if p.Pkg != "" {
			pkg = p.Pkg
		}
		cmd := exec.CommandContext(ctx, "go", "test", pkg)
		cmd.Dir = s.root
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("Tests failed:\n%s", string(out)), true
		}
		return fmt.Sprintf("Tests passed:\n%s", string(out)), false

	case "kiw_context":
		projectCtx, err := kiwcontext.Load(s.root)
		if err != nil {
			return fmt.Sprintf("failed to load project context: %v", err), true
		}
		b, err := json.MarshalIndent(projectCtx, "", "  ")
		if err != nil {
			return fmt.Sprintf("failed to serialize context: %v", err), true
		}
		return string(b), false

	case "kiw_verify":
		var vp struct {
			TaskID     string `json:"task_id"`
			AgentID    string `json:"agent_id"`
			SaveLedger bool   `json:"save_ledger"`
		}
		if len(argsRaw) > 0 {
			_ = json.Unmarshal(argsRaw, &vp)
		}
		if vp.TaskID == "" {
			vp.TaskID = "mcp-task"
		}
		if vp.AgentID == "" {
			vp.AgentID = "mcp-agent"
		}

		gitCommit := "unversioned"
		workingTreeClean := false
		if out, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
			gitCommit = strings.TrimSpace(string(out))
		}
		if out, err := exec.Command("git", "status", "--porcelain").Output(); err == nil {
			workingTreeClean = len(strings.TrimSpace(string(out))) == 0
		}

		gateDefs := []struct {
			ID          string
			Type        model.QualityGateType
			Description string
			Command     string
			Args        []string
			Required    bool
		}{
			{ID: "fmt", Type: model.GateTypeFormat, Description: "Go code formatting check (gofmt -l .)", Command: "gofmt", Args: []string{"-l", "."}, Required: true},
			{ID: "vet", Type: model.GateTypeStatic, Description: "Go static analysis and vet checks (go vet ./...)", Command: "go", Args: []string{"vet", "./..."}, Required: true},
			{ID: "test", Type: model.GateTypeTest, Description: "Automated unit and integration test suite (go test ./...)", Command: "go", Args: []string{"test", "./..."}, Required: true},
			{ID: "build", Type: model.GateTypeBuild, Description: "Pure-Go static compilation check (go build ./...)", Command: "go", Args: []string{"build", "./..."}, Required: true},
		}

		if _, err := os.Stat(filepath.Join(s.root, "layering_test.go")); err == nil {
			gateDefs = append(gateDefs, struct {
				ID          string
				Type        model.QualityGateType
				Description string
				Command     string
				Args        []string
				Required    bool
			}{ID: "layering", Type: model.GateTypeLayering, Description: "Downward-only layering architecture guard (layering_test.go)", Command: "go", Args: []string{"test", "-v", "./layering_test.go"}, Required: true})
		}

		var evidences []model.VerificationEvidence
		for _, def := range gateDefs {
			cmdStr := def.Command + " " + strings.Join(def.Args, " ")
			ev := model.VerificationEvidence{
				GateID:      def.ID,
				Type:        def.Type,
				Description: def.Description,
				Command:     cmdStr,
				Required:    def.Required,
				Status:      model.GateAttempted,
			}
			start := time.Now()
			cmd := exec.CommandContext(ctx, def.Command, def.Args...)
			cmd.Dir = s.root
			cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
			out, cmdErr := cmd.CombinedOutput()
			ev.DurationMs = time.Since(start).Milliseconds()
			rawOutput := strings.TrimSpace(string(out))

			if def.ID == "fmt" && len(rawOutput) > 0 {
				ev.Status = model.GateFailed
				ev.ExitCode = 1
				ev.Output = fmt.Sprintf("unformatted files detected:\n%s", rawOutput)
			} else if cmdErr != nil {
				ev.Status = model.GateFailed
				if exitErr, ok := cmdErr.(*exec.ExitError); ok {
					ev.ExitCode = exitErr.ExitCode()
				} else {
					ev.ExitCode = 1
				}
				ev.Output = sec.RedactSensitiveOutput(rawOutput)
			} else {
				ev.Status = model.GatePassed
				ev.ExitCode = 0
				ev.Output = sec.RedactSensitiveOutput(rawOutput)
			}
			ev.EvidenceHash = ev.ComputeHash(gitCommit)
			evidences = append(evidences, ev)
		}

		record := model.NewTaskVerificationRecord(vp.TaskID, vp.AgentID, gitCommit, workingTreeClean, evidences)
		if vp.SaveLedger {
			ledgerPath := filepath.Join(s.root, ".krewire", "ledger.json")
			var l *model.VerificationLedger
			if loaded, err := model.LoadLedgerFile(ledgerPath); err == nil {
				l = loaded
			} else {
				l = model.NewVerificationLedger(vp.TaskID)
			}
			if _, err := l.AppendBlock(gitCommit, evidences); err == nil {
				_ = l.SaveFile(ledgerPath)
			}
		}

		b, err := json.MarshalIndent(record, "", "  ")
		if err != nil {
			return fmt.Sprintf("failed to serialize record: %v", err), true
		}
		return string(b), !record.OverallVerified

	default:
		return fmt.Sprintf("unknown tool %q", name), true
	}
}

func (s *Server) collectMetrics() map[string]any {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	uptime := int64(time.Since(s.startTime).Seconds())

	return map[string]any{
		"goroutines":     runtime.NumGoroutine(),
		"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
		"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
		"sys_mb":         float64(m.Sys) / 1024 / 1024,
		"num_gc":         m.NumGC,
		"uptime_seconds": uptime,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}
}

func (s *Server) collectInfo() map[string]any {
	name := ""
	if s.config != nil {
		name = s.config.Title
	}
	if name == "" {
		name = filepath.Base(s.root)
	}

	explicit := ""
	if s.config != nil {
		explicit = s.config.Kind()
	}
	res, _ := shape.Detect(s.root, explicit)

	var svcs []map[string]any
	if s.config != nil {
		for _, svc := range s.config.Services {
			svcs = append(svcs, map[string]any{
				"name": svc.Name,
				"port": svc.Port,
			})
		}
	}

	return map[string]any{
		"name":       name,
		"kind":       res.String(),
		"root":       s.root,
		"go_version": runtime.Version(),
		"platform":   runtime.GOOS + "/" + runtime.GOARCH,
		"version":    func() string { v, _ := buildinfo.ResolveVersion(buildinfo.ModKrewire); return v }(),
		"services":   svcs,
	}
}

func (s *Server) collectRoutes() []map[string]string {
	var routes []map[string]string
	// Scans pages/
	pagesDir := filepath.Join(s.root, "pages")
	_ = filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".kiw" || ext == ".html" || ext == ".md" {
			rel, _ := filepath.Rel(pagesDir, path)
			route := "/" + filepath.ToSlash(rel)
			route = strings.TrimSuffix(route, ext)
			if strings.HasSuffix(route, "/index") {
				route = strings.TrimSuffix(route, "index")
			}
			if route == "" {
				route = "/"
			}
			routes = append(routes, map[string]string{
				"path": route,
				"kind": "page",
				"file": filepath.Join("pages", rel),
			})
		}
		return nil
	})
	return routes
}
