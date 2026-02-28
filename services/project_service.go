package services

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ProjectService handles project scanning and management
type ProjectService struct {
	ctx context.Context
}

// NewProjectService creates a new project service
func NewProjectService() *ProjectService {
	return &ProjectService{}
}

// SetContext sets the wails context
func (s *ProjectService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// ProjectInfo holds project information
type ProjectInfo struct {
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	HasClaudeMd   bool      `json:"hasClaudeMd"`
	HasSettings   bool      `json:"hasSettings"`
	HasMcp        bool      `json:"hasMcp"`
	HasHooks      bool      `json:"hasHooks"`
	HasCommands   bool      `json:"hasCommands"`
	HasAgents     bool      `json:"hasAgents"`
	HasSkills     bool      `json:"hasSkills"`
	LastModified  time.Time `json:"lastModified"`
	ConfigCount   int       `json:"configCount"`
}

// ScanProjects scans for existing Claude Code projects
func (s *ProjectService) ScanProjects() ([]ProjectInfo, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	projectsDir := filepath.Join(home, ".claude", "projects")
	var projects []ProjectInfo

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return projects, nil // return empty if directory doesn't exist
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Decode the encoded path back to real path
		name := entry.Name()
		realPath := decodeProjectPath(name)

		if realPath == "" {
			continue
		}

		// Check if the real path exists
		if _, err := os.Stat(realPath); os.IsNotExist(err) {
			continue
		}

		info := ProjectInfo{
			Name: filepath.Base(realPath),
			Path: realPath,
		}

		// Check for various config files
		claudeDir := filepath.Join(realPath, ".claude")
		info.HasClaudeMd = fileExists(filepath.Join(claudeDir, "CLAUDE.md")) ||
			fileExists(filepath.Join(realPath, "CLAUDE.md"))
		info.HasSettings = fileExists(filepath.Join(claudeDir, "settings.json"))
		info.HasMcp = fileExists(filepath.Join(claudeDir, ".mcp.json"))
		info.HasHooks = dirHasFiles(filepath.Join(claudeDir, "hooks"))
		info.HasCommands = dirHasFiles(filepath.Join(claudeDir, "commands"))
		info.HasAgents = dirHasFiles(filepath.Join(claudeDir, "agents"))
		info.HasSkills = dirHasFiles(filepath.Join(claudeDir, "skills"))

		// Count config items
		count := 0
		if info.HasClaudeMd { count++ }
		if info.HasSettings { count++ }
		if info.HasMcp { count++ }
		if info.HasHooks { count++ }
		if info.HasCommands { count++ }
		if info.HasAgents { count++ }
		if info.HasSkills { count++ }
		info.ConfigCount = count

		// Get last modification time
		if fi, err := entry.Info(); err == nil {
			info.LastModified = fi.ModTime()
		}

		projects = append(projects, info)
	}

	// Sort by modification time (newest first)
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastModified.After(projects[j].LastModified)
	})

	return projects, nil
}

// GetProjectDetail returns detailed project configuration
func (s *ProjectService) GetProjectDetail(projectPath string) (*ProjectInfo, error) {
	info := &ProjectInfo{
		Name: filepath.Base(projectPath),
		Path: projectPath,
	}

	claudeDir := filepath.Join(projectPath, ".claude")
	info.HasClaudeMd = fileExists(filepath.Join(claudeDir, "CLAUDE.md")) ||
		fileExists(filepath.Join(projectPath, "CLAUDE.md"))
	info.HasSettings = fileExists(filepath.Join(claudeDir, "settings.json"))
	info.HasMcp = fileExists(filepath.Join(claudeDir, ".mcp.json"))
	info.HasHooks = dirHasFiles(filepath.Join(claudeDir, "hooks"))
	info.HasCommands = dirHasFiles(filepath.Join(claudeDir, "commands"))
	info.HasAgents = dirHasFiles(filepath.Join(claudeDir, "agents"))
	info.HasSkills = dirHasFiles(filepath.Join(claudeDir, "skills"))

	count := 0
	if info.HasClaudeMd { count++ }
	if info.HasSettings { count++ }
	if info.HasMcp { count++ }
	if info.HasHooks { count++ }
	if info.HasCommands { count++ }
	if info.HasAgents { count++ }
	if info.HasSkills { count++ }
	info.ConfigCount = count

	return info, nil
}

