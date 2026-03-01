package luoshu

import (
	"context"
	"testing"
)

// mockLLM simulates an LLM Provider
type mockLLM struct {
	response string
	err      error
}

func (m *mockLLM) Chat(_ context.Context, _ ChatRequest) (*ChatResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ChatResponse{Content: m.response}, nil
}

func (m *mockLLM) Name() string { return "mock" }

// mockEmbedder simulates an Embedding Provider
type mockEmbedder struct {
	vectors [][]float32
	err     error
}

func (m *mockEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.vectors, nil
}

func (m *mockEmbedder) Dimensions() int { return 3 }
func (m *mockEmbedder) Name() string    { return "mock" }

func TestRecall_WithLLM(t *testing.T) {
	store := newTestStore(t)
	if err := store.Append(MemoryEntry{ID: "mem-1", Content: "auth system uses JWT token"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(MemoryEntry{ID: "mem-2", Content: "database chose PostgreSQL"}); err != nil {
		t.Fatal(err)
	}

	idx := newTestIndex(t)
	cache := &EmbeddingCache{dir: t.TempDir()}

	embedder := &mockEmbedder{vectors: [][]float32{{1, 0, 0}}}
	llm := &mockLLM{response: "Auth system uses JWT, database uses PostgreSQL."}

	searcher := NewSearcher(store, idx, cache, embedder, "test-model")
	recaller := NewRecaller(searcher, llm)

	result, err := recaller.Recall(context.Background(), "auth", SearchOptions{MaxResults: 5, Mode: "keyword"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary != "Auth system uses JWT, database uses PostgreSQL." {
		t.Errorf("unexpected summary: %s", result.Summary)
	}
	if result.SourceCount != 1 {
		t.Errorf("expected 1 source, got %d", result.SourceCount)
	}
}

func TestRecall_WithoutLLM(t *testing.T) {
	store := newTestStore(t)
	if err := store.Append(MemoryEntry{ID: "mem-1", Content: "auth system uses JWT token"}); err != nil {
		t.Fatal(err)
	}

	idx := newTestIndex(t)
	cache := &EmbeddingCache{dir: t.TempDir()}
	embedder := &NoopEmbeddingProvider{}
	llm := &NoopLLMProvider{}

	searcher := NewSearcher(store, idx, cache, embedder, "")
	recaller := NewRecaller(searcher, llm)

	result, err := recaller.Recall(context.Background(), "auth", SearchOptions{MaxResults: 5, Mode: "keyword"})
	if err != nil {
		t.Fatal(err)
	}
	// Without LLM, should use fallback summary
	if result.SourceCount != 1 {
		t.Errorf("expected 1 source, got %d", result.SourceCount)
	}
	if result.SearchMethod != "keyword" {
		t.Errorf("expected keyword search method, got %s", result.SearchMethod)
	}
}

func TestRecall_NoResults(t *testing.T) {
	store := newTestStore(t)
	idx := newTestIndex(t)
	cache := &EmbeddingCache{dir: t.TempDir()}
	embedder := &NoopEmbeddingProvider{}
	llm := &NoopLLMProvider{}

	searcher := NewSearcher(store, idx, cache, embedder, "")
	recaller := NewRecaller(searcher, llm)

	result, err := recaller.Recall(context.Background(), "nonexistent content", SearchOptions{MaxResults: 5})
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceCount != 0 {
		t.Errorf("expected 0 sources, got %d", result.SourceCount)
	}
	if result.Summary != "No related memories found." {
		t.Errorf("unexpected summary: %s", result.Summary)
	}
}

func TestRecall_LLMError_FallsBack(t *testing.T) {
	store := newTestStore(t)
	if err := store.Append(MemoryEntry{ID: "mem-1", Content: "auth system uses JWT token"}); err != nil {
		t.Fatal(err)
	}

	idx := newTestIndex(t)
	cache := &EmbeddingCache{dir: t.TempDir()}
	embedder := &NoopEmbeddingProvider{}
	llm := &mockLLM{err: ErrNotConfigured}

	searcher := NewSearcher(store, idx, cache, embedder, "")
	recaller := NewRecaller(searcher, llm)

	result, err := recaller.Recall(context.Background(), "auth", SearchOptions{MaxResults: 5, Mode: "keyword"})
	if err != nil {
		t.Fatal(err)
	}
	// LLM failure should fall back to fallback summary
	if result.SourceCount != 1 {
		t.Errorf("expected 1 source, got %d", result.SourceCount)
	}
}

func TestRecall_WithFileIndex_NilSafe(t *testing.T) {
	store := newTestStore(t)
	if err := store.Append(MemoryEntry{ID: "mem-1", Content: "auth system uses JWT token"}); err != nil {
		t.Fatal(err)
	}

	idx := newTestIndex(t)
	cache := &EmbeddingCache{dir: t.TempDir()}
	embedder := &NoopEmbeddingProvider{}
	llm := &NoopLLMProvider{}

	searcher := NewSearcher(store, idx, cache, embedder, "")
	recaller := NewRecaller(searcher, llm)
	// fileIndex is nil — should not panic
	result, err := recaller.Recall(context.Background(), "auth", SearchOptions{MaxResults: 5, Mode: "keyword"})
	if err != nil {
		t.Fatal(err)
	}
	if result.FileSources != nil {
		t.Errorf("expected nil FileSources when fileIndex is nil, got %v", result.FileSources)
	}
}

func TestRecall_WithFileIndex_Attached(t *testing.T) {
	store := newTestStore(t)
	if err := store.Append(MemoryEntry{ID: "mem-1", Content: "auth system uses JWT token"}); err != nil {
		t.Fatal(err)
	}

	idx := newTestIndex(t)
	cache := &EmbeddingCache{dir: t.TempDir()}
	embedder := &NoopEmbeddingProvider{}
	llm := &NoopLLMProvider{}

	searcher := NewSearcher(store, idx, cache, embedder, "")

	// Create a ClaudeIndex with noop embedder (won't actually search but exercises the code path)
	fileIdx := NewClaudeIndex(embedder, cache, "")

	recaller := NewRecaller(searcher, llm).WithFileIndex(fileIdx)
	result, err := recaller.Recall(context.Background(), "auth", SearchOptions{MaxResults: 5, Mode: "keyword"})
	if err != nil {
		t.Fatal(err)
	}
	// Noop embedder means file search returns empty (error), but JSONL search should still work
	if result.SourceCount < 1 {
		t.Errorf("expected at least 1 source from JSONL, got %d", result.SourceCount)
	}
}

func TestBuildUnifiedFallbackSummary(t *testing.T) {
	memResults := []SearchResult{
		{Entry: MemoryEntry{Content: "first memory"}},
		{Entry: MemoryEntry{Content: "second memory"}},
	}
	summary := buildUnifiedFallbackSummary(memResults, nil)
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
}

func TestBuildUnifiedFallbackSummary_WithFiles(t *testing.T) {
	memResults := []SearchResult{
		{Entry: MemoryEntry{Content: "memory entry"}},
	}
	fileResults := []FileSearchResult{
		{FilePath: "/test/MEMORY.md", Context: "some context", Content: "file content"},
	}
	summary := buildUnifiedFallbackSummary(memResults, fileResults)
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
}

func TestBuildUnifiedFallbackSummary_Truncation(t *testing.T) {
	longContent := ""
	for i := 0; i < 50; i++ {
		longContent += "This is a long content string used for testing truncation"
	}
	results := []SearchResult{
		{Entry: MemoryEntry{Content: longContent}},
	}
	summary := buildUnifiedFallbackSummary(results, nil)
	if len(summary) > 300 {
		t.Errorf("expected summary to be truncated, got length %d", len(summary))
	}
}

func TestBuildUnifiedRecallPrompt(t *testing.T) {
	memResults := []SearchResult{
		{Entry: MemoryEntry{Content: "JWT authentication", Tags: []string{"decision"}}},
	}
	prompt := buildUnifiedRecallPrompt("authentication approach", memResults, nil)
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
}

func TestBuildUnifiedRecallPrompt_WithFiles(t *testing.T) {
	memResults := []SearchResult{
		{Entry: MemoryEntry{Content: "JWT authentication"}},
	}
	fileResults := []FileSearchResult{
		{FilePath: "/test/rules/auth.md", Content: "Use OAuth2 for external APIs"},
	}
	prompt := buildUnifiedRecallPrompt("auth approach", memResults, fileResults)
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
}
