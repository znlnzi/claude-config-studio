package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SkillService manages Skills
type SkillService struct {
	ctx context.Context
}

func NewSkillService() *SkillService {
	return &SkillService{}
}

func (s *SkillService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// SkillInfo represents Skill information
type SkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`      // Source plugin (e.g. document-skills@anthropic-agent-skills)
	Marketplace string `json:"marketplace"` // Marketplace name
	PluginName  string `json:"pluginName"`  // Plugin name
	Type        string `json:"type"`        // skill or agent
	FilePath    string `json:"filePath"`    // Path to SKILL.md or agent .md
}

// GetAllSkills retrieves all skills provided by installed plugins
func (s *SkillService) GetAllSkills() ([]SkillInfo, error) {
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".claude", "plugins", "cache")

	var skills []SkillInfo

	// Iterate marketplace directories
	marketplaces, err := os.ReadDir(cacheDir)
	if err != nil {
		return skills, nil
	}

	for _, mp := range marketplaces {
		if !mp.IsDir() {
			continue
		}
		mpName := mp.Name()
		mpPath := filepath.Join(cacheDir, mpName)

		// Iterate plugin directories
		plugins, err := os.ReadDir(mpPath)
		if err != nil {
			continue
		}

		for _, plugin := range plugins {
			if !plugin.IsDir() {
				continue
			}
			pluginName := plugin.Name()
			pluginPath := filepath.Join(mpPath, pluginName)

			// Find version directories (use the latest)
			versions, err := os.ReadDir(pluginPath)
			if err != nil {
				continue
			}

			for _, ver := range versions {
				if !ver.IsDir() {
					continue
				}
				versionPath := filepath.Join(pluginPath, ver.Name())
				source := pluginName + "@" + mpName

				// Scan skills/ directory
				skillsDir := filepath.Join(versionPath, "skills")
				if entries, err := os.ReadDir(skillsDir); err == nil {
					for _, entry := range entries {
						if !entry.IsDir() {
							continue
						}
						skillMd := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
						if _, err := os.Stat(skillMd); err == nil {
							info := s.parseSkillFile(skillMd)
							if info.Name == "" {
								info.Name = entry.Name()
							}
							info.Source = source
							info.Marketplace = mpName
							info.PluginName = pluginName
							info.Type = "skill"
							info.FilePath = skillMd
							skills = append(skills, info)
						}
					}
				}

				// Scan agents/ directory
				agentsDir := filepath.Join(versionPath, "agents")
				if entries, err := os.ReadDir(agentsDir); err == nil {
					for _, entry := range entries {
						if entry.IsDir() {
							continue
						}
						name := entry.Name()
						if !strings.HasSuffix(name, ".md") {
							continue
						}
						agentMd := filepath.Join(agentsDir, name)
						info := s.parseSkillFile(agentMd)
						if info.Name == "" {
							info.Name = strings.TrimSuffix(name, ".md")
						}
						info.Source = source
						info.Marketplace = mpName
						info.PluginName = pluginName
						info.Type = "agent"
						info.FilePath = agentMd
						skills = append(skills, info)
					}
				}
			}
		}
	}

	return skills, nil
}

// GetSkillContent reads skill file content
func (s *SkillService) GetSkillContent(filePath string) (string, error) {
	// Security check: only allow reading files under .claude/plugins/cache
	home, _ := os.UserHomeDir()
	allowedPrefix := filepath.Join(home, ".claude", "plugins", "cache")
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(absPath, allowedPrefix) {
		return "", os.ErrPermission
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ========== User Custom Skills CRUD ==========

// UserSkillInfo represents user custom Skill information
type UserSkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Scope       string `json:"scope"`   // "global" or project path
	DirName     string `json:"dirName"` // Directory name (subdirectory format) or filename without .md (flat format)
	IsFlat      bool   `json:"isFlat"`  // true=flat file format(name.md), false=subdirectory format(name/SKILL.md)
}

// getSkillsBaseDir gets the base path for skills directory
func (s *SkillService) getSkillsBaseDir(scope string) string {
	if scope == "global" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".claude", "skills")
	}
	return filepath.Join(scope, ".claude", "skills")
}

// ListUserSkills lists user custom Skills
// Supports two formats:
//  1. Subdirectory format: skills/<name>/SKILL.md
//  2. Flat file format: skills/<name>.md
func (s *SkillService) ListUserSkills(scope string) ([]UserSkillInfo, error) {
	baseDir := s.getSkillsBaseDir(scope)
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return []UserSkillInfo{}, nil
	}

	var skills []UserSkillInfo
	for _, entry := range entries {
		if entry.IsDir() {
			// Subdirectory format: skills/<name>/SKILL.md
			skillMd := filepath.Join(baseDir, entry.Name(), "SKILL.md")
			if _, err := os.Stat(skillMd); err != nil {
				continue
			}
			info := s.parseSkillFile(skillMd)
			skills = append(skills, UserSkillInfo{
				Name:        info.Name,
				Description: info.Description,
				Scope:       scope,
				DirName:     entry.Name(),
				IsFlat:      false,
			})
		} else if strings.HasSuffix(entry.Name(), ".md") {
			// Flat file format: skills/<name>.md
			mdPath := filepath.Join(baseDir, entry.Name())
			info := s.parseSkillFile(mdPath)
			baseName := strings.TrimSuffix(entry.Name(), ".md")
			if info.Name == "" {
				info.Name = baseName
			}
			skills = append(skills, UserSkillInfo{
				Name:        info.Name,
				Description: info.Description,
				Scope:       scope,
				DirName:     baseName,
				IsFlat:      true,
			})
		}
	}
	return skills, nil
}

// GetUserSkill reads the full content of a user custom Skill
// isFlat is passed via the fourth parameter ("true"/"false") to distinguish between two storage formats
func (s *SkillService) GetUserSkill(scope, dirName string) (map[string]string, error) {
	baseDir := s.getSkillsBaseDir(scope)

	// Try subdirectory format first
	skillMd := filepath.Join(baseDir, dirName, "SKILL.md")
	isFlat := false
	if _, err := os.Stat(skillMd); err != nil {
		// Try flat file format
		skillMd = filepath.Join(baseDir, dirName+".md")
		isFlat = true
	}

	data, err := os.ReadFile(skillMd)
	if err != nil {
		return nil, err
	}

	rawContent := string(data)
	info := s.parseSkillFile(skillMd)

	// Extract content after frontmatter
	content := rawContent
	if strings.HasPrefix(strings.TrimSpace(rawContent), "---") {
		// Find the second ---
		rest := strings.TrimSpace(rawContent)
		rest = rest[3:] // Skip the first ---
		idx := strings.Index(rest, "---")
		if idx >= 0 {
			content = strings.TrimLeft(rest[idx+3:], "\n")
		}
	}

	flatStr := "false"
	if isFlat {
		flatStr = "true"
	}

	return map[string]string{
		"name":        info.Name,
		"description": info.Description,
		"content":     content,
		"rawContent":  rawContent,
		"isFlat":      flatStr,
	}, nil
}

// validateSkillFields validates Skill name and description fields
func (s *SkillService) validateSkillFields(name, description string) error {
	// name validation
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("name length cannot exceed 64 characters")
	}
	namePattern := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`)
	if !namePattern.MatchString(name) {
		return fmt.Errorf("name can only contain lowercase letters, numbers, and hyphens, and cannot start or end with a hyphen")
	}
	xmlTagPattern := regexp.MustCompile(`<[^>]+>`)
	if xmlTagPattern.MatchString(name) {
		return fmt.Errorf("name cannot contain XML tags")
	}
	nameLower := strings.ToLower(name)
	if strings.Contains(nameLower, "anthropic") || strings.Contains(nameLower, "claude") {
		return fmt.Errorf("name cannot contain \"anthropic\" or \"claude\"")
	}

	// description validation
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("description cannot be empty")
	}
	if len(description) > 1024 {
		return fmt.Errorf("description length cannot exceed 1024 characters")
	}
	if xmlTagPattern.MatchString(description) {
		return fmt.Errorf("description cannot contain XML tags")
	}

	return nil
}

