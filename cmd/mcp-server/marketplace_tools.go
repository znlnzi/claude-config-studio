package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/znlnzi/claude-config-studio/internal/marketplace"
	"github.com/znlnzi/claude-config-studio/internal/templatedata"
)

// ─── Tool Definitions ─────────────────────────────────────────

func buildMarketplaceSearchTool() mcp.Tool {
	return mcp.NewTool(
		"marketplace_search",
		mcp.WithDescription("Search community-contributed templates from the marketplace registry. Returns metadata including id, name, description, author, version, tags, downloads, and stars."),
		mcp.WithString("query",
			mcp.Description("Search keyword to filter templates by name, description, tags, or author"),
		),
		mcp.WithString("category",
			mcp.Description("Filter by category (e.g., 'Frontend', 'Backend')"),
		),
	)
}

func buildMarketplaceInstallTool() mcp.Tool {
	return mcp.NewTool(
		"marketplace_install",
		mcp.WithDescription("Install a community template from the marketplace registry to a project or global scope."),
		mcp.WithString("template_id",
			mcp.Required(),
			mcp.Description("Template ID from the marketplace registry (e.g., 'community/react-testing')"),
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

func buildMarketplaceInfoTool() mcp.Tool {
	return mcp.NewTool(
		"marketplace_info",
		mcp.WithDescription("Get detailed information about a community template from the marketplace, including content summary (agents, commands, rules, skills)."),
		mcp.WithString("template_id",
			mcp.Required(),
			mcp.Description("Template ID from the marketplace registry (e.g., 'community/react-testing')"),
		),
	)
}

// ─── Handlers ─────────────────────────────────────────────────

func handleMarketplaceSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := req.GetString("query", "")
	category := req.GetString("category", "")

	client := newMarketplaceClient()
	idx, err := client.FetchIndex(ctx)
	if err != nil {
		return mcp.NewToolResultError(errMarketplaceFetch(err)), nil
	}

	results := idx.Templates
	if category != "" {
		results = marketplace.FilterByCategory(results, category)
	}
	if query != "" {
		results = marketplace.Search(results, query)
	}

	output := map[string]interface{}{
		"templates":  results,
		"total":      len(results),
		"updated_at": idx.UpdatedAt,
	}
	out, _ := json.Marshal(output)
	return mcp.NewToolResultText(string(out)), nil
}

func handleMarketplaceInstall(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	templateID, err := req.RequireString("template_id")
	if err != nil {
		return mcp.NewToolResultError("template_id is required"), nil
	}
	scope := req.GetString("scope", "project")
	projectPath := req.GetString("project_path", "")
	overwrite := req.GetString("overwrite", "false") == "true"

	// Resolve .claude directory
	claudeDir, errMsg := resolveClaudeDir(scope, projectPath)
	if errMsg != "" {
		return mcp.NewToolResultError(errMsg), nil
	}

	// Find the template in the registry index
	client := newMarketplaceClient()
	idx, err := client.FetchIndex(ctx)
	if err != nil {
		return mcp.NewToolResultError(errMarketplaceFetch(err)), nil
	}

	var entry *marketplace.IndexEntry
	for i := range idx.Templates {
		if idx.Templates[i].ID == templateID {
			entry = &idx.Templates[i]
			break
		}
	}
	if entry == nil {
		return mcp.NewToolResultError(errMarketplaceTemplateNotFound(templateID)), nil
	}

	// Fetch the full template
	tmpl, err := client.FetchTemplate(ctx, entry.URL)
	if err != nil {
		return mcp.NewToolResultError(errMarketplaceInvalidTemplate(templateID, err)), nil
	}

	// Install using the shared function
	installedFiles, installErr := installTemplate(claudeDir, tmpl, sanitizeTemplateID(templateID), overwrite)
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
		"source":          "marketplace",
	}
	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

func handleMarketplaceInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	templateID, err := req.RequireString("template_id")
	if err != nil {
		return mcp.NewToolResultError("template_id is required"), nil
	}

	client := newMarketplaceClient()
	idx, err := client.FetchIndex(ctx)
	if err != nil {
		return mcp.NewToolResultError(errMarketplaceFetch(err)), nil
	}

	var entry *marketplace.IndexEntry
	for i := range idx.Templates {
		if idx.Templates[i].ID == templateID {
			entry = &idx.Templates[i]
			break
		}
	}
	if entry == nil {
		return mcp.NewToolResultError(errMarketplaceTemplateNotFound(templateID)), nil
	}

	// Fetch the full template for content summary
	tmpl, err := client.FetchTemplate(ctx, entry.URL)
	if err != nil {
		return mcp.NewToolResultError(errMarketplaceInvalidTemplate(templateID, err)), nil
	}

	// Build content summary
	contents := map[string]interface{}{}
	if tmpl.ClaudeMd != "" {
		contents["has_rules_file"] = true
	}
	if len(tmpl.Agents) > 0 {
		contents["agents"] = mapKeys(tmpl.Agents)
	}
	if len(tmpl.Commands) > 0 {
		contents["commands"] = mapKeys(tmpl.Commands)
	}
	if len(tmpl.Skills) > 0 {
		contents["skills"] = mapKeys(tmpl.Skills)
	}
	if len(tmpl.Rules) > 0 {
		contents["rules"] = mapKeys(tmpl.Rules)
	}
	if len(tmpl.Scripts) > 0 {
		contents["scripts"] = mapKeys(tmpl.Scripts)
	}
	if tmpl.Settings != nil {
		contents["has_settings"] = true
	}

	result := map[string]interface{}{
		"id":          entry.ID,
		"name":        entry.Name,
		"description": entry.Description,
		"author":      entry.Author,
		"version":     entry.Version,
		"category":    entry.Category,
		"tags":        entry.Tags,
		"downloads":   entry.Downloads,
		"stars":       entry.Stars,
		"contents":    contents,
	}
	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

// ─── Shared Helpers ───────────────────────────────────────────

// resolveClaudeDir determines the .claude directory path based on scope and project path.
// Returns (claudeDir, errorMessage). errorMessage is empty on success.
func resolveClaudeDir(scope, projectPath string) (string, string) {
	if scope == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errHomeDir(err)
		}
		return filepath.Join(home, ".claude"), ""
	}
	if projectPath == "" {
		return "", "project_path is required for project scope"
	}
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return "", errPathNotFound(projectPath)
	}
	return filepath.Join(projectPath, ".claude"), ""
}

