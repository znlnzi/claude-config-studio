package luoshu

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ciMockEmbedder is a test Embedding Provider that returns fixed vectors
type ciMockEmbedder struct {
	callCount int
}

func (m *ciMockEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	m.callCount += len(texts)
	vectors := make([][]float32, len(texts))
	for i, text := range texts {
		// Simple hash: generate different vectors using the first few characters
		vec := make([]float32, 4)
		for j := 0; j < len(text) && j < 4; j++ {
			vec[j] = float32(text[j]) / 256.0
		}
		vectors[i] = vec
	}
	return vectors, nil
}

func (m *ciMockEmbedder) Dimensions() int { return 4 }
func (m *ciMockEmbedder) Name() string    { return "mock" }

// setupTestProject creates a test project directory structure
func setupTestProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	memoryDir := filepath.Join(root, ".claude", "memory")
	rulesDir := filepath.Join(root, ".claude", "rules")
	os.MkdirAll(memoryDir, 0755)
	os.MkdirAll(rulesDir, 0755)

	os.WriteFile(filepath.Join(memoryDir, "MEMORY.md"), []byte(`# Project Memory

## Architecture
This project uses a microservices architecture with Go backend and React frontend.

## Key Decisions
- Chose PostgreSQL for database
- Using Redis for caching
`), 0644)

	os.WriteFile(filepath.Join(rulesDir, "coding-style.md"), []byte(`# Coding Style

## Naming
- Use camelCase for variables
- Use PascalCase for types
- Use UPPER_SNAKE_CASE for constants

## Error Handling
- Always handle errors explicitly
- Never silently swallow errors
`), 0644)

	os.WriteFile(filepath.Join(rulesDir, "security.md"), []byte(`# Security Rules

## API Keys
- Never hardcode API keys
- Use environment variables for secrets

## Input Validation
- Validate all user input
- Use parameterized queries
`), 0644)

	return root
}

func TestClaudeIndex_ScanFiles(t *testing.T) {
	root := setupTestProject(t)
	embedder := &ciMockEmbedder{}
	ci := NewClaudeIndex(embedder, nil, "test-model")

	files := ci.scanFiles(root)
	if len(files) != 3 {
		t.Fatalf("expected 3 .md files, got %d", len(files))
	}

	// Verify file paths are absolute
	for path := range files {
		if !filepath.IsAbs(path) {
			t.Errorf("expected absolute path, got %s", path)
		}
	}
}

func TestClaudeIndex_ScanFiles_Empty(t *testing.T) {
	root := t.TempDir()
	ci := NewClaudeIndex(&ciMockEmbedder{}, nil, "test-model")

	files := ci.scanFiles(root)
	if len(files) != 0 {
		t.Fatalf("expected 0 files for empty project, got %d", len(files))
	}
}

func TestClaudeIndex_ScanFiles_RootMemory(t *testing.T) {
	root := setupTestProject(t)
	// Create official auto-memory MEMORY.md at the project root
	os.WriteFile(filepath.Join(root, "MEMORY.md"), []byte(`# Official Auto-Memory

## User Preferences
- Prefers functional programming style
- Uses dark theme
`), 0644)

	ci := NewClaudeIndex(&ciMockEmbedder{}, nil, "test-model")
	files := ci.scanFiles(root)

	// Should have 3 files under .claude/ + 1 root MEMORY.md = 4
	if len(files) != 4 {
		t.Fatalf("expected 4 files (3 in .claude + 1 root MEMORY.md), got %d", len(files))
	}

	// Verify root MEMORY.md was scanned
	rootMemoryPath := filepath.Join(root, "MEMORY.md")
	if _, ok := files[rootMemoryPath]; !ok {
		t.Errorf("root MEMORY.md not found in scanned files")
	}
}

func TestClaudeIndex_ScanFiles_RootMemory_GlobalScope(t *testing.T) {
	// global scope should not scan root MEMORY.md
	ci := NewClaudeIndex(&ciMockEmbedder{}, nil, "test-model")
	files := ci.scanFiles("global")
	// Should not crash, return normally
	_ = files
}

