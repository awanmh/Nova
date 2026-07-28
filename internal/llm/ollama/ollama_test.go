package ollama_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/llm/ollama"
)

func TestOllamaProvider_HealthAndListModels(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"version":"0.1.0"}`))
		case "/api/tags":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"models":[{"name":"qwen3-coder:latest"},{"name":"deepseek-r1:latest"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	provider := ollama.NewProvider(ts.URL)
	if err := provider.Health(context.Background()); err != nil {
		t.Fatalf("expected Health to pass, got: %v", err)
	}

	models, err := provider.ListModels(context.Background())
	if err != nil {
		t.Fatalf("expected ListModels to pass, got: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}

	// Verify deepseek-r1 capability inference
	var r1 *llm.Model
	for i := range models {
		if models[i].Name == "deepseek-r1:latest" {
			r1 = &models[i]
		}
	}
	if r1 == nil || !r1.Capability.Reasoning {
		t.Fatalf("expected deepseek-r1 to have reasoning capability")
	}
}

func TestOllamaProvider_Chat(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"hello"},"done":true,"done_reason":"stop","prompt_eval_count":10,"eval_count":5}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	provider := ollama.NewProvider(ts.URL)
	resp, err := provider.Chat(context.Background(), &llm.ChatRequest{
		Model: "qwen3-coder",
		Messages: []llm.Message{
			{Role: "user", Content: "hi"},
		},
	})
	if err != nil {
		t.Fatalf("expected Chat to pass, got: %v", err)
	}
	if resp.Message.Content != "hello" {
		t.Fatalf("expected 'hello', got '%s'", resp.Message.Content)
	}
	if resp.CompletionTokens != 5 {
		t.Fatalf("expected 5 completion tokens, got %d", resp.CompletionTokens)
	}
}

func TestOllamaProvider_ChatStream(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/chat" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{\"message\":{\"role\":\"assistant\",\"content\":\"hello \"},\"done\":false}\n{\"message\":{\"role\":\"assistant\",\"content\":\"world\"},\"done\":true,\"done_reason\":\"stop\"}\n"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	provider := ollama.NewProvider(ts.URL)
	ch, err := provider.ChatStream(context.Background(), &llm.ChatRequest{
		Model: "qwen3-coder",
		Messages: []llm.Message{
			{Role: "user", Content: "hi"},
		},
		Stream: true,
	})
	if err != nil {
		t.Fatalf("expected ChatStream to pass, got: %v", err)
	}

	var content string
	for chunk := range ch {
		if chunk.Error != nil {
			t.Fatalf("unexpected stream chunk error: %v", chunk.Error)
		}
		content += chunk.Delta.Content
	}

	if content != "hello world" {
		t.Fatalf("expected 'hello world', got '%s'", content)
	}
}
