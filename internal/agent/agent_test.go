package agent_test

import (
	"context"
	"strings"
	"testing"

	"github.com/awanmh/Nova/internal/agent"
	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/memory"
	"github.com/awanmh/Nova/internal/permission"
	"github.com/awanmh/Nova/internal/tools"
)

type mockLLMProvider struct {
	calls int
}

func (m *mockLLMProvider) Name() string { return "mock" }
func (m *mockLLMProvider) Health(ctx context.Context) error { return nil }
func (m *mockLLMProvider) ListModels(ctx context.Context) ([]llm.Model, error) {
	return []llm.Model{{Name: "mock-model", Provider: "mock", Status: "READY"}}, nil
}
func (m *mockLLMProvider) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	m.calls++
	if m.calls == 1 {
		// First call: request tool call read_file
		return &llm.ChatResponse{
			Message: llm.Message{
				Role: "assistant",
				ToolCalls: []llm.ToolCall{
					{
						ID:        "call-1",
						Name:      "read_file",
						Arguments: `{"path":"hello.txt"}`,
					},
				},
			},
			FinishReason: "tool_calls",
		}, nil
	}
	// Second call: finish with text answer
	return &llm.ChatResponse{
		Message: llm.Message{
			Role:    "assistant",
			Content: "File read successfully. Task complete.",
		},
		FinishReason: "stop",
	}, nil
}
func (m *mockLLMProvider) ChatStream(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	return nil, nil
}

func TestGuard_IterationAndInfiniteLoop(t *testing.T) {
	guard := agent.NewGuard(3)

	// Test iteration cap
	if err := guard.CheckIteration(); err != nil {
		t.Fatalf("step 1 failed: %v", err)
	}
	_ = guard.CheckIteration()
	_ = guard.CheckIteration()
	if err := guard.CheckIteration(); err == nil {
		t.Fatalf("expected step 4 to exceed max iterations (3)")
	}

	// Test infinite loop detector
	guard2 := agent.NewGuard(10)
	_ = guard2.CheckInfiniteLoop("read_file", `{"path":"foo.go"}`)
	_ = guard2.CheckInfiniteLoop("read_file", `{"path":"foo.go"}`)
	err := guard2.CheckInfiniteLoop("read_file", `{"path":"foo.go"}`)
	if err == nil || !strings.Contains(err.Error(), "infinite loop detected") {
		t.Fatalf("expected infinite loop error on 3rd identical call, got: %v", err)
	}
}

func TestRunner_AutonomousLoopWithToolExecution(t *testing.T) {
	tempDir := t.TempDir()
	reg := tools.NewRegistry()
	_ = tools.RegisterStandardTools(tempDir, reg)

	perm := permission.NewEngine(permission.PolicyAllow, nil)
	auditLog, _ := permission.NewFileAuditLogger(tempDir)
	executor := tools.NewExecutor(tempDir, reg, perm, auditLog)

	store, err := memory.NewFileStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create file store: %v", err)
	}

	provider := &mockLLMProvider{}
	runner := agent.NewRunner(
		provider,
		"mock-model",
		executor,
		store,
		"sess-test",
		"general",
		"You are NOVA.",
		5,
	)

	ctx := context.Background()
	err = runner.Run(ctx, "Read hello.txt")
	if err != nil {
		t.Fatalf("runner Run failed: %v", err)
	}

	if runner.State() != agent.StateCompleted {
		t.Fatalf("expected state COMPLETED, got %s", runner.State())
	}
	if provider.calls != 2 {
		t.Fatalf("expected 2 LLM turns, got %d", provider.calls)
	}

	// Verify history stored in memory
	history, err := store.GetHistory(ctx, "sess-test")
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}
	// Expect messages: User prompt, Assistant tool call, Tool result, Assistant final answer
	if len(history) != 4 {
		t.Fatalf("expected 4 messages in session history, got %d", len(history))
	}
}
