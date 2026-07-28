package permission_test

import (
	"context"
	"testing"

	"github.com/awanmh/Nova/internal/permission"
	"github.com/awanmh/Nova/internal/tools"
)

func TestPermissionEngine_ReadOnly(t *testing.T) {
	engine := permission.NewEngine(permission.PolicyAsk, nil)
	req := &permission.Request{
		ToolName:  "fs.read",
		RiskLevel: tools.RiskReadOnly,
	}

	allowed, err := engine.Check(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error for read_only tool: %v", err)
	}
	if !allowed {
		t.Fatalf("expected read_only tool to be allowed automatically")
	}
}

func TestPermissionEngine_Deny(t *testing.T) {
	engine := permission.NewEngine(permission.PolicyDeny, nil)
	req := &permission.Request{
		ToolName:  "shell.exec",
		RiskLevel: tools.RiskHighImpact,
	}

	allowed, err := engine.Check(context.Background(), req)
	if err == nil {
		t.Fatalf("expected deny policy error, got nil")
	}
	if allowed {
		t.Fatalf("expected tool to be denied")
	}
}

func TestPermissionEngine_SessionApproval(t *testing.T) {
	engine := permission.NewEngine(permission.PolicyAsk, nil)
	engine.ApproveSession("fs.write")

	req := &permission.Request{
		ToolName:  "fs.write",
		RiskLevel: tools.RiskLowModify,
	}

	allowed, err := engine.Check(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error after session approval: %v", err)
	}
	if !allowed {
		t.Fatalf("expected tool to be allowed via session approval")
	}
}
