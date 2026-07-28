package tools_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/permission"
	"github.com/awanmh/Nova/internal/tools"
)

func TestStandardTools_FileOperations(t *testing.T) {
	tempDir := t.TempDir()
	reg := tools.NewRegistry()
	if err := tools.RegisterStandardTools(tempDir, reg); err != nil {
		t.Fatalf("failed to register standard tools: %v", err)
	}

	if len(reg.List()) != 8 {
		t.Fatalf("expected 8 standard tools registered, got %d", len(reg.List()))
	}

	ctx := context.Background()

	// Test write_file
	writeTool, _ := reg.Get("write_file")
	_, err := writeTool.Execute(ctx, `{"path":"hello.txt","content":"hello world"}`)
	if err != nil {
		t.Fatalf("write_file failed: %v", err)
	}

	// Test read_file
	readTool, _ := reg.Get("read_file")
	out, err := readTool.Execute(ctx, `{"path":"hello.txt"}`)
	if err != nil {
		t.Fatalf("read_file failed: %v", err)
	}
	if !strings.Contains(out, "hello world") {
		t.Fatalf("read_file output mismatch: %s", out)
	}

	// Test apply_patch
	patchTool, _ := reg.Get("apply_patch")
	_, err = patchTool.Execute(ctx, `{"path":"hello.txt","target_content":"world","replacement_content":"NOVA"}`)
	if err != nil {
		t.Fatalf("apply_patch failed: %v", err)
	}

	out2, _ := readTool.Execute(ctx, `{"path":"hello.txt"}`)
	if !strings.Contains(out2, "hello NOVA") {
		t.Fatalf("expected patched content 'hello NOVA', got: %s", out2)
	}

	// Test list_dir
	listTool, _ := reg.Get("list_dir")
	listOut, err := listTool.Execute(ctx, `{"path":"."}`)
	if err != nil {
		t.Fatalf("list_dir failed: %v", err)
	}
	if !strings.Contains(listOut, "hello.txt") {
		t.Fatalf("expected list_dir to show hello.txt")
	}

	// Test search_files
	searchTool, _ := reg.Get("search_files")
	searchOut, err := searchTool.Execute(ctx, `{"query":"NOVA"}`)
	if err != nil {
		t.Fatalf("search_files failed: %v", err)
	}
	if !strings.Contains(searchOut, "hello.txt") {
		t.Fatalf("expected search_files to find NOVA in hello.txt")
	}
}

func TestExecutor_PermissionAndAudit(t *testing.T) {
	tempDir := t.TempDir()
	reg := tools.NewRegistry()
	_ = tools.RegisterStandardTools(tempDir, reg)

	auditLog, err := permission.NewFileAuditLogger(tempDir)
	if err != nil {
		t.Fatalf("failed to create audit logger: %v", err)
	}

	// Permission engine that DENIES everything except read-only
	perm := permission.NewEngine(permission.PolicyDeny, nil)
	executor := tools.NewExecutor(tempDir, reg, perm, auditLog)

	ctx := context.Background()

	// 1. Try write_file -> should be DENIED
	respDeny := executor.ExecuteTool(ctx, llm.ToolCall{
		ID:        "call-1",
		Name:      "write_file",
		Arguments: `{"path":"forbidden.txt","content":"no"}`,
	})
	if respDeny.Error == "" || !strings.Contains(respDeny.Error, "Permission Denied") {
		t.Fatalf("expected write_file to be denied, got: %+v", respDeny)
	}

	// 2. Allow session for write_file -> try again -> should SUCCEED
	perm.ApproveSession("write_file")
	respAllow := executor.ExecuteTool(ctx, llm.ToolCall{
		ID:        "call-2",
		Name:      "write_file",
		Arguments: `{"path":"allowed.txt","content":"yes"}`,
	})
	if respAllow.Error != "" {
		t.Fatalf("expected allowed write_file to succeed, got error: %s", respAllow.Error)
	}

	// Verify audit log has 2 records (1 DENIED, 1 SUCCESS)
	records, err := auditLog.ReadRecent(ctx, 10)
	if err != nil {
		t.Fatalf("failed to read audit logs: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 audit log records, got %d", len(records))
	}
	if records[0].Status != "DENIED" || records[1].Status != "SUCCESS" {
		t.Fatalf("unexpected audit statuses: %s, %s", records[0].Status, records[1].Status)
	}
}

func TestToolSecurity_SensitiveFileBlocked(t *testing.T) {
	tempDir := t.TempDir()
	reg := tools.NewRegistry()
	_ = tools.RegisterStandardTools(tempDir, reg)

	// Attempt to read/write .env file
	writeTool, _ := reg.Get("write_file")
	_, err := writeTool.Execute(context.Background(), `{"path":".env","content":"SECRET=foo"}`)
	if err == nil {
		t.Fatalf("expected writing to .env to fail with access denied")
	}

	// Create .env manually via os to test read_file blocking
	_ = os.WriteFile(filepath.Join(tempDir, ".env"), []byte("SECRET=foo"), 0644)
	readTool, _ := reg.Get("read_file")
	_, err = readTool.Execute(context.Background(), `{"path":".env"}`)
	if err == nil {
		t.Fatalf("expected reading .env to fail with access denied")
	}
}
