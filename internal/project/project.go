package project

import (
	"context"
)

// FileClass categorizes a file in the workspace.
type FileClass string

const (
	FileClassSource   FileClass = "source"
	FileClassTest     FileClass = "test"
	FileClassConfig   FileClass = "config"
	FileClassDoc      FileClass = "documentation"
	FileClassSecret   FileClass = "secret"
	FileClassIgnored  FileClass = "ignored"
)

// FileMeta holds metadata for a single workspace file.
type FileMeta struct {
	Path      string    `json:"path"`
	Class     FileClass `json:"class"`
	Language  string    `json:"language,omitempty"`
	SizeBytes int64     `json:"size_bytes"`
	Hash      string    `json:"hash,omitempty"`
}

// Snapshot represents the architectural understanding of a project.
type Snapshot struct {
	Name         string     `json:"name"`
	RootPath     string     `json:"root_path"`
	Languages    []string   `json:"languages"`
	Frameworks   []string   `json:"frameworks"`
	Dependencies []string   `json:"dependencies"`
	EntryPoints  []string   `json:"entry_points"`
	Architecture string     `json:"architecture"`
	Summary      string     `json:"summary"`
	Files        []FileMeta `json:"files,omitempty"`
}

// Scanner defines the interface for analyzing a workspace filesystem and building a project snapshot.
type Scanner interface {
	Scan(ctx context.Context, rootPath string) (*Snapshot, error)
	GetFileMeta(path string) (*FileMeta, bool)
	RescanIncremental(ctx context.Context, changedPaths []string) (*Snapshot, error)
}
