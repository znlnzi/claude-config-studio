package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/znlnzi/claude-config-studio/internal/luoshu"

	"github.com/mark3labs/mcp-go/mcp"
)

func buildSaveMemoryTool() mcp.Tool {
	return mcp.NewTool(
		"save_memory",
		mcp.WithDescription("Save a memory entry to the project's .claude/memory/ directory. Use 'global' as project_path for global memory (~/.claude/memory/)."),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("Absolute project path, or 'global' for global memory"),
		),
		mcp.WithString("filename",
			mcp.Required(),
			mcp.Description("Filename such as MEMORY.md, session-state.md, decisions.md"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("File content to save"),
		),
		mcp.WithString("mode",
			mcp.Description("Write mode: 'overwrite' (default) or 'append'"),
		),
	)
}

func buildLoadMemoryTool() mcp.Tool {
	return mcp.NewTool(
		"load_memory",
		mcp.WithDescription("Load memory files from a project's .claude/memory/ directory. If filename is omitted, returns all .md files."),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("Absolute project path, or 'global' for global memory"),
		),
		mcp.WithString("filename",
			mcp.Description("Specific filename to load; omit to list all memory files"),
		),
	)
}

func buildSearchMemoryTool() mcp.Tool {
	return mcp.NewTool(
		"search_memory",
		mcp.WithDescription("Keyword search over memory files (case-insensitive exact match). For conceptual or semantic queries, prefer ov_search which finds related content even with different wording."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search keyword (case-insensitive)"),
		),
		mcp.WithString("project_path",
			mcp.Description("Limit search to a specific project path; omit to search all"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of matching lines to return (default 20)"),
		),
	)
}

// handleSaveMemory handles the save_memory tool call
func handleSaveMemory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError("project_path is required"), nil
	}
	filename, err := req.RequireString("filename")
	if err != nil {
		return mcp.NewToolResultError("filename is required"), nil
	}
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("content is required"), nil
	}
	mode := req.GetString("mode", "overwrite")

	if !isSafeFilename(filename) {
		return mcp.NewToolResultError("unsafe filename: must not contain '..' or path separators"), nil
	}

	memoryDir, err := resolveMemoryDir(projectPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to resolve path: %v", err)), nil
	}

	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create directory: %v", err)), nil
	}

	filePath := filepath.Join(memoryDir, filename)

	if mode == "append" {
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to open file: %v", err)), nil
		}
		defer f.Close()
		if _, err := f.WriteString(content); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to write: %v", err)), nil
		}
	} else {
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to write: %v", err)), nil
		}
	}

	result := map[string]interface{}{
		"success": true,
		"path":    filePath,
	}
	data, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(data)), nil
}

// handleLoadMemory handles the load_memory tool call
func handleLoadMemory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError("project_path is required"), nil
	}
	filename := req.GetString("filename", "")

	memoryDir, err := resolveMemoryDir(projectPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to resolve path: %v", err)), nil
	}

	// Filename specified: return single file content
	if filename != "" {
		if !isSafeFilename(filename) {
			return mcp.NewToolResultError("unsafe filename"), nil
		}
		filePath := filepath.Join(memoryDir, filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to read %s: %v", filename, err)), nil
		}
		fi, _ := os.Stat(filePath)
		modTime := ""
		if fi != nil {
			modTime = fi.ModTime().Format(time.RFC3339)
		}
		result := map[string]interface{}{
			"filename":    filename,
			"content":     string(data),
			"path":        filePath,
			"modified_at": modTime,
		}
		out, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(out)), nil
	}

	// No filename specified: list all .md files in the directory
	entries, err := os.ReadDir(memoryDir)
	if err != nil {
		result := map[string]interface{}{
			"files":   []interface{}{},
			"message": "memory directory does not exist or is empty",
		}
		out, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(out)), nil
	}

	type fileEntry struct {
		Filename   string `json:"filename"`
		Content    string `json:"content"`
		Size       int64  `json:"size"`
		ModifiedAt string `json:"modified_at"`
	}

	var files []fileEntry
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		filePath := filepath.Join(memoryDir, e.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		fi, _ := e.Info()
		modTime := ""
		size := int64(0)
		if fi != nil {
			modTime = fi.ModTime().Format(time.RFC3339)
			size = fi.Size()
		}
		files = append(files, fileEntry{
			Filename:   e.Name(),
			Content:    string(data),
			Size:       size,
			ModifiedAt: modTime,
		})
	}

	result := map[string]interface{}{
		"files": files,
	}
	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

