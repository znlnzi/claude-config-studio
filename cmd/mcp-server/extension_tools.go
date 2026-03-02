package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func buildListExtensionsTool() mcp.Tool {
	return mcp.NewTool(
		"extension_list",
		mcp.WithDescription("List extension files (agents, rules, skills, or commands) in a given scope."),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Extension type: 'agents', 'rules', 'skills', or 'commands'"),
		),
		mcp.WithString("scope",
			mcp.Description("Scope: 'global' (default) or absolute project path"),
		),
	)
}

func buildReadExtensionTool() mcp.Tool {
	return mcp.NewTool(
		"extension_read",
		mcp.WithDescription("Read the content of a specific extension file."),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Extension type: 'agents', 'rules', 'skills', or 'commands'"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Extension name (without .md extension)"),
		),
		mcp.WithString("scope",
			mcp.Description("Scope: 'global' (default) or absolute project path"),
		),
	)
}

func buildSaveExtensionTool() mcp.Tool {
	return mcp.NewTool(
		"extension_save",
		mcp.WithDescription("Create or update an extension file (agent, rule, skill, or command)."),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Extension type: 'agents', 'rules', 'skills', or 'commands'"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Extension name (without .md extension)"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("File content to write"),
		),
		mcp.WithString("scope",
			mcp.Description("Scope: 'global' (default) or absolute project path"),
		),
	)
}

func buildDeleteExtensionTool() mcp.Tool {
	return mcp.NewTool(
		"extension_delete",
		mcp.WithDescription("Delete an extension file."),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Extension type: 'agents', 'rules', 'skills', or 'commands'"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Extension name (without .md extension)"),
		),
		mcp.WithString("scope",
			mcp.Description("Scope: 'global' (default) or absolute project path"),
		),
	)
}

// resolveExtensionDir resolves the extension file directory.
// extType: "agents", "rules", "skills", "commands"
// scope: "global" or an absolute project path
func resolveExtensionDir(extType, scope string) (string, error) {
	if extType != "agents" && extType != "rules" && extType != "skills" && extType != "commands" {
		return "", fmt.Errorf("invalid extension type: %s (must be agents, rules, skills, or commands)", extType)
	}

	var claudeDir string
	if scope == "" || scope == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		claudeDir = filepath.Join(home, ".claude")
	} else {
		claudeDir = filepath.Join(scope, ".claude")
	}

	return filepath.Join(claudeDir, extType), nil
}

// handleListExtensions lists extension files
func handleListExtensions(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	extType, err := req.RequireString("type")
	if err != nil {
		return mcp.NewToolResultError("type is required (agents, rules, skills, commands)"), nil
	}
	scope := req.GetString("scope", "global")

	extDir, err := resolveExtensionDir(extType, scope)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	entries, err := os.ReadDir(extDir)
	if err != nil {
		result := map[string]interface{}{
			"files": []interface{}{},
			"type":  extType,
			"scope": scope,
		}
		out, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(out)), nil
	}

	type extensionEntry struct {
		Name       string `json:"name"`
		FileName   string `json:"filename"`
		IsDir      bool   `json:"is_dir"`
		Size       int64  `json:"size"`
		ModifiedAt string `json:"modified_at"`
	}

	var files []extensionEntry
	for _, e := range entries {
		fi, err := e.Info()
		if err != nil {
			continue
		}

		name := e.Name()

		// Skills use directory-based format
		if e.IsDir() && extType == "skills" {
			skillFile := filepath.Join(extDir, name, "SKILL.md")
			if skillFi, err := os.Stat(skillFile); err == nil {
				files = append(files, extensionEntry{
					Name:       name,
					FileName:   name + "/SKILL.md",
					IsDir:      true,
					Size:       skillFi.Size(),
					ModifiedAt: skillFi.ModTime().Format(time.RFC3339),
				})
			}
			continue
		}

		// Regular .md files
		if !e.IsDir() && strings.HasSuffix(name, ".md") {
			files = append(files, extensionEntry{
				Name:       strings.TrimSuffix(name, ".md"),
				FileName:   name,
				IsDir:      false,
				Size:       fi.Size(),
				ModifiedAt: fi.ModTime().Format(time.RFC3339),
			})
		}
	}

	result := map[string]interface{}{
		"files": files,
		"type":  extType,
		"scope": scope,
		"total": len(files),
	}
	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

