package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const serverInstructions = `Use this server to manage Claude Code configuration and luoshu (洛书) cross-session intelligent memory.

Product name: luoshu (洛书) - cross-session intelligent memory system.
Skill prefix: /luoshu.setup, /luoshu.config

When starting a new session in a project without .claude/ directory, proactively suggest running /luoshu.setup to initialize the project configuration.

For projects with existing .claude/ configuration, check if .claude/.setup-meta.json exists. If missing or outdated (>30 days), suggest running /luoshu.setup to sync latest features.

Available tool categories:
- Memory: save_memory, load_memory, search_memory
- Config: config_get_global, config_save_global, config_save_project, get_project_config, list_projects
- Templates: template_list, template_install, template_installed, template_uninstall
- Extensions: extension_list, extension_read, extension_save, extension_delete
- Hooks: hooks_list, hooks_save
- Evolution: evolve_status, evolve_analyze, evolve_apply
- Luoshu Config: luoshu_config_get, luoshu_config_set, luoshu_config_validate, luoshu_provider_list
- Luoshu Memory: memory_extract, memory_semantic_search, luoshu_recall
- Luoshu Status: luoshu_status, luoshu_reindex
- Import/Export: export_config, import_config
- File Semantic Search: ov_search, ov_index, ov_status

Luoshu recall (luoshu_recall):
- Intelligent recall: semantic search + LLM synthesis into coherent answers
- Falls back to formatted search results when LLM is unavailable
- Use for natural language queries like "what decisions were made about auth?"

Search memory enhancement (search_memory):
- Automatically supplements keyword results with semantic search when luoshu embedding is configured
- When keyword matches < 3, semantic search runs automatically
- Results include both keyword matches and semantic matches with source attribution

Luoshu configuration guidance:
- When LLM is not configured (luoshu_config_get shows empty api_key):
  - In /luoshu.setup: strongly recommend configuring as Step 4 (can be skipped)
  - In /luoshu.config: show status and guide configuration
- Always mask API Keys in responses (show only first 3 and last 4 characters)
- Keys are stored locally in ~/.luoshu/config.json
- Support environment variable override: LUOSHU_LLM_API_KEY

Graceful degradation:
- Without LLM config: keyword search works, semantic search unavailable
- With LLM config: full semantic search + auto-extraction enabled
- On API failure: silently fall back to keyword-only mode

File semantic search (ov_search, ov_index, ov_status):
- ov_search: semantic search over .claude/memory/*.md, .claude/rules/*.md, and project root MEMORY.md (official auto-memory)
- Preferred over search_memory for conceptual queries using natural language
- Auto-indexes files on first search, incremental sync on subsequent calls
- ov_index: manually trigger index rebuild (force=true for full rebuild)
- ov_status: check index status and configuration
- Requires embedding configuration (same as luoshu semantic search)

MCP Resources (read-only access to configuration files):
- claude://global/claude-md — Global CLAUDE.md instructions
- claude://global/memory/{filename} — Global memory files (~/.claude/memory/*.md)
- claude://project/{project_path}/memory/{filename} — Project memory files

Community resources:
- aimtpl.com — 1500+ Claude Code template analysis and discovery
- github.com/anthropics/claude-code — Official Claude Code repository
- github.com/VoltAgent/awesome-claude-code-subagents — 100+ specialized subagents`

const serverVersion = "0.7.2"

