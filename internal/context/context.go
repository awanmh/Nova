package context

import (
	"context"
)

// ItemType describes the source of a context item.
type ItemType string

const (
	ItemTypeFile   ItemType = "file"
	ItemTypeSymbol ItemType = "symbol"
	ItemTypeSearch ItemType = "search"
	ItemTypeRule   ItemType = "rule"
)

// Item represents a chunk of code or text injected into the LLM prompt.
type Item struct {
	Type     ItemType `json:"type"`
	Path     string   `json:"path"`
	Content  string   `json:"content"`
	Score    float64  `json:"score"`
	TokenCnt int      `json:"token_cnt"`
}

// Bundle represents the curated collection of context items for a prompt.
type Bundle struct {
	Items      []Item `json:"items"`
	TotalTokens int   `json:"total_tokens"`
}

// Request represents criteria for retrieving context.
type Request struct {
	Query       string   `json:"query"`
	TargetPaths []string `json:"target_paths,omitempty"`
	MaxTokens   int      `json:"max_tokens"`
}

// Engine defines the interface for selecting, ranking, and compressing context.
type Engine interface {
	Retrieve(ctx context.Context, req *Request) (*Bundle, error)
	Rank(items []Item, query string) []Item
	Compress(bundle *Bundle, maxTokens int) *Bundle
}

// EstimateTokens provides a fast approximation of LLM token count (~4 characters per token).
func EstimateTokens(content string) int {
	if len(content) == 0 {
		return 0
	}
	tokens := len(content) / 4
	if tokens == 0 {
		return 1
	}
	return tokens
}
