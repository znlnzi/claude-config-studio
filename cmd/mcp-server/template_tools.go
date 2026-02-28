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

// handleInstallTemplate installs a template to project or global scope
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
		return mcp.NewToolResultError(fmt.Sprintf("template not found: %s", templateID)), nil
	}

	// Determine .claude directory
	var claudeDir string
	if scope == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get home dir: %v", err)), nil
		}
		claudeDir = filepath.Join(home, ".claude")
	} else {
		if projectPath == "" {
			return mcp.NewToolResultError("project_path is required for project scope"), nil
		}
		if _, err := os.Stat(projectPath); os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("project path does not exist: %s", projectPath)), nil
		}
		claudeDir = filepath.Join(projectPath, ".claude")
	}

	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create .claude dir: %v", err)), nil
	}
	var installedFiles []string

	// Install rules file (tpl-{id}.md)
	if tmpl.ClaudeMd != "" {
		rulesDir := filepath.Join(claudeDir, "rules")
		if err := os.MkdirAll(rulesDir, 0755); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create rules dir: %v", err)), nil
		}
		header := fmt.Sprintf("<!-- template: %s | %s -->\n\n", tmpl.ID, tmpl.Name)
		content := header + tmpl.ClaudeMd
		filePath := filepath.Join(rulesDir, "tpl-"+tmpl.ID+".md")
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to write template rule: %v", err)), nil
		}
		installedFiles = append(installedFiles, "rules/tpl-"+tmpl.ID+".md")
	}

	// Install agents
	if len(tmpl.Agents) > 0 {
		if err := templatedata.WriteExtensionFiles(claudeDir, "agents", tmpl.Agents, overwrite); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to write agents: %v", err)), nil
		}
		for name := range tmpl.Agents {
			installedFiles = append(installedFiles, "agents/"+name+".md")
		}
	}

	// Install commands
	if len(tmpl.Commands) > 0 {
		if err := templatedata.WriteExtensionFiles(claudeDir, "commands", tmpl.Commands, overwrite); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to write commands: %v", err)), nil
		}
		for name := range tmpl.Commands {
			installedFiles = append(installedFiles, "commands/"+name+".md")
		}
	}

	// Install skills
	if len(tmpl.Skills) > 0 {
		if err := templatedata.WriteSkillFiles(claudeDir, tmpl.Skills, overwrite); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to write skills: %v", err)), nil
		}
		for name := range tmpl.Skills {
			installedFiles = append(installedFiles, "skills/"+name+"/SKILL.md")
		}
	}

	// Install additional rules bundled with the template
	if len(tmpl.Rules) > 0 {
		if err := templatedata.WriteExtensionFiles(claudeDir, "rules", tmpl.Rules, overwrite); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to write rules: %v", err)), nil
		}
		for name := range tmpl.Rules {
			installedFiles = append(installedFiles, "rules/"+name+".md")
		}
	}

	// Install scripts
	if len(tmpl.Scripts) > 0 {
		scriptsDir := filepath.Join(claudeDir, "scripts")
		if err := os.MkdirAll(scriptsDir, 0755); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create scripts dir: %v", err)), nil
		}
		for name, content := range tmpl.Scripts {
			scriptPath := filepath.Join(scriptsDir, name)
			if !overwrite {
				if _, err := os.Stat(scriptPath); err == nil {
					continue
				}
			}
			if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to write script %s: %v", name, err)), nil
			}
			installedFiles = append(installedFiles, "scripts/"+name)
		}
	}

	// Merge settings.json
	if tmpl.Settings != nil {
		if err := templatedata.MergeAndWriteJSON(filepath.Join(claudeDir, "settings.json"), tmpl.Settings); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to merge settings: %v", err)), nil
		}
		installedFiles = append(installedFiles, "settings.json (merged)")
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
			return mcp.NewToolResultError(fmt.Sprintf("failed to get home dir: %v", err)), nil
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
			return mcp.NewToolResultError(fmt.Sprintf("failed to get home dir: %v", err)), nil
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