// handleReadExtension reads the content of a single extension file
func handleReadExtension(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	extType, err := req.RequireString("type")
	if err != nil {
		return mcp.NewToolResultError("type is required (agents, rules, skills, commands)"), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("name is required"), nil
	}
	scope := req.GetString("scope", "global")

	if !isSafeFilename(name) {
		return mcp.NewToolResultError("unsafe name: must not contain '..' or path separators"), nil
	}

	extDir, err := resolveExtensionDir(extType, scope)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Try reading: directory-based (skills) first, then flat file
	var content string
	var filePath string

	if extType == "skills" {
		skillPath := filepath.Join(extDir, name, "SKILL.md")
		if data, err := os.ReadFile(skillPath); err == nil {
			content = string(data)
			filePath = skillPath
		}
	}

	if content == "" {
		flatPath := filepath.Join(extDir, name+".md")
		if data, err := os.ReadFile(flatPath); err == nil {
			content = string(data)
			filePath = flatPath
		}
	}

	if content == "" {
		return mcp.NewToolResultError(fmt.Sprintf("extension not found: %s/%s", extType, name)), nil
	}

	result := map[string]interface{}{
		"name":    name,
		"type":    extType,
		"scope":   scope,
		"content": content,
		"path":    filePath,
	}
	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

// handleSaveExtension creates or updates an extension file
func handleSaveExtension(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	extType, err := req.RequireString("type")
	if err != nil {
		return mcp.NewToolResultError("type is required (agents, rules, skills, commands)"), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("name is required"), nil
	}
	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("content is required"), nil
	}
	scope := req.GetString("scope", "global")

	if !isSafeFilename(name) {
		return mcp.NewToolResultError("unsafe name: must not contain '..' or path separators"), nil
	}

	extDir, err := resolveExtensionDir(extType, scope)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := os.MkdirAll(extDir, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create directory: %v", err)), nil
	}

	var filePath string
	if extType == "skills" {
		// Skills use directory-based format
		skillDir := filepath.Join(extDir, name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create skill dir: %v", err)), nil
		}
		filePath = filepath.Join(skillDir, "SKILL.md")
	} else {
		filePath = filepath.Join(extDir, name+".md")
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to write: %v", err)), nil
	}

	result := map[string]interface{}{
		"success": true,
		"name":    name,
		"type":    extType,
		"scope":   scope,
		"path":    filePath,
	}
	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}

// handleDeleteExtension deletes an extension file
func handleDeleteExtension(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	extType, err := req.RequireString("type")
	if err != nil {
		return mcp.NewToolResultError("type is required (agents, rules, skills, commands)"), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("name is required"), nil
	}
	scope := req.GetString("scope", "global")

	if !isSafeFilename(name) {
		return mcp.NewToolResultError("unsafe name: must not contain '..' or path separators"), nil
	}

	extDir, err := resolveExtensionDir(extType, scope)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var removed bool

	// Try deleting directory-based (skills)
	if extType == "skills" {
		skillDir := filepath.Join(extDir, name)
		if _, err := os.Stat(skillDir); err == nil {
			if err := os.RemoveAll(skillDir); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to delete skill dir: %v", err)), nil
			}
			removed = true
		}
	}

	// Try deleting flat file
	if !removed {
		flatPath := filepath.Join(extDir, name+".md")
		if _, err := os.Stat(flatPath); err == nil {
			if err := os.Remove(flatPath); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to delete: %v", err)), nil
			}
			removed = true
		}
	}

	if !removed {
		return mcp.NewToolResultError(fmt.Sprintf("extension not found: %s/%s", extType, name)), nil
	}

	result := map[string]interface{}{
		"success": true,
		"name":    name,
		"type":    extType,
		"scope":   scope,
	}
	out, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(out)), nil
}
