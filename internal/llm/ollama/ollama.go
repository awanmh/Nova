package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/awanmh/Nova/internal/llm"
)

// Provider implements llm.Provider for the Ollama local inference server.
type Provider struct {
	endpoint string
	client   *http.Client
}

// NewProvider creates a new Ollama provider instance targeting the given endpoint.
func NewProvider(endpoint string) *Provider {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	endpoint = strings.TrimRight(endpoint, "/")
	return &Provider{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *Provider) Name() string {
	return "ollama"
}

// Health checks if Ollama server is reachable and running by calling /api/version.
func (p *Provider) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint+"/api/version", nil)
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

type tagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// ListModels discovers installed models by querying /api/tags.
func (p *Provider) ListModels(ctx context.Context) ([]llm.Model, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint+"/api/tags", nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", llm.ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list ollama models: status %d", resp.StatusCode)
	}

	var data tagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode ollama tags: %w", err)
	}

	models := make([]llm.Model, 0, len(data.Models))
	for _, m := range data.Models {
		lower := strings.ToLower(m.Name)
		cap := llm.Capability{
			Chat:             true,
			Streaming:        true,
			ToolCalling:      true,
			StructuredOutput: true,
			MaxTokens:        128000,
		}
		if strings.Contains(lower, "r1") || strings.Contains(lower, "reason") {
			cap.Reasoning = true
		}
		models = append(models, llm.Model{
			Name:       m.Name,
			Provider:   p.Name(),
			Capability: cap,
			Status:     "READY",
		})
	}
	return models, nil
}

type chatPayload struct {
	Model    string        `json:"model"`
	Messages []llm.Message `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatRespPayload struct {
	Message          llm.Message `json:"message"`
	Done             bool        `json:"done"`
	DoneReason       string      `json:"done_reason"`
	PromptEvalCount  int         `json:"prompt_eval_count"`
	EvalCount        int         `json:"eval_count"`
}

// Chat sends a non-streaming chat completion request to /api/chat.
func (p *Provider) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	payload := chatPayload{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   false,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint+"/api/chat", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", llm.ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, llm.ErrModelNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama chat error (%d): %s", resp.StatusCode, string(body))
	}

	var data chatRespPayload
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode ollama chat response: %w", err)
	}

	return &llm.ChatResponse{
		Message:          data.Message,
		FinishReason:     data.DoneReason,
		PromptTokens:     data.PromptEvalCount,
		CompletionTokens: data.EvalCount,
	}, nil
}

// ChatStream sends a streaming chat completion request to /api/chat.
func (p *Provider) ChatStream(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	payload := chatPayload{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   true,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint+"/api/chat", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", llm.ErrProviderUnavailable, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, llm.ErrModelNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ollama stream error (%d): %s", resp.StatusCode, string(body))
	}

	ch := make(chan llm.StreamChunk)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			var data chatRespPayload
			if err := json.Unmarshal(line, &data); err != nil {
				ch <- llm.StreamChunk{Error: err}
				return
			}
			chunk := llm.StreamChunk{
				Delta:        data.Message,
				FinishReason: data.DoneReason,
			}
			select {
			case <-ctx.Done():
				return
			case ch <- chunk:
			}
			if data.Done {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- llm.StreamChunk{Error: err}
		}
	}()

	return ch, nil
}
