package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/znlnzi/claude-config-studio/internal/evolution"

	"github.com/mark3labs/mcp-go/mcp"
)

// ─── Evolution tool definitions ─────────────────────────────

func buildEvolveStatusTool() mcp.Tool {
	return mcp.NewTool(
		"evolve_status",
		mcp.WithDescription("Get the current status of the evolution system: pending suggestions count, last analysis time, and history statistics."),
	)
}

func buildEvolveAnalyzeTool() mcp.Tool {
	return mcp.NewTool(
		"evolve_analyze",
		mcp.WithDescription("Trigger a local analysis of rules to find duplicates, missing coverage, and health issues. Does not require LLM API. Results are stored as suggestions for review."),
		mcp.WithString("scope",
			mcp.Description("Scope to analyze: 'global' (default) or absolute project path"),
		),
	)
}

func buildEvolveApplyTool() mcp.Tool {
	return mcp.NewTool(
		"evolve_apply",
		mcp.WithDescription("Approve or reject a suggestion by its ID. Approved suggestions are marked for action; rejected ones are dismissed."),
		mcp.WithString("suggestion_id",
			mcp.Required(),
			mcp.Description("The suggestion ID to approve or reject"),
		),
		mcp.WithString("action",
			mcp.Required(),
			mcp.Description("Action to take: 'approve' or 'reject'"),
		),
	)
}

// ─── Evolution tool handlers ─────────────────────────────

func handleEvolveStatus(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	store := getEvolveStore()
	status, err := store.GetStatus()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get status: %v", err)), nil
	}

	// Append pending suggestion summaries
	pending, _ := store.GetPendingSuggestions()
	type summaryItem struct {
		ID         string  `json:"id"`
		Type       string  `json:"type"`
		Title      string  `json:"title"`
		Confidence float64 `json:"confidence"`
	}
	var pendingList []summaryItem
	for _, s := range pending {
		pendingList = append(pendingList, summaryItem{
			ID:         s.ID,
			Type:       s.Type,
			Title:      s.Title,
			Confidence: s.Confidence,
		})
	}

	result := map[string]interface{}{
		"status":              status,
		"pending_suggestions": pendingList,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleEvolveAnalyze(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scope := request.GetString("scope", "global")

	claudeDir, err := resolveClaudeDirForScope(scope)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Check if rules directory exists
	rulesDir := filepath.Join(claudeDir, "rules")
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("no rules directory found at %s", rulesDir)), nil
	}

	// Execute analysis
	analyzer := evolution.NewAnalyzer(claudeDir, scope)
	suggestions, record, err := analyzer.Analyze()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("analysis failed: %v", err)), nil
	}

	// Save results
	store := getEvolveStore()
	if len(suggestions) > 0 {
		if err := store.AddSuggestions(suggestions); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to save suggestions: %v", err)), nil
		}
	}
	if err := store.AddHistory(record); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to save history: %v", err)), nil
	}

	result := map[string]interface{}{
		"scope":             scope,
		"rules_scanned":     record.RulesScanned,
		"suggestions_found": record.SuggestionsFound,
		"duration_ms":       record.DurationMs,
		"suggestions":       suggestions,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleEvolveApply(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	suggestionID, err := request.RequireString("suggestion_id")
	if err != nil {
		return mcp.NewToolResultError("suggestion_id is required"), nil
	}
	action, err := request.RequireString("action")
	if err != nil {
		return mcp.NewToolResultError("action is required"), nil
	}

	if action != "approve" && action != "reject" {
		return mcp.NewToolResultError("action must be 'approve' or 'reject'"), nil
	}

	store := getEvolveStore()

	// Find the suggestion
	suggestions, err := store.LoadSuggestions()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to load suggestions: %v", err)), nil
	}

	var target *evolution.Suggestion
	for i := range suggestions {
		if suggestions[i].ID == suggestionID {
			target = &suggestions[i]
			break
		}
	}

	if target == nil {
		return mcp.NewToolResultError(fmt.Sprintf("suggestion not found: %s", suggestionID)), nil
	}

	if target.Status != "pending" {
		return mcp.NewToolResultError(fmt.Sprintf("suggestion is already %s", target.Status)), nil
	}

	// Update status
	newStatus := "rejected"
	if action == "approve" {
		newStatus = "approved"
	}

	if err := store.UpdateSuggestionStatus(suggestionID, newStatus); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update: %v", err)), nil
	}

	result := map[string]interface{}{
		"success":       true,
		"suggestion_id": suggestionID,
		"action":        action,
		"new_status":    newStatus,
		"title":         target.Title,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ─── Helper functions ─────────────────────────────

// getEvolveStore returns the global evolution store
func getEvolveStore() *evolution.Store {
	home, _ := os.UserHomeDir()
	return evolution.NewStore(filepath.Join(home, ".claude", "evolution"))
}

// resolveClaudeDirForScope resolves the .claude directory based on scope
func resolveClaudeDirForScope(scope string) (string, error) {
	if scope == "" || scope == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".claude"), nil
	}
	if _, err := os.Stat(scope); os.IsNotExist(err) {
		return "", fmt.Errorf("path does not exist: %s", scope)
	}
	return filepath.Join(scope, ".claude"), nil
}
