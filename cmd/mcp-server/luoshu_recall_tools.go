package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/znlnzi/claude-config-studio/internal/luoshu"

	"github.com/mark3labs/mcp-go/mcp"
)

// ─── luoshu_recall tool ─────────────────────────────

func buildLuoshuRecallTool() mcp.Tool {
	return mcp.NewTool(
		"luoshu_recall",
		mcp.WithDescription("Intelligent recall: query related memories using natural language and automatically synthesize a coherent answer. Combines semantic search with LLM synthesis. Falls back to keyword search when LLM is not configured."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Recall query (natural language, e.g. 'What was the last decision about the authentication system')"),
		),
		mcp.WithString("project_path",
			mcp.Description("Limit search scope to a specific project"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of search sources (default 5)"),
		),
	)
}

func handleLuoshuRecall(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query parameter is required"), nil
	}
	projectPath := req.GetString("project_path", "")
	maxResults := req.GetInt("max_results", 5)

	cfg, err := luoshu.Load()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load config: %v", err)), nil
	}

	llm, embedder := luoshu.NewProviders(cfg)

	store, err := luoshu.NewMemoryStore()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to initialize store: %v", err)), nil
	}
	index, err := luoshu.NewVectorIndex()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to initialize index: %v", err)), nil
	}
	cache, err := luoshu.NewEmbeddingCache()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to initialize cache: %v", err)), nil
	}

	searcher := luoshu.NewSearcher(store, index, cache, embedder, cfg.Embedding.Model)
	recaller := luoshu.NewRecaller(searcher, llm)

	// Attach file index for unified search across JSONL memories + file content
	fileIndex := luoshu.NewClaudeIndex(embedder, cache, cfg.Embedding.Model)
	recaller.WithFileIndex(fileIndex)

	result, err := recaller.Recall(ctx, query, luoshu.SearchOptions{
		ProjectPath: projectPath,
		MaxResults:  maxResults,
		Mode:        "auto",
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Recall failed: %v", err)), nil
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