// SaveUserSkill creates or updates a user custom Skill
// isFlat parameter: "true" means flat file format, other values mean subdirectory format
func (s *SkillService) SaveUserSkill(scope, dirName, name, description, content, isFlat string) error {
	if dirName == "" {
		return fmt.Errorf("directory name cannot be empty")
	}
	if err := s.validateSkillFields(name, description); err != nil {
		return err
	}
	baseDir := s.getSkillsBaseDir(scope)

	// Generate file content
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("name: " + name + "\n")
	sb.WriteString("description: " + description + "\n")
	sb.WriteString("---\n")
	if content != "" {
		sb.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			sb.WriteString("\n")
		}
	}

	if isFlat == "true" {
		// Flat file format: skills/<name>.md
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		skillMd := filepath.Join(baseDir, dirName+".md")
		return os.WriteFile(skillMd, []byte(sb.String()), 0644)
	}

	// Subdirectory format: skills/<name>/SKILL.md
	skillDir := filepath.Join(baseDir, dirName)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	skillMd := filepath.Join(skillDir, "SKILL.md")
	return os.WriteFile(skillMd, []byte(sb.String()), 0644)
}

// DeleteUserSkill deletes a user custom Skill
// isFlat: "true" means flat file format
func (s *SkillService) DeleteUserSkill(scope, dirName, isFlat string) error {
	if dirName == "" {
		return fmt.Errorf("directory name cannot be empty")
	}
	baseDir := s.getSkillsBaseDir(scope)
	absBase, _ := filepath.Abs(baseDir)

	if isFlat == "true" {
		// Flat file format: skills/<name>.md
		filePath := filepath.Join(baseDir, dirName+".md")
		absFile, _ := filepath.Abs(filePath)
		if !strings.HasPrefix(absFile, absBase) {
			return os.ErrPermission
		}
		return os.Remove(filePath)
	}

	// Subdirectory format: skills/<name>/
	skillDir := filepath.Join(baseDir, dirName)
	absDir, _ := filepath.Abs(skillDir)
	if !strings.HasPrefix(absDir, absBase) {
		return os.ErrPermission
	}
	return os.RemoveAll(skillDir)
}

