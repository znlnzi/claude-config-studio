package luoshu

import (
	"context"
	"fmt"
	"strings"
)

// Recaller is an intelligent recall engine: semantic search + LLM synthesis
type Recaller struct {
	searcher  *Searcher
	llm       LLMProvider
	fileIndex *ClaudeIndex // optional: file-based search (auto-memory, rules)
}

// NewRecaller creates an intelligent recall engine
func NewRecaller(searcher *Searcher, llm LLMProvider) *Recaller {
	return &Recaller{searcher: searcher, llm: llm}
}

// WithFileIndex attaches a ClaudeIndex for unified search across JSONL memories and file-based content
func (r *Recaller) WithFileIndex(idx *ClaudeIndex) *Recaller {
	r.fileIndex = idx
	return r
}

// RecallResult represents a recall result
type RecallResult struct {
	Summary      string             `json:"summary"`                  // LLM-synthesized answer
	Sources      []SearchResult     `json:"sources"`                  // Raw search results (JSONL memories)
	FileSources  []FileSearchResult `json:"file_sources,omitempty"`   // File search results (auto-memory, rules)
	SearchMethod string             `json:"search_method"`            // Search method used
	SourceCount  int                `json:"source_count"`             // Total number of sources
}

// Recall performs intelligent recall: searches related memories and synthesizes a coherent answer via LLM
func (r *Recaller) Recall(ctx context.Context, query string, opts SearchOptions) (*RecallResult, error) {
	if opts.MaxResults <= 0 {
		opts.MaxResults = 5
	}

	// 1. Semantic search over JSONL memories
	results, method, err := r.searcher.Search(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// 2. File-based search (if fileIndex is available)
	var fileResults []FileSearchResult
	if r.fileIndex != nil {
		scope := opts.ProjectPath
		if scope == "" {
			scope = "global"
		}
		// Auto-reconcile before search
		_, _ = r.fileIndex.Reconcile(ctx, scope)
		fileResults, _ = r.fileIndex.Search(ctx, query, scope, opts.MaxResults)
	}

	totalSources := len(results) + len(fileResults)

	if totalSources == 0 {
		return &RecallResult{
			Summary:      "No related memories found.",
			Sources:      nil,
			FileSources:  nil,
			SearchMethod: method,
			SourceCount:  0,
		}, nil
	}

	// 3. If LLM is unavailable, return search results directly (no synthesis)
	if r.llm.Name() == "noop" {
		return &RecallResult{
			Summary:      buildUnifiedFallbackSummary(results, fileResults),
			Sources:      results,
			FileSources:  fileResults,
			SearchMethod: method,
			SourceCount:  totalSources,
		}, nil
	}

	// 4. Synthesize all search results via LLM
	summary, err := r.synthesizeUnified(ctx, query, results, fileResults)
	if err != nil {
		// LLM synthesis failed, fall back to raw results
		return &RecallResult{
			Summary:      buildUnifiedFallbackSummary(results, fileResults),
			Sources:      results,
			FileSources:  fileResults,
			SearchMethod: method,
			SourceCount:  totalSources,
		}, nil
	}

	searchMethod := method + "+llm"
	if len(fileResults) > 0 {
		searchMethod = method + "+files+llm"
	}

	return &RecallResult{
		Summary:      summary,
		Sources:      results,
		FileSources:  fileResults,
		SearchMethod: searchMethod,
		SourceCount:  totalSources,
	}, nil
}

// synthesizeUnified uses LLM to combine both JSONL and file search results into a coherent answer
func (r *Recaller) synthesizeUnified(ctx context.Context, query string, memResults []SearchResult, fileResults []FileSearchResult) (string, error) {
	prompt := buildUnifiedRecallPrompt(query, memResults, fileResults)

	resp, err := r.llm.Chat(ctx, ChatRequest{
		Messages:    []Message{{Role: "user", Content: prompt}},
		MaxTokens:   1000,
		Temperature: 0.3,
	})
	if err != nil {
		return "", fmt.Errorf("LLM synthesis failed: %w", err)
	}

	return resp.Content, nil
}

const unifiedRecallPromptTemplate = `You are a memory recall assistant. Based on the user's question and the retrieved memory entries, generate a concise, coherent answer.

User question: %s

%s
Requirements:
- Answer the user's question directly, do not say "Based on memories..."
- Synthesize information from multiple sources, do not list them one by one
- If there are contradictions among the sources, point them out
- If the information is insufficient for a complete answer, indicate which parts are known
- Keep it concise, 1-3 paragraphs`

func buildUnifiedRecallPrompt(query string, memResults []SearchResult, fileResults []FileSearchResult) string {
	var sources strings.Builder

	if len(memResults) > 0 {
		sources.WriteString("Retrieved memories (sorted by relevance):\n")
		for i, r := range memResults {
			sources.WriteString(fmt.Sprintf("\n--- Memory %d ---\n", i+1))
			sources.WriteString(r.Entry.Content)
			if len(r.Entry.Tags) > 0 {
				sources.WriteString(fmt.Sprintf("\nTags: %s", strings.Join(r.Entry.Tags, ", ")))
			}
			sources.WriteString("\n")
		}
	}

	if len(fileResults) > 0 {
		sources.WriteString("\nRetrieved file content (auto-memory, rules, etc.):\n")
		for i, r := range fileResults {
			content := r.Content
			if len(content) > 500 {
				content = content[:500] + "..."
			}
			sources.WriteString(fmt.Sprintf("\n--- File %d: %s ---\n", i+1, r.FilePath))
			sources.WriteString(content)
			sources.WriteString("\n")
		}
	}

	return fmt.Sprintf(unifiedRecallPromptTemplate, query, sources.String())
}

// buildUnifiedFallbackSummary generates a simple summary when LLM is unavailable
func buildUnifiedFallbackSummary(memResults []SearchResult, fileResults []FileSearchResult) string {
	total := len(memResults) + len(fileResults)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d related sources:\n\n", total))

	if len(memResults) > 0 {
		sb.WriteString("**Memories:**\n")
		for i, r := range memResults {
			content := r.Entry.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, content))
		}
	}

	if len(fileResults) > 0 {
		sb.WriteString("\n**Files:**\n")
		for i, r := range fileResults {
			context := r.Context
			if context == "" {
				context = r.FilePath
			}
			sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, r.FilePath, context))
		}
	}

	return sb.String()
}
