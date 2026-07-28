package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/permission"
	"github.com/awanmh/Nova/internal/security"
)

// Executor manages tool invocation requests from the LLM, enforcing safety policies,
// logging audit trails, and redacting secret output data.
type Executor struct {
	rootDir  string
	registry *Registry
	perm     permission.Engine
	audit    permission.AuditLogger
}

// NewExecutor creates a new safe tool executor.
func NewExecutor(rootDir string, registry *Registry, perm permission.Engine, audit permission.AuditLogger) *Executor {
	return &Executor{
		rootDir:  rootDir,
		registry: registry,
		perm:     perm,
		audit:    audit,
	}
}

// ExecuteTool safely runs a tool requested by the LLM and returns an LLM-compatible response.
func (e *Executor) ExecuteTool(ctx context.Context, call llm.ToolCall) llm.ToolResponse {
	start := time.Now()

	tool, found := e.registry.Get(call.Name)
	if !found {
		return llm.ToolResponse{
			ID:      call.ID,
			Name:    call.Name,
			Content: "",
			Error:   fmt.Sprintf("tool '%s' is not registered", call.Name),
		}
	}

	risk := permission.ClassifyRisk(call.Name, call.Arguments)

	// Permission check
	allowed, err := e.perm.Check(ctx, &permission.Request{
		ToolName:  call.Name,
		RiskLevel: risk,
		Arguments: call.Arguments,
	})

	if !allowed || err != nil {
		denyReason := "access denied by security policy"
		if err != nil {
			denyReason = err.Error()
		}
		e.logAudit(ctx, call.Name, risk, call.Arguments, false, "DENIED", time.Since(start).Milliseconds())
		return llm.ToolResponse{
			ID:      call.ID,
			Name:    call.Name,
			Content: "",
			Error:   fmt.Sprintf("Permission Denied: %s", denyReason),
		}
	}

	// Execute tool action
	out, execErr := tool.Execute(ctx, call.Arguments)

	// Redact secrets from output
	redactedOut := security.Redact(out)

	status := "SUCCESS"
	var errStr string
	if execErr != nil {
		status = "ERROR"
		errStr = security.Redact(execErr.Error())
	}

	e.logAudit(ctx, call.Name, risk, call.Arguments, true, status, time.Since(start).Milliseconds())

	return llm.ToolResponse{
		ID:      call.ID,
		Name:    call.Name,
		Content: redactedOut,
		Error:   errStr,
	}
}

func (e *Executor) logAudit(ctx context.Context, name string, risk RiskLevel, args string, allowed bool, status string, durMs int64) {
	if e.audit == nil {
		return
	}
	_ = e.audit.Log(ctx, permission.AuditRecord{
		Timestamp:  time.Now().UTC(),
		ToolName:   name,
		RiskLevel:  risk,
		Arguments:  args,
		Allowed:    allowed,
		Status:     status,
		DurationMs: durMs,
	})
}
