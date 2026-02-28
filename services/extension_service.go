package services

import (
	"context"
	"encoding/base64"
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

// ExtensionService manages extension files such as Commands/Agents/Skills
type ExtensionService struct {
	ctx context.Context
}

func NewExtensionService() *ExtensionService {
	return &ExtensionService{}
}

func (s *ExtensionService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// ExtensionFile represents extension file information
type ExtensionFile struct {
	Name         string `json:"name"`
	FileName     string `json:"fileName"`
	Path         string `json:"path"`
	Content      string `json:"content"`
	LastModified string `json:"lastModified"`
	Size         int64  `json:"size"`
}

// ListExtensions lists extension files of the specified type
// extType: "commands", "agents", "skills"
// scope: "global" or project path
func (s *ExtensionService) ListExtensions(extType string, scope string) ([]ExtensionFile, error) {
	var dir string
	if scope == "global" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".claude", extType)
	} else {
		dir = filepath.Join(scope, ".claude", extType)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return []ExtensionFile{}, nil
	}

	var files []ExtensionFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		filePath := filepath.Join(dir, name)
		info, _ := entry.Info()

		ext := ExtensionFile{
			Name:     strings.TrimSuffix(name, ".md"),
			FileName: name,
			Path:     filePath,
		}
		if info != nil {
			ext.LastModified = info.ModTime().Format(time.RFC3339)
			ext.Size = info.Size()
		}

		files = append(files, ext)
	}

	return files, nil
}

// GetExtension reads the content of an extension file
func (s *ExtensionService) GetExtension(extType string, scope string, fileName string) (*ExtensionFile, error) {
	var dir string
	if scope == "global" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".claude", extType)
	} else {
		dir = filepath.Join(scope, ".claude", extType)
	}

	filePath := filepath.Join(dir, fileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	info, _ := os.Stat(filePath)
	ext := &ExtensionFile{
		Name:     strings.TrimSuffix(fileName, ".md"),
		FileName: fileName,
		Path:     filePath,
		Content:  string(data),
	}
	if info != nil {
		ext.LastModified = info.ModTime().Format(time.RFC3339)
		ext.Size = info.Size()
	}

	return ext, nil
}

// SaveExtension saves an extension file
func (s *ExtensionService) SaveExtension(extType string, scope string, fileName string, content string) error {
	var dir string
	if scope == "global" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".claude", extType)
	} else {
		dir = filepath.Join(scope, ".claude", extType)
	}

	os.MkdirAll(dir, 0755)
	filePath := filepath.Join(dir, fileName)
	return os.WriteFile(filePath, []byte(content), 0644)
}

// DeleteExtension deletes an extension file
func (s *ExtensionService) DeleteExtension(extType string, scope string, fileName string) error {
	var dir string
	if scope == "global" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".claude", extType)
	} else {
		dir = filepath.Join(scope, ".claude", extType)
	}

	filePath := filepath.Join(dir, fileName)
	return os.Remove(filePath)
}

// RenameExtension renames an extension file
func (s *ExtensionService) RenameExtension(extType string, scope string, oldName string, newName string) error {
	var dir string
	if scope == "global" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".claude", extType)
	} else {
		dir = filepath.Join(scope, ".claude", extType)
	}

	oldPath := filepath.Join(dir, oldName)
	newPath := filepath.Join(dir, newName)
	return os.Rename(oldPath, newPath)
}

// ============ Online Marketplace ============

// OnlineExtension represents online extension information
type OnlineExtension struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Source      string `json:"source"` // "builtin" | "github"
	RepoURL     string `json:"repoUrl"`
	DownloadURL string `json:"downloadUrl"` // raw content URL
	ExtType     string `json:"extType"`     // "agents" | "skills"
}

// OnlineExtensionResult represents online marketplace search results
type OnlineExtensionResult struct {
	Extensions []OnlineExtension `json:"extensions"`
	Total      int               `json:"total"`
}

