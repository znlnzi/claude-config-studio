package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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

// serverVersion is set at build time via ldflags:
//
//	go build -ldflags "-X main.serverVersion=0.8.0" ./cmd/mcp-server/
var serverVersion = "dev"

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
