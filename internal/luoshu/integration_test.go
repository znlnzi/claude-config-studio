package luoshu

import (
	"context"
	"os"
	"testing"
	"time"
)

// Integration tests require a real API Key, provided via environment variable or ~/.luoshu/config.json.
// How to run:
//   LUOSHU_LLM_API_KEY=xxx go test -v -run=TestIntegration -timeout=120s ./internal/luoshu/
//   Or configure ~/.luoshu/config.json and run directly.

func loadTestConfig(t *testing.T) *Config {
	t.Helper()

	cfg := DefaultConfig()

	// Prefer environment variables
	if key := os.Getenv("LUOSHU_LLM_API_KEY"); key != "" {
		cfg.LLM.APIKey = key
		cfg.Embedding.APIKey = key
		return cfg
	}

	// Fall back to ~/.luoshu/config.json
	loaded, err := Load()
	if err != nil {
		t.Skipf("unable to load config: %v", err)
	}
	if loaded.LLM.APIKey == "" {
		t.Skip("skipping integration test: API Key not configured (set LUOSHU_LLM_API_KEY or ~/.luoshu/config.json)")
	}
	return loaded
}

// ─── TestConnection ─────────────────────────────────

func TestIntegration_TestConnection(t *testing.T) {
	cfg := loadTestConfig(t)

	connected, status, err := TestConnection(cfg)
	if !connected {
		t.Fatalf("connection failed: status=%s, err=%v", status, err)
	}
	if status != "ok" {
		t.Errorf("expected status=ok, got %s", status)
	}
}

func TestIntegration_TestConnection_InvalidKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LLM.APIKey = "invalid-key-for-testing"

	connected, status, _ := TestConnection(cfg)
	if connected {
		t.Error("should not connect successfully with invalid key")
	}
	if status != "auth_failed" {
		t.Errorf("expected status=auth_failed, got %s", status)
	}
}

// ─── Chat (LLM) ────────────────────────────────────

func TestIntegration_Chat_Simple(t *testing.T) {
	cfg := loadTestConfig(t)
	llm, _ := NewProviders(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := llm.Chat(ctx, ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "Answer in one word: what is 1+1?"},
		},
		MaxTokens:   10,
		Temperature: 0,
	})
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp.Content == "" {
		t.Error("Chat returned empty content")
	}
	if resp.TokensUsed <= 0 {
		t.Errorf("TokensUsed should be > 0, got %d", resp.TokensUsed)
	}
	t.Logf("Chat response: %q (tokens: %d)", resp.Content, resp.TokensUsed)
}

