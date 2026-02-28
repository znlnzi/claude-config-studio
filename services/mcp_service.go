package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MCPService manages MCP server configurations
type MCPService struct {
	ctx context.Context
}

func NewMCPService() *MCPService {
	return &MCPService{}
}

func (s *MCPService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// MCPServer represents MCP server configuration
type MCPServer struct {
	Name    string            `json:"name"`
	Type    string            `json:"type,omitempty"`    // "http" or "" (command line)
	URL     string            `json:"url,omitempty"`     // HTTP type
	Headers map[string]string `json:"headers,omitempty"` // HTTP headers
	Command string            `json:"command,omitempty"` // Command line type
	Args    []string          `json:"args,omitempty"`    // Command arguments
	Timeout int               `json:"timeout,omitempty"` // Timeout in seconds
}

// MCPConfig represents the complete MCP configuration (.mcp.json format)
type MCPConfig struct {
	Servers []MCPServer `json:"servers"`
}

// GetGlobalMCPServers retrieves global MCP configuration (merged from multiple sources)
func (s *MCPService) GetGlobalMCPServers() ([]MCPServer, error) {
	home, _ := os.UserHomeDir()
	seen := make(map[string]bool)
	var servers []MCPServer

	// 1. Read ~/.claude/.mcp.json
	mcpPath := filepath.Join(home, ".claude", ".mcp.json")
	if list, err := s.readMCPFile(mcpPath); err == nil {
		for _, srv := range list {
			if !seen[srv.Name] {
				seen[srv.Name] = true
				servers = append(servers, srv)
			}
		}
	}

	// 2. Read projects.*.mcpServers from ~/.claude.json (Claude Code CLI configuration)
	claudeJsonPath := filepath.Join(home, ".claude.json")
	if data, err := os.ReadFile(claudeJsonPath); err == nil {
		var claudeConfig struct {
			Projects map[string]struct {
				McpServers map[string]json.RawMessage `json:"mcpServers"`
			} `json:"projects"`
		}
		if err := json.Unmarshal(data, &claudeConfig); err == nil {
			for _, proj := range claudeConfig.Projects {
				for name, rawServer := range proj.McpServers {
					if seen[name] {
						continue
					}
					var srv MCPServer
					if err := json.Unmarshal(rawServer, &srv); err != nil {
						continue
					}
					srv.Name = name
					seen[name] = true
					servers = append(servers, srv)
				}
			}
		}
	}

	// 3. Read Claude Desktop configuration
	desktopConfigPath := filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	if list, err := s.readMCPFile(desktopConfigPath); err == nil {
		for _, srv := range list {
			if !seen[srv.Name] {
				seen[srv.Name] = true
				servers = append(servers, srv)
			}
		}
	}

	return servers, nil
}

// SaveGlobalMCPServers saves global MCP configuration
func (s *MCPService) SaveGlobalMCPServers(servers []MCPServer) error {
	home, _ := os.UserHomeDir()
	mcpPath := filepath.Join(home, ".claude", ".mcp.json")
	return s.writeMCPFile(mcpPath, servers)
}

// GetProjectMCPServers retrieves project-level MCP configuration
func (s *MCPService) GetProjectMCPServers(projectPath string) ([]MCPServer, error) {
	mcpPath := filepath.Join(projectPath, ".claude", ".mcp.json")
	return s.readMCPFile(mcpPath)
}

// SaveProjectMCPServers saves project-level MCP configuration
func (s *MCPService) SaveProjectMCPServers(projectPath string, servers []MCPServer) error {
	claudeDir := filepath.Join(projectPath, ".claude")
	os.MkdirAll(claudeDir, 0755)
	mcpPath := filepath.Join(claudeDir, ".mcp.json")
	return s.writeMCPFile(mcpPath, servers)
}

func (s *MCPService) readMCPFile(path string) ([]MCPServer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return []MCPServer{}, nil
	}

	// .mcp.json format: { "mcpServers": { "name": { ... } } } or { "name": { ... } }
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return []MCPServer{}, nil
	}

	// Check if wrapped with mcpServers
	serversMap := raw
	if mcpServersRaw, ok := raw["mcpServers"]; ok {
		var inner map[string]json.RawMessage
		if err := json.Unmarshal(mcpServersRaw, &inner); err == nil {
			serversMap = inner
		}
	} else {
		// Filter out non-server top-level keys
		delete(serversMap, "$schema")
	}

	var servers []MCPServer
	for name, rawServer := range serversMap {
		var srv MCPServer
		if err := json.Unmarshal(rawServer, &srv); err != nil {
			continue
		}
		srv.Name = name
		servers = append(servers, srv)
	}

	return servers, nil
}

