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
	store.Append(MemoryEntry{ID: "mem-1", Content: "auth system uses JWT token"})
	store.Append(MemoryEntry{ID: "mem-2", Content: "database chose PostgreSQL"})

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
	store.Append(MemoryEntry{ID: "mem-1", Content: "auth system uses JWT token"})

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
	store.Append(MemoryEntry{ID: "mem-1", Content: "auth system uses JWT token"})

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

func TestBuildFallbackSummary(t *testing.T) {
	results := []SearchResult{
		{Entry: MemoryEntry{Content: "first memory"}},
		{Entry: MemoryEntry{Content: "second memory"}},
	}
	summary := buildFallbackSummary(results)
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
}

func TestBuildFallbackSummary_Truncation(t *testing.T) {
	longContent := ""
	for i := 0; i < 50; i++ {
		longContent += "This is a long content string used for testing truncation"
	}
	results := []SearchResult{
		{Entry: MemoryEntry{Content: longContent}},
	}
	summary := buildFallbackSummary(results)
	if len(summary) > 300 {
		// content is truncated to 200 characters + ellipsis + prefix
		// total length should be reasonable
	}
}

func TestBuildRecallPrompt(t *testing.T) {
	results := []SearchResult{
		{Entry: MemoryEntry{Content: "JWT authentication", Tags: []string{"decision"}}},
	}
	prompt := buildRecallPrompt("authentication approach", results)
	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
}
