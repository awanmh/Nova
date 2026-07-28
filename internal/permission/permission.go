package permission

import (
	"context"
	"fmt"
	"sync"
)

// RiskLevel defines the safety risk associated with a tool execution.
type RiskLevel string

const (
	RiskReadOnly   RiskLevel = "read_only"
	RiskLowModify  RiskLevel = "low_modify"
	RiskHighImpact RiskLevel = "high_impact"
)

// Policy defines action policy: "allow", "ask", or "deny".
type Policy string

const (
	PolicyAllow Policy = "allow"
	PolicyAsk   Policy = "ask"
	PolicyDeny  Policy = "deny"
)

// Request represents a permission request for a tool call.
type Request struct {
	ToolName  string
	RiskLevel RiskLevel
	Arguments string
}

// ApprovalHandler is a callback for UI or TUI to prompt the user for permission.
type ApprovalHandler func(ctx context.Context, req *Request) (bool, error)

// Engine evaluates safety policies and prompts the user when needed.
type Engine interface {
	Check(ctx context.Context, req *Request) (bool, error)
	SetPolicy(toolName string, p Policy)
	ApproveSession(toolName string)
}

type engine struct {
	mu             sync.RWMutex
	defaultPolicy  Policy
	policies       map[string]Policy // custom policy per tool
	sessionAllowed map[string]bool   // approved for the current session
	handler        ApprovalHandler
}

// NewEngine creates a new permission engine with a default policy and approval handler.
func NewEngine(defaultPolicy Policy, handler ApprovalHandler) Engine {
	return &engine{
		defaultPolicy:  defaultPolicy,
		policies:       make(map[string]Policy),
		sessionAllowed: make(map[string]bool),
		handler:        handler,
	}
}

func (e *engine) Check(ctx context.Context, req *Request) (bool, error) {
	e.mu.RLock()
	// Read-only tools are allowed automatically unless explicitly denied
	if req.RiskLevel == RiskReadOnly {
		if p, ok := e.policies[req.ToolName]; ok && p == PolicyDeny {
			e.mu.RUnlock()
			return false, fmt.Errorf("tool '%s' is explicitly denied by policy", req.ToolName)
		}
		e.mu.RUnlock()
		return true, nil
	}

	// Check if already approved for this session
	if e.sessionAllowed[req.ToolName] {
		e.mu.RUnlock()
		return true, nil
	}

	p, exists := e.policies[req.ToolName]
	if !exists {
		p = e.defaultPolicy
	}
	e.mu.RUnlock()

	switch p {
	case PolicyAllow:
		return true, nil
	case PolicyDeny:
		return false, fmt.Errorf("tool '%s' is denied by policy", req.ToolName)
	case PolicyAsk:
		if e.handler == nil {
			return false, fmt.Errorf("permission required for '%s' but no interactive approval handler is available", req.ToolName)
		}
		return e.handler(ctx, req)
	default:
		return false, fmt.Errorf("unknown policy '%s'", p)
	}
}

func (e *engine) SetPolicy(toolName string, p Policy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.policies[toolName] = p
}

func (e *engine) ApproveSession(toolName string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sessionAllowed[toolName] = true
}
