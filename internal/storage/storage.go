package storage

import (
	"context"
	"time"
)

// Session represents an interactive session in NOVA.
type Session struct {
	ID          string    `json:"id"`
	Workspace   string    `json:"workspace"`
	Model       string    `json:"model"`
	Persona     string    `json:"persona"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Store defines the persistence layer interface for sessions, messages, plans, and events.
type Store interface {
	Init(ctx context.Context) error
	Close() error

	SaveSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, id string) (*Session, error)
	ListSessions(ctx context.Context, workspace string) ([]Session, error)
}
