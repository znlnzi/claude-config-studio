package evolution

import "time"

// Suggestion represents a self-evolution suggestion
type Suggestion struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`        // duplicate_rules, missing_coverage, health_issue, conflict
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Files       []string `json:"files,omitempty"`
	Suggestion  string   `json:"suggestion"`
	Confidence  float64  `json:"confidence"` // 0-1
	Source      string   `json:"source"`     // local_analysis, llm_analysis, web_search
	CreatedAt   string   `json:"created_at"`
	Status      string   `json:"status"` // pending, approved, rejected
}

// AnalysisRecord records the history of a single analysis run
type AnalysisRecord struct {
	ID              string `json:"id"`
	Scope           string `json:"scope"` // global or project path
	Timestamp       string `json:"timestamp"`
	RulesScanned    int    `json:"rules_scanned"`
	SuggestionsFound int   `json:"suggestions_found"`
	DurationMs      int64  `json:"duration_ms"`
}

// EvolveStatus represents the current state of the evolution system
type EvolveStatus struct {
	Initialized       bool   `json:"initialized"`
	PendingSuggestions int    `json:"pending_suggestions"`
	ApprovedTotal      int    `json:"approved_total"`
	RejectedTotal      int    `json:"rejected_total"`
	LastAnalysis       string `json:"last_analysis,omitempty"`
	TotalAnalyses      int    `json:"total_analyses"`
}

// NewSuggestionID generates a suggestion ID
func NewSuggestionID() string {
	now := time.Now()
	return now.Format("sugg-20060102-150405")
}
