package luoshu

import (
	"context"
	"fmt"
	"strings"
)

// Recaller is an intelligent recall engine: semantic search + LLM synthesis
type Recaller struct {
	searcher *Searcher
	llm      LLMProvider
}

// NewRecaller creates an intelligent recall engine
func NewRecaller(searcher *Searcher, llm LLMProvider) *Recaller {
	return &Recaller{searcher: searcher, llm: llm}
}

// RecallResult represents a recall result
type RecallResult struct {
	Summary      string         `json:"summary"`       // LLM-synthesized answer
	Sources      []SearchResult `json:"sources"`        // Raw search results
	SearchMethod string         `json:"search_method"`  // Search method used
	SourceCount  int            `json:"source_count"`   // Number of sources
}

// Recall performs intelligent recall: searches related memories and synthesizes a coherent answer via LLM
func (r *Recaller) Recall(ctx context.Context, query string, opts SearchOptions) (*RecallResult, error) {
	if opts.MaxResults <= 0 {
		opts.MaxResults = 5
	}

	// 1. Semantic search
	results, method, err := r.searcher.Search(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		return &RecallResult{
			Summary:      "No related memories found.",
			Sources:      nil,
			SearchMethod: method,
			SourceCount:  0,
		}, nil
	}

	// 2. If LLM is unavailable, return search results directly (no synthesis)
	if r.llm.Name() == "noop" {
		return &RecallResult{
			Summary:      buildFallbackSummary(results),
			Sources:      results,
			SearchMethod: method,
			SourceCount:  len(results),
		}, nil
	}

	// 3. Synthesize search results via LLM
	summary, err := r.synthesize(ctx, query, results)
	if err != nil {
		// LLM synthesis failed, fall back to raw results
		return &RecallResult{
			Summary:      buildFallbackSummary(results),
			Sources:      results,
			SearchMethod: method,
			SourceCount:  len(results),
		}, nil
	}

	return &RecallResult{
		Summary:      summary,
		Sources:      results,
		SearchMethod: method + "+llm",
		SourceCount:  len(results),
	}, nil
}

// synthesize uses LLM to combine search results into a coherent answer
func (r *Recaller) synthesize(ctx context.Context, query string, results []SearchResult) (string, error) {
	prompt := buildRecallPrompt(query, results)

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

const recallPromptTemplate = `You are a memory recall assistant. Based on the user's question and the retrieved memory entries, generate a concise, coherent answer.

User question: %s

Retrieved memories (sorted by relevance):
%s

Requirements:
- Answer the user's question directly, do not say "Based on memories..."
- Synthesize information from multiple memories, do not list them one by one
- If there are contradictions among the memories, point them out
- If the memories are insufficient for a complete answer, indicate which parts are known
- Keep it concise, 1-3 paragraphs`

func buildRecallPrompt(query string, results []SearchResult) string {
	var memories strings.Builder
	for i, r := range results {
		memories.WriteString(fmt.Sprintf("\n--- Memory %d ---\n", i+1))
		memories.WriteString(r.Entry.Content)
		if len(r.Entry.Tags) > 0 {
			memories.WriteString(fmt.Sprintf("\nTags: %s", strings.Join(r.Entry.Tags, ", ")))
		}
		memories.WriteString("\n")
	}
	return fmt.Sprintf(recallPromptTemplate, query, memories.String())
}

// buildFallbackSummary generates a simple summary when LLM is unavailable
func buildFallbackSummary(results []SearchResult) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d related memories:\n\n", len(results)))
	for i, r := range results {
		content := r.Entry.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, content))
	}
	return sb.String()
}
