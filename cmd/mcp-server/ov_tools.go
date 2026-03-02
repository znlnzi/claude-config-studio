package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/znlnzi/claude-config-studio/internal/luoshu"

	"github.com/mark3labs/mcp-go/mcp"
)

// ─── ov_search tool ──────────────────────────────────

func buildOvSearchTool() mcp.Tool {
	return mcp.NewTool(
		"ov_search",
		mcp.WithDescription(`Semantic search over Claude Code memory and rules files.

    PREFERRED over keyword search (search_memory) when the query is
    conceptual or uses natural language. For example, searching
    "error handling" will also match "exception patterns", and
    "testing best practices" will find testing-related rules even if the exact
    words don't appear.

    Automatically indexes and syncs files — no setup required.

    Args:
        query: Natural language search query.
        scope: "global" for ~/.claude/, or absolute project path.
        limit: Maximum results to return (default 10).`),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Natural language search query"),
		),
		mcp.WithString("scope",
			mcp.Description(`"global" for ~/.claude/, or absolute project path`),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum results to return (default 10)"),
		),
	)
}

func handleOvSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query parameter is required"), nil
	}
	scope := req.GetString("scope", "global")
	limit := req.GetInt("limit", 10)

	cfg, err := luoshu.Load()
	if err != nil {
		return mcp.NewToolResultError(errConfigLoad(err)), nil
	}

	_, embedder := luoshu.NewProviders(cfg)
	cache, err := luoshu.NewEmbeddingCache()
	if err != nil {
		return mcp.NewToolResultError(errInitFailed("embedding cache", err)), nil
	}

	ci := luoshu.NewClaudeIndex(embedder, cache, cfg.Embedding.Model)

	// Auto reconcile (with cooldown mechanism)
	_, _ = ci.Reconcile(ctx, scope)

	results, err := ci.Search(ctx, query, scope, limit)
	if err != nil {
		return mcp.NewToolResultError(errSearchFailed(err)), nil
	}

	output := map[string]interface{}{
		"results": results,
		"total":   len(results),
		"query":   query,
		"scope":   scope,
	}
	data, _ := json.MarshalIndent(output, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ─── ov_index tool ──────────────────────────────────

func buildOvIndexTool() mcp.Tool {
	return mcp.NewTool(
		"ov_index",
		mcp.WithDescription(`Build or update the vector index for semantic search.

    Scans memory/*.md and rules/*.md files and creates vector embeddings.
    Usually NOT needed — ov_search automatically detects file changes.
    Use force=True to rebuild the entire index from scratch if search
    results seem stale or inconsistent.

    Args:
        scope: "global" for ~/.claude/, or absolute project path.
        force: Delete manifest and rebuild index (default False).`),
		mcp.WithString("scope",
			mcp.Description(`"global" for ~/.claude/, or absolute project path`),
		),
		mcp.WithBoolean("force",
			mcp.Description("Delete manifest and rebuild index (default false)"),
		),
	)
}

func handleOvIndex(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scope := req.GetString("scope", "global")
	force := req.GetBool("force", false)

	cfg, err := luoshu.Load()
	if err != nil {
		return mcp.NewToolResultError(errConfigLoad(err)), nil
	}

	_, embedder := luoshu.NewProviders(cfg)
	if embedder.Name() == "noop" {
		return mcp.NewToolResultError("Embedding not configured. Please run /luoshu.config to set the API Key first."), nil
	}

	cache, err := luoshu.NewEmbeddingCache()
	if err != nil {
		return mcp.NewToolResultError(errInitFailed("embedding cache", err)), nil
	}

	ci := luoshu.NewClaudeIndex(embedder, cache, cfg.Embedding.Model)

	start := time.Now()
	var changes int

	if force {
		if err := ci.Reindex(ctx, scope); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to rebuild index: %v. Check embedding API connection with luoshu_config_validate", err)), nil
		}
		status, _ := ci.Status(scope)
		changes = status.IndexedFiles
	} else {
		var err error
		changes, err = ci.Reconcile(ctx, scope)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("incremental sync failed: %v. Check embedding API connection with luoshu_config_validate", err)), nil
		}
	}

	duration := time.Since(start).Milliseconds()
	status, _ := ci.Status(scope)

	result := map[string]interface{}{
		"changes":        changes,
		"total_indexed":  status.IndexedFiles,
		"vector_entries": status.VectorEntries,
		"duration_ms":    duration,
		"force":          force,
		"scope":          scope,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ─── ov_status tool ──────────────────────────────────

func buildOvStatusTool() mcp.Tool {
	return mcp.NewTool(
		"ov_status",
		mcp.WithDescription(`Check OpenViking MCP server status and index information.

    Returns initialization state, indexed scopes, and configuration details.
    Use this to diagnose issues or verify the server is working correctly.`),
	)
}

func handleOvStatus(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg, err := luoshu.Load()
	if err != nil {
		return mcp.NewToolResultError(errConfigLoad(err)), nil
	}

	embedConfigured := cfg.Embedding.APIKey != ""
	if !embedConfigured && cfg.LLM.APIKey != "" {
		embedConfigured = true
	}

	_, embedder := luoshu.NewProviders(cfg)
	cache, err := luoshu.NewEmbeddingCache()
	if err != nil {
		return mcp.NewToolResultError(errInitFailed("embedding cache", err)), nil
	}

	ci := luoshu.NewClaudeIndex(embedder, cache, cfg.Embedding.Model)

	globalStatus, _ := ci.Status("global")

	result := map[string]interface{}{
		"initialized":          true,
		"embedding_configured": embedConfigured,
		"embedding_model":      cfg.Embedding.Model,
		"global_index": map[string]interface{}{
			"indexed_files":   globalStatus.IndexedFiles,
			"vector_entries":  globalStatus.VectorEntries,
			"last_reconciled": globalStatus.LastReconciled,
			"index_dir":       globalStatus.IndexDir,
		},
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
