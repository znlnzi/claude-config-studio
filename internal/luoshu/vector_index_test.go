package luoshu

import (
	"math"
	"testing"
)

func newTestIndex(t *testing.T) *VectorIndex {
	t.Helper()
	dir := t.TempDir()
	return &VectorIndex{dir: dir}
}

func TestCosineSimilarity_Identical(t *testing.T) {
	a := []float32{1, 2, 3}
	score := cosineSimilarity(a, a)
	if math.Abs(score-1.0) > 0.001 {
		t.Errorf("expected ~1.0, got %f", score)
	}
}

func TestCosineSimilarity_Orthogonal(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{0, 1}
	score := cosineSimilarity(a, b)
	if math.Abs(score) > 0.001 {
		t.Errorf("expected ~0.0, got %f", score)
	}
}

func TestCosineSimilarity_Opposite(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{-1, 0}
	score := cosineSimilarity(a, b)
	if math.Abs(score+1.0) > 0.001 {
		t.Errorf("expected ~-1.0, got %f", score)
	}
}

func TestCosineSimilarity_DifferentLength(t *testing.T) {
	a := []float32{1, 2}
	b := []float32{1, 2, 3}
	score := cosineSimilarity(a, b)
	if score != 0 {
		t.Errorf("expected 0 for different lengths, got %f", score)
	}
}

func TestCosineSimilarity_Empty(t *testing.T) {
	score := cosineSimilarity(nil, nil)
	if score != 0 {
		t.Errorf("expected 0 for nil vectors, got %f", score)
	}
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	a := []float32{0, 0, 0}
	b := []float32{1, 2, 3}
	score := cosineSimilarity(a, b)
	if score != 0 {
		t.Errorf("expected 0 for zero vector, got %f", score)
	}
}

func TestVectorIndex_AddAndSearch(t *testing.T) {
	idx := newTestIndex(t)

	entries := []VectorEntry{
		{MemoryID: "mem-1", ChunkID: 0, Text: "auth system", Vector: []float32{1, 0, 0}},
		{MemoryID: "mem-2", ChunkID: 0, Text: "cache system", Vector: []float32{0, 1, 0}},
		{MemoryID: "mem-3", ChunkID: 0, Text: "similar auth", Vector: []float32{0.9, 0.1, 0}},
	}
	if err := idx.Add(entries); err != nil {
		t.Fatal(err)
	}

	// Search for vectors similar to [1,0,0]
	matches := idx.Search([]float32{1, 0, 0}, 2, 0.5)
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].Entry.MemoryID != "mem-1" {
		t.Errorf("expected mem-1 as top match, got %s", matches[0].Entry.MemoryID)
	}
}

func TestVectorIndex_Search_MinScore(t *testing.T) {
	idx := newTestIndex(t)

	idx.Add([]VectorEntry{
		{MemoryID: "mem-1", Vector: []float32{1, 0}},
		{MemoryID: "mem-2", Vector: []float32{0, 1}},
	})

	matches := idx.Search([]float32{1, 0}, 10, 0.9)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match with minScore 0.9, got %d", len(matches))
	}
}

func TestVectorIndex_Remove(t *testing.T) {
	idx := newTestIndex(t)
	idx.Add([]VectorEntry{
		{MemoryID: "mem-1", ChunkID: 0, Vector: []float32{1, 0}},
		{MemoryID: "mem-1", ChunkID: 1, Vector: []float32{0.9, 0.1}},
		{MemoryID: "mem-2", ChunkID: 0, Vector: []float32{0, 1}},
	})

	if idx.Count() != 3 {
		t.Fatalf("expected 3, got %d", idx.Count())
	}

	idx.Remove("mem-1")
	if idx.Count() != 1 {
		t.Fatalf("expected 1 after remove, got %d", idx.Count())
	}
}

func TestVectorIndex_Persistence(t *testing.T) {
	dir := t.TempDir()

	// Write
	idx1 := &VectorIndex{dir: dir}
	idx1.Add([]VectorEntry{
		{MemoryID: "mem-1", Vector: []float32{1, 2, 3}},
	})

	// Reload
	idx2 := &VectorIndex{dir: dir}
	idx2.load()

	if idx2.Count() != 1 {
		t.Fatalf("expected 1 entry after reload, got %d", idx2.Count())
	}
}

func TestSortVectorMatches(t *testing.T) {
	matches := []VectorMatch{
		{Score: 0.5},
		{Score: 0.9},
		{Score: 0.7},
	}
	sortVectorMatches(matches)
	if matches[0].Score != 0.9 || matches[1].Score != 0.7 || matches[2].Score != 0.5 {
		t.Errorf("not sorted descending: %v", matches)
	}
}
