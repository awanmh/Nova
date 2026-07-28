package memory

import (
	"context"
	"time"
)

// Scope defines the lifespan of a memory item.
type Scope string

const (
	ScopeSession Scope = "session"
	ScopeProject Scope = "project"
	ScopeUser    Scope = "user"
)

// Item represents a single stored memory or preference.
type Item struct {
	ID        string    `json:"id"`
	Scope     Scope     `json:"scope"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

// Engine defines the interface for persisting and retrieving context memory across turns and sessions.
type Engine interface {
	Store(ctx context.Context, scope Scope, key, value string) error
	Retrieve(ctx context.Context, scope Scope, key string) (*Item, error)
	List(ctx context.Context, scope Scope) ([]Item, error)
	Delete(ctx context.Context, id string) error
}