// ========== Multi-file Skill Management ==========

// SkillFileInfo represents file information within a Skill directory
type SkillFileInfo struct {
	RelativePath string `json:"relativePath"`
	IsDir        bool   `json:"isDir"`
	Size         int64  `json:"size"`
	IsMain       bool   `json:"isMain"` // SKILL.md
}

// ListSkillFiles lists all files in a Skill directory
func (s *SkillService) ListSkillFiles(scope, dirName string) ([]SkillFileInfo, error) {
	baseDir := s.getSkillsBaseDir(scope)
	skillDir := filepath.Join(baseDir, dirName)

	// Security check
	absBase, _ := filepath.Abs(baseDir)
	absDir, _ := filepath.Abs(skillDir)
	if !strings.HasPrefix(absDir, absBase) {
		return nil, os.ErrPermission
	}

	var files []SkillFileInfo
	err := filepath.Walk(skillDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(skillDir, path)
		if rel == "." {
			return nil
		}
		files = append(files, SkillFileInfo{
			RelativePath: rel,
			IsDir:        info.IsDir(),
			Size:         info.Size(),
			IsMain:       rel == "SKILL.md",
		})
		return nil
	})
	if err != nil {
		return []SkillFileInfo{}, nil
	}
	return files, nil
}

// ReadSkillFile reads a specific file in a Skill directory
func (s *SkillService) ReadSkillFile(scope, dirName, relativePath string) (string, error) {
	baseDir := s.getSkillsBaseDir(scope)
	filePath := filepath.Join(baseDir, dirName, relativePath)

	absBase, _ := filepath.Abs(filepath.Join(baseDir, dirName))
	absFile, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFile, absBase) {
		return "", os.ErrPermission
	}

	data, err := os.ReadFile(absFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveSkillFile saves a specific file in a Skill directory (not SKILL.md)
func (s *SkillService) SaveSkillFile(scope, dirName, relativePath, content string) error {
	baseDir := s.getSkillsBaseDir(scope)
	filePath := filepath.Join(baseDir, dirName, relativePath)

	absBase, _ := filepath.Abs(filepath.Join(baseDir, dirName))
	absFile, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFile, absBase) {
		return os.ErrPermission
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(absFile), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(absFile, []byte(content), 0644)
}

// DeleteSkillFile deletes a specific file in a Skill directory (SKILL.md cannot be deleted)
func (s *SkillService) DeleteSkillFile(scope, dirName, relativePath string) error {
	if relativePath == "SKILL.md" {
		return fmt.Errorf("cannot delete the main file SKILL.md")
	}

	baseDir := s.getSkillsBaseDir(scope)
	filePath := filepath.Join(baseDir, dirName, relativePath)

	absBase, _ := filepath.Abs(filepath.Join(baseDir, dirName))
	absFile, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFile, absBase) {
		return os.ErrPermission
	}

	return os.Remove(absFile)
}