func TestClaudeIndex_DiffFiles(t *testing.T) {
	ci := NewClaudeIndex(&ciMockEmbedder{}, nil, "test-model")

	manifest := map[string]FileState{
		"/a.md": {Mtime: 100, Size: 50},
		"/b.md": {Mtime: 200, Size: 60},
		"/c.md": {Mtime: 300, Size: 70},
	}
	current := map[string]FileState{
		"/a.md": {Mtime: 100, Size: 50},  // no change
		"/b.md": {Mtime: 250, Size: 65},  // modified
		"/d.md": {Mtime: 400, Size: 80},  // new
	}

	toAdd, toRemove := ci.diffFiles(manifest, current)

	if len(toAdd) != 2 {
		t.Errorf("expected 2 files to add (modified + new), got %d", len(toAdd))
	}
	if len(toRemove) != 1 {
		t.Errorf("expected 1 file to remove (deleted), got %d", len(toRemove))
	}

	// Verify toRemove contains /c.md
	found := false
	for _, p := range toRemove {
		if p == "/c.md" {
			found = true
		}
	}
	if !found {
		t.Error("expected /c.md in toRemove")
	}
}

func TestClaudeIndex_Manifest_ReadWrite(t *testing.T) {
	dir := t.TempDir()
	ci := NewClaudeIndex(&ciMockEmbedder{}, nil, "test-model")

	// Write
	m := &FileManifest{
		Version: 1,
		Scope:   "test",
		Files: map[string]FileState{
			"/test.md": {Mtime: 100, Size: 50, IndexedAt: "2026-01-01T00:00:00Z"},
		},
		LastReconciled: "2026-01-01T00:00:00Z",
	}
	if err := ci.saveManifest(dir, m); err != nil {
		t.Fatal(err)
	}

	// Read
	loaded := ci.loadManifest(dir)
	if loaded.Version != 1 {
		t.Errorf("expected version 1, got %d", loaded.Version)
	}
	if len(loaded.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(loaded.Files))
	}
	if loaded.Files["/test.md"].Mtime != 100 {
		t.Errorf("expected mtime 100, got %f", loaded.Files["/test.md"].Mtime)
	}
}

func TestClaudeIndex_Manifest_NotExist(t *testing.T) {
	dir := t.TempDir()
	ci := NewClaudeIndex(&ciMockEmbedder{}, nil, "test-model")

	m := ci.loadManifest(filepath.Join(dir, "nonexistent"))
	if m.Version != 1 {
		t.Errorf("expected default version 1, got %d", m.Version)
	}
	if len(m.Files) != 0 {
		t.Errorf("expected empty files, got %d", len(m.Files))
	}
}

