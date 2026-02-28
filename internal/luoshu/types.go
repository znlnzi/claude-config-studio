package luoshu

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
	ID        string            `json:"id"`
	Content   string            `json:"content"`
	Summary   string            `json:"summary,omitempty"`
	Source    MemorySource      `json:"source"`
	Tags      []string          `json:"tags,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

// MemorySource represents memory source information
type MemorySource struct {
	Type      string `json:"type"`       // "auto_extract" | "manual" | "session_state"
	Project   string `json:"project"`
	SessionID string `json:"session_id,omitempty"`
	Filename  string `json:"filename,omitempty"`
}

// VectorEntry represents a vector index entry
type VectorEntry struct {
	MemoryID string    `json:"memory_id"`
	ChunkID  int       `json:"chunk_id"`
	Text     string    `json:"text"`
	Vector   []float32 `json:"-"` // stored separately via gob
}

// SearchResult represents a search result
type SearchResult struct {
	Entry     MemoryEntry `json:"entry"`
	Score     float64     `json:"score"`
	MatchType string      `json:"match_type"` // "semantic" | "keyword" | "hybrid"
	Highlight string      `json:"highlight,omitempty"`
}

// VectorMatch represents a vector search match
type VectorMatch struct {
	Entry VectorEntry
	Score float64
}

// NewMemoryID generates a unique memory ID
// Format: mem-20260218-150405-a1b2c3d4
func NewMemoryID() string {
	now := time.Now()
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("mem-%s-%s", now.Format("20060102-150405"), hex.EncodeToString(b))
}
