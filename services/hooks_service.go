package services

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

// HooksService manages Claude Code Hooks configuration
type HooksService struct {
	ctx context.Context
}

func NewHooksService() *HooksService {
	return &HooksService{}
}

func (s *HooksService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// HookCommand represents a single hook command
type HookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

// HookEntry represents an entry under a hook event
type HookEntry struct {
	Matcher string        `json:"matcher,omitempty"` // tool matcher (for PreToolUse/PostToolUse)
	Hooks   []HookCommand `json:"hooks"`
}

// HooksConfig represents the complete hooks configuration
type HooksConfig struct {
	Event   string      `json:"event"`
	Entries []HookEntry `json:"entries"`
}

// GetGlobalHooks returns global hooks configuration (read from settings.json)
func (s *HooksService) GetGlobalHooks() ([]HooksConfig, error) {
	home, _ := os.UserHomeDir()
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	return s.readHooksFromSettings(settingsPath)
}

// SaveGlobalHooks saves global hooks configuration (writes to settings.json)
func (s *HooksService) SaveGlobalHooks(hooks []HooksConfig) error {
	home, _ := os.UserHomeDir()
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	return s.writeHooksToSettings(settingsPath, hooks)
}

// GetProjectHooks returns project-level hooks configuration
func (s *HooksService) GetProjectHooks(projectPath string) ([]HooksConfig, error) {
	settingsPath := filepath.Join(projectPath, ".claude", "settings.json")
	return s.readHooksFromSettings(settingsPath)
}

// SaveProjectHooks saves project-level hooks configuration
func (s *HooksService) SaveProjectHooks(projectPath string, hooks []HooksConfig) error {
	claudeDir := filepath.Join(projectPath, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")
	return s.writeHooksToSettings(settingsPath, hooks)
}

func (s *HooksService) readHooksFromSettings(path string) ([]HooksConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return []HooksConfig{}, nil
	}

	var settings map[string]json.RawMessage
	if err := json.Unmarshal(data, &settings); err != nil {
		return []HooksConfig{}, nil
	}

	hooksRaw, ok := settings["hooks"]
	if !ok {
		return []HooksConfig{}, nil
	}

	// hooks format: { "EventName": [ { "matcher": "...", "hooks": [...] } ] }
	var hooksMap map[string][]json.RawMessage
	if err := json.Unmarshal(hooksRaw, &hooksMap); err != nil {
		return []HooksConfig{}, nil
	}

	var result []HooksConfig
	eventOrder := []string{"PreToolUse", "PostToolUse", "SessionStart", "Stop", "UserPromptSubmit"}
	for _, event := range eventOrder {
		entries, ok := hooksMap[event]
		if !ok || len(entries) == 0 {
			continue
		}

		config := HooksConfig{Event: event}
		for _, entryRaw := range entries {
			var entry HookEntry
			if err := json.Unmarshal(entryRaw, &entry); err != nil {
				continue
			}
			config.Entries = append(config.Entries, entry)
		}
		result = append(result, config)
	}

	// Handle events not in eventOrder
	for event, entries := range hooksMap {
		found := false
		for _, e := range eventOrder {
			if e == event {
				found = true
				break
			}
		}
		if found {
			continue
		}
		config := HooksConfig{Event: event}
		for _, entryRaw := range entries {
			var entry HookEntry
			if err := json.Unmarshal(entryRaw, &entry); err != nil {
				continue
			}
			config.Entries = append(config.Entries, entry)
		}
		result = append(result, config)
	}

	return result, nil
}

func (s *HooksService) writeHooksToSettings(path string, hooks []HooksConfig) error {
	// Read existing settings.json
	var settings map[string]json.RawMessage
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &settings)
	}
	if settings == nil {
		settings = make(map[string]json.RawMessage)
	}

	// Build hooks object
	hooksMap := make(map[string][]HookEntry)
	for _, h := range hooks {
		if len(h.Entries) > 0 {
			hooksMap[h.Event] = h.Entries
		}
	}

	if len(hooksMap) > 0 {
		hooksJSON, err := json.Marshal(hooksMap)
		if err != nil {
			return err
		}
		settings["hooks"] = json.RawMessage(hooksJSON)
	} else {
		delete(settings, "hooks")
	}

	formatted, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, formatted, 0644)
}
