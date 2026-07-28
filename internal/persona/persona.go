package persona

import (
	"fmt"
	"sort"
	"sync"
)

// Persona represents an engineering specialty or behavioral mode for NOVA.
type Persona struct {
	Name                string   `json:"name" yaml:"name"`
	Description         string   `json:"description" yaml:"description"`
	SystemRule          string   `json:"system_rule" yaml:"system_rule"`
	DefaultTools        []string `json:"default_tools,omitempty" yaml:"default_tools,omitempty"`
	TokenBudgetModifier float64  `json:"token_budget_modifier,omitempty" yaml:"token_budget_modifier,omitempty"`
}

// Builtin returns the default set of engineering personas for NOVA.
func Builtin() map[string]Persona {
	return map[string]Persona{
		"general": {
			Name:        "general",
			Description: "General Software Engineering Assistant",
			SystemRule:  "You are a versatile, pragmatic Software Engineer. Focus on writing clean, maintainable, and correct code that solves the user's task effectively.",
			DefaultTools: []string{
				"read_file", "write_file", "list_dir", "search_files",
				"run_command", "git_status", "git_diff", "apply_patch",
			},
			TokenBudgetModifier: 1.0,
		},
		"architect": {
			Name:        "architect",
			Description: "System Architecture & Design Specialist",
			SystemRule:  "You are a Senior System Architect. Prioritize high-level component boundaries, decoupling, scalability, and clean modular architecture. Always analyze tradeoffs before making implementation changes.",
			DefaultTools: []string{
				"read_file", "list_dir", "search_files", "git_status", "git_diff",
			},
			TokenBudgetModifier: 1.2,
		},
		"security": {
			Name:        "security",
			Description: "Application Security & Hardening Specialist",
			SystemRule:  "You are an Application Security Specialist. Prioritize defensive coding, vulnerability screening, strict input sanitization, and secret protection. Never expose or print private keys or credentials.",
			DefaultTools: []string{
				"read_file", "list_dir", "search_files", "git_status", "git_diff",
			},
			TokenBudgetModifier: 1.0,
		},
		"tdd": {
			Name:        "tdd",
			Description: "Test-Driven Development (TDD) Specialist",
			SystemRule:  "You are a strict Test-Driven Development (TDD) Practitioner. Follow the Red-Green-Refactor cycle: always write failing unit tests before implementing production code, then make them pass with minimal changes.",
			DefaultTools: []string{
				"read_file", "write_file", "list_dir", "search_files",
				"run_command", "git_status", "git_diff", "apply_patch",
			},
			TokenBudgetModifier: 1.0,
		},
	}
}

// BuiltinPersonas is a backwards-compatible alias for Builtin.
func BuiltinPersonas() map[string]Persona {
	return Builtin()
}

// Manager manages built-in and custom personas.
type Manager struct {
	mu       sync.RWMutex
	personas map[string]Persona
}

// NewManager creates a new Persona Manager initialized with builtin personas.
func NewManager() *Manager {
	m := &Manager{
		personas: make(map[string]Persona),
	}
	for k, v := range Builtin() {
		m.personas[k] = v
	}
	return m
}

// Get retrieves a persona by name. Defaults to "general" if not found.
func (m *Manager) Get(name string) Persona {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if p, ok := m.personas[name]; ok {
		return p
	}
	// default fallback
	return m.personas["general"]
}

// Register adds or overrides a persona.
func (m *Manager) Register(p Persona) error {
	if p.Name == "" {
		return fmt.Errorf("persona name cannot be empty")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.personas[p.Name] = p
	return nil
}

// List returns a sorted slice of all registered personas.
func (m *Manager) List() []Persona {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []Persona
	for _, p := range m.personas {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})
	return list
}