func TestIntegration_Chat_SystemPrompt(t *testing.T) {
	cfg := loadTestConfig(t)
	llm, _ := NewProviders(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := llm.Chat(ctx, ChatRequest{
		Messages: []Message{
			{Role: "system", Content: "You are an assistant that only responds with numbers. Do not respond with any text."},
			{Role: "user", Content: "2+3=?"},
		},
		MaxTokens:   10,
		Temperature: 0,
	})
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if resp.Content == "" {
		t.Error("Chat returned empty content")
	}
	t.Logf("Chat response: %q", resp.Content)
}

// ─── Embed (Embedding) ─────────────────────────────

func TestIntegration_Embed_Single(t *testing.T) {
	cfg := loadTestConfig(t)
	_, embed := NewProviders(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	vectors, err := embed.Embed(ctx, []string{"hello world"})
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(vectors) != 1 {
		t.Fatalf("expected 1 vector, got %d", len(vectors))
	}
	if len(vectors[0]) != cfg.Embedding.Dimensions {
		t.Errorf("expected dimensions %d, got %d", cfg.Embedding.Dimensions, len(vectors[0]))
	}
	t.Logf("vector dimensions: %d, first 3 values: %v", len(vectors[0]), vectors[0][:3])
}

func TestIntegration_Embed_Batch(t *testing.T) {
	cfg := loadTestConfig(t)
	_, embed := NewProviders(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	texts := []string{"apple is a fruit", "programming is a skill", "the weather is nice today"}
	vectors, err := embed.Embed(ctx, texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(vectors) != len(texts) {
		t.Fatalf("expected %d vectors, got %d", len(texts), len(vectors))
	}
	for i, v := range vectors {
		if len(v) != cfg.Embedding.Dimensions {
			t.Errorf("vector[%d] dimensions %d != %d", i, len(v), cfg.Embedding.Dimensions)
		}
	}
}

func TestIntegration_Embed_Empty(t *testing.T) {
	cfg := loadTestConfig(t)
	_, embed := NewProviders(cfg)

	ctx := context.Background()
	vectors, err := embed.Embed(ctx, []string{})
	if err != nil {
		t.Fatalf("empty input should not error: %v", err)
	}
	if vectors != nil {
		t.Errorf("empty input should return nil, got %d vectors", len(vectors))
	}
}

// ─── Semantic Similarity End-to-End ──────────────────

func TestIntegration_SemanticSimilarity(t *testing.T) {
	cfg := loadTestConfig(t)
	_, embed := NewProviders(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	texts := []string{
		"Go 语言的错误处理机制",     // 0: programming
		"Golang 的 error 处理方式", // 1: programming (semantically similar)
		"今天的晚餐吃什么",         // 2: daily life (unrelated)
	}

	vectors, err := embed.Embed(ctx, texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	sim01 := cosineSimilarity(vectors[0], vectors[1]) // programming vs programming
	sim02 := cosineSimilarity(vectors[0], vectors[2]) // programming vs daily life

	t.Logf("programming vs programming: %.4f", sim01)
	t.Logf("programming vs daily life: %.4f", sim02)

	if sim01 <= sim02 {
		t.Errorf("semantically similar texts similarity (%.4f) should be higher than unrelated texts (%.4f)", sim01, sim02)
	}
	if sim01 < 0.5 {
		t.Errorf("semantically similar texts similarity should be >= 0.5, got %.4f", sim01)
	}
}

// ─── VectorIndex End-to-End ──────────────────────────

func TestIntegration_VectorIndex_EndToEnd(t *testing.T) {
	cfg := loadTestConfig(t)
	_, embed := NewProviders(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Prepare test data (Chinese text as intentional test input for embedding API)
	entries := []struct {
		id   string
		text string
	}{
		{"mem-001", "项目使用 React 18 和 TypeScript 开发前端"},
		{"mem-002", "后端 API 采用 Go 语言的 Gin 框架"},
		{"mem-003", "数据库选择了 PostgreSQL"},
		{"mem-004", "用户认证使用 JWT Token"},
		{"mem-005", "今天中午吃了麻辣烫"},
	}

	// Embedding
	texts := make([]string, len(entries))
	for i, e := range entries {
		texts[i] = e.text
	}
	vectors, err := embed.Embed(ctx, texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	// Build index
	idx := &VectorIndex{dir: t.TempDir()}
	for i, e := range entries {
		if err := idx.Add([]VectorEntry{{MemoryID: e.id, ChunkID: 0, Text: e.text, Vector: vectors[i]}}); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Query: search for "frontend tech stack" (Chinese test input for embedding API)
	queryVec, err := embed.Embed(ctx, []string{"前端技术栈是什么"})
	if err != nil {
		t.Fatalf("query Embed failed: %v", err)
	}

	results := idx.Search(queryVec[0], 3, 0)
	if len(results) == 0 {
		t.Fatal("search results are empty")
	}

	t.Logf("query: '前端技术栈是什么'")
	for _, r := range results {
		t.Logf("  %s (score: %.4f)", r.Entry.MemoryID, r.Score)
	}

	// The most relevant result should be the React/TypeScript entry
	if results[0].Entry.MemoryID != "mem-001" {
		t.Errorf("expected most relevant result to be mem-001 (React/TS), got %s", results[0].Entry.MemoryID)
	}
}

// ─── Recall End-to-End ───────────────────────────────

func TestIntegration_Recall_EndToEnd(t *testing.T) {
	cfg := loadTestConfig(t)
	llm, embed := NewProviders(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Prepare memory data and write to store (Chinese text as intentional test input)
	tmpDir := t.TempDir()
	store := &MemoryStore{dir: tmpDir}
	now := time.Now().Format(time.RFC3339)

	memories := []MemoryEntry{
		{ID: "mem-r1", Content: "Decided to use JWT for user authentication because stateless tokens are better suited for decoupled frontend/backend architecture", Tags: []string{"auth", "decision"}, CreatedAt: now, UpdatedAt: now},
		{ID: "mem-r2", Content: "Chose PostgreSQL for database considering future needs for JSONB support and full-text search", Tags: []string{"database", "decision"}, CreatedAt: now, UpdatedAt: now},
		{ID: "mem-r3", Content: "API endpoints use RESTful style uniformly with versioned paths /api/v1/", Tags: []string{"api", "decision"}, CreatedAt: now, UpdatedAt: now},
	}

	for _, m := range memories {
		if err := store.Append(m); err != nil {
			t.Fatalf("Append failed: %v", err)
		}
	}

	// Embedding + build index
	texts := make([]string, len(memories))
	for i, m := range memories {
		texts[i] = m.Content
	}
	vectors, err := embed.Embed(ctx, texts)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	idx := &VectorIndex{dir: tmpDir}
	for i, m := range memories {
		if err := idx.Add([]VectorEntry{{MemoryID: m.ID, ChunkID: 0, Text: m.Content, Vector: vectors[i]}}); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Build Searcher → Recaller
	cache := &EmbeddingCache{dir: tmpDir, cache: make(map[string][]float32)}
	searcher := NewSearcher(store, idx, cache, embed, cfg.Embedding.Model)
	recaller := NewRecaller(searcher, llm)

	result, err := recaller.Recall(ctx, "JWT", SearchOptions{MaxResults: 5, Mode: "semantic"})
	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}

	if result.Summary == "" {
		t.Error("Recall returned empty answer")
	}

	t.Logf("Recall answer: %s", result.Summary)
	t.Logf("search method: %s, source count: %d", result.SearchMethod, result.SourceCount)

	if result.SourceCount == 0 {
		t.Error("expected to find at least 1 source")
	}
}

// ─── Extractor End-to-End ────────────────────────────

func TestIntegration_Extract_EndToEnd(t *testing.T) {
	cfg := loadTestConfig(t)
	llm, _ := NewProviders(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	store := &MemoryStore{dir: t.TempDir()}
	extractor := &Extractor{llm: llm, store: store}

	conversation := `User: What authentication method should we use for our API?
Assistant: I suggest using JWT because the project has a decoupled frontend/backend architecture, and JWT's stateless nature is a good fit.
User: OK, what about the database?
Assistant: I recommend PostgreSQL. It supports JSONB and full-text search, and has good future extensibility.
User: Let's go with that.`

	entries, err := extractor.Extract(ctx, conversation, "/tmp/test-project", []string{})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	t.Logf("extracted %d memory entries", len(entries))
	for _, e := range entries {
		t.Logf("  [%s] %s (tags: %v)", e.ID, truncate(e.Content, 60), e.Tags)
	}

	if len(entries) == 0 {
		t.Error("should extract at least 1 memory entry")
	}

	// Verify persistence
	all, _ := store.LoadAll()
	if len(all) != len(entries) {
		t.Errorf("persisted entry count %d != extracted entry count %d", len(all), len(entries))
	}
}

// ─── Helper Functions ────────────────────────────────

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