// MarketplaceServer represents an MCP server in the marketplace
type MarketplaceServer struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	DescriptionCN string   `json:"descriptionCN"`
	RepoURL       string   `json:"repoUrl"`
	Package       string   `json:"package"`
	Transport     string   `json:"transport"`
	Command       string   `json:"command"`
	Args          []string `json:"args"`
	Version       string   `json:"version"`
	PublishedAt   string   `json:"publishedAt"`
}

// MarketplaceResult represents marketplace search results (with pagination)
type MarketplaceResult struct {
	Servers    []MarketplaceServer `json:"servers"`
	NextCursor string              `json:"nextCursor"`
	Total      int                 `json:"total"`
}

// registryResponse represents Registry API response
type registryResponse struct {
	Servers []struct {
		Server registryServerInfo `json:"server"`
		Meta   struct {
			Official *struct {
				PublishedAt string `json:"publishedAt"`
				IsLatest    bool   `json:"isLatest"`
			} `json:"io.modelcontextprotocol.registry/official"`
		} `json:"_meta"`
	} `json:"servers"`
	Metadata struct {
		NextCursor string `json:"nextCursor"`
		Count      int    `json:"count"`
	} `json:"metadata"`
}

// SearchMarketplace searches MCP marketplace (supports multiple sources + pagination)
// source: "official" | "smithery" | "glama"
func (s *MCPService) SearchMarketplace(source string, query string, inputCursor string) (MarketplaceResult, error) {
	switch source {
	case "smithery":
		return s.searchSmithery(query, inputCursor)
	case "glama":
		return s.searchGlama(query, inputCursor)
	default:
		return s.searchOfficial(query, inputCursor)
	}
}

// --- Official Registry ---

func (s *MCPService) searchOfficial(query string, inputCursor string) (MarketplaceResult, error) {
	const targetCount = 30
	const maxPages = 5
	const pageSize = 100

	seen := make(map[string]int)
	var internal []marketplaceServerInternal
	cursor := inputCursor
	lastCursor := ""

	client := &http.Client{Timeout: 15 * time.Second}

	for page := 0; page < maxPages && len(internal) < targetCount; page++ {
		apiURL := fmt.Sprintf("https://registry.modelcontextprotocol.io/v0/servers?limit=%d", pageSize)
		if query != "" {
			apiURL += "&search=" + url.QueryEscape(query)
		}
		if cursor != "" {
			apiURL += "&cursor=" + url.QueryEscape(cursor)
		}

		resp, err := client.Get(apiURL)
		if err != nil {
			if page == 0 {
				return MarketplaceResult{}, fmt.Errorf("request failed: %w", err)
			}
			break
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			break
		}

		var registry registryResponse
		if err := json.Unmarshal(body, &registry); err != nil {
			break
		}

		if len(registry.Servers) == 0 {
			break
		}

		for _, item := range registry.Servers {
			srv := item.Server
			displayName := srv.Name
			if parts := strings.Split(displayName, "/"); len(parts) > 1 {
				displayName = parts[len(parts)-1]
			}

			isLatest := false
			publishedAt := ""
			if item.Meta.Official != nil {
				isLatest = item.Meta.Official.IsLatest
				publishedAt = item.Meta.Official.PublishedAt
			}

			if idx, exists := seen[displayName]; exists {
				if isLatest && !internal[idx].isLatest {
					ms := buildMarketplaceServer(srv, displayName, publishedAt, isLatest)
					internal[idx] = ms
				}
				continue
			}

			ms := buildMarketplaceServer(srv, displayName, publishedAt, isLatest)
			seen[displayName] = len(internal)
			internal = append(internal, ms)
		}

		lastCursor = registry.Metadata.NextCursor
		if lastCursor == "" {
			break
		}
		cursor = lastCursor
	}

	results := make([]MarketplaceServer, len(internal))
	for i, ms := range internal {
		results[i] = ms.MarketplaceServer
	}

	return MarketplaceResult{
		Servers:    results,
		NextCursor: lastCursor,
		Total:      len(results),
	}, nil
}

