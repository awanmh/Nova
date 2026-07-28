package execution

import (
	"context"

	"github.com/awanmh/Nova/internal/llm"
)

// Result holds the normalized output of a tool execution.
type Result struct {
	ToolID   string `json:"tool_id"`
	ToolName string `json:"tool_name"`
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
}

// Orchestrator executes tool calls requested by the model, subject to safety policy checks.
type Orchestrator interface {
	Execute(ctx context.Context, calls []llm.ToolCall) ([]Result, error)
	ExecuteSingle(ctx context.Context, call llm.ToolCall) (*Result, error)
}