func TestClaudeIndex_Reconcile(t *testing.T) {
	root := setupTestProject(t)

	// Override index path with temp dir
	cacheDir := t.TempDir()
	cache := &EmbeddingCache{dir: cacheDir, cache: make(map[string][]float32)}
	embedder := &ciMockEmbedder{}

	ci := NewClaudeIndex(embedder, cache, "test-model")
	// Override indexDir to use temp directory
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	ctx := context.Background()
	changes, err := ci.Reconcile(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	if changes != 3 {
		t.Errorf("expected 3 changes on first reconcile, got %d", changes)
	}
	if embedder.callCount == 0 {
		t.Error("expected embedder to be called")
	}

	// Reconcile again (during cooldown) should return 0
	changes2, err := ci.Reconcile(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	if changes2 != 0 {
		t.Errorf("expected 0 changes during cooldown, got %d", changes2)
	}
}

func TestClaudeIndex_ReconcileCooldown(t *testing.T) {
	ci := NewClaudeIndex(&NoopEmbeddingProvider{}, nil, "test-model")

	// Set last reconcile to just now
	ci.lastReconcile["test-scope"] = time.Now()

	changes, err := ci.Reconcile(context.Background(), "test-scope")
	if err != nil {
		t.Fatal(err)
	}
	if changes != 0 {
		t.Errorf("expected 0 during cooldown, got %d", changes)
	}
}

func TestClaudeIndex_Reindex(t *testing.T) {
	root := setupTestProject(t)

	cacheDir := t.TempDir()
	cache := &EmbeddingCache{dir: cacheDir, cache: make(map[string][]float32)}
	embedder := &ciMockEmbedder{}

	ci := NewClaudeIndex(embedder, cache, "test-model")

	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	ctx := context.Background()
	if err := ci.Reindex(ctx, root); err != nil {
		t.Fatal(err)
	}

	status, err := ci.Status(root)
	if err != nil {
		t.Fatal(err)
	}
	if status.IndexedFiles != 3 {
		t.Errorf("expected 3 indexed files, got %d", status.IndexedFiles)
	}
	if status.VectorEntries == 0 {
		t.Error("expected vector entries > 0")
	}
}

func TestClaudeIndex_Search(t *testing.T) {
	root := setupTestProject(t)

	cacheDir := t.TempDir()
	cache := &EmbeddingCache{dir: cacheDir, cache: make(map[string][]float32)}
	embedder := &ciMockEmbedder{}

	ci := NewClaudeIndex(embedder, cache, "test-model")

	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	ctx := context.Background()

	// Index first
	if err := ci.Reindex(ctx, root); err != nil {
		t.Fatal(err)
	}

	// Search
	results, err := ci.Search(ctx, "error handling", root, 10)
	if err != nil {
		t.Fatal(err)
	}

	// Should have results (vectors from mock embedder have some similarity)
	if len(results) == 0 {
		t.Error("expected at least one search result")
	}

	// Verify result structure
	for _, r := range results {
		if r.FilePath == "" {
			t.Error("expected non-empty file path")
		}
		if r.Score <= 0 {
			t.Errorf("expected positive score, got %f", r.Score)
		}
		if r.Content == "" {
			t.Error("expected non-empty content")
		}
	}
}

func TestClaudeIndex_Search_NoopEmbedder(t *testing.T) {
	ci := NewClaudeIndex(&NoopEmbeddingProvider{}, nil, "test-model")

	_, err := ci.Search(context.Background(), "test", "global", 10)
	if err == nil {
		t.Error("expected error with noop embedder")
	}
}

func TestClaudeIndex_Status(t *testing.T) {
	ci := NewClaudeIndex(&ciMockEmbedder{}, nil, "test-model")

	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	status, err := ci.Status("global")
	if err != nil {
		t.Fatal(err)
	}
	if status.Scope != "global" {
		t.Errorf("expected scope 'global', got %s", status.Scope)
	}
	if status.IndexedFiles != 0 {
		t.Errorf("expected 0 indexed files initially, got %d", status.IndexedFiles)
	}
}

func TestClaudeIndex_ScopeIsolation(t *testing.T) {
	ci := NewClaudeIndex(&ciMockEmbedder{}, nil, "test-model")

	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	globalDir := ci.indexDir("global")
	projectDir := ci.indexDir("/some/project")

	if globalDir == projectDir {
		t.Error("global and project index dirs should be different")
	}
}

func TestEncodePath(t *testing.T) {
	// Same path should produce same encoding
	a := encodePath("/Users/test/project")
	b := encodePath("/Users/test/project")
	if a != b {
		t.Errorf("expected same encoding, got %s vs %s", a, b)
	}

	// Different paths should produce different encoding
	c := encodePath("/Users/test/other")
	if a == c {
		t.Error("expected different encoding for different paths")
	}
}

func TestClaudeIndex_FileID(t *testing.T) {
	ci := NewClaudeIndex(&ciMockEmbedder{}, nil, "test-model")

	id1 := ci.fileID("/path/to/file.md")
	id2 := ci.fileID("/path/to/file.md")
	id3 := ci.fileID("/path/to/other.md")

	if id1 != id2 {
		t.Error("same path should produce same ID")
	}
	if id1 == id3 {
		t.Error("different paths should produce different IDs")
	}
	if len(id1) < 10 {
		t.Errorf("ID too short: %s", id1)
	}
}

func TestClaudeIndex_ExtractContext(t *testing.T) {
	ci := NewClaudeIndex(&ciMockEmbedder{}, nil, "test-model")

	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{"h1", "# Title\nContent", "Title"},
		{"h2", "## Section\nContent", "Section"},
		{"no heading", "Just some content", "Just some content"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ci.extractContext(tt.text)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
