package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/znlnzi/claude-config-studio/internal/luoshu"

	"github.com/mark3labs/mcp-go/mcp"
)

// ─── luoshu_status tool ──────────────────────────────

func buildLuoshuStatusTool() mcp.Tool {
	return mcp.NewTool(
		"luoshu_status",
		mcp.WithDescription("Return the overall luoshu system status: configuration state, memory statistics, and index status."),
	)
}

func handleLuoshuStatus(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg, err := luoshu.Load()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load config: %v", err)), nil
	}

	llmConfigured := cfg.LLM.APIKey != ""
	embedConfigured := cfg.Embedding.APIKey != ""
	if !embedConfigured && llmConfigured {
		embedConfigured = true
	}

	memoryEntries := 0
	if store, err := luoshu.NewMemoryStore(); err == nil {
		memoryEntries, _ = store.Count()
	}

	vectorEntries := 0
	if index, err := luoshu.NewVectorIndex(); err == nil {
		vectorEntries = index.Count()
	}

	cacheEntries := 0
	if cache, err := luoshu.NewEmbeddingCache(); err == nil {
		cacheEntries = cache.Count()
	}

	result := map[string]interface{}{
		"version":              serverVersion,
		"llm_configured":       llmConfigured,
		"embedding_configured": embedConfigured,
		"memory_entries":       memoryEntries,
		"vector_index_entries": vectorEntries,
		"embedding_cache_size": cacheEntries,
		"config_path":          "~/.luoshu/config.json",
		"features": map[string]bool{
			"semantic_search": embedConfigured,
			"auto_extract":    llmConfigured && cfg.Memory.AutoExtract,
			"keyword_search":  true,
		},
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ─── luoshu_reindex tool ─────────────────────────────

func buildLuoshuReindexTool() mcp.Tool {
	return mcp.NewTool(
		"luoshu_reindex",
		mcp.WithDescription("Rebuild the vector index. Iterates over all memory entries and regenerates embeddings. Requires Embedding configuration."),
	)
}

func handleLuoshuReindex(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg, err := luoshu.Load()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load config: %v", err)), nil
	}

	_, embedder := luoshu.NewProviders(cfg)
	if embedder.Name() == "noop" {
		return mcp.NewToolResultError("Embedding not configured. Please run /luoshu.config to set the API Key first."), nil
	}

	store, err := luoshu.NewMemoryStore()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to initialize store: %v", err)), nil
	}

	entries, err := store.LoadAll()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load memories: %v", err)), nil
	}

	if len(entries) == 0 {
		return mcp.NewToolResultText(`{"reindexed": 0, "message": "No memory entries to index"}`), nil
	}

	index, err := luoshu.NewVectorIndex()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to initialize index: %v", err)), nil
	}

	cache, err := luoshu.NewEmbeddingCache()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to initialize cache: %v", err)), nil
	}

	start := time.Now()
	indexed := 0
	model := cfg.Embedding.Model

	for _, entry := range entries {
		chunks := luoshu.ChunkText(entry.Content, 2000)
		for _, chunk := range chunks {
			vec, ok := cache.Get(chunk.Text, model)
			if !ok {
				vectors, err := embedder.Embed(ctx, []string{chunk.Text})
				if err != nil {
					continue
				}
				if len(vectors) > 0 {
					vec = vectors[0]
					cache.Set(chunk.Text, model, vec)
				}
			}
			if vec != nil {
				ve := luoshu.VectorEntry{
					MemoryID: entry.ID,
					ChunkID:  chunk.Index,
					Text:     chunk.Text,
					Vector:   vec,
				}
				index.Add([]luoshu.VectorEntry{ve})
				indexed++
			}
		}
	}

	cache.Save()

	duration := time.Since(start).Milliseconds()
	result := map[string]interface{}{
		"reindexed":   indexed,
		"total":       len(entries),
		"duration_ms": duration,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