func main() {
	transport := flag.String("transport", "stdio", "Transport mode: stdio or http")
	httpAddr := flag.String("http-addr", "localhost:8080", "HTTP listen address (http mode only)")
	flag.Parse()

	s := server.NewMCPServer(
		"claude-config-mcp",
		serverVersion,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(false, true),
		server.WithInstructions(serverInstructions),
	)

	// ─── Memory Management ─────────────────────────────
	s.AddTool(buildSaveMemoryTool(), handleSaveMemory)
	s.AddTool(buildLoadMemoryTool(), handleLoadMemory)
	s.AddTool(buildSearchMemoryTool(), handleSearchMemory)

	// ─── Configuration Management ─────────────────────────────
	s.AddTool(buildGetProjectConfigTool(), handleGetProjectConfig)
	s.AddTool(buildListProjectsTool(), handleListProjects)
	s.AddTool(buildGetGlobalConfigTool(), handleGetGlobalConfig)
	s.AddTool(buildSaveGlobalConfigTool(), handleSaveGlobalConfig)
	s.AddTool(buildSaveProjectConfigTool(), handleSaveProjectConfig)

	// ─── Template Management ─────────────────────────────
	s.AddTool(buildListTemplatesTool(), handleListTemplates)
	s.AddTool(buildInstallTemplateTool(), handleInstallTemplate)
	s.AddTool(buildUninstallTemplateTool(), handleUninstallTemplate)
	s.AddTool(buildGetInstalledTemplatesTool(), handleGetInstalledTemplates)

	// ─── Extension Management (agents/rules/skills/commands) ───
	s.AddTool(buildListExtensionsTool(), handleListExtensions)
	s.AddTool(buildReadExtensionTool(), handleReadExtension)
	s.AddTool(buildSaveExtensionTool(), handleSaveExtension)
	s.AddTool(buildDeleteExtensionTool(), handleDeleteExtension)

	// ─── Hooks Management ─────────────────────────────
	s.AddTool(buildHooksListTool(), handleHooksList)
	s.AddTool(buildHooksSaveTool(), handleHooksSave)

	// ─── Evolution Engine ─────────────────────────────────
	s.AddTool(buildEvolveStatusTool(), handleEvolveStatus)
	s.AddTool(buildEvolveAnalyzeTool(), handleEvolveAnalyze)
	s.AddTool(buildEvolveApplyTool(), handleEvolveApply)

	// ─── Luoshu Memory Config ──────────────────────
	s.AddTool(buildLuoshuConfigGetTool(), handleLuoshuConfigGet)
	s.AddTool(buildLuoshuConfigSetTool(), handleLuoshuConfigSet)
	s.AddTool(buildLuoshuConfigValidateTool(), handleLuoshuConfigValidate)
	s.AddTool(buildLuoshuProviderListTool(), handleLuoshuProviderList)

	// ─── Luoshu Memory Extract & Semantic Search ─────────────────
	s.AddTool(buildMemoryExtractTool(), handleMemoryExtract)
	s.AddTool(buildMemorySemanticSearchTool(), handleMemorySemanticSearch)

	// ─── Luoshu Intelligent Recall ────────────────────────────
	s.AddTool(buildLuoshuRecallTool(), handleLuoshuRecall)

	// ─── Luoshu Status & Index ──────────────────────────
	s.AddTool(buildLuoshuStatusTool(), handleLuoshuStatus)
	s.AddTool(buildLuoshuReindexTool(), handleLuoshuReindex)

	// ─── OpenViking File Semantic Search ─────────────────────
	s.AddTool(buildOvSearchTool(), handleOvSearch)
	s.AddTool(buildOvIndexTool(), handleOvIndex)
	s.AddTool(buildOvStatusTool(), handleOvStatus)

	// ─── Config Import/Export ────────────────────────────────
	s.AddTool(buildExportConfigTool(), handleExportConfig)
	s.AddTool(buildImportConfigTool(), handleImportConfig)

	// ─── MCP Resources ──────────────────────────────────────
	registerResources(s)

	switch *transport {
	case "http":
		httpServer := server.NewStreamableHTTPServer(s,
			server.WithEndpointPath("/mcp"),
		)
		log.Printf("MCP HTTP server listening on %s/mcp", *httpAddr)
		if err := httpServer.Start(*httpAddr); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	default:
		if err := server.ServeStdio(s); err != nil {
			fmt.Fprintf(os.Stderr, "MCP Server error: %v\n", err)
			os.Exit(1)
		}
	}
}

// ─── Memory Management Tool Definitions ─────────────────────────────

func buildSaveMemoryTool() mcp.Tool {
	return mcp.NewTool(
		"save_memory",
		mcp.WithDescription("Save a memory entry to the project's .claude/memory/ directory. Use 'global' as project_path for global memory (~/.claude/memory/)."),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("Absolute project path, or 'global' for global memory"),
		),
		mcp.WithString("filename",
			mcp.Required(),
			mcp.Description("Filename such as MEMORY.md, session-state.md, decisions.md"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("File content to save"),
		),
		mcp.WithString("mode",
			mcp.Description("Write mode: 'overwrite' (default) or 'append'"),
		),
	)
}

func buildLoadMemoryTool() mcp.Tool {
	return mcp.NewTool(
		"load_memory",
		mcp.WithDescription("Load memory files from a project's .claude/memory/ directory. If filename is omitted, returns all .md files."),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("Absolute project path, or 'global' for global memory"),
		),
		mcp.WithString("filename",
			mcp.Description("Specific filename to load; omit to list all memory files"),
		),
	)
}

func buildSearchMemoryTool() mcp.Tool {
	return mcp.NewTool(
		"search_memory",
		mcp.WithDescription("Keyword search over memory files (case-insensitive exact match). For conceptual or semantic queries, prefer ov_search which finds related content even with different wording."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search keyword (case-insensitive)"),
		),
		mcp.WithString("project_path",
			mcp.Description("Limit search to a specific project path; omit to search all"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of matching lines to return (default 20)"),
		),
	)
}

// ─── Configuration Management Tool Definitions ─────────────────────────────

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

func buildGetGlobalConfigTool() mcp.Tool {
	return mcp.NewTool(
		"config_get_global",
		mcp.WithDescription("Get global Claude Code configuration overview including CLAUDE.md, settings.json, and .mcp.json contents."),
	)
}

func buildSaveGlobalConfigTool() mcp.Tool {
	return mcp.NewTool(
		"config_save_global",
		mcp.WithDescription("Save a global Claude Code configuration field. Validates JSON for settings and mcp fields."),
		mcp.WithString("field",
			mcp.Required(),
			mcp.Description("Config field to save: 'claude_md', 'settings', or 'mcp'"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Content to write to the field"),
		),
	)
}

func buildSaveProjectConfigTool() mcp.Tool {
	return mcp.NewTool(
		"config_save_project",
		mcp.WithDescription("Save a project-level Claude Code configuration field. Validates JSON for settings and mcp fields."),
		mcp.WithString("project_path",
			mcp.Required(),
			mcp.Description("Absolute project path"),
		),
		mcp.WithString("field",
			mcp.Required(),
			mcp.Description("Config field to save: 'claude_md', 'settings', or 'mcp'"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Content to write to the field"),
		),
	)
}

// ─── Template Management Tool Definitions ─────────────────────────────

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

// ─── Extension Management Tool Definitions ─────────────────────────────

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
