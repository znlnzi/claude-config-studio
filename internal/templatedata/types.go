package templatedata

// Template defines a configuration template.
type Template struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Category    string            `json:"category"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	ClaudeMd    string            `json:"claudeMd,omitempty"`
	Settings    interface{}       `json:"settings,omitempty"`
	McpServers  interface{}       `json:"mcpServers,omitempty"`
	Hooks       interface{}       `json:"hooks,omitempty"`
	Agents      map[string]string `json:"agents,omitempty"`
	Commands    map[string]string `json:"commands,omitempty"`
	Skills      map[string]string `json:"skills,omitempty"`
	Rules       map[string]string `json:"rules,omitempty"`
	Scripts     map[string]string `json:"scripts,omitempty"` // .claude/scripts/{name} executable scripts
}

// TemplateCategory defines a category of templates.
type TemplateCategory struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Icon      string     `json:"icon"`
	Templates []Template `json:"templates"`
}

// InstalledTemplateInfo holds information about an installed template.
type InstalledTemplateInfo struct {
	TemplateID string `json:"templateId"`
	Scope      string `json:"scope"`    // "global" | "project"
	FilePath   string `json:"filePath"` // actual file path
}