// handleSearchMemory handles the search_memory tool call
func handleSearchMemory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query is required"), nil
	}
	projectPath := req.GetString("project_path", "")
	maxResults := req.GetInt("max_results", 20)

	queryLower := strings.ToLower(query)

	type matchEntry struct {
		Line    int    `json:"line"`
		Content string `json:"content"`
	}
	type searchResult struct {
		Project  string       `json:"project"`
		Filename string       `json:"filename"`
		Matches  []matchEntry `json:"matches"`
	}

	var results []searchResult
	totalMatches := 0

	// Collect the list of projects to search
	var projectPaths []string
	if projectPath != "" && projectPath != "global" {
		projectPaths = []string{projectPath}
	} else {
		projectPaths = append(projectPaths, "global")
		allProjects, _ := scanAllProjects()
		projectPaths = append(projectPaths, allProjects...)
	}

	for _, pp := range projectPaths {
		if totalMatches >= maxResults {
			break
		}

		memDir, err := resolveMemoryDir(pp)
		if err != nil {
			continue
		}

		entries, err := os.ReadDir(memDir)
		if err != nil {
			continue
		}

		for _, e := range entries {
			if totalMatches >= maxResults {
				break
			}
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}

			data, err := os.ReadFile(filepath.Join(memDir, e.Name()))
			if err != nil {
				continue
			}

			var matches []matchEntry
			lines := strings.Split(string(data), "\n")
			for i, line := range lines {
				if totalMatches >= maxResults {
					break
				}
				if strings.Contains(strings.ToLower(line), queryLower) {
					matches = append(matches, matchEntry{
						Line:    i + 1,
						Content: line,
					})
					totalMatches++
				}
			}

			if len(matches) > 0 {
				label := pp
				if pp == "global" {
					label = "~/.claude (global)"
				}
				results = append(results, searchResult{
					Project:  label,
					Filename: e.Name(),
					Matches:  matches,
				})
			}
		}
	}

	// Semantic search supplement: automatically add semantic results when keyword matches are few and luoshu embedding is available
	searchMethod := "keyword"
	var semanticResults []map[string]interface{}

	if totalMatches < 3 {
		semanticResults, searchMethod = trySemanticSupplement(ctx, query, projectPath, maxResults-totalMatches, totalMatches > 0)
	}

	output := map[string]interface{}{
		"results":       results,
		"total_matches": totalMatches,
		"query":         query,
		"search_method": searchMethod,
	}
	if len(semanticResults) > 0 {
		output["semantic_results"] = semanticResults
		output["total_matches"] = totalMatches + len(semanticResults)
	}
	out, _ := json.Marshal(output)
	return mcp.NewToolResultText(string(out)), nil
}

// trySemanticSupplement attempts to supplement keyword results with semantic search.
// Returns semantic results and the final search method identifier.
func trySemanticSupplement(ctx context.Context, query, projectPath string, limit int, hasKeywordResults bool) ([]map[string]interface{}, string) {
	if limit <= 0 {
		limit = 5
	}

	cfg, err := luoshu.Load()
	if err != nil {
		return nil, "keyword"
	}

	_, embedder := luoshu.NewProviders(cfg)
	if embedder.Name() == "noop" {
		return nil, "keyword"
	}

	store, err := luoshu.NewMemoryStore()
	if err != nil {
		return nil, "keyword"
	}
	index, err := luoshu.NewVectorIndex()
	if err != nil {
		return nil, "keyword"
	}
	cache, err := luoshu.NewEmbeddingCache()
	if err != nil {
		return nil, "keyword"
	}

	searcher := luoshu.NewSearcher(store, index, cache, embedder, cfg.Embedding.Model)
	results, _, err := searcher.Search(ctx, query, luoshu.SearchOptions{
		ProjectPath: projectPath,
		MaxResults:  limit,
		Mode:        "semantic",
	})
	if err != nil || len(results) == 0 {
		return nil, "keyword"
	}

	var semanticResults []map[string]interface{}
	for _, r := range results {
		semanticResults = append(semanticResults, map[string]interface{}{
			"content":   r.Entry.Content,
			"score":     r.Score,
			"source":    r.Entry.Source,
			"tags":      r.Entry.Tags,
			"highlight": r.Highlight,
		})
	}

	method := "keyword+semantic"
	if !hasKeywordResults {
		method = "semantic"
	}
	return semanticResults, method
}
