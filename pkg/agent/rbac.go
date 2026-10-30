package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/krewire/auth/pkg/audit"
	"github.com/krewire/auth/pkg/rbac"
)

// ErrPermissionDenied is returned when a caller or agent lacks necessary RBAC permissions.
var ErrPermissionDenied = errors.New("agent: permission denied")

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
func CheckToolPermission(eval *rbac.Evaluator, roles []string, toolName string) error {
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
	Evaluator *rbac.Evaluator
	Roles     []string
	TenantID  string
	UserID    string
	Audit     audit.AuditLogger
	OnDenial  func(toolName string, err error)
}

// AgentApprover returns an Approver callback using the provided RBAC options.
// Denied tool invocations are logged to the audit ledger.
func AgentApprover(opts RBACApproverOptions) Approver {
	return func(toolName string, _ json.RawMessage) bool {
		if opts.Evaluator == nil {
			return true
		}
		err := CheckToolPermission(opts.Evaluator, opts.Roles, toolName)
		if err != nil {
			if opts.Audit != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_ = opts.Audit.Record(ctx, &audit.Event{
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
func NewAgentApprover(eval *rbac.Evaluator, roles []string, tenantID, userID string, auditor audit.AuditLogger) Approver {
	return AgentApprover(RBACApproverOptions{
		Evaluator: eval,
		Roles:     roles,
		TenantID:  tenantID,
		UserID:    userID,
		Audit:     auditor,
	})
}

// DefaultPersonaRole returns the canonical RBAC role definition for a given persona name.
func DefaultPersonaRole(personaName string) (rbac.Role, bool) {
	switch strings.ToLower(strings.TrimSpace(personaName)) {
	case "coding", "codingassistant":
		return rbac.RolePersonaCodingAssistant, true
	case "devops", "devopsrunner":
		return rbac.RolePersonaDevOpsRunner, true
	case "viewer", "viewerassistant":
		return rbac.RolePersonaViewerAssistant, true
	default:
		return rbac.Role{}, false
	}
}
