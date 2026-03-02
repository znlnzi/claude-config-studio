package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/znlnzi/claude-config-studio/internal/exporter"

	"github.com/mark3labs/mcp-go/mcp"
)

// ─── export_config tool ─────────────────────────────

func buildExportConfigTool() mcp.Tool {
	return mcp.NewTool(
		"export_config",
		mcp.WithDescription("Export Claude Code configuration as a base64-encoded ZIP. Use scope='global' for ~/.claude/ or scope='project' with project_path for project-level config."),
		mcp.WithString("scope",
			mcp.Required(),
			mcp.Description("Export scope: 'global' or 'project'"),
		),
		mcp.WithString("project_path",
			mcp.Description("Absolute project path (required when scope is 'project')"),
		),
	)
}

func handleExportConfig(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scope, err := req.RequireString("scope")
	if err != nil {
		return mcp.NewToolResultError("scope parameter is required"), nil
	}

	var result *exporter.ExportResult

	switch scope {
	case "global":
		result, err = exporter.ExportGlobalConfig()
	case "project":
		projectPath := req.GetString("project_path", "")
		if projectPath == "" {
			return mcp.NewToolResultError("project_path is required when scope is 'project'"), nil
		}
		if _, statErr := os.Stat(projectPath); os.IsNotExist(statErr) {
			return mcp.NewToolResultError(errPathNotFound(projectPath)), nil
		}
		result, err = exporter.ExportProjectConfig(projectPath)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("invalid scope: %s (must be 'global' or 'project')", scope)), nil
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("export failed: %v. Check that the .claude/ directory exists and has readable files", err)), nil
	}

	output := map[string]interface{}{
		"success": true,
		"format":  "base64-zip",
		"size":    result.Size,
		"files":   result.Files,
		"data":    result.Data,
	}
	data, _ := json.MarshalIndent(output, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ─── import_config tool ─────────────────────────────

func buildImportConfigTool() mcp.Tool {
	return mcp.NewTool(
		"import_config",
		mcp.WithDescription("Import Claude Code configuration from a base64-encoded ZIP. Provide target_path to specify where to extract (e.g., ~/.claude/ for global, or project root for project-level)."),
		mcp.WithString("target_path",
			mcp.Required(),
			mcp.Description("Absolute path to import into (e.g., '~/.claude/' for global config, or project root path)"),
		),
		mcp.WithString("data",
			mcp.Required(),
			mcp.Description("Base64-encoded ZIP data (from export_config)"),
		),
	)
}

func handleImportConfig(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetPath, err := req.RequireString("target_path")
	if err != nil {
		return mcp.NewToolResultError("target_path parameter is required"), nil
	}
	base64Data, err := req.RequireString("data")
	if err != nil {
		return mcp.NewToolResultError("data parameter is required"), nil
	}

	// Resolve ~ to home directory
	if targetPath == "~" || targetPath == "~/" {
		home, _ := os.UserHomeDir()
		targetPath = home
	} else if len(targetPath) > 2 && targetPath[:2] == "~/" {
		home, _ := os.UserHomeDir()
		targetPath = filepath.Join(home, targetPath[2:])
	}

	result, err := exporter.ImportConfig(targetPath, base64Data)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("import failed: %v. Check that the target path exists and the base64 data is valid", err)), nil
	}

	output := map[string]interface{}{
		"success":        true,
		"files_imported": result.FilesImported,
		"file_names":     result.FileNames,
	}
	data, _ := json.MarshalIndent(output, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
