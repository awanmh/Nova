package memory_test

import (
	"context"
	"strings"
	"testing"

	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/memory"
)

func TestFileStore_SessionAndMessages(t *testing.T) {
	tempDir := t.TempDir()
	store, err := memory.NewFileStore(tempDir)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}

	ctx := context.Background()

	// 1. Create Session
	session, err := store.CreateSession(ctx, "sess-1", "Test Conversation", "tdd")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if session.ID != "sess-1" || session.Persona != "tdd" {
		t.Fatalf("unexpected session data: %+v", session)
	}

	// 2. Append Messages
	err = store.AppendMessage(ctx, "sess-1", llm.Message{Role: "user", Content: "Hello NOVA"})
	if err != nil {
		t.Fatalf("AppendMessage 1 failed: %v", err)
	}
	err = store.AppendMessage(ctx, "sess-1", llm.Message{Role: "assistant", Content: "Hello! Ready for TDD."})
	if err != nil {
		t.Fatalf("AppendMessage 2 failed: %v", err)
	}

	// 3. Get History
	hist, err := store.GetHistory(ctx, "sess-1")
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}
	if len(hist) != 2 {
		t.Fatalf("expected 2 messages in history, got %d", len(hist))
	}

	// 4. Export Markdown
	md, err := memory.ExportMarkdown(ctx, store, "sess-1")
	if err != nil {
		t.Fatalf("ExportMarkdown failed: %v", err)
	}
	if !strings.Contains(md, "Test Conversation") || !strings.Contains(md, "Ready for TDD.") {
		t.Fatalf("unexpected markdown export: %s", md)
	}

	// 5. Clear History
	if err := store.ClearHistory(ctx, "sess-1"); err != nil {
		t.Fatalf("ClearHistory failed: %v", err)
	}
	hist2, _ := store.GetHistory(ctx, "sess-1")
	if len(hist2) != 0 {
		t.Fatalf("expected 0 messages after ClearHistory, got %d", len(hist2))
	}
}

func TestFileStore_MemoryItems(t *testing.T) {
	tempDir := t.TempDir()
	store, _ := memory.NewFileStore(tempDir)
	ctx := context.Background()

	err := store.Store(ctx, memory.ScopeProject, "db_port", "5432")
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	it, err := store.Retrieve(ctx, memory.ScopeProject, "db_port")
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	if it.Value != "5432" {
		t.Fatalf("expected '5432', got '%s'", it.Value)
	}

	list, _ := store.List(ctx, memory.ScopeProject)
	if len(list) != 1 {
		t.Fatalf("expected 1 project memory item, got %d", len(list))
	}
}
