package luoshu

import (
	"encoding/gob"
	"math"
	"os"
	"path/filepath"
	"sync"
)

// VectorIndex uses gob storage + brute-force cosine similarity search
type VectorIndex struct {
	mu      sync.RWMutex
	dir     string
	entries []VectorEntry
}

// NewVectorIndex creates a vector index instance (default ~/.luoshu/vectors/)
func NewVectorIndex() (*VectorIndex, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}
	return NewVectorIndexAt(filepath.Join(dir, "vectors"))
}

// NewVectorIndexAt creates a vector index instance at the specified directory
func NewVectorIndexAt(dir string) (*VectorIndex, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	idx := &VectorIndex{dir: dir}
	idx.load()
	return idx, nil
}

// Clear removes all vector entries and persists the change
func (idx *VectorIndex) Clear() error {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.entries = nil
	f, err := os.Create(idx.gobPath())
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(idx.entries)
}

func (idx *VectorIndex) gobPath() string {
	return filepath.Join(idx.dir, "index.gob")
}

func (idx *VectorIndex) load() {
	f, err := os.Open(idx.gobPath())
	if err != nil {
		return
	}
	defer f.Close()
	var entries []VectorEntry
	if err := gob.NewDecoder(f).Decode(&entries); err != nil {
		return
	}
	idx.entries = entries
}

// Save persists the vector index to disk
func (idx *VectorIndex) Save() error {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	f, err := os.Create(idx.gobPath())
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(idx.entries)
}

// Add appends vector entries and persists them
func (idx *VectorIndex) Add(entries []VectorEntry) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.entries = append(idx.entries, entries...)
	f, err := os.Create(idx.gobPath())
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(idx.entries)
}

// Search performs brute-force search for the most similar vectors
func (idx *VectorIndex) Search(queryVector []float32, topK int, minScore float64) []VectorMatch {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var matches []VectorMatch
	for _, entry := range idx.entries {
		score := cosineSimilarity(queryVector, entry.Vector)
		if score >= minScore {
			matches = append(matches, VectorMatch{Entry: entry, Score: score})
		}
	}
	sortVectorMatches(matches)
	if len(matches) > topK {
		matches = matches[:topK]
	}
	return matches
}

// Remove deletes all vectors associated with the given memoryID
func (idx *VectorIndex) Remove(memoryID string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	var filtered []VectorEntry
	for _, e := range idx.entries {
		if e.MemoryID != memoryID {
			filtered = append(filtered, e)
		}
	}
	idx.entries = filtered
	return nil
}

// Count returns the number of vector entries
func (idx *VectorIndex) Count() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.entries)
}

// cosineSimilarity computes cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// sortVectorMatches sorts matches by Score in descending order (insertion sort, small data set)
func sortVectorMatches(matches []VectorMatch) {
	for i := 1; i < len(matches); i++ {
		j := i
		for j > 0 && matches[j].Score > matches[j-1].Score {
			matches[j], matches[j-1] = matches[j-1], matches[j]
			j--
		}
	}
}
