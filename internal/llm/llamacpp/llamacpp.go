package llamacpp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/llm/openai"
)

// Provider implements llm.Provider for a local llama.cpp HTTP server (usually http://localhost:8080).
type Provider struct {
	endpoint string
	client   *http.Client
	delegate *openai.Provider
}

// NewProvider creates a new llama.cpp provider targeting the given endpoint.
func NewProvider(endpoint string) *Provider {
	if endpoint == "" {
		endpoint = "http://localhost:8080"
	}
	endpoint = strings.TrimRight(endpoint, "/")
	// llama.cpp server implements OpenAI-compatible /v1 endpoints
	delegateEndpoint := endpoint + "/v1"
	return &Provider{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		delegate: openai.NewProvider(delegateEndpoint, ""),
	}
}

func (p *Provider) Name() string {
	return "llamacpp"
}

// Health checks if llama.cpp server is reachable via /health.
func (p *Provider) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", llm.ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: unexpected status code %d", llm.ErrProviderUnavailable, resp.StatusCode)
	}
	return nil
}

// ListModels delegates to OpenAI-compatible /v1/models on llama.cpp server.
func (p *Provider) ListModels(ctx context.Context) ([]llm.Model, error) {
	models, err := p.delegate.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	// Re-stamp provider name to "llamacpp"
	for i := range models {
		models[i].Provider = p.Name()
	}
	return models, nil
}

// Chat delegates to OpenAI-compatible /v1/chat/completions on llama.cpp server.
func (p *Provider) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	return p.delegate.Chat(ctx, req)
}

// ChatStream delegates to OpenAI-compatible /v1/chat/completions SSE on llama.cpp server.
func (p *Provider) ChatStream(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	return p.delegate.ChatStream(ctx, req)
}
