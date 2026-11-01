package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ErrPermissionDenied is returned when a caller or agent lacks necessary RBAC permissions.
var ErrPermissionDenied = errors.New("agent: permission denied")

// Permission represents an allowed action on a resource.
type Permission struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
}

// Role defines a set of permissions identified by name.
type Role struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Permissions []Permission `json:"permissions"`
}

// Authorizer evaluates whether caller roles have permission to perform action on resource.
type Authorizer interface {
	HasPermission(roles []string, action, resource string) bool
}

// AuditEvent represents an auditable event.
type AuditEvent struct {
	ActorID   string                 `json:"actor_id"`
	TenantID  string                 `json:"tenant_id"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	TargetID  string                 `json:"target_id"`
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Auditor records audit events.
type Auditor interface {
	Record(ctx context.Context, ev *AuditEvent) error
}

// NoopAuditor discards audit events.
type NoopAuditor struct{}

func (n *NoopAuditor) Record(_ context.Context, _ *AuditEvent) error { return nil }

// MemoryAuditor stores audit events in memory for testing and standalone usage.
type MemoryAuditor struct {
	mu     sync.Mutex
	Events []*AuditEvent
}

func NewMemoryAuditor() *MemoryAuditor {
	return &MemoryAuditor{}
}

func (m *MemoryAuditor) Record(_ context.Context, ev *AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = append(m.Events, ev)
	return nil
}

func (m *MemoryAuditor) All() []*AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*AuditEvent, len(m.Events))
	copy(out, m.Events)
	return out
}

func (m *MemoryAuditor) Query(_ context.Context, tenantID, action string) []*AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []*AuditEvent
	for _, ev := range m.Events {
		if (tenantID == "" || ev.TenantID == tenantID) && (action == "" || ev.Action == action) {
			res = append(res, ev)
		}
	}
	return res
}

// Pre-defined persona roles
var (
	RolePersonaCodingAssistant = Role{
		Name:        "CodingAssistant",
		Description: "Autonomous coding assistant with full workspace tool access",
		Permissions: []Permission{
			{Action: "run", Resource: "agent"},
			{Action: "chat", Resource: "agent"},
			{Action: "fs:read", Resource: "tools"},
			{Action: "fs:write", Resource: "tools"},
			{Action: "exec", Resource: "tools"},
			{Action: "web", Resource: "tools"},
		},
	}

	RolePersonaDevOpsRunner = Role{
		Name:        "DevOpsRunner",
		Description: "DevOps runner with execution, filesystem, and deployment permissions",
		Permissions: []Permission{
			{Action: "run", Resource: "agent"},
			{Action: "fs:read", Resource: "tools"},
			{Action: "fs:write", Resource: "tools"},
			{Action: "exec", Resource: "tools"},
			{Action: "web", Resource: "tools"},
		},
	}

	RolePersonaViewerAssistant = Role{
		Name:        "ViewerAssistant",
		Description: "Read-only assistant restricted from file modifications and command execution",
		Permissions: []Permission{
			{Action: "chat", Resource: "agent"},
			{Action: "fs:read", Resource: "tools"},
			{Action: "web", Resource: "tools"},
		},
	}
)

// MemoryAuthorizer is an in-memory authorizer implementing Authorizer.
type MemoryAuthorizer struct {
	mu    sync.RWMutex
	roles map[string]Role
}

func NewMemoryAuthorizer() *MemoryAuthorizer {
	m := &MemoryAuthorizer{roles: make(map[string]Role)}
	m.RegisterRole(RolePersonaCodingAssistant)
	m.RegisterRole(RolePersonaDevOpsRunner)
	m.RegisterRole(RolePersonaViewerAssistant)
	return m
}

func (m *MemoryAuthorizer) RegisterRole(r Role) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[strings.ToLower(r.Name)] = r
	m.roles[strings.ToLower(strings.TrimSuffix(r.Name, "Assistant"))] = r
	m.roles[strings.ToLower(strings.TrimSuffix(r.Name, "Runner"))] = r
}

func (m *MemoryAuthorizer) HasPermission(roles []string, action, resource string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, rName := range roles {
		role, ok := m.roles[strings.ToLower(rName)]
		if !ok {
			// check alias (e.g. developer -> codingassistant)
			if strings.EqualFold(rName, "developer") {
				role, ok = m.roles["codingassistant"]
			}
		}
		if !ok {
			continue
		}
		for _, perm := range role.Permissions {
			if (perm.Action == "*" || perm.Action == action) &&
				(perm.Resource == "*" || perm.Resource == resource) {
				return true
			}
		}
	}
	return false
}

// ToolPermission maps a tool name to its required RBAC action and resource.
func ToolPermission(toolName string) (action string, resource string) {
	switch toolName {
	case "run_command":
		return "exec", "tools"
	case "fetch_url":
		return "web", "tools"
	case "write_file", "edit_file", "patch_file":
		return "fs:write", "tools"
	case "read_file", "list_dir", "file_search", "content_search":
		return "fs:read", "tools"
	case "subagent":
		return "run", "agent"
	default:
		return toolName, "tools"
	}
}

// CheckToolPermission checks if caller with roles has permission to invoke toolName.
func CheckToolPermission(eval Authorizer, roles []string, toolName string) error {
	if eval == nil {
		return nil
	}
	action, resource := ToolPermission(toolName)
	if !eval.HasPermission(roles, action, resource) {
		return fmt.Errorf("%w: missing %s:%s for tool %s", ErrPermissionDenied, resource, action, toolName)
	}
	return nil
}

// RBACApproverOptions configures role-based tool gating.
type RBACApproverOptions struct {
	Authorizer Authorizer
	Roles      []string
	TenantID   string
	UserID     string
	Audit      Auditor
	OnDenial   func(toolName string, err error)
}

// AgentApprover returns an Approver callback using the provided RBAC options.
// Denied tool invocations are logged to the audit ledger.
func AgentApprover(opts RBACApproverOptions) Approver {
	return func(toolName string, _ json.RawMessage) bool {
		if opts.Authorizer == nil {
			return true
		}
		err := CheckToolPermission(opts.Authorizer, opts.Roles, toolName)
		if err != nil {
			if opts.Audit != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_ = opts.Audit.Record(ctx, &AuditEvent{
					ActorID:   opts.UserID,
					TenantID:  opts.TenantID,
					Action:    "agent.tool_denied",
					Resource:  "tools",
					TargetID:  toolName,
					Status:    "failure",
					Timestamp: time.Now().UTC(),
					Metadata: map[string]interface{}{
						"tool":  toolName,
						"roles": opts.Roles,
						"error": err.Error(),
					},
				})
			}
			if opts.OnDenial != nil {
				opts.OnDenial(toolName, err)
			}
			return false
		}
		return true
	}
}

// NewAgentApprover creates an Approver callback with standard parameters.
func NewAgentApprover(auth Authorizer, roles []string, tenantID, userID string, auditor Auditor) Approver {
	return AgentApprover(RBACApproverOptions{
		Authorizer: auth,
		Roles:      roles,
		TenantID:   tenantID,
		UserID:     userID,
		Audit:      auditor,
	})
}

// DefaultPersonaRole returns the canonical RBAC role definition for a given persona name.
func DefaultPersonaRole(personaName string) (Role, bool) {
	switch strings.ToLower(strings.TrimSpace(personaName)) {
	case "coding", "codingassistant":
		return RolePersonaCodingAssistant, true
	case "devops", "devopsrunner":
		return RolePersonaDevOpsRunner, true
	case "viewer", "viewerassistant":
		return RolePersonaViewerAssistant, true
	default:
		return Role{}, false
	}
}
