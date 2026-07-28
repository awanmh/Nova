package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/awanmh/Nova/internal/llm"
)

// Session represents a chat conversation session.
type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Persona   string    `json:"persona"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SessionRecord stores session metadata and its message history.
type SessionRecord struct {
	Session  Session       `json:"session"`
	Messages []llm.Message `json:"messages"`
}

// SessionStore defines CRUD operations for sessions and message histories.
type SessionStore interface {
	CreateSession(ctx context.Context, id, title, persona string) (Session, error)
	GetSession(ctx context.Context, id string) (Session, error)
	ListSessions(ctx context.Context) ([]Session, error)
	DeleteSession(ctx context.Context, id string) error
	AppendMessage(ctx context.Context, sessionID string, msg llm.Message) error
	GetHistory(ctx context.Context, sessionID string) ([]llm.Message, error)
	ClearHistory(ctx context.Context, sessionID string) error
	ClearAll(ctx context.Context) error
}

// FileStore implements SessionStore and Engine using a persistent JSON file storage in .nova/memory.json.
type FileStore struct {
	mu       sync.RWMutex
	filePath string
	sessions map[string]*SessionRecord
	items    map[string]Item // for key-value memory engine
}

type fileStoreData struct {
	Sessions map[string]*SessionRecord `json:"sessions"`
	Items    map[string]Item           `json:"items"`
}

// NewFileStore creates or loads a memory file store at <rootDir>/.nova/memory.json.
func NewFileStore(rootDir string) (*FileStore, error) {
	novaDir := filepath.Join(rootDir, ".nova")
	if err := os.MkdirAll(novaDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .nova directory: %w", err)
	}
	fs := &FileStore{
		filePath: filepath.Join(novaDir, "memory.json"),
		sessions: make(map[string]*SessionRecord),
		items:    make(map[string]Item),
	}
	if err := fs.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return fs, nil
}

func (fs *FileStore) load() error {
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		return err
	}
	var d fileStoreData
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	if d.Sessions != nil {
		fs.sessions = d.Sessions
	}
	if d.Items != nil {
		fs.items = d.Items
	}
	return nil
}

func (fs *FileStore) save() error {
	d := fileStoreData{
		Sessions: fs.sessions,
		Items:    fs.items,
	}
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.filePath, data, 0644)
}

// SessionStore implementations
func (fs *FileStore) CreateSession(ctx context.Context, id, title, persona string) (Session, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	now := time.Now().UTC()
	s := Session{
		ID:        id,
		Title:     title,
		Persona:   persona,
		CreatedAt: now,
		UpdatedAt: now,
	}
	fs.sessions[id] = &SessionRecord{
		Session:  s,
		Messages: []llm.Message{},
	}
	return s, fs.save()
}

func (fs *FileStore) GetSession(ctx context.Context, id string) (Session, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	rec, ok := fs.sessions[id]
	if !ok {
		return Session{}, fmt.Errorf("session '%s' not found", id)
	}
	return rec.Session, nil
}

func (fs *FileStore) ListSessions(ctx context.Context) ([]Session, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	var list []Session
	for _, rec := range fs.sessions {
		list = append(list, rec.Session)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].UpdatedAt.After(list[j].UpdatedAt)
	})
	return list, nil
}

func (fs *FileStore) DeleteSession(ctx context.Context, id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	delete(fs.sessions, id)
	return fs.save()
}

func (fs *FileStore) AppendMessage(ctx context.Context, sessionID string, msg llm.Message) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	rec, ok := fs.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session '%s' not found", sessionID)
	}
	rec.Messages = append(rec.Messages, msg)
	rec.Session.UpdatedAt = time.Now().UTC()
	return fs.save()
}

func (fs *FileStore) GetHistory(ctx context.Context, sessionID string) ([]llm.Message, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	rec, ok := fs.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session '%s' not found", sessionID)
	}
	return append([]llm.Message(nil), rec.Messages...), nil
}

func (fs *FileStore) ClearHistory(ctx context.Context, sessionID string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	rec, ok := fs.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session '%s' not found", sessionID)
	}
	rec.Messages = []llm.Message{}
	rec.Session.UpdatedAt = time.Now().UTC()
	return fs.save()
}

func (fs *FileStore) ClearAll(ctx context.Context) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.sessions = make(map[string]*SessionRecord)
	return fs.save()
}

// Engine implementations for key-value memory item store
func (fs *FileStore) Store(ctx context.Context, scope Scope, key, value string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	id := fmt.Sprintf("%s:%s", scope, key)
	fs.items[id] = Item{
		ID:        id,
		Scope:     scope,
		Key:       key,
		Value:     value,
		CreatedAt: time.Now().UTC(),
	}
	return fs.save()
}

func (fs *FileStore) Retrieve(ctx context.Context, scope Scope, key string) (*Item, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	id := fmt.Sprintf("%s:%s", scope, key)
	it, ok := fs.items[id]
	if !ok {
		return nil, fmt.Errorf("memory item '%s' not found", key)
	}
	return &it, nil
}

func (fs *FileStore) List(ctx context.Context, scope Scope) ([]Item, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	var list []Item
	for _, it := range fs.items {
		if it.Scope == scope {
			list = append(list, it)
		}
	}
	return list, nil
}

func (fs *FileStore) Delete(ctx context.Context, id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	delete(fs.items, id)
	return fs.save()
}
