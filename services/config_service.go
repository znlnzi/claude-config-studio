package services

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ConfigService manages reading and writing Claude Code configuration files
type ConfigService struct {
	ctx context.Context
}

// NewConfigService creates a new config service
func NewConfigService() *ConfigService {
	return &ConfigService{}
}

// SetContext sets the wails context
func (s *ConfigService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// GlobalConfig represents the global configuration structure
type GlobalConfig struct {
	ClaudeHome string          `json:"claudeHome"`
	ClaudeMd   string          `json:"claudeMd"`
	Settings   json.RawMessage `json:"settings"`
	HasClaudeMd bool           `json:"hasClaudeMd"`
	HasSettings bool           `json:"hasSettings"`
}

// SettingsJSON represents the settings.json structure
type SettingsJSON struct {
	Env                   map[string]string      `json:"env,omitempty"`
	Hooks                 map[string]interface{} `json:"hooks,omitempty"`
	StatusLine            interface{}            `json:"statusLine,omitempty"`
	EnabledPlugins        map[string]interface{} `json:"enabledPlugins,omitempty"`
	Language              string                 `json:"language,omitempty"`
	AlwaysThinkingEnabled bool                   `json:"alwaysThinkingEnabled,omitempty"`
}

// ProjectConfig represents the project-level configuration structure
type ProjectConfig struct {
	Path        string          `json:"path"`
	ClaudeMd    string          `json:"claudeMd"`
	Settings    json.RawMessage `json:"settings"`
	McpConfig   json.RawMessage `json:"mcpConfig"`
	HasClaudeMd bool            `json:"hasClaudeMd"`
	HasSettings bool            `json:"hasSettings"`
	HasMcp      bool            `json:"hasMcp"`
}

// GetClaudeHome returns the Claude configuration home directory
func (s *ConfigService) GetClaudeHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude")
}

// GetGlobalConfig returns the global configuration
func (s *ConfigService) GetGlobalConfig() (*GlobalConfig, error) {
	claudeHome := s.GetClaudeHome()
	config := &GlobalConfig{
		ClaudeHome: claudeHome,
	}

	// Read CLAUDE.md
	claudeMdPath := filepath.Join(claudeHome, "CLAUDE.md")
	if data, err := os.ReadFile(claudeMdPath); err == nil {
		config.ClaudeMd = string(data)
		config.HasClaudeMd = true
	}

	// Read settings.json
	settingsPath := filepath.Join(claudeHome, "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		config.Settings = json.RawMessage(data)
		config.HasSettings = true
	}

	return config, nil
}

// SaveGlobalClaudeMd saves the global CLAUDE.md
func (s *ConfigService) SaveGlobalClaudeMd(content string) error {
	claudeHome := s.GetClaudeHome()
	claudeMdPath := filepath.Join(claudeHome, "CLAUDE.md")
	return os.WriteFile(claudeMdPath, []byte(content), 0644)
}

// SaveGlobalSettings saves the global settings.json
func (s *ConfigService) SaveGlobalSettings(content string) error {
	// Validate JSON format
	var js json.RawMessage
	if err := json.Unmarshal([]byte(content), &js); err != nil {
		return err
	}

	// Format JSON
	formatted, err := json.MarshalIndent(js, "", "  ")
	if err != nil {
		return err
	}

	claudeHome := s.GetClaudeHome()
	settingsPath := filepath.Join(claudeHome, "settings.json")
	return os.WriteFile(settingsPath, formatted, 0644)
}

// GetProjectConfig returns the project-level configuration
func (s *ConfigService) GetProjectConfig(projectPath string) (*ProjectConfig, error) {
	config := &ProjectConfig{
		Path: projectPath,
	}

	claudeDir := filepath.Join(projectPath, ".claude")

	// Read project-level CLAUDE.md (prefer .claude/CLAUDE.md, fallback to root CLAUDE.md)
	claudeMdPath := filepath.Join(claudeDir, "CLAUDE.md")
	if data, err := os.ReadFile(claudeMdPath); err == nil {
		config.ClaudeMd = string(data)
		config.HasClaudeMd = true
	} else {
		rootClaudeMd := filepath.Join(projectPath, "CLAUDE.md")
		if data, err := os.ReadFile(rootClaudeMd); err == nil {
			config.ClaudeMd = string(data)
			config.HasClaudeMd = true
		}
	}

	// Read project-level settings.json
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		config.Settings = json.RawMessage(data)
		config.HasSettings = true
	}

	// Read MCP configuration
	mcpPath := filepath.Join(claudeDir, ".mcp.json")
	if data, err := os.ReadFile(mcpPath); err == nil {
		config.McpConfig = json.RawMessage(data)
		config.HasMcp = true
	}

	return config, nil
}

// SaveProjectClaudeMd saves the project-level CLAUDE.md
func (s *ConfigService) SaveProjectClaudeMd(projectPath string, content string) error {
	claudeDir := filepath.Join(projectPath, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}
	claudeMdPath := filepath.Join(claudeDir, "CLAUDE.md")
	return os.WriteFile(claudeMdPath, []byte(content), 0644)
}

// SaveProjectSettings saves the project-level settings.json
func (s *ConfigService) SaveProjectSettings(projectPath string, content string) error {
	var js json.RawMessage
	if err := json.Unmarshal([]byte(content), &js); err != nil {
		return err
	}

	formatted, err := json.MarshalIndent(js, "", "  ")
	if err != nil {
		return err
	}

	claudeDir := filepath.Join(projectPath, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	return os.WriteFile(settingsPath, formatted, 0644)
}

// SaveProjectMcp saves the project-level MCP configuration
func (s *ConfigService) SaveProjectMcp(projectPath string, content string) error {
	var js json.RawMessage
	if err := json.Unmarshal([]byte(content), &js); err != nil {
		return err
	}

	formatted, err := json.MarshalIndent(js, "", "  ")
	if err != nil {
		return err
	}

	claudeDir := filepath.Join(projectPath, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}
	mcpPath := filepath.Join(claudeDir, ".mcp.json")
	return os.WriteFile(mcpPath, formatted, 0644)
}

// SelectDirectory opens the native directory picker
func (s *ConfigService) SelectDirectory() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(s.ctx, runtime.OpenDialogOptions{
		Title: "Select Project Directory",
	})
	return dir, err
}
