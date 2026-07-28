package tools

import (
	"context"
	"fmt"
	"sync"

	"github.com/awanmh/Nova/internal/permission"
)

// RiskLevel defines the safety risk associated with a tool execution.
type RiskLevel = permission.RiskLevel

const (
	RiskReadOnly   = permission.RiskReadOnly
	RiskLowModify  = permission.RiskLowModify
	RiskHighImpact = permission.RiskHighImpact
)

// Schema represents JSON Schema for tool input arguments.
type Schema struct {
	Type       string              `json:"type"` // usually "object"
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// Tool defines the interface that every NOVA action tool must implement.
type Tool interface {
	Name() string
	Description() string
	Schema() Schema
	RiskLevel() RiskLevel
	Execute(ctx context.Context, argsJSON string) (string, error)
}

// Registry manages the set of available tools.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (r *Registry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool '%s' is already registered", name)
	}
	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// List returns all registered tools.
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		res = append(res, t)
	}
	return res
}
