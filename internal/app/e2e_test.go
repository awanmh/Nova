package app_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awanmh/Nova/internal/agent"
	"github.com/awanmh/Nova/internal/config"
	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/memory"
	"github.com/awanmh/Nova/internal/permission"
	"github.com/awanmh/Nova/internal/persona"
	"github.com/awanmh/Nova/internal/project"
	"github.com/awanmh/Nova/internal/security"
	"github.com/awanmh/Nova/internal/tools"
	"github.com/awanmh/Nova/internal/tui"
)

type mockE2ELLMProvider struct {
	calls int
}

func (m *mockE2ELLMProvider) Name() string { return "mock-e2e" }
func (m *mockE2ELLMProvider) Health(ctx context.Context) error { return nil }
func (m *mockE2ELLMProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
	return []llm.Model{{Name: "mock-model", Provider: "mock-e2e", Status: "READY"}}, nil
}
func (m *mockE2ELLMProvider) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	m.calls++
	if m.calls == 1 {
		// First turn: model requests write_file tool call
		return &llm.ChatResponse{
			Message: llm.Message{
				Role: "assistant",
				ToolCalls: []llm.ToolCall{
					{
						ID:        "call-write-1",
						Name:      "write_file",
						Arguments: `{"path":"main.go","content":"package main\n\nfunc main() {}\n"}`,
					},
				},
			},
			FinishReason: "tool_calls",
		}, nil
	}
	// Second turn: model confirms file creation
	return &llm.ChatResponse{
		Message: llm.Message{
			Role:    "assistant",
			Content: "Created main.go successfully via TDD persona.",
		},
		FinishReason: "stop",
	}, nil
}
func (m *mockE2ELLMProvider) ChatStream(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	return nil, nil
}

// TestNOVA_EndToEndStack verifies the full 13-module NOVA architecture working together.
func TestNOVA_EndToEndStack(t *testing.T) {
	tempDir := t.TempDir()
	ctx := context.Background()

	// 1. Config System
	cfg := config.DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("config validation failed: %v", err)
	}

	// 2. Security & Redactor
	secret := "sk-test123456789012345678901234567890"
	redacted := security.Redact(secret)
	if strings.Contains(redacted, secret) {
		t.Fatalf("redactor failed to scrub secret")
	}

	// 3. Project Intelligence
	scanner := project.NewWorkspaceScanner()
	info, err := scanner.Scan(ctx, tempDir)
	if err != nil {
		t.Fatalf("project scan failed: %v", err)
	}

	// 4. Tools & Permission Engine
	reg := tools.NewRegistry()
	if err := tools.RegisterStandardTools(tempDir, reg); err != nil {
		t.Fatalf("tool registration failed: %v", err)
	}
	perm := permission.NewEngine(permission.PolicyAllow, nil)
	auditLog, err := permission.NewFileAuditLogger(tempDir)
	if err != nil {
		t.Fatalf("audit logger init failed: %v", err)
	}
	executor := tools.NewExecutor(tempDir, reg, perm, auditLog)

	// 5. Memory & Session Store
	store, err := memory.NewFileStore(tempDir)
	if err != nil {
		t.Fatalf("memory store init failed: %v", err)
	}

	// 6. Persona System
	pMgr := persona.NewManager()
	tddPersona := pMgr.Get("tdd")

	// 7. Agent Execution Loop
	provider := &mockE2ELLMProvider{}
	runner := agent.NewRunner(
		provider,
		cfg.DefaultModel,
		executor,
		store,
		"sess-e2e",
		tddPersona.Name,
		tddPersona.SystemRule,
		10,
	)

	if err := runner.Run(ctx, "Create main.go"); err != nil {
		t.Fatalf("agent run failed: %v", err)
	}
	if runner.State() != agent.StateCompleted {
		t.Fatalf("expected agent state COMPLETED, got %s", runner.State())
	}

	// 8. Verify tool execution result on disk
	mainPath := filepath.Join(tempDir, "main.go")
	data, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("expected main.go created on disk: %v", err)
	}
	if !strings.Contains(string(data), "func main()") {
		t.Fatalf("unexpected main.go content: %s", string(data))
	}

	// 9. Verify session transcript in Memory Store
	hist, err := store.GetHistory(ctx, "sess-e2e")
	if err != nil || len(hist) != 4 {
		t.Fatalf("expected 4 messages in session history, got %d (err=%v)", len(hist), err)
	}
	md, err := memory.ExportMarkdown(ctx, store, "sess-e2e")
	if err != nil || !strings.Contains(md, "Created main.go successfully via TDD persona.") {
		t.Fatalf("unexpected markdown export: %s", md)
	}

	// 10. Verify TUI model initialization and rendering
	tuiModel := tui.NewModel("llama3", tempDir, "tdd")
	if tuiModel.Persona != "tdd" || tuiModel.Workspace != tempDir {
		t.Fatalf("unexpected TUI model state: %+v", tuiModel)
	}
	view := tuiModel.View()
	if !strings.Contains(view, "NOVA") || !strings.Contains(view, "tdd") {
		t.Fatalf("unexpected TUI view output: %s", view)
	}

	t.Logf("✅ Complete E2E integration test passed across all 13 modules (workspace: %s, info: %+v)", tempDir, info.Languages)
}
