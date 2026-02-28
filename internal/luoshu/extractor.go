package luoshu

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Extractor uses LLM to extract key points from sessions
type Extractor struct {
	llm   LLMProvider
	store *MemoryStore
}

// NewExtractor creates an extractor
func NewExtractor(llm LLMProvider, store *MemoryStore) *Extractor {
	return &Extractor{llm: llm, store: store}
}

// Extract extracts key points from a session summary and saves them
func (e *Extractor) Extract(ctx context.Context, sessionSummary, projectPath string, tags []string) ([]MemoryEntry, error) {
	prompt := buildExtractPrompt(sessionSummary)

	resp, err := e.llm.Chat(ctx, ChatRequest{
		Messages:    []Message{{Role: "user", Content: prompt}},
		MaxTokens:   2000,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	var extracted []extractedPoint
	if err := json.Unmarshal([]byte(resp.Content), &extracted); err != nil {
		jsonStr := extractJSON(resp.Content)
		if err2 := json.Unmarshal([]byte(jsonStr), &extracted); err2 != nil {
			return nil, fmt.Errorf("failed to parse extraction results: %w", err)
		}
	}

	now := time.Now().Format(time.RFC3339)
	var entries []MemoryEntry
	for _, ep := range extracted {
		allTags := make([]string, 0, len(ep.Tags)+len(tags))
		allTags = append(allTags, ep.Tags...)
		allTags = append(allTags, tags...)
		entry := MemoryEntry{
			ID:      NewMemoryID(),
			Content: ep.Content,
			Source: MemorySource{
				Type:    "auto_extract",
				Project: projectPath,
			},
			Tags:      allTags,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := e.store.Append(entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

type extractedPoint struct {
	Content    string   `json:"content"`
	Tags       []string `json:"tags"`
	Importance int      `json:"importance"`
}

const extractPromptTemplate = `You are a memory extraction assistant. Extract key points worth remembering long-term from the following session summary.

Output a JSON array, each point containing:
- content: The key point (concise, 1-3 sentences)
- tags: Classification tag array (possible values: decision, preference, solution, context, architecture)
- importance: Importance level 1-5

Only extract information with long-term value, ignore:
- Temporary debugging steps
- One-time operational commands
- Pure chat/small talk content

Session summary:
%s`

func buildExtractPrompt(summary string) string {
	return fmt.Sprintf(extractPromptTemplate, summary)
}

// extractJSON attempts to extract a JSON array from markdown code blocks or text
func extractJSON(text string) string {
	start := strings.Index(text, "[")
	end := strings.LastIndex(text, "]")
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return text
}