// --- Smithery ---

type smitheryResponse struct {
	Servers []struct {
		QualifiedName string `json:"qualifiedName"`
		DisplayName   string `json:"displayName"`
		Description   string `json:"description"`
		Homepage      string `json:"homepage"`
		UseCount      int    `json:"useCount"`
		CreatedAt     string `json:"createdAt"`
	} `json:"servers"`
	Pagination struct {
		CurrentPage int `json:"currentPage"`
		TotalPages  int `json:"totalPages"`
		TotalCount  int `json:"totalCount"`
	} `json:"pagination"`
}

func (s *MCPService) searchSmithery(query string, inputCursor string) (MarketplaceResult, error) {
	page := 1
	if inputCursor != "" {
		fmt.Sscanf(inputCursor, "%d", &page)
	}

	apiURL := fmt.Sprintf("https://registry.smithery.ai/servers?pageSize=30&page=%d", page)
	if query != "" {
		apiURL += "&q=" + url.QueryEscape(query)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return MarketplaceResult{}, fmt.Errorf("Smithery request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return MarketplaceResult{}, fmt.Errorf("failed to read response: %w", err)
	}

	var sr smitheryResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return MarketplaceResult{}, fmt.Errorf("failed to parse response: %w", err)
	}

	var results []MarketplaceServer
	for _, srv := range sr.Servers {
		// Extract package name from qualifiedName: @owner/repo -> npm package name
		displayName := srv.DisplayName
		if displayName == "" {
			displayName = srv.QualifiedName
		}

		ms := MarketplaceServer{
			Name:          displayName,
			Description:   srv.Description,
			DescriptionCN: translateDescription(displayName, srv.Description),
			RepoURL:       srv.Homepage,
			Package:       srv.QualifiedName,
			Transport:     "stdio",
			Command:       "npx",
			Args:          []string{"-y", "@smithery/cli@latest", "run", srv.QualifiedName},
			PublishedAt:   srv.CreatedAt,
		}
		results = append(results, ms)
	}

	nextCursor := ""
	if page < sr.Pagination.TotalPages {
		nextCursor = fmt.Sprintf("%d", page+1)
	}

	return MarketplaceResult{
		Servers:    results,
		NextCursor: nextCursor,
		Total:      sr.Pagination.TotalCount,
	}, nil
}

// --- Glama ---

type glamaResponse struct {
	Servers []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Namespace   string `json:"namespace"`
		Description string `json:"description"`
		Repository  *struct {
			URL string `json:"url"`
		} `json:"repository"`
		URL string `json:"url"`
	} `json:"servers"`
	PageInfo struct {
		EndCursor   string `json:"endCursor"`
		HasNextPage bool   `json:"hasNextPage"`
	} `json:"pageInfo"`
}

