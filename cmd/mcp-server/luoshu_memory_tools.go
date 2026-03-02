package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/znlnzi/claude-config-studio/internal/luoshu"

	"github.com/mark3labs/mcp-go/mcp"
)

// ─── memory_extract tool ─────────────────────────────

func buildMemoryExtractTool() mcp.Tool {
	return mcp.NewTool(
		"memory_extract",
		mcp.WithDescription("Extract key points from a session summary and save them as memory entries. Requires LLM configuration. Recommended to call proactively during PreCompact (before context compression) and at session end to avoid losing critical information."),
		mcp.WithString("session_summary",
			mcp.Required(),
			mcp.Description("Session summary text (organized by Claude)"),
		),
		mcp.WithString("project_path",
			mcp.Description("Associated project path"),
		),
		mcp.WithString("tags",
			mcp.Description("Additional tags, comma-separated"),
		),
	)
}

func handleMemoryExtract(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	summary, err := req.RequireString("session_summary")
	if err != nil {
		return mcp.NewToolResultError("session_summary parameter is required"), nil
	}
	projectPath := req.GetString("project_path", "")
	tagsStr := req.GetString("tags", "")

	var tags []string
	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	cfg, err := luoshu.Load()
	if err != nil {
		return mcp.NewToolResultError(errConfigLoad(err)), nil
	}

	llm, _ := luoshu.NewProviders(cfg)
	if llm.Name() == "noop" {
		return mcp.NewToolResultError("LLM not configured. Please run /luoshu.config to set the API Key first."), nil
	}

	store, err := luoshu.NewMemoryStore()
	if err != nil {
		return mcp.NewToolResultError(errInitFailed("memory store", err)), nil
	}

	extractor := luoshu.NewExtractor(llm, store)
	entries, err := extractor.Extract(ctx, summary, projectPath, tags)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("extraction failed: %v. Check LLM API connection with luoshu_config_validate", err)), nil
	}

	result := map[string]interface{}{
		"extracted": len(entries),
		"entries":   entries,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ─── memory_semantic_search tool ─────────────────────

func buildMemorySemanticSearchTool() mcp.Tool {
	return mcp.NewTool(
		"memory_semantic_search",
		mcp.WithDescription("Semantic search over memories. Find related memories using natural language. Automatically falls back to keyword search when Embedding is not configured."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query (natural language)"),
		),
		mcp.WithString("project_path",
			mcp.Description("Limit search scope to a specific project"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of results to return (default 10)"),
		),
		mcp.WithString("mode",
			mcp.Description("Search mode: 'auto' (default, hybrid), 'semantic', 'keyword'"),
		),
	)
}

func handleMemorySemanticSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query parameter is required"), nil
	}
	projectPath := req.GetString("project_path", "")
	maxResults := req.GetInt("max_results", 10)
	mode := req.GetString("mode", "auto")

	cfg, err := luoshu.Load()
	if err != nil {
		return mcp.NewToolResultError(errConfigLoad(err)), nil
	}

	_, embedder := luoshu.NewProviders(cfg)

	store, err := luoshu.NewMemoryStore()
	if err != nil {
		return mcp.NewToolResultError(errInitFailed("memory store", err)), nil
	}
	index, err := luoshu.NewVectorIndex()
	if err != nil {
		return mcp.NewToolResultError(errInitFailed("vector index", err)), nil
	}
	cache, err := luoshu.NewEmbeddingCache()
	if err != nil {
		return mcp.NewToolResultError(errInitFailed("embedding cache", err)), nil
	}

	searcher := luoshu.NewSearcher(store, index, cache, embedder, cfg.Embedding.Model)
	results, searchMethod, err := searcher.Search(ctx, query, luoshu.SearchOptions{
		ProjectPath: projectPath,
		MaxResults:  maxResults,
		Mode:        mode,
	})
	if err != nil {
		return mcp.NewToolResultError(errSearchFailed(err)), nil
	}

	output := map[string]interface{}{
		"results":       results,
		"total":         len(results),
		"search_method": searchMethod,
		"query":         query,
	}
	data, _ := json.MarshalIndent(output, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
