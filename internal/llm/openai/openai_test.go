package openai_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/llm/openai"
)

func TestOpenAIProvider_HealthAndListModels(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/models" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"},{"id":"o1-mini"}]}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	provider := openai.NewProvider(ts.URL, "test-key")
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

	// Verify o1-mini reasoning capability
	var o1 *llm.Model
	for i := range models {
		if models[i].Name == "o1-mini" {
			o1 = &models[i]
		}
	}
	if o1 == nil || !o1.Capability.Reasoning {
		t.Fatalf("expected o1-mini to have reasoning capability")
	}
}

func TestOpenAIProvider_Chat(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chat/completions" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"openai response"},"finish_reason":"stop"}],"usage":{"prompt_tokens":12,"completion_tokens":4}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	provider := openai.NewProvider(ts.URL, "key")
	resp, err := provider.Chat(context.Background(), &llm.ChatRequest{
		Model: "gpt-4o",
		Messages: []llm.Message{
			{Role: "user", Content: "hello"},
		},
	})
	if err != nil {
		t.Fatalf("expected Chat to pass, got: %v", err)
	}
	if resp.Message.Content != "openai response" {
		t.Fatalf("expected 'openai response', got '%s'", resp.Message.Content)
	}
}

func TestOpenAIProvider_ChatStream(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chat/completions" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hello \"}}]}\n\ndata: {\"choices\":[{\"delta\":{\"content\":\"stream\"},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	provider := openai.NewProvider(ts.URL, "key")
	ch, err := provider.ChatStream(context.Background(), &llm.ChatRequest{
		Model: "gpt-4o",
		Messages: []llm.Message{
			{Role: "user", Content: "hello"},
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

	if content != "hello stream" {
		t.Fatalf("expected 'hello stream', got '%s'", content)
	}
}