func (s *MCPService) searchGlama(query string, inputCursor string) (MarketplaceResult, error) {
	apiURL := "https://glama.ai/api/mcp/v1/servers?limit=30"
	if query != "" {
		apiURL += "&search=" + url.QueryEscape(query)
	}
	if inputCursor != "" {
		apiURL += "&cursor=" + url.QueryEscape(inputCursor)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return MarketplaceResult{}, fmt.Errorf("Glama request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return MarketplaceResult{}, fmt.Errorf("failed to read response: %w", err)
	}

	var gr glamaResponse
	if err := json.Unmarshal(body, &gr); err != nil {
		return MarketplaceResult{}, fmt.Errorf("failed to parse response: %w", err)
	}

	var results []MarketplaceServer
	for _, srv := range gr.Servers {
		displayName := srv.Name
		if displayName == "" {
			displayName = srv.Slug
		}

		// Construct npm package name: namespace/name
		pkg := srv.Namespace + "/" + srv.Slug

		repoURL := ""
		if srv.Repository != nil {
			repoURL = srv.Repository.URL
		}

		ms := MarketplaceServer{
			Name:          displayName,
			Description:   srv.Description,
			DescriptionCN: translateDescription(displayName, srv.Description),
			RepoURL:       repoURL,
			Package:       pkg,
			Transport:     "stdio",
		}
		results = append(results, ms)
	}

	nextCursor := ""
	if gr.PageInfo.HasNextPage {
		nextCursor = gr.PageInfo.EndCursor
	}

	return MarketplaceResult{
		Servers:    results,
		NextCursor: nextCursor,
		Total:      len(results),
	}, nil
}

// marketplaceServerInternal is for internal use, with isLatest flag
type marketplaceServerInternal struct {
	MarketplaceServer
	isLatest bool
}

type registryServerInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Repository  *struct {
		URL string `json:"url"`
	} `json:"repository"`
	Packages []struct {
		RegistryType string `json:"registryType"`
		Identifier   string `json:"identifier"`
		Transport    struct {
			Type string `json:"type"`
		} `json:"transport"`
	} `json:"packages"`
}

func buildMarketplaceServer(srv registryServerInfo, displayName, publishedAt string, isLatest bool) marketplaceServerInternal {
	ms := marketplaceServerInternal{}
	ms.Name = displayName
	ms.Description = srv.Description
	ms.DescriptionCN = translateDescription(displayName, srv.Description)
	ms.Version = srv.Version
	ms.PublishedAt = publishedAt
	ms.isLatest = isLatest

	if srv.Repository != nil {
		ms.RepoURL = srv.Repository.URL
	}

	if len(srv.Packages) > 0 {
		pkg := srv.Packages[0]
		ms.Package = pkg.Identifier
		ms.Transport = pkg.Transport.Type
		if pkg.RegistryType == "npm" {
			ms.Command = "npx"
			ms.Args = []string{"-y", pkg.Identifier}
		} else if pkg.RegistryType == "oci" {
			ms.Command = "docker"
			ms.Args = []string{"run", "-i", pkg.Identifier}
		}
	}

	return ms
}

