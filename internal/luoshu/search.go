package luoshu

import (
	"context"
	"sort"
	"strings"
)

// SearchOptions defines search options
type SearchOptions struct {
	ProjectPath string  // Limit to project scope (empty = all)
	MaxResults  int     // Maximum number of results (default 10)
	MinScore    float64 // Minimum similarity threshold (default 0.6)
	Mode        string  // "auto" | "semantic" | "keyword"
}

// Searcher is a hybrid search engine
type Searcher struct {
	store    *MemoryStore
	index    *VectorIndex
	cache    *EmbeddingCache
	embedder EmbeddingProvider
	model    string
}

// NewSearcher creates a hybrid search engine
func NewSearcher(store *MemoryStore, index *VectorIndex, cache *EmbeddingCache, embedder EmbeddingProvider, model string) *Searcher {
	return &Searcher{
		store:    store,
		index:    index,
		cache:    cache,
		embedder: embedder,
		model:    model,
	}
}

// Search performs a hybrid search
// Returns: results, searchMethod("semantic"|"keyword"|"hybrid"), error
func (s *Searcher) Search(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, string, error) {
	if opts.MaxResults <= 0 {
		opts.MaxResults = 10
	}
	if opts.MinScore <= 0 {
		opts.MinScore = 0.6
	}
	if opts.Mode == "" {
		opts.Mode = "auto"
	}

	if opts.Mode == "keyword" {
		return s.keywordSearch(query, opts), "keyword", nil
	}

	isEmbedAvailable := s.embedder.Name() != "noop"

	if opts.Mode == "semantic" {
		if !isEmbedAvailable {
			return s.keywordSearch(query, opts), "keyword", nil
		}
		results, err := s.semanticSearch(ctx, query, opts)
		if err != nil {
			return s.keywordSearch(query, opts), "keyword", nil
		}
		return results, "semantic", nil
	}

	// auto mode: keyword + semantic, merge via RRF
	kwResults := s.keywordSearch(query, opts)
	if !isEmbedAvailable {
		return kwResults, "keyword", nil
	}
	semResults, err := s.semanticSearch(ctx, query, opts)
	if err != nil {
		return kwResults, "keyword", nil
	}
	merged := mergeRRF(semResults, kwResults, opts.MaxResults)
	return merged, "hybrid", nil
}

// keywordSearch performs case-insensitive keyword search
func (s *Searcher) keywordSearch(query string, opts SearchOptions) []SearchResult {
	entries, err := s.store.LoadAll()
	if err != nil || len(entries) == 0 {
		return nil
	}

	queryLower := strings.ToLower(query)
	var results []SearchResult

	for _, entry := range entries {
		if opts.ProjectPath != "" && entry.Source.Project != opts.ProjectPath {
			continue
		}
		contentLower := strings.ToLower(entry.Content)
		if !strings.Contains(contentLower, queryLower) {
			continue
		}
		highlight := extractHighlight(entry.Content, query)
		results = append(results, SearchResult{
			Entry:     entry,
			Score:     keywordScore(contentLower, queryLower),
			MatchType: "keyword",
			Highlight: highlight,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if len(results) > opts.MaxResults {
		results = results[:opts.MaxResults]
	}
	return results
}

// semanticSearch performs semantic search
func (s *Searcher) semanticSearch(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	// Get query vector (check cache first)
	queryVec, ok := s.cache.Get(query, s.model)
	if !ok {
		vectors, err := s.embedder.Embed(ctx, []string{query})
		if err != nil {
			return nil, err
		}
		if len(vectors) == 0 || len(vectors[0]) == 0 {
			return nil, nil
		}
		queryVec = vectors[0]
		s.cache.Set(query, s.model, queryVec)
		_ = s.cache.Save()
	}

	// Vector search
	matches := s.index.Search(queryVec, opts.MaxResults*2, opts.MinScore)
	if len(matches) == 0 {
		return nil, nil
	}

	// Map back to full MemoryEntry by MemoryID
	entries, _ := s.store.LoadAll()
	entryMap := make(map[string]MemoryEntry, len(entries))
	for _, e := range entries {
		entryMap[e.ID] = e
	}

	seen := make(map[string]bool)
	var results []SearchResult
	for _, m := range matches {
		if seen[m.Entry.MemoryID] {
			continue
		}
		seen[m.Entry.MemoryID] = true
		entry, ok := entryMap[m.Entry.MemoryID]
		if !ok {
			continue
		}
		if opts.ProjectPath != "" && entry.Source.Project != opts.ProjectPath {
			continue
		}
		results = append(results, SearchResult{
			Entry:     entry,
			Score:     m.Score,
			MatchType: "semantic",
			Highlight: m.Entry.Text,
		})
	}
	if len(results) > opts.MaxResults {
		results = results[:opts.MaxResults]
	}
	return results, nil
}

// mergeRRF merges two result sets using Reciprocal Rank Fusion
// score(d) = Σ 1/(k + rank_i), k=60
func mergeRRF(semantic, keyword []SearchResult, maxResults int) []SearchResult {
	const k = 60.0

	scores := make(map[string]float64)
	entries := make(map[string]SearchResult)

	for rank, r := range semantic {
		id := r.Entry.ID
		scores[id] += 1.0 / (k + float64(rank+1))
		entries[id] = r
	}
	for rank, r := range keyword {
		id := r.Entry.ID
		scores[id] += 1.0 / (k + float64(rank+1))
		if _, exists := entries[id]; !exists {
			entries[id] = r
		}
	}

	type scored struct {
		id    string
		score float64
	}
	var ranked []scored
	for id, s := range scores {
		ranked = append(ranked, scored{id: id, score: s})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	var results []SearchResult
	for _, r := range ranked {
		if len(results) >= maxResults {
			break
		}
		entry := entries[r.id]
		entry.Score = r.score
		entry.MatchType = "hybrid"
		results = append(results, entry)
	}
	return results
}

// keywordScore computes a simple keyword match score (occurrence count / content length)
func keywordScore(content, query string) float64 {
	count := strings.Count(content, query)
	if count == 0 {
		return 0
	}
	return float64(count) / float64(len(content)+1)
}

// extractHighlight extracts a context snippet containing the query term
func extractHighlight(content, query string) string {
	idx := strings.Index(strings.ToLower(content), strings.ToLower(query))
	if idx < 0 {
		return ""
	}
	start := idx - 50
	if start < 0 {
		start = 0
	}
	end := idx + len(query) + 50
	if end > len(content) {
		end = len(content)
	}
	return content[start:end]
}
