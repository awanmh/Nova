package context

import (
	"fmt"
	"strings"

	"github.com/awanmh/Nova/internal/persona"
)

// SystemPromptBuilder constructs the master system message for the LLM prompt.
type SystemPromptBuilder struct {
	baseRule       string
	persona        persona.Persona
	projectSummary string
	bundle         *Bundle
}

// NewSystemPromptBuilder creates a new SystemPromptBuilder with default base rules.
func NewSystemPromptBuilder() *SystemPromptBuilder {
	return &SystemPromptBuilder{
		baseRule: "You are NOVA, a Local-First AI Software Engineering Agent. You understand, modify, execute, and verify code safely on local repositories.",
	}
}

// WithPersona sets the active engineering persona.
func (b *SystemPromptBuilder) WithPersona(p persona.Persona) *SystemPromptBuilder {
	b.persona = p
	return b
}

// WithProjectSummary sets the natural language project architecture overview.
func (b *SystemPromptBuilder) WithProjectSummary(summary string) *SystemPromptBuilder {
	b.projectSummary = summary
	return b
}

// WithBundle sets the token-budgeted context bundle.
func (b *SystemPromptBuilder) WithBundle(bundle *Bundle) *SystemPromptBuilder {
	b.bundle = bundle
	return b
}

// Build compiles all components into a cohesive system prompt string.
func (b *SystemPromptBuilder) Build() string {
	var sb strings.Builder

	// 1. Base rule & persona
	sb.WriteString(b.baseRule)
	sb.WriteString("\n\n")

	if b.persona.SystemRule != "" {
		sb.WriteString(fmt.Sprintf("Persona (%s): %s\n\n", b.persona.Name, b.persona.SystemRule))
	}

	// 2. Project architecture summary
	if b.projectSummary != "" {
		sb.WriteString("Project Architecture Overview:\n")
		sb.WriteString(b.projectSummary)
		sb.WriteString("\n\n")
	}

	// 3. Formatted context items
	if b.bundle != nil && len(b.bundle.Items) > 0 {
		sb.WriteString("Retrieved Code & Context:\n")
		for _, item := range b.bundle.Items {
			sb.WriteString(fmt.Sprintf("--- [%s] %s (score: %.2f) ---\n", strings.ToUpper(string(item.Type)), item.Path, item.Score))
			sb.WriteString(item.Content)
			sb.WriteString("\n\n")
		}
	}

	return strings.TrimSpace(sb.String())
}
