package llm

import (
	"context"
	"errors"
)

var (
	ErrProviderUnavailable = errors.New("provider is unavailable or unreachable")
	ErrModelNotFound       = errors.New("requested model was not found")
	ErrRequestTimeout      = errors.New("request to provider timed out")
)

// Capability defines what a model supports.
type Capability struct {
	Chat             bool `json:"chat"`
	Streaming        bool `json:"streaming"`
	ToolCalling      bool `json:"tool_calling"`
	StructuredOutput bool `json:"structured_output"`
	Vision           bool `json:"vision"`
	Reasoning        bool `json:"reasoning"`
	MaxTokens        int  `json:"max_tokens"`
}

// Model represents a discovered AI model.
type Model struct {
	Name       string     `json:"name"`
	Provider   string     `json:"provider"`
	Capability Capability `json:"capability"`
	Status     string     `json:"status"` // e.g., "READY", "UNAVAILABLE"
}

// Message represents a single chat message.
type Message struct {
	Role      string     `json:"role"` // "system", "user", "assistant", "tool"
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	ToolID    string     `json:"tool_id,omitempty"`
}

// ToolCall represents a tool invocation requested by the model.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON-encoded string
}

// ToolResponse represents the result of executing a ToolCall.
type ToolResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// ChatRequest represents an invocation of the LLM.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream"`
}

// ChatResponse represents the result of an LLM invocation.
type ChatResponse struct {
	Message          Message `json:"message"`
	FinishReason     string  `json:"finish_reason"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
}

// StreamChunk represents a chunk in a streaming response.
type StreamChunk struct {
	Delta        Message `json:"delta"`
	FinishReason string  `json:"finish_reason,omitempty"`
	Error        error   `json:"-"`
}

// Provider defines the interface that all LLM provider backends (Ollama, OpenAI, llama.cpp) must implement.
type Provider interface {
	Name() string
	Health(ctx context.Context) error
	ListModels(ctx context.Context) ([]Model, error)
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error)
}