// Category display name mapping
var agentCategoryNames = map[string]string{
	"01-core-development":    "Core Development",
	"02-language-specialists": "Language Specialists",
	"03-infrastructure":      "Infrastructure",
	"04-quality-security":    "Quality & Security",
	"05-data-ai":             "Data & AI",
	"06-developer-experience": "Developer Experience",
	"07-specialized-domains":  "Specialized Domains",
	"08-business-product":     "Business & Product",
	"09-meta-orchestration":   "Meta Orchestration",
	"10-research-analysis":    "Research & Analysis",
}

// SearchOnlineExtensions searches online extensions (agents or skills)
// extType: "agents" | "skills"
// source: "builtin" | "github" | ""(all)
func (s *ExtensionService) SearchOnlineExtensions(extType string, source string, query string) (OnlineExtensionResult, error) {
	var all []OnlineExtension

	// Built-in data
	if source == "" || source == "builtin" {
		if extType == "agents" {
			all = append(all, getBuiltinAgents()...)
		} else if extType == "skills" {
			all = append(all, getBuiltinSkills()...)
		}
	}

	// GitHub dynamic loading
	if source == "" || source == "github" {
		ghExts, err := s.fetchGitHubExtensions(extType)
		if err == nil {
			// Merge and deduplicate (built-in takes priority)
			existing := make(map[string]bool)
			for _, e := range all {
				existing[e.Name] = true
			}
			for _, e := range ghExts {
				if !existing[e.Name] {
					all = append(all, e)
				}
			}
		}
	}

	// Search filter
	if query != "" {
		q := strings.ToLower(query)
		var filtered []OnlineExtension
		for _, e := range all {
			if strings.Contains(strings.ToLower(e.Name), q) ||
				strings.Contains(strings.ToLower(e.Description), q) ||
				strings.Contains(strings.ToLower(e.Category), q) {
				filtered = append(filtered, e)
			}
		}
		all = filtered
	}

	return OnlineExtensionResult{
		Extensions: all,
		Total:      len(all),
	}, nil
}

// InstallOnlineExtension installs an extension from the online marketplace
// scope: "global" or project path
func (s *ExtensionService) InstallOnlineExtension(extType string, ext OnlineExtension, scope string) error {
	if ext.DownloadURL == "" {
		return fmt.Errorf("no download URL")
	}
	if scope == "" {
		scope = "global"
	}

	// Download content
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(ext.DownloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	// GitHub API returns JSON, need to parse the content field
	if strings.Contains(ext.DownloadURL, "api.github.com") {
		var ghFile struct {
			Content  string `json:"content"`
			Encoding string `json:"encoding"`
		}
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &ghFile); err != nil {
			return fmt.Errorf("failed to parse GitHub response: %w", err)
		}
		if ghFile.Encoding == "base64" {
			decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(ghFile.Content, "\n", ""))
			if err != nil {
				return fmt.Errorf("decoding failed: %w", err)
			}
			return s.SaveExtension(extType, scope, ext.Name+".md", string(decoded))
		}
		return s.SaveExtension(extType, scope, ext.Name+".md", string(body))
	}

	// Direct raw content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}
	return s.SaveExtension(extType, scope, ext.Name+".md", string(body))
}

// fetchGitHubExtensions fetches extension list from GitHub repository
func (s *ExtensionService) fetchGitHubExtensions(extType string) ([]OnlineExtension, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	if extType == "agents" {
		return s.fetchGitHubAgents(client)
	}
	return s.fetchGitHubSkills(client)
}

func (s *ExtensionService) fetchGitHubAgents(client *http.Client) ([]OnlineExtension, error) {
	apiURL := "https://api.github.com/repos/VoltAgent/awesome-claude-code-subagents/git/trees/main?recursive=1"
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tree struct {
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
		} `json:"tree"`
	}
	if err := json.Unmarshal(body, &tree); err != nil {
		return nil, err
	}

	var exts []OnlineExtension
	for _, item := range tree.Tree {
		if item.Type != "blob" || !strings.HasPrefix(item.Path, "categories/") || !strings.HasSuffix(item.Path, ".md") {
			continue
		}
		if strings.Contains(item.Path, "README") || strings.Contains(item.Path, ".claude-plugin") {
			continue
		}
		parts := strings.Split(item.Path, "/")
		if len(parts) < 3 {
			continue
		}
		cat := parts[1]
		name := strings.TrimSuffix(parts[2], ".md")

		catCN := agentCategoryNames[cat]
		if catCN == "" {
			catCN = cat
		}

		exts = append(exts, OnlineExtension{
			Name:        name,
			Description: catCN + " - " + formatAgentName(name),
			Category:    catCN,
			Source:      "github",
			RepoURL:     "https://github.com/VoltAgent/awesome-claude-code-subagents",
			DownloadURL: "https://api.github.com/repos/VoltAgent/awesome-claude-code-subagents/contents/" + url.PathEscape(item.Path),
			ExtType:     "agents",
		})
	}
	return exts, nil
}

