package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/znlnzi/claude-config-studio/internal/templatedata"
)

func buildListTemplatesTool() mcp.Tool {
	return mcp.NewTool(
		"template_list",
		mcp.WithDescription("List all available Claude Code configuration templates with categories, names, and descriptions."),
	)
}

func buildInstallTemplateTool() mcp.Tool {
	return mcp.NewTool(
		"template_install",
		mcp.WithDescription("Install a configuration template to a project or global scope. Writes rules, agents, skills, and commands files."),
		mcp.WithString("template_id",
			mcp.Required(),
			mcp.Description("Template ID to install (e.g., 'hackathon-core', 'cross-session-memory')"),
		),
		mcp.WithString("scope",
			mcp.Description("Install scope: 'project' (default) or 'global'"),
		),
		mcp.WithString("project_path",
			mcp.Description("Absolute project path (required for project scope)"),
		),
		mcp.WithString("overwrite",
			mcp.Description("Set to 'true' to overwrite existing files (default: 'false')"),
		),
	)
}

func buildUninstallTemplateTool() mcp.Tool {
	return mcp.NewTool(
		"template_uninstall",
		mcp.WithDescription("Uninstall a template by removing its rules file (tpl-{id}.md)."),
		mcp.WithString("template_id",
			mcp.Required(),
			mcp.Description("Template ID to uninstall"),
		),
		mcp.WithString("scope",
			mcp.Description("Scope: 'project' (default) or 'global'"),
		),
		mcp.WithString("project_path",
			mcp.Description("Absolute project path (required for project scope)"),
		),
	)
}

func buildGetInstalledTemplatesTool() mcp.Tool {
	return mcp.NewTool(
		"template_installed",
		mcp.WithDescription("List installed templates in a project or global scope by scanning for tpl-*.md files in the rules directory."),
		mcp.WithString("scope",
			mcp.Description("Scope: 'project' (default) or 'global'"),
		),
		mcp.WithString("project_path",
			mcp.Description("Absolute project path (required for project scope)"),
		),
	)
}

// handleListTemplates lists all available templates
func handleListTemplates(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	categories := templatedata.GetAllTemplates()

	type templateMeta struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Category    string   `json:"category"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}

	type categoryMeta struct {
		ID        string         `json:"id"`
		Name      string         `json:"name"`
		Icon      string         `json:"icon"`
		Templates []templateMeta `json:"templates"`
	}

	var result []categoryMeta
	totalTemplates := 0

	for _, cat := range categories {
		var templates []templateMeta
		for _, t := range cat.Templates {
			templates = append(templates, templateMeta{
				ID:          t.ID,
				Name:        t.Name,
				Category:    t.Category,
				Description: t.Description,
				Tags:        t.Tags,
			})
			totalTemplates++
		}
		result = append(result, categoryMeta{
			ID:        cat.ID,
			Name:      cat.Name,
			Icon:      cat.Icon,
			Templates: templates,
		})
	}

	output := map[string]interface{}{
		"categories":      result,
		"total_templates": totalTemplates,
	}
	out, _ := json.Marshal(output)
	return mcp.NewToolResultText(string(out)), nil
}

// handleInstallTemplate installs a builtin template to project or global scope
func handleInstallTemplate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	templateID, err := req.RequireString("template_id")
	if err != nil {
		return mcp.NewToolResultError("template_id is required"), nil
	}
	scope := req.GetString("scope", "project")
	projectPath := req.GetString("project_path", "")
	overwrite := req.GetString("overwrite", "false") == "true"

	tmpl := templatedata.GetTemplateByID(templateID)
	if tmpl == nil {
		return mcp.NewToolResultError(errTemplateNotFound(templateID)), nil
	}

	claudeDir, errMsg := resolveClaudeDir(scope, projectPath)
	if errMsg != "" {
		return mcp.NewToolResultError(errMsg), nil
	}

	installedFiles, installErr := installTemplate(claudeDir, tmpl, tmpl.ID, overwrite)
	if installErr != "" {
		return mcp.NewToolResultError(installErr), nil
	}

	result := map[string]interface{}{
		"success":         true,
		"template_id":     templateID,
		"template_name":   tmpl.Name,
		"scope":           scope,
		"installed_files": installedFiles,
		"total_files":     len(installedFiles),
	}
	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

// handleUninstallTemplate uninstalls a template
func handleUninstallTemplate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	templateID, err := req.RequireString("template_id")
	if err != nil {
		return mcp.NewToolResultError("template_id is required"), nil
	}
	scope := req.GetString("scope", "project")
	projectPath := req.GetString("project_path", "")

	// Determine rules directory
	var rulesDir string
	if scope == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return mcp.NewToolResultError(errHomeDir(err)), nil
		}
		rulesDir = filepath.Join(home, ".claude", "rules")
	} else {
		if projectPath == "" {
			return mcp.NewToolResultError("project_path is required for project scope"), nil
		}
		rulesDir = filepath.Join(projectPath, ".claude", "rules")
	}

	// Delete tpl-{id}.md file
	filePath := filepath.Join(rulesDir, "tpl-"+templateID+".md")
	var removed bool
	if _, err := os.Stat(filePath); err == nil {
		os.Remove(filePath)
		removed = true
	}

	result := map[string]interface{}{
		"success":     removed,
		"template_id": templateID,
		"scope":       scope,
		"removed":     removed,
	}
	if !removed {
		result["message"] = fmt.Sprintf("template rule file not found: tpl-%s.md", templateID)
	}

	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

// handleGetInstalledTemplates returns the list of installed templates
func handleGetInstalledTemplates(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scope := req.GetString("scope", "project")
	projectPath := req.GetString("project_path", "")

	var rulesDir string
	if scope == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return mcp.NewToolResultError(errHomeDir(err)), nil
		}
		rulesDir = filepath.Join(home, ".claude", "rules")
	} else {
		if projectPath == "" {
			return mcp.NewToolResultError("project_path is required for project scope"), nil
		}
		rulesDir = filepath.Join(projectPath, ".claude", "rules")
	}

	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		result := map[string]interface{}{
			"installed": []interface{}{},
			"scope":     scope,
		}
		out, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(out)), nil
	}

	type installedInfo struct {
		TemplateID string `json:"template_id"`
		Name       string `json:"name"`
		FilePath   string `json:"file_path"`
	}

	var installed []installedInfo
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() && strings.HasPrefix(name, "tpl-") && strings.HasSuffix(name, ".md") {
			tid := strings.TrimSuffix(strings.TrimPrefix(name, "tpl-"), ".md")
			displayName := tid
			if tmpl := templatedata.GetTemplateByID(tid); tmpl != nil {
				displayName = tmpl.Name
			}
			installed = append(installed, installedInfo{
				TemplateID: tid,
				Name:       displayName,
				FilePath:   filepath.Join(rulesDir, name),
			})
		}
	}

	result := map[string]interface{}{
		"installed": installed,
		"scope":     scope,
		"total":     len(installed),
	}
	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}
