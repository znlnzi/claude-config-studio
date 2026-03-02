package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerResources registers all MCP Resources (static and templates) on the server
func registerResources(s *server.MCPServer) {
	// Static resource: global CLAUDE.md
	s.AddResource(
		mcp.NewResource(
			"claude://global/claude-md",
			"Global CLAUDE.md",
			mcp.WithResourceDescription("Global Claude Code instructions (~/.claude/CLAUDE.md)"),
			mcp.WithMIMEType("text/markdown"),
		),
		handleGlobalClaudeMD,
	)

	// Template: global memory files
	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"claude://global/memory/{filename}",
			"Global Memory File",
			mcp.WithTemplateDescription("Memory files in ~/.claude/memory/ directory"),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		handleGlobalMemoryFile,
	)

	// Template: project memory files
	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"claude://project/{project_path}/memory/{filename}",
			"Project Memory File",
			mcp.WithTemplateDescription("Memory files in a project's .claude/memory/ directory"),
			mcp.WithTemplateMIMEType("text/markdown"),
		),
		handleProjectMemoryFile,
	)
}

// handleGlobalClaudeMD reads ~/.claude/CLAUDE.md
func handleGlobalClaudeMD(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve home directory: %w. Check $HOME environment variable", err)
	}

	filePath := filepath.Join(home, ".claude", "CLAUDE.md")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      "claude://global/claude-md",
					MIMEType: "text/markdown",
					Text:     "# Global CLAUDE.md\n\n(File does not exist yet. Create ~/.claude/CLAUDE.md to add global instructions.)",
				},
			}, nil
		}
		return nil, fmt.Errorf("failed to read CLAUDE.md: %w. Check file permissions", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "claude://global/claude-md",
			MIMEType: "text/markdown",
			Text:     string(data),
		},
	}, nil
}

// handleGlobalMemoryFile reads a file from ~/.claude/memory/
func handleGlobalMemoryFile(_ context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	filename := extractURIParam(req.Params.URI, "claude://global/memory/")
	if !isResourceSafeFilename(filename) {
		return nil, fmt.Errorf("invalid filename: %s. Filename must end in .md and cannot contain '..' or path separators", filename)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve home directory: %w. Check $HOME environment variable", err)
	}

	filePath := filepath.Join(home, ".claude", "memory", filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read memory file %s: %w. Check that the file exists", filename, err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "text/markdown",
			Text:     string(data),
		},
	}, nil
}

// handleProjectMemoryFile reads a file from {project_path}/.claude/memory/
func handleProjectMemoryFile(_ context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Parse URI: claude://project/{project_path}/memory/{filename}
	uri := req.Params.URI
	const prefix = "claude://project/"
	if !strings.HasPrefix(uri, prefix) {
		return nil, fmt.Errorf("invalid URI format: %s. Expected claude://project/{path}/memory/{filename}", uri)
	}

	rest := uri[len(prefix):]
	const memorySegment = "/memory/"
	memIdx := strings.LastIndex(rest, memorySegment)
	if memIdx < 0 {
		return nil, fmt.Errorf("invalid URI format, missing /memory/ segment: %s. Expected claude://project/{path}/memory/{filename}", uri)
	}

	projectPath := rest[:memIdx]
	filename := rest[memIdx+len(memorySegment):]

	if !isResourceSafeFilename(filename) {
		return nil, fmt.Errorf("invalid filename: %s. Filename must end in .md and cannot contain '..' or path separators", filename)
	}
	if !isResourceSafePath(projectPath) {
		return nil, fmt.Errorf("invalid project path: %s. Path must be absolute and cannot contain '..'", projectPath)
	}

	filePath := filepath.Join(projectPath, ".claude", "memory", filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read memory file %s: %w. Check that the file exists", filename, err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/markdown",
			Text:     string(data),
		},
	}, nil
}

// extractURIParam extracts the parameter portion after a prefix from a URI
func extractURIParam(uri, prefix string) string {
	if strings.HasPrefix(uri, prefix) {
		return uri[len(prefix):]
	}
	return ""
}

// isResourceSafeFilename validates a filename for resource access
func isResourceSafeFilename(name string) bool {
	if name == "" {
		return false
	}
	if strings.Contains(name, "..") {
		return false
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	if !strings.HasSuffix(name, ".md") {
		return false
	}
	return true
}

// isResourceSafePath validates a project path for resource access
func isResourceSafePath(path string) bool {
	if path == "" {
		return false
	}
	if strings.Contains(path, "..") {
		return false
	}
	if !strings.HasPrefix(path, "/") {
		return false
	}
	return true
}
