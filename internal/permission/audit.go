package permission

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditRecord represents a single tool execution audit trail entry.
type AuditRecord struct {
	Timestamp  time.Time `json:"timestamp"`
	ToolName   string    `json:"tool_name"`
	RiskLevel  RiskLevel `json:"risk_level"`
	Arguments  string    `json:"arguments"`
	Allowed    bool      `json:"allowed"`
	Status     string    `json:"status"` // "SUCCESS", "ERROR", "DENIED"
	DurationMs int64     `json:"duration_ms"`
}

// AuditLogger defines the interface for recording and querying tool execution audit logs.
type AuditLogger interface {
	Log(ctx context.Context, rec AuditRecord) error
	ReadRecent(ctx context.Context, limit int) ([]AuditRecord, error)
}

type FileAuditLogger struct {
	mu      sync.Mutex
	logPath string
}

// NewFileAuditLogger creates a new audit logger writing to <rootDir>/.nova/audit.log.
func NewFileAuditLogger(rootDir string) (*FileAuditLogger, error) {
	novaDir := filepath.Join(rootDir, ".nova")
	if err := os.MkdirAll(novaDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .nova directory: %w", err)
	}
	return &FileAuditLogger{
		logPath: filepath.Join(novaDir, "audit.log"),
	}, nil
}

func (l *FileAuditLogger) Log(ctx context.Context, rec AuditRecord) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if rec.Timestamp.IsZero() {
		rec.Timestamp = time.Now().UTC()
	}

	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	_, err = f.Write(append(data, '\n'))
	return err
}

func (l *FileAuditLogger) ReadRecent(ctx context.Context, limit int) ([]AuditRecord, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.Open(l.logPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var all []AuditRecord
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		var rec AuditRecord
		if err := json.Unmarshal([]byte(line), &rec); err == nil {
			all = append(all, rec)
		}
	}

	if len(all) > limit && limit > 0 {
		return all[len(all)-limit:], nil
	}
	return all, nil
}
