package context_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awanmh/Nova/internal/context"
	"github.com/awanmh/Nova/internal/persona"
)

func TestEstimateTokens(t *testing.T) {
	if context.EstimateTokens("") != 0 {
		t.Fatalf("expected 0 for empty string")
	}
	if context.EstimateTokens("a") != 1 {
		t.Fatalf("expected 1 for small string")
	}
	tokens := context.EstimateTokens("1234567890123456") // 16 chars -> 4 tokens
	if tokens != 4 {
		t.Fatalf("expected 4 tokens, got %d", tokens)
	}
}

func TestLoadFile(t *testing.T) {
	tempDir := t.TempDir()

	validFile := filepath.Join(tempDir, "main.go")
	_ = os.WriteFile(validFile, []byte("package main\nfunc main() {}\n"), 0644)

	item, err := context.LoadFile(tempDir, "main.go")
	if err != nil {
		t.Fatalf("expected LoadFile to pass, got: %v", err)
	}
	if item.Type != context.ItemTypeFile || item.Path != "main.go" {
		t.Fatalf("unexpected item fields: %+v", item)
	}

	// Test sensitive file rejection
	envFile := filepath.Join(tempDir, ".env")
	_ = os.WriteFile(envFile, []byte("SECRET=123"), 0644)

	_, err = context.LoadFile(tempDir, ".env")
	if err != context.ErrSensitiveFile {
		t.Fatalf("expected ErrSensitiveFile, got: %v", err)
	}
}

func TestExtractGoSymbols(t *testing.T) {
	code := `package sample

type User struct {
	Name string
}

func GetUser(id string) (*User, error) {
	return nil, nil
}
`
	symbols := context.ExtractGoSymbols("sample.go", code)
	if len(symbols) < 2 {
		t.Fatalf("expected at least 2 symbols (struct and func), got %d", len(symbols))
	}
}

func TestRankAndCompress(t *testing.T) {
	items := []context.Item{
		{Type: context.ItemTypeFile, Path: "utils/helper.go", Content: "package utils\n// helper func", Score: 1.0, TokenCnt: 100},
		{Type: context.ItemTypeFile, Path: "auth/login.go", Content: "package auth\nfunc Login()", Score: 1.0, TokenCnt: 100},
	}

	ranked := context.Rank(items, "login")
	if ranked[0].Path != "auth/login.go" {
		t.Fatalf("expected auth/login.go to rank highest for query 'login'")
	}

	// Compress to maxTokens = 150 (only 1 full item should fit)
	bundle := context.Compress(ranked, 150)
	if len(bundle.Items) == 0 {
		t.Fatalf("expected at least 1 item in compressed bundle")
	}
	if bundle.TotalTokens > 150 {
		t.Fatalf("expected total tokens <= 150, got %d", bundle.TotalTokens)
	}
}

func TestSystemPromptBuilder(t *testing.T) {
	builder := context.NewSystemPromptBuilder().
		WithPersona(persona.BuiltinPersonas()["security"]).
		WithProjectSummary("Go CLI repository").
		WithBundle(&context.Bundle{
			Items: []context.Item{
				{Type: context.ItemTypeFile, Path: "main.go", Content: "package main", Score: 1.0},
			},
		})

	prompt := builder.Build()
	if !strings.Contains(prompt, "Persona (security):") {
		t.Fatalf("expected prompt to contain security persona")
	}
	if !strings.Contains(prompt, "Project Architecture Overview:") {
		t.Fatalf("expected prompt to contain project summary")
	}
	if !strings.Contains(prompt, "--- [FILE] main.go (score: 1.00) ---") {
		t.Fatalf("expected prompt to contain formatted file block")
	}
}
