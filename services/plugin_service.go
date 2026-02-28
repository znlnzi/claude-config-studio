package services

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

// PluginService manages plugin operations
type PluginService struct {
	ctx context.Context
}

func NewPluginService() *PluginService {
	return &PluginService{}
}

func (s *PluginService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// PluginInfo holds plugin information
type PluginInfo struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
}

// GetEnabledPlugins returns the list of enabled plugins
func (s *PluginService) GetEnabledPlugins() ([]PluginInfo, error) {
	home, _ := os.UserHomeDir()
	settingsPath := filepath.Join(home, ".claude", "settings.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return []PluginInfo{}, nil
	}

	var settings map[string]json.RawMessage
	if err := json.Unmarshal(data, &settings); err != nil {
		return []PluginInfo{}, nil
	}

	pluginsRaw, ok := settings["enabledPlugins"]
	if !ok {
		return []PluginInfo{}, nil
	}

	var plugins map[string]interface{}
	if err := json.Unmarshal(pluginsRaw, &plugins); err != nil {
		return []PluginInfo{}, nil
	}

	var result []PluginInfo
	for name, val := range plugins {
		info := PluginInfo{
			Name:    name,
			Enabled: true,
		}

		// Parse source (name format: plugin-name@source)
		parts := splitPluginName(name)
		info.Name = parts[0]
		if len(parts) > 1 {
			info.Source = parts[1]
		}

		// Try to read plugin metadata
		if desc := s.getPluginDescription(name); desc != "" {
			info.Description = desc
		}

		_ = val
		result = append(result, info)
	}

	return result, nil
}

// TogglePlugin enables/disables a plugin
func (s *PluginService) TogglePlugin(pluginKey string, enabled bool) error {
	home, _ := os.UserHomeDir()
	settingsPath := filepath.Join(home, ".claude", "settings.json")

	// Read existing settings
	var settings map[string]json.RawMessage
	if data, err := os.ReadFile(settingsPath); err == nil {
		json.Unmarshal(data, &settings)
	}
	if settings == nil {
		settings = make(map[string]json.RawMessage)
	}

	// Read enabledPlugins
	var plugins map[string]interface{}
	if raw, ok := settings["enabledPlugins"]; ok {
		json.Unmarshal(raw, &plugins)
	}
	if plugins == nil {
		plugins = make(map[string]interface{})
	}

	if enabled {
		plugins[pluginKey] = true
	} else {
		delete(plugins, pluginKey)
	}

	pluginsJSON, _ := json.Marshal(plugins)
	settings["enabledPlugins"] = json.RawMessage(pluginsJSON)

	formatted, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, formatted, 0644)
}

// GetInstalledPlugins returns the list of installed plugins (read from plugins directory)
func (s *PluginService) GetInstalledPlugins() ([]PluginInfo, error) {
	home, _ := os.UserHomeDir()
	pluginsDir := filepath.Join(home, ".claude", "plugins", "installed_plugins.json")

	data, err := os.ReadFile(pluginsDir)
	if err != nil {
		return []PluginInfo{}, nil
	}

	var installed map[string]interface{}
	if err := json.Unmarshal(data, &installed); err != nil {
		return []PluginInfo{}, nil
	}

	// Get enabled status
	enabledMap := s.getEnabledMap()

	var result []PluginInfo
	for key := range installed {
		parts := splitPluginName(key)
		info := PluginInfo{
			Name:    parts[0],
			Enabled: enabledMap[key],
		}
		if len(parts) > 1 {
			info.Source = parts[1]
		}
		result = append(result, info)
	}

	return result, nil
}

func (s *PluginService) getEnabledMap() map[string]bool {
	home, _ := os.UserHomeDir()
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return map[string]bool{}
	}
	var settings map[string]json.RawMessage
	json.Unmarshal(data, &settings)
	raw, ok := settings["enabledPlugins"]
	if !ok {
		return map[string]bool{}
	}
	var plugins map[string]interface{}
	json.Unmarshal(raw, &plugins)
	result := make(map[string]bool)
	for k := range plugins {
		result[k] = true
	}
	return result
}

func (s *PluginService) getPluginDescription(name string) string {
	home, _ := os.UserHomeDir()
	// Try to read description from plugin cache
	parts := splitPluginName(name)
	if len(parts) < 2 {
		return ""
	}
	pluginJSON := filepath.Join(home, ".claude", "plugins", "cache", parts[1], parts[0], ".claude-plugin", "plugin.json")
	data, err := os.ReadFile(pluginJSON)
	if err != nil {
		return ""
	}
	var meta map[string]interface{}
	json.Unmarshal(data, &meta)
	if desc, ok := meta["description"].(string); ok {
		return desc
	}
	return ""
}

func splitPluginName(name string) []string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '@' {
			return []string{name[:i], name[i+1:]}
		}
	}
	return []string{name}
}
