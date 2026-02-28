package luoshu

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ClaudeIndex is a file indexer that performs semantic search over .claude/memory/*.md and .claude/rules/*.md
type ClaudeIndex struct {
	embedder EmbeddingProvider
	cache    *EmbeddingCache
	model    string

	mu            sync.Mutex
	lastReconcile map[string]time.Time // scope -> last reconcile time
}

// FileManifest tracks file state for indexing
type FileManifest struct {
	Version        int                  `json:"version"`
	Scope          string               `json:"scope"`
	Files          map[string]FileState `json:"files"`
	LastReconciled string               `json:"last_reconciled"`
}

// FileState represents the state of a single file
type FileState struct {
	Mtime     float64 `json:"mtime"`
	Size      int64   `json:"size"`
	IndexedAt string  `json:"indexed_at"`
}

// FileSearchResult represents a file search result
type FileSearchResult struct {
	FilePath string  `json:"file_path"`
	Score    float64 `json:"score"`
	Content  string  `json:"content"`
	Context  string  `json:"context"`
}

// IndexStatus represents index status information
type IndexStatus struct {
	Scope          string `json:"scope"`
	IndexedFiles   int    `json:"indexed_files"`
	VectorEntries  int    `json:"vector_entries"`
	LastReconciled string `json:"last_reconciled"`
	IndexDir       string `json:"index_dir"`
}

// NewClaudeIndex creates a file indexer
func NewClaudeIndex(embedder EmbeddingProvider, cache *EmbeddingCache, model string) *ClaudeIndex {
	return &ClaudeIndex{
		embedder:      embedder,
		cache:         cache,
		model:         model,
		lastReconcile: make(map[string]time.Time),
	}
}

// reconcileCooldown is the minimum interval between two Reconcile calls
const reconcileCooldown = 5 * time.Second

// Reconcile scans files and incrementally syncs the index, returning the number of changed files
func (ci *ClaudeIndex) Reconcile(ctx context.Context, scope string) (int, error) {
	ci.mu.Lock()
	if t, ok := ci.lastReconcile[scope]; ok && time.Since(t) < reconcileCooldown {
		ci.mu.Unlock()
		return 0, nil
	}
	ci.lastReconcile[scope] = time.Now()
	ci.mu.Unlock()

	indexDir := ci.indexDir(scope)
	if err := os.MkdirAll(indexDir, 0700); err != nil {
		return 0, fmt.Errorf("failed to create index directory: %w", err)
	}

	manifest := ci.loadManifest(indexDir)

	// Scan current files
	currentFiles := ci.scanFiles(scope)

	// Detect changes
	toAdd, toRemove := ci.diffFiles(manifest.Files, currentFiles)
	if len(toAdd) == 0 && len(toRemove) == 0 {
		return 0, nil
	}

	// Load vector index
	vecDir := filepath.Join(indexDir, "vectors")
	index, err := NewVectorIndexAt(vecDir)
	if err != nil {
		return 0, fmt.Errorf("failed to load vector index: %w", err)
	}

	// Remove old vectors for deleted/modified files
	for _, path := range toRemove {
		fileID := ci.fileID(path)
		index.Remove(fileID)
	}
	for _, path := range toAdd {
		fileID := ci.fileID(path)
		index.Remove(fileID)
	}

	// Index new/modified files
	isEmbedAvailable := ci.embedder.Name() != "noop"
	for _, path := range toAdd {
		if !isEmbedAvailable {
			break
		}
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		ci.indexFile(ctx, index, path, string(content))
	}

	// Save index
	if err := index.Save(); err != nil {
		return 0, fmt.Errorf("failed to save vector index: %w", err)
	}

	// Update manifest
	now := time.Now().UTC().Format(time.RFC3339)
	for _, path := range toRemove {
		delete(manifest.Files, path)
	}
	for _, path := range toAdd {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		manifest.Files[path] = FileState{
			Mtime:     float64(info.ModTime().Unix()),
			Size:      info.Size(),
			IndexedAt: now,
		}
	}
	manifest.LastReconciled = now
	ci.saveManifest(indexDir, manifest)

	if ci.cache != nil {
		ci.cache.Save()
	}

	return len(toAdd) + len(toRemove), nil
}

// Search performs semantic search over file contents
func (ci *ClaudeIndex) Search(ctx context.Context, query string, scope string, limit int) ([]FileSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	if ci.embedder.Name() == "noop" {
		return nil, fmt.Errorf("embedding not configured, cannot perform semantic search")
	}

	indexDir := ci.indexDir(scope)
	vecDir := filepath.Join(indexDir, "vectors")
	index, err := NewVectorIndexAt(vecDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load vector index: %w", err)
	}

	// Get query vector
	queryVec, ok := ci.cache.Get(query, ci.model)
	if !ok {
		vectors, err := ci.embedder.Embed(ctx, []string{query})
		if err != nil {
			return nil, fmt.Errorf("failed to generate query vector: %w", err)
		}
		if len(vectors) == 0 || len(vectors[0]) == 0 {
			return nil, nil
		}
		queryVec = vectors[0]
		ci.cache.Set(query, ci.model, queryVec)
		ci.cache.Save()
	}

	matches := index.Search(queryVec, limit*3, 0.3)

	// Deduplicate by file, keeping the highest score
	seen := make(map[string]bool)
	var results []FileSearchResult
	for _, m := range matches {
		filePath := m.Entry.MemoryID // We use MemoryID to store the hashed file path ID
		if seen[filePath] {
			continue
		}
		seen[filePath] = true

		content := m.Entry.Text
		if len(content) > 2000 {
			content = content[:2000]
		}

		results = append(results, FileSearchResult{
			FilePath: ci.resolveFilePath(m.Entry.MemoryID, scope),
			Score:    m.Score,
			Content:  content,
			Context:  ci.extractContext(m.Entry.Text),
		})
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// Reindex forces a full index rebuild
func (ci *ClaudeIndex) Reindex(ctx context.Context, scope string) error {
	indexDir := ci.indexDir(scope)
	if err := os.MkdirAll(indexDir, 0700); err != nil {
		return fmt.Errorf("failed to create index directory: %w", err)
	}

	vecDir := filepath.Join(indexDir, "vectors")
	index, err := NewVectorIndexAt(vecDir)
	if err != nil {
		return fmt.Errorf("failed to load vector index: %w", err)
	}

	// Clear existing index
	if err := index.Clear(); err != nil {
		return fmt.Errorf("failed to clear index: %w", err)
	}

	files := ci.scanFiles(scope)
	manifest := &FileManifest{
		Version: 1,
		Scope:   scope,
		Files:   make(map[string]FileState),
	}

	now := time.Now().UTC().Format(time.RFC3339)
	isEmbedAvailable := ci.embedder.Name() != "noop"

	for path, info := range files {
		if !isEmbedAvailable {
			break
		}
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		ci.indexFile(ctx, index, path, string(content))
		manifest.Files[path] = FileState{
			Mtime:     info.Mtime,
			Size:      info.Size,
			IndexedAt: now,
		}
	}

	if err := index.Save(); err != nil {
		return fmt.Errorf("failed to save vector index: %w", err)
	}

	manifest.LastReconciled = now
	ci.saveManifest(indexDir, manifest)

	if ci.cache != nil {
		ci.cache.Save()
	}

	// Reset cooldown
	ci.mu.Lock()
	ci.lastReconcile[scope] = time.Now()
	ci.mu.Unlock()

	return nil
}

// Status returns the index status
func (ci *ClaudeIndex) Status(scope string) (*IndexStatus, error) {
	indexDir := ci.indexDir(scope)
	manifest := ci.loadManifest(indexDir)

	vecDir := filepath.Join(indexDir, "vectors")
	vectorCount := 0
	if index, err := NewVectorIndexAt(vecDir); err == nil {
		vectorCount = index.Count()
	}

	return &IndexStatus{
		Scope:          scope,
		IndexedFiles:   len(manifest.Files),
		VectorEntries:  vectorCount,
		LastReconciled: manifest.LastReconciled,
		IndexDir:       indexDir,
	}, nil
}

// ─── Internal methods ───────────────────────────────────────

// indexDir returns the index storage directory
func (ci *ClaudeIndex) indexDir(scope string) string {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".luoshu", "file-index")
	if scope == "global" || scope == "" {
		return filepath.Join(base, "global")
	}
	// project scope: use path hash as directory name
	return filepath.Join(base, "projects", encodePath(scope))
}

// encodePath encodes a path into a safe directory name
func encodePath(path string) string {
	h := sha256.Sum256([]byte(path))
	return hex.EncodeToString(h[:8])
}

// scanFiles scans .md files under the target scope
func (ci *ClaudeIndex) scanFiles(scope string) map[string]FileState {
	files := make(map[string]FileState)

	var dirs []string
	if scope == "global" || scope == "" {
		home, _ := os.UserHomeDir()
		dirs = []string{
			filepath.Join(home, ".claude", "memory"),
			filepath.Join(home, ".claude", "rules"),
		}
	} else {
		dirs = []string{
			filepath.Join(scope, ".claude", "memory"),
			filepath.Join(scope, ".claude", "rules"),
		}
		// Official auto-memory: MEMORY.md at project root
		rootMemory := filepath.Join(scope, "MEMORY.md")
		if info, err := os.Stat(rootMemory); err == nil && !info.IsDir() {
			files[rootMemory] = FileState{
				Mtime: float64(info.ModTime().Unix()),
				Size:  info.Size(),
			}
		}
	}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			info, err := e.Info()
			if err != nil {
				continue
			}
			files[path] = FileState{
				Mtime: float64(info.ModTime().Unix()),
				Size:  info.Size(),
			}
		}
	}

	return files
}

// diffFiles compares manifest with current files and returns lists of files to add/remove
func (ci *ClaudeIndex) diffFiles(manifest map[string]FileState, current map[string]FileState) (toAdd, toRemove []string) {
	// New or modified files
	for path, state := range current {
		old, exists := manifest[path]
		if !exists || old.Mtime != state.Mtime || old.Size != state.Size {
			toAdd = append(toAdd, path)
		}
	}
	// Deleted files
	for path := range manifest {
		if _, exists := current[path]; !exists {
			toRemove = append(toRemove, path)
		}
	}
	return
}

// indexFile chunks and vectorizes a single file
func (ci *ClaudeIndex) indexFile(ctx context.Context, index *VectorIndex, path string, content string) {
	fileID := ci.fileID(path)
	chunks := ChunkText(content, 2000)

	for _, chunk := range chunks {
		vec, ok := ci.cache.Get(chunk.Text, ci.model)
		if !ok {
			vectors, err := ci.embedder.Embed(ctx, []string{chunk.Text})
			if err != nil || len(vectors) == 0 {
				continue
			}
			vec = vectors[0]
			ci.cache.Set(chunk.Text, ci.model, vec)
		}
		if vec != nil {
			entry := VectorEntry{
				MemoryID: fileID,
				ChunkID:  chunk.Index,
				Text:     chunk.Text,
				Vector:   vec,
			}
			index.Add([]VectorEntry{entry})
		}
	}
}

// fileID generates a stable ID for a file path
func (ci *ClaudeIndex) fileID(path string) string {
	h := sha256.Sum256([]byte(path))
	return "file-" + hex.EncodeToString(h[:8])
}

// resolveFilePath resolves a file path from a fileID (via manifest lookup)
func (ci *ClaudeIndex) resolveFilePath(fileID string, scope string) string {
	indexDir := ci.indexDir(scope)
	manifest := ci.loadManifest(indexDir)
	for path := range manifest.Files {
		if ci.fileID(path) == fileID {
			return path
		}
	}
	return fileID // fallback
}

// extractContext extracts the context heading from a matched chunk
func (ci *ClaudeIndex) extractContext(text string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") {
			return strings.TrimSpace(strings.TrimLeft(line, "# "))
		}
	}
	if len(lines) > 0 {
		first := strings.TrimSpace(lines[0])
		if len(first) > 80 {
			first = first[:80]
		}
		return first
	}
	return ""
}

// loadManifest loads the manifest file
func (ci *ClaudeIndex) loadManifest(indexDir string) *FileManifest {
	path := filepath.Join(indexDir, "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return &FileManifest{Version: 1, Files: make(map[string]FileState)}
	}
	var m FileManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return &FileManifest{Version: 1, Files: make(map[string]FileState)}
	}
	if m.Files == nil {
		m.Files = make(map[string]FileState)
	}
	return &m
}

// saveManifest saves the manifest file
func (ci *ClaudeIndex) saveManifest(indexDir string, m *FileManifest) error {
	path := filepath.Join(indexDir, "manifest.json")
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