func (s *ExtensionService) fetchGitHubSkills(client *http.Client) ([]OnlineExtension, error) {
	apiURL := "https://api.github.com/repos/anthropics/skills/contents/skills"
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var dirs []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &dirs); err != nil {
		return nil, err
	}

	var exts []OnlineExtension
	for _, d := range dirs {
		if d.Type != "dir" {
			continue
		}
		exts = append(exts, OnlineExtension{
			Name:        d.Name,
			Description: "Anthropic Official Skill - " + formatAgentName(d.Name),
			Category:    "Anthropic Official",
			Source:      "github",
			RepoURL:     "https://github.com/anthropics/skills/tree/main/skills/" + d.Name,
			DownloadURL: "https://api.github.com/repos/anthropics/skills/contents/skills/" + d.Name + "/SKILL.md",
			ExtType:     "skills",
		})
	}
	return exts, nil
}

// formatAgentName formats name: kebab-case -> Title Case
func formatAgentName(name string) string {
	parts := strings.Split(name, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// ============ Built-in Data ============

func getBuiltinAgents() []OnlineExtension {
	agents := []struct {
		name     string
		category string
		desc     string
	}{
		// Core Development
		{"frontend-developer", "Core Development", "Build React components, responsive layouts, client-side state management"},
		{"backend-developer", "Core Development", "Design RESTful APIs, microservice architecture, database schemas"},
		{"fullstack-developer", "Core Development", "Full-stack web application development, frontend and backend integration"},
		{"mobile-developer", "Core Development", "React Native/Flutter cross-platform mobile development"},
		{"api-designer", "Core Development", "API interface design and documentation standards"},
		{"ui-designer", "Core Development", "User interface design and interaction experience optimization"},
		// Language Specialists
		{"typescript-pro", "Language Specialists", "TypeScript advanced types, generics, and strict type safety"},
		{"python-pro", "Language Specialists", "Python 3.12+ async programming and performance optimization"},
		{"golang-pro", "Language Specialists", "Go concurrency patterns, microservices, and performance optimization"},
		{"rust-engineer", "Language Specialists", "Rust systems programming, memory safety, and async patterns"},
		{"java-architect", "Language Specialists", "Java 21+ virtual threads, Spring Boot microservices"},
		{"react-specialist", "Language Specialists", "React 19 + Next.js 15 modern frontend architecture"},
		{"vue-expert", "Language Specialists", "Vue 3 Composition API and ecosystem"},
		{"swift-expert", "Language Specialists", "Swift/SwiftUI iOS native development"},
		// Infrastructure
		{"devops-engineer", "Infrastructure", "CI/CD pipelines, containerization, and automated deployment"},
		{"kubernetes-specialist", "Infrastructure", "K8s cluster management, Helm Charts, service mesh"},
		{"cloud-architect", "Infrastructure", "AWS/Azure/GCP multi-cloud architecture design"},
		{"terraform-engineer", "Infrastructure", "Terraform/OpenTofu infrastructure as code"},
		{"security-engineer", "Infrastructure", "Security architecture design, vulnerability assessment, and compliance"},
		// Quality & Security
		{"code-reviewer", "Quality & Security", "Code review, security scanning, and quality assurance"},
		{"test-automator", "Quality & Security", "Automated testing strategies, CI/CD test integration"},
		{"debugger", "Quality & Security", "Error debugging, stack analysis, and root cause identification"},
		{"performance-engineer", "Quality & Security", "Performance optimization, load testing, and observability"},
		{"security-auditor", "Quality & Security", "Security auditing, OWASP standards, and compliance checks"},
		// Data & AI
		{"ai-engineer", "Data & AI", "LLM application development, RAG systems, and intelligent agents"},
		{"data-scientist", "Data & AI", "Data analysis, machine learning, and statistical modeling"},
		{"data-engineer", "Data & AI", "Data pipelines, ETL, and data warehouse architecture"},
		{"prompt-engineer", "Data & AI", "Prompt engineering, LLM optimization, and AI system design"},
		{"ml-engineer", "Data & AI", "ML model training, deployment, and monitoring"},
		// Developer Experience
		{"documentation-engineer", "Developer Experience", "Technical documentation writing and API doc generation"},
		{"dx-optimizer", "Developer Experience", "Developer experience optimization, toolchain improvement"},
		{"legacy-modernizer", "Developer Experience", "Legacy system refactoring, framework migration, and technical debt"},
		// Specialized Domains
		{"blockchain-developer", "Specialized Domains", "Smart contracts, DeFi protocols, and Web3 development"},
		{"game-developer", "Specialized Domains", "Unity/game engine development and performance optimization"},
		{"payment-integration", "Specialized Domains", "Stripe/PayPal payment integration and billing systems"},
		// Business & Product
		{"business-analyst", "Business & Product", "Business analysis, KPI frameworks, and data-driven decisions"},
		{"product-manager", "Business & Product", "Product planning, requirements analysis, and roadmap management"},
		{"technical-writer", "Business & Product", "Technical writing, user manuals, and API documentation"},
	}

	var result []OnlineExtension
	for _, a := range agents {
		result = append(result, OnlineExtension{
			Name:        a.name,
			Description: a.desc,
			Category:    a.category,
			Source:      "builtin",
			RepoURL:     "https://github.com/VoltAgent/awesome-claude-code-subagents",
			DownloadURL: "https://api.github.com/repos/VoltAgent/awesome-claude-code-subagents/contents/categories/" + getCategoryDir(a.category) + "/" + a.name + ".md",
			ExtType:     "agents",
		})
	}
	return result
}

func getCategoryDir(catCN string) string {
	for dir, cn := range agentCategoryNames {
		if cn == catCN {
			return dir
		}
	}
	return ""
}

func getBuiltinSkills() []OnlineExtension {
	skills := []struct {
		name string
		desc string
	}{
		{"algorithmic-art", "Create algorithmic and generative art using p5.js"},
		{"brand-guidelines", "Apply brand colors and typography to design artifacts"},
		{"canvas-design", "Create beautiful visual designs and posters"},
		{"doc-coauthoring", "Structured document collaborative writing workflow"},
		{"docx", "Word document creation, editing, and formatting"},
		{"frontend-design", "Production-grade frontend UI design and component development"},
		{"internal-comms", "Internal communication templates (status reports, updates, etc.)"},
		{"mcp-builder", "Guide for building high-quality MCP servers"},
		{"pdf", "PDF document processing, form filling, and generation"},
		{"pptx", "PowerPoint presentation creation and editing"},
		{"skill-creator", "Guided workflow for creating new Skills"},
		{"slack-gif-creator", "Create optimized GIF animations for Slack"},
		{"theme-factory", "Theme system toolkit with 10 preset themes"},
		{"web-artifacts-builder", "Build complex web components with React + Tailwind"},
		{"webapp-testing", "Test local web applications using Playwright"},
		{"xlsx", "Excel spreadsheet creation, data analysis, and formulas"},
	}

	var result []OnlineExtension
	for _, sk := range skills {
		result = append(result, OnlineExtension{
			Name:        sk.name,
			Description: sk.desc,
			Category:    "Anthropic Official",
			Source:      "builtin",
			RepoURL:     "https://github.com/anthropics/skills/tree/main/skills/" + sk.name,
			DownloadURL: "https://api.github.com/repos/anthropics/skills/contents/skills/" + sk.name + "/SKILL.md",
			ExtType:     "skills",
		})
	}
	return result
}