// InitProjectConfig initializes project configuration (creates .claude directory)
func (s *ProjectService) InitProjectConfig(projectPath string) error {
	claudeDir := filepath.Join(projectPath, ".claude")
	return os.MkdirAll(claudeDir, 0755)
}

// GetGlobalStats returns global statistics
func (s *ProjectService) GetGlobalStats() (map[string]interface{}, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	claudeHome := filepath.Join(home, ".claude")
	stats := map[string]interface{}{
		"hasGlobalClaudeMd":  fileExists(filepath.Join(claudeHome, "CLAUDE.md")),
		"hasGlobalSettings":  fileExists(filepath.Join(claudeHome, "settings.json")),
		"hasLspConfig":       fileExists(filepath.Join(claudeHome, "cclsp.json")),
		"globalAgentCount":   countFiles(filepath.Join(claudeHome, "agents"), ".md"),
		"globalCommandCount": countFiles(filepath.Join(claudeHome, "commands"), ".md"),
	}

	// Count projects
	projects, _ := s.ScanProjects()
	stats["projectCount"] = len(projects)

	// Read settings.json to get enabled plugin count
	settingsPath := filepath.Join(claudeHome, "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		var settings map[string]json.RawMessage
		if err := json.Unmarshal(data, &settings); err == nil {
			if plugins, ok := settings["enabledPlugins"]; ok {
				var pluginMap map[string]interface{}
				if err := json.Unmarshal(plugins, &pluginMap); err == nil {
					stats["enabledPluginCount"] = len(pluginMap)
				}
			}
		}
	}

	return stats, nil
}

// Helper functions

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func dirHasFiles(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	return len(entries) > 0
}

func countFiles(dir string, ext string) int {
	count := 0
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ext) {
			count++
		}
	}
	return count
}

// decodeProjectPath converts encoded directory names under ~/.claude/projects/ to real paths
// Claude Code encoding rule: replaces /, _, spaces, and non-ASCII characters with -
// Therefore decoding requires recursive filesystem matching to determine the correct path
func decodeProjectPath(encoded string) string {
	if encoded == "" || !strings.HasPrefix(encoded, "-") {
		return ""
	}
	return resolveEncodedPath("/", encoded[1:])
}

// resolveEncodedPath recursively matches the filesystem to decode the path
func resolveEncodedPath(dir, remaining string) string {
	if remaining == "" {
		return dir
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	type candidate struct {
		name    string
		encLen  int // encoded length consumed by matching (including separator)
		isExact bool
	}

	var candidates []candidate
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		enc := encodePathSegment(e.Name())
		if remaining == enc {
			candidates = append(candidates, candidate{e.Name(), len(enc), true})
		} else if strings.HasPrefix(remaining, enc+"-") {
			candidates = append(candidates, candidate{e.Name(), len(enc) + 1, false})
		}
	}

	// Sort by encoded length descending (greedy matching, try longest first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].encLen > candidates[j].encLen
	})

	for _, c := range candidates {
		nextDir := filepath.Join(dir, c.name)
		if c.isExact {
			if _, err := os.Stat(nextDir); err == nil {
				return nextDir
			}
		} else {
			result := resolveEncodedPath(nextDir, remaining[c.encLen:])
			if result != "" {
				return result
			}
		}
	}

	return ""
}

// encodePathSegment encodes a directory name according to Claude Code rules
// Characters not in [a-zA-Z0-9.-] are replaced with -
func encodePathSegment(name string) string {
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '.' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	return b.String()
}
