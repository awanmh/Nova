package llm_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/llm/ollama"
	"github.com/awanmh/Nova/internal/llm/openai"
)

func TestRegistry_DiscoverAndSelect(t *testing.T) {
	tsOllama := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"models":[{"name":"qwen3-coder:latest"}]}`))
	}))
	defer tsOllama.Close()

	tsOpenAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"id":"o1-mini"}]}`))
	}))
	defer tsOpenAI.Close()

	registry := llm.NewRegistry()
	_ = registry.Register(ollama.NewProvider(tsOllama.URL))
	_ = registry.Register(openai.NewProvider(tsOpenAI.URL, "key"))

	models, err := registry.DiscoverModels(context.Background())
	if err != nil {
		t.Fatalf("expected DiscoverModels to pass, got: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 models discovered across providers, got %d", len(models))
	}

	// Test SelectModel
	p, m, err := registry.SelectModel(context.Background(), "qwen3-coder:latest")
	if err != nil {
		t.Fatalf("expected to select qwen3-coder:latest, got: %v", err)
	}
	if p.Name() != "ollama" {
		t.Fatalf("expected ollama provider, got '%s'", p.Name())
	}
	if m.Name != "qwen3-coder:latest" {
		t.Fatalf("expected qwen3-coder:latest, got '%s'", m.Name)
	}

	// Test SelectByCapability (reasoning)
	pReason, mReason, err := registry.SelectByCapability(context.Background(), llm.Capability{Reasoning: true})
	if err != nil {
		t.Fatalf("expected to find reasoning model, got: %v", err)
	}
	if mReason.Name != "o1-mini" || pReason.Name() != "openai" {
		t.Fatalf("expected o1-mini from openai, got '%s' from '%s'", mReason.Name, pReason.Name())
	}

	// Test FallbackModel
	_, mFallback, err := registry.FallbackModel(context.Background(), "qwen3-coder:latest")
	if err != nil {
		t.Fatalf("expected fallback model, got: %v", err)
	}
	if mFallback.Name != "o1-mini" {
		t.Fatalf("expected fallback to 'o1-mini', got '%s'", mFallback.Name)
	}
}
