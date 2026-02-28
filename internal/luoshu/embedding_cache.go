package luoshu

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
)

// EmbeddingCache caches embedding results, SHA-256(text+model) -> vector
type EmbeddingCache struct {
	mu    sync.RWMutex
	dir   string
	cache map[string][]float32
}

// NewEmbeddingCache creates an embedding cache instance
func NewEmbeddingCache() (*EmbeddingCache, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}
	cacheDir := filepath.Join(dir, "cache")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, err
	}
	c := &EmbeddingCache{dir: cacheDir, cache: make(map[string][]float32)}
	c.load()
	return c, nil
}

func (c *EmbeddingCache) gobPath() string {
	return filepath.Join(c.dir, "embeddings.gob")
}

func (c *EmbeddingCache) load() {
	f, err := os.Open(c.gobPath())
	if err != nil {
		return
	}
	defer f.Close()
	gob.NewDecoder(f).Decode(&c.cache)
}

// Get retrieves from cache
func (c *EmbeddingCache) Get(text, model string) ([]float32, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.cache[cacheKey(text, model)]
	return v, ok
}

// Set writes to cache
func (c *EmbeddingCache) Set(text, model string, vector []float32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[cacheKey(text, model)] = vector
}

// Save persists the cache to disk
func (c *EmbeddingCache) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	f, err := os.Create(c.gobPath())
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(c.cache)
}

// Count returns the number of cached entries
func (c *EmbeddingCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

func cacheKey(text, model string) string {
	h := sha256.Sum256([]byte(text + "|" + model))
	return hex.EncodeToString(h[:])
}