// translateDescription generates a localized description based on keywords
func translateDescription(name, desc string) string {
	lowerName := strings.ToLower(name)
	lowerDesc := strings.ToLower(desc)
	combined := lowerName + " " + lowerDesc

	// Keyword to description mapping (ordered by priority)
	keywords := []struct {
		keys []string
		cn   string
	}{
		{[]string{"filesystem", "file system", "file-system"}, "File system read/write and management"},
		{[]string{"github"}, "GitHub repository and Issue/PR management"},
		{[]string{"gitlab"}, "GitLab project management and CI/CD"},
		{[]string{"git"}, "Git version control operations"},
		{[]string{"postgres", "postgresql"}, "PostgreSQL database query and management"},
		{[]string{"mysql"}, "MySQL database operations"},
		{[]string{"sqlite"}, "SQLite lightweight database"},
		{[]string{"mongodb", "mongo"}, "MongoDB document database"},
		{[]string{"redis"}, "Redis cache and data operations"},
		{[]string{"database", "sql"}, "Database query and management"},
		{[]string{"docker", "container"}, "Docker container management"},
		{[]string{"kubernetes", "k8s"}, "Kubernetes cluster management"},
		{[]string{"aws", "amazon"}, "AWS cloud service integration"},
		{[]string{"azure"}, "Azure cloud platform integration"},
		{[]string{"gcp", "google cloud"}, "Google Cloud integration"},
		{[]string{"slack"}, "Slack messaging and workflows"},
		{[]string{"discord"}, "Discord bot integration"},
		{[]string{"notion"}, "Notion notes and knowledge base"},
		{[]string{"jira"}, "Jira project and task management"},
		{[]string{"linear"}, "Linear task tracking"},
		{[]string{"confluence"}, "Confluence document collaboration"},
		{[]string{"browser", "puppeteer", "playwright", "selenium"}, "Browser automation and web operations"},
		{[]string{"search", "web search"}, "Web search service"},
		{[]string{"scrape", "scraping", "crawl"}, "Web content scraping"},
		{[]string{"fetch", "http"}, "HTTP requests and API calls"},
		{[]string{"email", "smtp", "gmail"}, "Email send/receive service"},
		{[]string{"pdf"}, "PDF document processing"},
		{[]string{"image", "vision"}, "Image processing and analysis"},
		{[]string{"openai", "llm", "ai"}, "AI/LLM model integration"},
		{[]string{"stripe"}, "Stripe payment integration"},
		{[]string{"sentry"}, "Sentry error monitoring"},
		{[]string{"datadog"}, "Datadog monitoring and analytics"},
		{[]string{"elasticsearch", "elastic"}, "Elasticsearch search engine"},
		{[]string{"graphql"}, "GraphQL API integration"},
		{[]string{"rest", "api"}, "REST API service integration"},
		{[]string{"s3", "storage", "blob"}, "Object storage service"},
		{[]string{"memory", "knowledge"}, "Memory and knowledge management"},
		{[]string{"code", "lint", "format"}, "Code analysis and formatting"},
		{[]string{"test"}, "Testing tool integration"},
		{[]string{"ci", "cd", "deploy"}, "CI/CD continuous integration and deployment"},
		{[]string{"monitor", "observ"}, "Monitoring and observability"},
		{[]string{"log"}, "Log collection and analysis"},
		{[]string{"auth"}, "Authentication service"},
		{[]string{"map", "geo", "location"}, "Map and geolocation service"},
		{[]string{"calendar", "schedule"}, "Calendar and scheduling management"},
		{[]string{"chat", "message"}, "Instant messaging integration"},
		{[]string{"translate", "translation"}, "Translation service"},
		{[]string{"weather"}, "Weather information query"},
		{[]string{"crypto", "blockchain"}, "Cryptocurrency and blockchain"},
		{[]string{"finance", "stock"}, "Financial data and trading"},
		{[]string{"document", "doc"}, "Document processing and conversion"},
		{[]string{"exa"}, "Exa AI search engine"},
		{[]string{"context7"}, "Context7 real-time documentation query"},
	}

	for _, kw := range keywords {
		for _, key := range kw.keys {
			if strings.Contains(combined, key) {
				return kw.cn
			}
		}
	}

	return "MCP tool service"
}

// InstallFromMarketplace installs an MCP server from the marketplace to global configuration
func (s *MCPService) InstallFromMarketplace(displayName string, pkg string) error {
	servers, err := s.GetGlobalMCPServers()
	if err != nil {
		return err
	}

	// Check if already installed
	for _, srv := range servers {
		if srv.Name == displayName {
			return fmt.Errorf("server %s already exists", displayName)
		}
	}

	newServer := MCPServer{
		Name:    displayName,
		Command: "npx",
		Args:    []string{"-y", pkg},
	}
	servers = append(servers, newServer)

	return s.SaveGlobalMCPServers(servers)
}

func (s *MCPService) writeMCPFile(path string, servers []MCPServer) error {
	mcpMap := make(map[string]interface{})
	mcpMap["$schema"] = "https://raw.githubusercontent.com/anthropics/claude-code/main/.mcp.schema.json"

	serversMap := make(map[string]interface{})
	for _, srv := range servers {
		entry := make(map[string]interface{})
		if srv.Type == "http" {
			entry["type"] = "http"
			entry["url"] = srv.URL
			if len(srv.Headers) > 0 {
				entry["headers"] = srv.Headers
			}
		} else {
			entry["command"] = srv.Command
			if len(srv.Args) > 0 {
				entry["args"] = srv.Args
			}
		}
		if srv.Timeout > 0 {
			entry["timeout"] = srv.Timeout
		}
		serversMap[srv.Name] = entry
	}

	mcpMap["mcpServers"] = serversMap
	formatted, err := json.MarshalIndent(mcpMap, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)
	return os.WriteFile(path, formatted, 0644)
}
