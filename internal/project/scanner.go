package project

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/awanmh/Nova/internal/security"
)

// WorkspaceScanner implements Scanner by walking the filesystem directory tree.
type WorkspaceScanner struct {
	mu        sync.RWMutex
	rootPath  string
	files     map[string]FileMeta
	snapshot  *Snapshot
}

// NewWorkspaceScanner creates a new project scanner.
func NewWorkspaceScanner() *WorkspaceScanner {
	return &WorkspaceScanner{
		files: make(map[string]FileMeta),
	}
}

// ClassifyFile determines the FileClass and programming language of a relative file path.
func ClassifyFile(relPath string) (FileClass, string) {
	if security.IsSensitiveFile(relPath) {
		return FileClassSecret, ""
	}

	base := strings.ToLower(filepath.Base(relPath))
	ext := strings.ToLower(filepath.Ext(relPath))

	// Tests
	if strings.HasSuffix(base, "_test.go") || strings.HasSuffix(base, ".spec.ts") || strings.HasSuffix(base, ".test.js") || strings.HasSuffix(base, "_test.py") {
		return FileClassTest, langFromExt(ext)
	}

	// Documentation
	if ext == ".md" || ext == ".txt" || base == "license" || base == "changelog" {
		return FileClassDoc, "Markdown"
	}

	// Configuration
	if base == "go.mod" || base == "go.sum" || base == "package.json" || base == "tsconfig.json" || base == "cargo.toml" || ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".toml" || base == "makefile" {
		return FileClassConfig, "Config"
	}

	// Source
	lang := langFromExt(ext)
	if lang != "" {
		return FileClassSource, lang
	}

	return FileClassIgnored, ""
}

func langFromExt(ext string) string {
	switch ext {
	case ".go":
		return "Go"
	case ".js":
		return "JavaScript"
	case ".ts", ".tsx":
		return "TypeScript"
	case ".py":
		return "Python"
	case ".rs":
		return "Rust"
	case ".java":
		return "Java"
	case ".c", ".h":
		return "C"
	case ".cpp", ".hpp":
		return "C++"
	default:
		return ""
	}
}

// isIgnoredDir checks if a directory should be excluded from scanning.
func isIgnoredDir(name string) bool {
	switch name {
	case ".git", ".idea", ".vscode", "node_modules", "vendor", "bin", "dist", ".nova":
		return true
	}
	return false
}

// Scan traverses the workspace directory tree and builds a project architectural snapshot.
func (s *WorkspaceScanner) Scan(ctx context.Context, rootPath string) (*Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	s.rootPath = absRoot
	s.files = make(map[string]FileMeta)

	var fileList []FileMeta

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // ignore inaccessible files
		}

		if d.IsDir() && isIgnoredDir(d.Name()) {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		info, err := d.Info()
		if err != nil {
			return nil
		}

		class, lang := ClassifyFile(rel)
		if class == FileClassIgnored || class == FileClassSecret {
			return nil
		}

		meta := FileMeta{
			Path:      rel,
			Class:     class,
			Language:  lang,
			SizeBytes: info.Size(),
		}

		s.files[rel] = meta
		fileList = append(fileList, meta)
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	languages := DetectLanguages(fileList)
	frameworks := DetectFrameworks(absRoot, fileList)
	entryPoints := DetectEntryPoints(fileList)

	snap := &Snapshot{
		Name:         filepath.Base(absRoot),
		RootPath:     absRoot,
		Languages:    languages,
		Frameworks:   frameworks,
		EntryPoints:  entryPoints,
		Architecture: "Hexagonal / Modular Architecture",
		Files:        fileList,
	}

	snap.Summary = GenerateSummary(snap)
	s.snapshot = snap
	return snap, nil
}

// GetFileMeta returns metadata for a file if present in the last scan.
func (s *WorkspaceScanner) GetFileMeta(path string) (*FileMeta, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	meta, ok := s.files[filepath.ToSlash(path)]
	return &meta, ok
}

// RescanIncremental updates specific changed paths in the snapshot.
func (s *WorkspaceScanner) RescanIncremental(ctx context.Context, changedPaths []string) (*Snapshot, error) {
	// For simplicity and accuracy, if snapshot doesn't exist, do a full scan
	s.mu.Lock()
	if s.snapshot == nil || s.rootPath == "" {
		s.mu.Unlock()
		return s.Scan(ctx, s.rootPath)
	}

	for _, p := range changedPaths {
		rel := filepath.ToSlash(p)
		abs := filepath.Join(s.rootPath, rel)
		info, err := os.Stat(abs)
		if err != nil {
			// file deleted
			delete(s.files, rel)
			continue
		}
		class, lang := ClassifyFile(rel)
		if class != FileClassIgnored {
			s.files[rel] = FileMeta{
				Path:      rel,
				Class:     class,
				Language:  lang,
				SizeBytes: info.Size(),
			}
		} else {
			delete(s.files, rel)
		}
	}

	// Rebuild slice and refresh summary
	var fileList []FileMeta
	for _, m := range s.files {
		fileList = append(fileList, m)
	}
	s.snapshot.Files = fileList
	s.snapshot.Languages = DetectLanguages(fileList)
	s.snapshot.Frameworks = DetectFrameworks(s.rootPath, fileList)
	s.snapshot.EntryPoints = DetectEntryPoints(fileList)
	s.snapshot.Summary = GenerateSummary(s.snapshot)
	s.mu.Unlock()

	return s.snapshot, nil
}
