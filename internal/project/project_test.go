package project_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awanmh/Nova/internal/project"
)

func TestClassifyFile(t *testing.T) {
	tests := []struct {
		path     string
		expClass project.FileClass
		expLang  string
	}{
		{"cmd/nova/main.go", project.FileClassSource, "Go"},
		{"internal/app/app_test.go", project.FileClassTest, "Go"},
		{"go.mod", project.FileClassConfig, "Config"},
		{"README.md", project.FileClassDoc, "Markdown"},
		{".env.local", project.FileClassSecret, ""},
	}

	for _, tc := range tests {
		class, lang := project.ClassifyFile(tc.path)
		if class != tc.expClass {
			t.Fatalf("expected class %s for %s, got %s", tc.expClass, tc.path, class)
		}
		if lang != tc.expLang {
			t.Fatalf("expected lang %s for %s, got %s", tc.expLang, tc.path, lang)
		}
	}
}

func TestWorkspaceScanner_Scan(t *testing.T) {
	tempDir := t.TempDir()

	_ = os.MkdirAll(filepath.Join(tempDir, "cmd", "nova"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "cmd", "nova", "main.go"), []byte("package main\nfunc main() {}\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte("module test\nrequire github.com/spf13/cobra v1.0.0\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# Test Repo\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, ".env"), []byte("SECRET=foo"), 0644)

	scanner := project.NewWorkspaceScanner()
	snap, err := scanner.Scan(context.Background(), tempDir)
	if err != nil {
		t.Fatalf("unexpected scan error: %v", err)
	}

	if len(snap.Files) != 3 {
		t.Fatalf("expected 3 files (excluding .env secret), got %d", len(snap.Files))
	}

	if len(snap.Languages) != 1 || snap.Languages[0] != "Go" {
		t.Fatalf("expected language Go, got: %v", snap.Languages)
	}

	if len(snap.Frameworks) != 1 || snap.Frameworks[0] != "Cobra (CLI)" {
		t.Fatalf("expected framework Cobra (CLI), got: %v", snap.Frameworks)
	}

	if len(snap.EntryPoints) != 1 || snap.EntryPoints[0] != "cmd/nova/main.go" {
		t.Fatalf("expected entrypoint cmd/nova/main.go, got: %v", snap.EntryPoints)
	}

	if !strings.Contains(snap.Summary, "Languages: Go") {
		t.Fatalf("expected summary to contain 'Languages: Go', got: %s", snap.Summary)
	}
}

func TestWorkspaceScanner_RescanIncremental(t *testing.T) {
	tempDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tempDir, "a.go"), []byte("package a\n"), 0644)

	scanner := project.NewWorkspaceScanner()
	snap, _ := scanner.Scan(context.Background(), tempDir)
	if len(snap.Files) != 1 {
		t.Fatalf("expected 1 file initially")
	}

	// Add b.go and rescan incrementally
	_ = os.WriteFile(filepath.Join(tempDir, "b.go"), []byte("package b\n"), 0644)
	snap2, err := scanner.RescanIncremental(context.Background(), []string{"b.go"})
	if err != nil {
		t.Fatalf("unexpected rescan error: %v", err)
	}
	if len(snap2.Files) != 2 {
		t.Fatalf("expected 2 files after incremental rescan, got %d", len(snap2.Files))
	}
}
