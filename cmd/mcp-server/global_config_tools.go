package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleGetGlobalConfig returns the global configuration overview
func handleGetGlobalConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get home dir: %v", err)), nil
	}

	claudeHome := filepath.Join(home, ".claude")

	result := map[string]interface{}{
		"claude_home": claudeHome,
	}

	// CLAUDE.md
	claudeMdPath := filepath.Join(claudeHome, "CLAUDE.md")
	if data, err := os.ReadFile(claudeMdPath); err == nil {
		result["claude_md"] = string(data)
		result["has_claude_md"] = true
	} else {
		result["has_claude_md"] = false
	}

	// settings.json
	settingsPath := filepath.Join(claudeHome, "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		result["settings"] = string(data)
		result["has_settings"] = true
	} else {
		result["has_settings"] = false
	}

	// .mcp.json
	mcpPath := filepath.Join(claudeHome, ".mcp.json")
	if data, err := os.ReadFile(mcpPath); err == nil {
		result["mcp_config"] = string(data)
		result["has_mcp"] = true
	} else {
		result["has_mcp"] = false
	}

	// Directory detection
	result["has_agents"] = dirHasFiles(filepath.Join(claudeHome, "agents"))
	result["has_rules"] = dirHasFiles(filepath.Join(claudeHome, "rules"))
	result["has_skills"] = dirHasFiles(filepath.Join(claudeHome, "skills"))
	result["has_commands"] = dirHasFiles(filepath.Join(claudeHome, "commands"))
	result["has_memory"] = dirHasFiles(filepath.Join(claudeHome, "memory"))

	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

// handleSaveGlobalConfig saves a global configuration field
func handleSaveGlobalConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	field, err := req.RequireString("field")
	if err != nil {
		return mcp.NewToolResultError("field is required (claude_md, settings, mcp)"), nil
	}
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("content is required"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get home dir: %v", err)), nil
	}

	claudeHome := filepath.Join(home, ".claude")
	os.MkdirAll(claudeHome, 0755)

	var targetPath string
	switch field {
	case "claude_md":
		targetPath = filepath.Join(claudeHome, "CLAUDE.md")
	case "settings":
		targetPath = filepath.Join(claudeHome, "settings.json")
		if !isValidJSON(content) {
			return mcp.NewToolResultError("invalid JSON content for settings"), nil
		}
		content = formatJSON(content)
	case "mcp":
		targetPath = filepath.Join(claudeHome, ".mcp.json")
		if !isValidJSON(content) {
			return mcp.NewToolResultError("invalid JSON content for mcp config"), nil
		}
		content = formatJSON(content)
	default:
		return mcp.NewToolResultError("field must be one of: claude_md, settings, mcp"), nil
	}

	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to write %s: %v", targetPath, err)), nil
	}

	result := map[string]interface{}{
		"success": true,
		"path":    targetPath,
		"field":   field,
	}
	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

// handleSaveProjectConfig saves a project-level configuration field
func handleSaveProjectConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError("project_path is required"), nil
	}
	field, err := req.RequireString("field")
	if err != nil {
		return mcp.NewToolResultError("field is required (claude_md, settings, mcp)"), nil
	}
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("content is required"), nil
	}

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("project path does not exist: %s", projectPath)), nil
	}

	claudeDir := filepath.Join(projectPath, ".claude")
	os.MkdirAll(claudeDir, 0755)

	var targetPath string
	switch field {
	case "claude_md":
		targetPath = filepath.Join(claudeDir, "CLAUDE.md")
	case "settings":
		targetPath = filepath.Join(claudeDir, "settings.json")
		if !isValidJSON(content) {
			return mcp.NewToolResultError("invalid JSON content for settings"), nil
		}
		content = formatJSON(content)
	case "mcp":
		targetPath = filepath.Join(claudeDir, ".mcp.json")
		if !isValidJSON(content) {
			return mcp.NewToolResultError("invalid JSON content for mcp config"), nil
		}
		content = formatJSON(content)
	default:
		return mcp.NewToolResultError("field must be one of: claude_md, settings, mcp"), nil
	}

	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to write %s: %v", targetPath, err)), nil
	}

	result := map[string]interface{}{
		"success":      true,
		"path":         targetPath,
		"field":        field,
		"project_path": projectPath,
	}
	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

// isValidJSON checks whether the string is valid JSON
func isValidJSON(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

// formatJSON formats a JSON string with indentation
func formatJSON(s string) string {
	var parsed interface{}
	if err := json.Unmarshal([]byte(s), &parsed); err != nil {
		return s
	}
	formatted, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return s
	}
	return string(formatted) + "\n"
}
