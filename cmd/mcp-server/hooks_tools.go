package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
)

// ─── Hooks Management Tool Definitions ─────────────────────────────

func buildHooksListTool() mcp.Tool {
	return mcp.NewTool(
		"hooks_list",
		mcp.WithDescription("List all configured hooks (global or project scope). Returns hooks grouped by event type (PreToolUse, PostToolUse, SessionStart, Stop, etc.)."),
		mcp.WithString("scope",
			mcp.Description("Scope: 'global' (default) or absolute project path"),
		),
	)
}

func buildHooksSaveTool() mcp.Tool {
	return mcp.NewTool(
		"hooks_save",
		mcp.WithDescription("Save hooks configuration to settings.json. Only modifies the 'hooks' field, preserving all other settings. Pass the complete hooks object as JSON."),
		mcp.WithString("hooks",
			mcp.Required(),
			mcp.Description("Complete hooks configuration as JSON object. Format: {\"EventName\": [{\"matcher\": \"...\", \"hooks\": [{\"type\": \"command\", \"command\": \"...\"}]}]}"),
		),
		mcp.WithString("scope",
			mcp.Description("Scope: 'global' (default) or absolute project path"),
		),
	)
}

// ─── Hooks Management Tool Handlers ─────────────────────────────

func handleHooksList(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scope := request.GetString("scope", "global")

	settingsPath, err := resolveSettingsPath(scope)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	events, err := readHooksFromSettings(settingsPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read hooks: %v", err)), nil
	}

	result := map[string]interface{}{
		"scope":        scope,
		"events":       events,
		"total_events": len(events),
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleHooksSave(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	hooksStr, err := request.RequireString("hooks")
	if err != nil {
		return mcp.NewToolResultError("hooks parameter is required"), nil
	}
	scope := request.GetString("scope", "global")

	// Validate hooks JSON format
	var hooksObj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(hooksStr), &hooksObj); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid hooks JSON: %v", err)), nil
	}

	// Validate event names
	validEvents := map[string]bool{
		"PreToolUse":       true,
		"PostToolUse":      true,
		"SessionStart":     true,
		"SessionEnd":       true,
		"Stop":             true,
		"PreCompact":       true,
		"UserPromptSubmit": true,
	}
	for event := range hooksObj {
		if !validEvents[event] {
			return mcp.NewToolResultError(fmt.Sprintf("unknown hook event: %q. Valid events: PreToolUse, PostToolUse, SessionStart, SessionEnd, Stop, PreCompact, UserPromptSubmit", event)), nil
		}
	}

	settingsPath, err := resolveSettingsPath(scope)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := writeHooksToSettings(settingsPath, hooksStr); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to save hooks: %v", err)), nil
	}

	result := map[string]interface{}{
		"success":      true,
		"scope":        scope,
		"events_saved": len(hooksObj),
		"path":         settingsPath,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ─── Helper Functions ─────────────────────────────

// resolveSettingsPath resolves the settings.json path based on scope
func resolveSettingsPath(scope string) (string, error) {
	if scope == "" || scope == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".claude", "settings.json"), nil
	}
	// project scope
	if _, err := os.Stat(scope); os.IsNotExist(err) {
		return "", fmt.Errorf("project path does not exist: %s", scope)
	}
	return filepath.Join(scope, ".claude", "settings.json"), nil
}

// hooksEvent represents hooks configuration for a single event type
type hooksEvent struct {
	Event   string            `json:"event"`
	Entries []json.RawMessage `json:"entries"`
}

// readHooksFromSettings reads hooks configuration from settings.json
func readHooksFromSettings(path string) ([]hooksEvent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return []hooksEvent{}, nil // return empty if file doesn't exist
	}

	var settings map[string]json.RawMessage
	if err := json.Unmarshal(data, &settings); err != nil {
		return []hooksEvent{}, nil
	}

	hooksRaw, ok := settings["hooks"]
	if !ok {
		return []hooksEvent{}, nil
	}

	var hooksMap map[string]json.RawMessage
	if err := json.Unmarshal(hooksRaw, &hooksMap); err != nil {
		return []hooksEvent{}, nil
	}

	// Output known events in a fixed order
	eventOrder := []string{"PreToolUse", "PostToolUse", "SessionStart", "SessionEnd", "Stop", "PreCompact", "UserPromptSubmit"}
	seen := make(map[string]bool)
	var result []hooksEvent

	for _, event := range eventOrder {
		raw, ok := hooksMap[event]
		if !ok {
			continue
		}
		seen[event] = true
		var entries []json.RawMessage
		if err := json.Unmarshal(raw, &entries); err != nil {
			continue
		}
		result = append(result, hooksEvent{Event: event, Entries: entries})
	}

	// Handle unknown events (potentially added in the future)
	for event, raw := range hooksMap {
		if seen[event] {
			continue
		}
		var entries []json.RawMessage
		if err := json.Unmarshal(raw, &entries); err != nil {
			continue
		}
		result = append(result, hooksEvent{Event: event, Entries: entries})
	}

	return result, nil
}

// writeHooksToSettings writes hooks JSON to the hooks field of settings.json
func writeHooksToSettings(path string, hooksJSON string) error {
	// Read existing settings
	var settings map[string]json.RawMessage
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &settings)
	}
	if settings == nil {
		settings = make(map[string]json.RawMessage)
	}

	// Replace hooks field
	settings["hooks"] = json.RawMessage(hooksJSON)

	formatted, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)
	return os.WriteFile(path, append(formatted, '\n'), 0644)
}