// installTemplate installs a template into the given .claude directory.
// filePrefix is used for the tpl-{prefix}.md rules file name.
// Returns (installedFiles, errorMessage). errorMessage is empty on success.
func installTemplate(claudeDir string, tmpl *templatedata.Template, filePrefix string, overwrite bool) ([]string, string) {
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return nil, errCreateDir(claudeDir, err)
	}

	var installedFiles []string

	// Install rules file (tpl-{prefix}.md)
	if tmpl.ClaudeMd != "" {
		rulesDir := filepath.Join(claudeDir, "rules")
		if err := os.MkdirAll(rulesDir, 0755); err != nil {
			return nil, errCreateDir(rulesDir, err)
		}
		header := fmt.Sprintf("<!-- template: %s | %s -->\n\n", tmpl.ID, tmpl.Name)
		content := header + tmpl.ClaudeMd
		filePath := filepath.Join(rulesDir, "tpl-"+filePrefix+".md")
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return nil, errWriteFailed(filePath, err)
		}
		installedFiles = append(installedFiles, "rules/tpl-"+filePrefix+".md")
	}

	// Install agents
	if len(tmpl.Agents) > 0 {
		if err := templatedata.WriteExtensionFiles(claudeDir, "agents", tmpl.Agents, overwrite); err != nil {
			return nil, fmt.Sprintf("failed to write agents: %v. Check directory permissions for %s/agents/", err, claudeDir)
		}
		for name := range tmpl.Agents {
			installedFiles = append(installedFiles, "agents/"+name+".md")
		}
	}

	// Install commands
	if len(tmpl.Commands) > 0 {
		if err := templatedata.WriteExtensionFiles(claudeDir, "commands", tmpl.Commands, overwrite); err != nil {
			return nil, fmt.Sprintf("failed to write commands: %v. Check directory permissions for %s/commands/", err, claudeDir)
		}
		for name := range tmpl.Commands {
			installedFiles = append(installedFiles, "commands/"+name+".md")
		}
	}

	// Install skills
	if len(tmpl.Skills) > 0 {
		if err := templatedata.WriteSkillFiles(claudeDir, tmpl.Skills, overwrite); err != nil {
			return nil, fmt.Sprintf("failed to write skills: %v. Check directory permissions for %s/skills/", err, claudeDir)
		}
		for name := range tmpl.Skills {
			installedFiles = append(installedFiles, "skills/"+name+"/SKILL.md")
		}
	}

	// Install additional rules
	if len(tmpl.Rules) > 0 {
		if err := templatedata.WriteExtensionFiles(claudeDir, "rules", tmpl.Rules, overwrite); err != nil {
			return nil, fmt.Sprintf("failed to write rules: %v. Check directory permissions for %s/rules/", err, claudeDir)
		}
		for name := range tmpl.Rules {
			installedFiles = append(installedFiles, "rules/"+name+".md")
		}
	}

	// Install scripts
	if len(tmpl.Scripts) > 0 {
		scriptsDir := filepath.Join(claudeDir, "scripts")
		if err := os.MkdirAll(scriptsDir, 0755); err != nil {
			return nil, errCreateDir(scriptsDir, err)
		}
		for name, content := range tmpl.Scripts {
			scriptPath := filepath.Join(scriptsDir, name)
			if !overwrite {
				if _, err := os.Stat(scriptPath); err == nil {
					continue
				}
			}
			if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
				return nil, errWriteFailed(scriptPath, err)
			}
			installedFiles = append(installedFiles, "scripts/"+name)
		}
	}

	// Merge settings.json
	if tmpl.Settings != nil {
		if err := templatedata.MergeAndWriteJSON(filepath.Join(claudeDir, "settings.json"), tmpl.Settings); err != nil {
			return nil, fmt.Sprintf("failed to merge settings.json: %v. Check file permissions for %s/settings.json", err, claudeDir)
		}
		installedFiles = append(installedFiles, "settings.json (merged)")
	}

	return installedFiles, ""
}

// sanitizeTemplateID replaces slashes in template IDs for safe filenames.
func sanitizeTemplateID(id string) string {
	return strings.ReplaceAll(id, "/", "-")
}

// newMarketplaceClient creates a marketplace client, respecting the env override.
func newMarketplaceClient() *marketplace.Client {
	url := os.Getenv("CLAUDE_MARKETPLACE_REGISTRY_URL")
	return marketplace.NewClient(url)
}

// mapKeys returns the keys of a map as a sorted slice.
func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