// CreateSkillFile creates a new file in a Skill directory (with initial content based on file extension)
func (s *SkillService) CreateSkillFile(scope, dirName, relativePath string) error {
	baseDir := s.getSkillsBaseDir(scope)
	filePath := filepath.Join(baseDir, dirName, relativePath)

	absBase, _ := filepath.Abs(filepath.Join(baseDir, dirName))
	absFile, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFile, absBase) {
		return os.ErrPermission
	}

	// Check if file already exists
	if _, err := os.Stat(absFile); err == nil {
		return fmt.Errorf("file already exists: %s", relativePath)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(absFile), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate initial content based on file extension
	ext := strings.ToLower(filepath.Ext(relativePath))
	var content string
	switch ext {
	case ".md":
		content = "# " + strings.TrimSuffix(filepath.Base(relativePath), ext) + "\n\n"
	case ".py":
		content = "# " + filepath.Base(relativePath) + "\n\n"
	case ".js", ".ts":
		content = "// " + filepath.Base(relativePath) + "\n\n"
	case ".json":
		content = "{}\n"
	case ".yaml", ".yml":
		content = "# " + filepath.Base(relativePath) + "\n"
	default:
		content = ""
	}

	return os.WriteFile(absFile, []byte(content), 0644)
}

// GetMCPServerNames retrieves all MCP server name list (global + all projects)
func (s *SkillService) GetMCPServerNames() ([]string, error) {
	home, _ := os.UserHomeDir()
	seen := make(map[string]bool)

	// Extract server names from .mcp.json files
	extractFromMCPFile := func(path string) {
		data, err := os.ReadFile(path)
		if err != nil {
			return
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return
		}
		serversMap := raw
		if mcpServersRaw, ok := raw["mcpServers"]; ok {
			var inner map[string]json.RawMessage
			if err := json.Unmarshal(mcpServersRaw, &inner); err == nil {
				serversMap = inner
			}
		}
		for name := range serversMap {
			if name == "$schema" || name == "mcpServers" {
				continue
			}
			seen[name] = true
		}
	}

	// 1. Read mcpServers from ~/.claude/settings.json
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		var settings map[string]json.RawMessage
		if err := json.Unmarshal(data, &settings); err == nil {
			if mcpRaw, ok := settings["mcpServers"]; ok {
				var mcpMap map[string]json.RawMessage
				if err := json.Unmarshal(mcpRaw, &mcpMap); err == nil {
					for name := range mcpMap {
						seen[name] = true
					}
				}
			}
		}
	}

	// 2. Read ~/.claude/.mcp.json
	extractFromMCPFile(filepath.Join(home, ".claude", ".mcp.json"))

	// 3. Read projects.*.mcpServers from ~/.claude.json (Claude Code CLI storage location)
	claudeJsonPath := filepath.Join(home, ".claude.json")
	if data, err := os.ReadFile(claudeJsonPath); err == nil {
		var claudeConfig struct {
			Projects map[string]struct {
				McpServers map[string]json.RawMessage `json:"mcpServers"`
			} `json:"projects"`
		}
		if err := json.Unmarshal(data, &claudeConfig); err == nil {
			for _, proj := range claudeConfig.Projects {
				for name := range proj.McpServers {
					seen[name] = true
				}
			}
		}
	}

	// 4. Scan all projects' .claude/.mcp.json
	projectsDir := filepath.Join(home, ".claude", "projects")
	entries, err := os.ReadDir(projectsDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			realPath := decodeProjectPath(entry.Name())
			if realPath == "" {
				continue
			}
			extractFromMCPFile(filepath.Join(realPath, ".claude", ".mcp.json"))
		}
	}

	// 5. Read Claude Desktop configuration
	desktopConfigPath := filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	extractFromMCPFile(desktopConfigPath)

	var names []string
	for name := range seen {
		names = append(names, name)
	}
	return names, nil
}

// parseSkillFile parses frontmatter from SKILL.md or agent .md files
func (s *SkillService) parseSkillFile(path string) SkillInfo {
	var info SkillInfo

	f, err := os.Open(path)
	if err != nil {
		return info
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		if lineCount == 1 && strings.TrimSpace(line) == "---" {
			inFrontmatter = true
			continue
		}

		if inFrontmatter {
			if strings.TrimSpace(line) == "---" {
				break
			}

			if strings.HasPrefix(line, "name:") {
				info.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			} else if strings.HasPrefix(line, "description:") {
				info.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			}
		}
	}

	return info
}
