package openai

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

// Provider implements llm.Provider for OpenAI and OpenAI-compatible endpoints (vLLM, LM Studio, etc.).
type Provider struct {
	endpoint string
	apiKey   string
	client   *http.Client
}

// NewProvider creates a new OpenAI provider instance.
func NewProvider(endpoint, apiKey string) *Provider {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}
	endpoint = strings.TrimRight(endpoint, "/")
	return &Provider{
		endpoint: endpoint,
		apiKey:   apiKey,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *Provider) Name() string {
	return "openai"
}

func (p *Provider) setAuth(req *http.Request) {
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
}

// Health checks if OpenAI endpoint is reachable.
func (p *Provider) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint+"/models", nil)
	if err != nil {
		return err
	}
	p.setAuth(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", llm.ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("%w: unexpected status code %d", llm.ErrProviderUnavailable, resp.StatusCode)
}

type modelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// ListModels discovers available models by querying /models.
func (p *Provider) ListModels(ctx context.Context) ([]llm.Model, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint+"/models", nil)
	if err != nil {
		return nil, err
	}
	p.setAuth(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", llm.ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list openai models: status %d", resp.StatusCode)
	}

	var data modelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode openai models: %w", err)
	}

	models := make([]llm.Model, 0, len(data.Data))
	for _, m := range data.Data {
		lower := strings.ToLower(m.ID)
		cap := llm.Capability{
			Chat:             true,
			Streaming:        true,
			ToolCalling:      true,
			StructuredOutput: true,
			MaxTokens:        128000,
		}
		if strings.Contains(lower, "o1") || strings.Contains(lower, "o3") || strings.Contains(lower, "reason") {
			cap.Reasoning = true
		}
		models = append(models, llm.Model{
			Name:       m.ID,
			Provider:   p.Name(),
			Capability: cap,
			Status:     "READY",
		})
	}
	return models, nil
}

type chatPayload struct {
	Model       string        `json:"model"`
	Messages    []llm.Message `json:"messages"`
	Temperature float64       `json:"temperature"`
	Stream      bool          `json:"stream"`
}

type chatRespPayload struct {
	Choices []struct {
		Message      llm.Message `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// Chat sends a non-streaming chat completion request to /chat/completions.
func (p *Provider) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	payload := chatPayload{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		Stream:      false,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	p.setAuth(httpReq)

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
		return nil, fmt.Errorf("openai chat error (%d): %s", resp.StatusCode, string(body))
	}

	var data chatRespPayload
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode openai chat response: %w", err)
	}
	if len(data.Choices) == 0 {
		return nil, fmt.Errorf("openai returned 0 choices")
	}

	return &llm.ChatResponse{
		Message:          data.Choices[0].Message,
		FinishReason:     data.Choices[0].FinishReason,
		PromptTokens:     data.Usage.PromptTokens,
		CompletionTokens: data.Usage.CompletionTokens,
	}, nil
}

type streamRespPayload struct {
	Choices []struct {
		Delta        llm.Message `json:"delta"`
		FinishReason *string     `json:"finish_reason"`
	} `json:"choices"`
}

// ChatStream sends a streaming SSE request to /chat/completions.
func (p *Provider) ChatStream(ctx context.Context, req *llm.ChatRequest) (<-chan llm.StreamChunk, error) {
	payload := chatPayload{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		Stream:      true,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	p.setAuth(httpReq)

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
		return nil, fmt.Errorf("openai stream error (%d): %s", resp.StatusCode, string(body))
	}

	ch := make(chan llm.StreamChunk)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payload == "[DONE]" {
				return
			}
			var data streamRespPayload
			if err := json.Unmarshal([]byte(payload), &data); err != nil {
				ch <- llm.StreamChunk{Error: err}
				return
			}
			if len(data.Choices) == 0 {
				continue
			}
			finishReason := ""
			if data.Choices[0].FinishReason != nil {
				finishReason = *data.Choices[0].FinishReason
			}
			chunk := llm.StreamChunk{
				Delta:        data.Choices[0].Delta,
				FinishReason: finishReason,
			}
			select {
			case <-ctx.Done():
				return
			case ch <- chunk:
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- llm.StreamChunk{Error: err}
		}
	}()

	return ch, nil
}
