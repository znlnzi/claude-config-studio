package luoshu

import "testing"

func newTestCache(t *testing.T) *EmbeddingCache {
	t.Helper()
	return &EmbeddingCache{dir: t.TempDir(), cache: make(map[string][]float32)}
}

func TestEmbeddingCache_SetAndGet(t *testing.T) {
	cache := newTestCache(t)

	vec := []float32{1, 2, 3}
	cache.Set("hello", "model-a", vec)

	got, ok := cache.Get("hello", "model-a")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != 3 || got[0] != 1 {
		t.Errorf("unexpected vector: %v", got)
	}
}

func TestEmbeddingCache_Miss(t *testing.T) {
	cache := newTestCache(t)

	_, ok := cache.Get("nonexistent", "model")
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestEmbeddingCache_DifferentModel(t *testing.T) {
	cache := newTestCache(t)

	cache.Set("hello", "model-a", []float32{1, 0})
	cache.Set("hello", "model-b", []float32{0, 1})

	va, _ := cache.Get("hello", "model-a")
	vb, _ := cache.Get("hello", "model-b")

	if va[0] != 1 || vb[0] != 0 {
		t.Errorf("different models should have different vectors")
	}
}

func TestEmbeddingCache_Count(t *testing.T) {
	cache := newTestCache(t)

	if cache.Count() != 0 {
		t.Fatalf("expected 0, got %d", cache.Count())
	}

	cache.Set("a", "m", []float32{1})
	cache.Set("b", "m", []float32{2})

	if cache.Count() != 2 {
		t.Fatalf("expected 2, got %d", cache.Count())
	}
}

func TestEmbeddingCache_Persistence(t *testing.T) {
	dir := t.TempDir()

	// Write
	c1 := &EmbeddingCache{dir: dir, cache: make(map[string][]float32)}
	c1.Set("test", "model", []float32{1, 2, 3})
	if err := c1.Save(); err != nil {
		t.Fatal(err)
	}

	// Reload
	c2 := &EmbeddingCache{dir: dir, cache: make(map[string][]float32)}
	c2.load()

	got, ok := c2.Get("test", "model")
	if !ok {
		t.Fatal("expected cache hit after reload")
	}
	if len(got) != 3 || got[0] != 1 {
		t.Errorf("unexpected vector after reload: %v", got)
	}
}

func TestCacheKey_Deterministic(t *testing.T) {
	k1 := cacheKey("hello", "model")
	k2 := cacheKey("hello", "model")
	if k1 != k2 {
		t.Error("same input should produce same key")
	}

	k3 := cacheKey("hello", "other-model")
	if k1 == k3 {
		t.Error("different model should produce different key")
	}
}
