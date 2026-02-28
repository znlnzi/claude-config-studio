package luoshu

import "testing"

func TestKeywordScore(t *testing.T) {
	score := keywordScore("hello world hello", "hello")
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}

	score2 := keywordScore("no match here", "xyz")
	if score2 != 0 {
		t.Errorf("expected 0 for no match, got %f", score2)
	}
}

func TestExtractHighlight(t *testing.T) {
	content := "This is a long text content that contains the keyword test in the middle, followed by more content for padding."
	highlight := extractHighlight(content, "keyword")
	if highlight == "" {
		t.Fatal("expected non-empty highlight")
	}
}

func TestExtractHighlight_NotFound(t *testing.T) {
	highlight := extractHighlight("hello world", "xyz")
	if highlight != "" {
		t.Errorf("expected empty highlight, got %q", highlight)
	}
}

func TestExtractHighlight_CaseInsensitive(t *testing.T) {
	highlight := extractHighlight("Hello World Test", "hello")
	if highlight == "" {
		t.Fatal("expected non-empty highlight for case-insensitive match")
	}
}

func TestMergeRRF_BothEmpty(t *testing.T) {
	result := mergeRRF(nil, nil, 10)
	if len(result) != 0 {
		t.Fatalf("expected 0, got %d", len(result))
	}
}

func TestMergeRRF_SemanticOnly(t *testing.T) {
	semantic := []SearchResult{
		{Entry: MemoryEntry{ID: "mem-1"}, Score: 0.9},
		{Entry: MemoryEntry{ID: "mem-2"}, Score: 0.8},
	}
	result := mergeRRF(semantic, nil, 10)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestMergeRRF_Overlap(t *testing.T) {
	semantic := []SearchResult{
		{Entry: MemoryEntry{ID: "mem-1"}, Score: 0.9},
		{Entry: MemoryEntry{ID: "mem-2"}, Score: 0.8},
	}
	keyword := []SearchResult{
		{Entry: MemoryEntry{ID: "mem-1"}, Score: 0.7},
		{Entry: MemoryEntry{ID: "mem-3"}, Score: 0.5},
	}
	result := mergeRRF(semantic, keyword, 10)
	if len(result) != 3 {
		t.Fatalf("expected 3 unique results, got %d", len(result))
	}
	// mem-1 should be first (present in both lists)
	if result[0].Entry.ID != "mem-1" {
		t.Errorf("expected mem-1 as top result (in both lists), got %s", result[0].Entry.ID)
	}
	if result[0].MatchType != "hybrid" {
		t.Errorf("expected hybrid match type, got %s", result[0].MatchType)
	}
}

func TestMergeRRF_MaxResults(t *testing.T) {
	semantic := []SearchResult{
		{Entry: MemoryEntry{ID: "mem-1"}},
		{Entry: MemoryEntry{ID: "mem-2"}},
		{Entry: MemoryEntry{ID: "mem-3"}},
	}
	result := mergeRRF(semantic, nil, 2)
	if len(result) != 2 {
		t.Fatalf("expected 2 (maxResults), got %d", len(result))
	}
}

func TestSearcher_KeywordSearch(t *testing.T) {
	store := newTestStore(t)
	store.Append(MemoryEntry{ID: "mem-1", Content: "auth system uses JWT"})
	store.Append(MemoryEntry{ID: "mem-2", Content: "database uses PostgreSQL"})
	store.Append(MemoryEntry{ID: "mem-3", Content: "cache uses Redis"})

	idx := newTestIndex(t)
	cache := &EmbeddingCache{dir: t.TempDir()}
	embedder := &NoopEmbeddingProvider{}

	searcher := NewSearcher(store, idx, cache, embedder, "")
	results, method, err := searcher.Search(nil, "auth", SearchOptions{MaxResults: 10, Mode: "keyword"})
	if err != nil {
		t.Fatal(err)
	}
	if method != "keyword" {
		t.Errorf("expected keyword method, got %s", method)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Entry.ID != "mem-1" {
		t.Errorf("expected mem-1, got %s", results[0].Entry.ID)
	}
}

func TestSearcher_AutoMode_NoEmbedder(t *testing.T) {
	store := newTestStore(t)
	store.Append(MemoryEntry{ID: "mem-1", Content: "test keyword search"})

	idx := newTestIndex(t)
	cache := &EmbeddingCache{dir: t.TempDir()}
	embedder := &NoopEmbeddingProvider{}

	searcher := NewSearcher(store, idx, cache, embedder, "")
	results, method, err := searcher.Search(nil, "keyword", SearchOptions{MaxResults: 10, Mode: "auto"})
	if err != nil {
		t.Fatal(err)
	}
	if method != "keyword" {
		t.Errorf("auto mode without embedder should fallback to keyword, got %s", method)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearcher_ProjectFilter(t *testing.T) {
	store := newTestStore(t)
	store.Append(MemoryEntry{ID: "mem-1", Content: "auth project-a", Source: MemorySource{Project: "/project-a"}})
	store.Append(MemoryEntry{ID: "mem-2", Content: "auth project-b", Source: MemorySource{Project: "/project-b"}})

	idx := newTestIndex(t)
	cache := &EmbeddingCache{dir: t.TempDir()}
	embedder := &NoopEmbeddingProvider{}

	searcher := NewSearcher(store, idx, cache, embedder, "")
	results, _, _ := searcher.Search(nil, "auth", SearchOptions{
		MaxResults:  10,
		Mode:        "keyword",
		ProjectPath: "/project-a",
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 filtered result, got %d", len(results))
	}
	if results[0].Entry.ID != "mem-1" {
		t.Errorf("expected mem-1, got %s", results[0].Entry.ID)
	}
}
