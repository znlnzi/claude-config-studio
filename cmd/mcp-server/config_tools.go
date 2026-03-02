package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func buildGetProjectConfigTool() mcp.Tool {
	return mcp.NewTool(
		"get_project_config",
		mcp.WithDescription("Get an overview of a project's Claude Code configuration (.claude/ directory contents)."),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("Absolute project path"),
		),
	)
}

func buildListProjectsTool() mcp.Tool {
	return mcp.NewTool(
		"list_projects",
		mcp.WithDescription("List all Claude Code managed projects found in ~/.claude/projects/."),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of projects to return (default 50)"),
		),
	)
}

// handleGetProjectConfig handles the get_project_config tool call
func handleGetProjectConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectPath, err := req.RequireString("project_path")
	if err != nil {
		return mcp.NewToolResultError("project_path is required"), nil
	}

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("project path does not exist: %s", projectPath)), nil
	}

	claudeDir := filepath.Join(projectPath, ".claude")

	hasClaudeMd := fileExists(filepath.Join(claudeDir, "CLAUDE.md")) ||
		fileExists(filepath.Join(projectPath, "CLAUDE.md"))
	claudeMdPreview := ""
	if hasClaudeMd {
		mdPath := filepath.Join(claudeDir, "CLAUDE.md")
		if !fileExists(mdPath) {
			mdPath = filepath.Join(projectPath, "CLAUDE.md")
		}
		if data, err := os.ReadFile(mdPath); err == nil {
			content := string(data)
			if len(content) > 500 {
				claudeMdPreview = content[:500] + "..."
			} else {
				claudeMdPreview = content
			}
		}
	}

	hasSettings := fileExists(filepath.Join(claudeDir, "settings.json"))
	hasMcp := fileExists(filepath.Join(claudeDir, ".mcp.json"))
	hasHooks := dirHasFiles(filepath.Join(claudeDir, "hooks"))
	hasAgents := dirHasFiles(filepath.Join(claudeDir, "agents"))
	hasCommands := dirHasFiles(filepath.Join(claudeDir, "commands"))
	hasSkills := dirHasFiles(filepath.Join(claudeDir, "skills"))
	hasMemory := dirHasFiles(filepath.Join(claudeDir, "memory"))
	hasRules := dirHasFiles(filepath.Join(claudeDir, "rules"))

	configCount := 0
	for _, has := range []bool{hasClaudeMd, hasSettings, hasMcp, hasHooks, hasAgents, hasCommands, hasSkills, hasRules} {
		if has {
			configCount++
		}
	}

	result := map[string]interface{}{
		"path":              projectPath,
		"name":              filepath.Base(projectPath),
		"has_claude_md":     hasClaudeMd,
		"claude_md_preview": claudeMdPreview,
		"has_settings":      hasSettings,
		"has_mcp":           hasMcp,
		"has_hooks":         hasHooks,
		"has_agents":        hasAgents,
		"has_commands":      hasCommands,
		"has_skills":        hasSkills,
		"has_rules":         hasRules,
		"has_memory":        hasMemory,
		"config_count":      configCount,
	}

	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

// handleListProjects handles the list_projects tool call
func handleListProjects(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := req.GetInt("limit", 50)

	projDir, err := getProjectsDir()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get projects dir: %v", err)), nil
	}

	entries, err := os.ReadDir(projDir)
	if err != nil {
		result := map[string]interface{}{
			"projects": []interface{}{},
			"total":    0,
		}
		out, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(out)), nil
	}

	type projectEntry struct {
		Name         string `json:"name"`
		Path         string `json:"path"`
		ConfigCount  int    `json:"config_count"`
		HasMemory    bool   `json:"has_memory"`
		LastModified string `json:"last_modified"`
	}

	var projects []projectEntry

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		realPath := decodeProjectPath(entry.Name())
		if realPath == "" {
			continue
		}
		if _, err := os.Stat(realPath); os.IsNotExist(err) {
			continue
		}

		claudeDir := filepath.Join(realPath, ".claude")

		configCount := 0
		for _, check := range []string{
			filepath.Join(claudeDir, "CLAUDE.md"),
			filepath.Join(claudeDir, "settings.json"),
			filepath.Join(claudeDir, ".mcp.json"),
		} {
			if fileExists(check) {
				configCount++
			}
		}
		if !fileExists(filepath.Join(claudeDir, "CLAUDE.md")) && fileExists(filepath.Join(realPath, "CLAUDE.md")) {
			configCount++
		}
		for _, subDir := range []string{"hooks", "agents", "commands", "skills", "rules"} {
			if dirHasFiles(filepath.Join(claudeDir, subDir)) {
				configCount++
			}
		}

		hasMemory := dirHasFiles(filepath.Join(claudeDir, "memory"))

		modTime := ""
		if fi, err := entry.Info(); err == nil {
			modTime = fi.ModTime().Format(time.RFC3339)
		}

		projects = append(projects, projectEntry{
			Name:         filepath.Base(realPath),
			Path:         realPath,
			ConfigCount:  configCount,
			HasMemory:    hasMemory,
			LastModified: modTime,
		})
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastModified > projects[j].LastModified
	})

	total := len(projects)
	if len(projects) > limit {
		projects = projects[:limit]
	}

	result := map[string]interface{}{
		"projects": projects,
		"total":    total,
	}
	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}
