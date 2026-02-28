package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zenglingzi/claude-config-studio/internal/templatedata"
)

// TemplateService manages template operations
type TemplateService struct {
	ctx context.Context
}

func NewTemplateService() *TemplateService {
	return &TemplateService{}
}

func (s *TemplateService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// Template definition (type alias pointing to shared package)
type Template = templatedata.Template

// TemplateCategory template category (type alias pointing to shared package)
type TemplateCategory = templatedata.TemplateCategory

// InstalledTemplateInfo installed template info (type alias pointing to shared package)
type InstalledTemplateInfo = templatedata.InstalledTemplateInfo

// GetTemplates returns all built-in templates
func (s *TemplateService) GetTemplates() []TemplateCategory {
	return templatedata.GetAllTemplates()
}

// GetTemplateByID returns a template by its ID
func (s *TemplateService) GetTemplateByID(id string) *Template {
	return templatedata.GetTemplateByID(id)
}

// ApplyTemplate applies a template to a project
// overwrite=false: incremental merge (append CLAUDE.md, merge JSON, skip existing extension files)
// overwrite=true: force overwrite (used for updating installed templates)
func (s *TemplateService) ApplyTemplate(projectPath string, templateID string, overwrite bool) error {
	tmpl := s.GetTemplateByID(templateID)
	if tmpl == nil {
		return nil
	}

	claudeDir := filepath.Join(projectPath, ".claude")
	os.MkdirAll(claudeDir, 0755)

	if tmpl.ClaudeMd != "" {
		claudeMdPath := filepath.Join(claudeDir, "CLAUDE.md")
		if overwrite {
			if err := os.WriteFile(claudeMdPath, []byte(tmpl.ClaudeMd), 0644); err != nil {
				return err
			}
		} else if existing, err := os.ReadFile(claudeMdPath); err == nil && len(strings.TrimSpace(string(existing))) > 0 {
			merged := strings.TrimRight(string(existing), "\n") + "\n\n" + tmpl.ClaudeMd
			if err := os.WriteFile(claudeMdPath, []byte(merged), 0644); err != nil {
				return err
			}
		} else {
			if err := os.WriteFile(claudeMdPath, []byte(tmpl.ClaudeMd), 0644); err != nil {
				return err
			}
		}
	}

	// settings.json and .mcp.json are always merged, unaffected by overwrite flag
	if tmpl.Settings != nil {
		if err := templatedata.MergeAndWriteJSON(filepath.Join(claudeDir, "settings.json"), tmpl.Settings); err != nil {
			return err
		}
	}

	if tmpl.McpServers != nil {
		if err := templatedata.MergeAndWriteJSON(filepath.Join(claudeDir, ".mcp.json"), tmpl.McpServers); err != nil {
			return err
		}
	}

	if err := templatedata.WriteExtensionFiles(claudeDir, "agents", tmpl.Agents, overwrite); err != nil {
		return err
	}
	if err := templatedata.WriteExtensionFiles(claudeDir, "commands", tmpl.Commands, overwrite); err != nil {
		return err
	}
	if err := templatedata.WriteSkillFiles(claudeDir, tmpl.Skills, overwrite); err != nil {
		return err
	}
	if err := templatedata.WriteExtensionFiles(claudeDir, "rules", tmpl.Rules, overwrite); err != nil {
		return err
	}

	return nil
}

// MigrateFlatSkills migrates flat-format skills to directory-based format, returns migration count
func (s *TemplateService) MigrateFlatSkills(projectPath string) (int, error) {
	skillsDir := filepath.Join(projectPath, ".claude", "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return 0, nil
	}
	return templatedata.MigrateFlatSkills(skillsDir), nil
}

// getRulesDir determines the rules directory based on scope
func getRulesDir(scope, targetPath string) (string, error) {
	if scope == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(home, ".claude", "rules"), nil
	}
	if targetPath == "" {
		return "", fmt.Errorf("project path cannot be empty")
	}
	return filepath.Join(targetPath, ".claude", "rules"), nil
}

// getClaudeDir determines the .claude directory based on scope
func getClaudeDir(scope, targetPath string) (string, error) {
	if scope == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(home, ".claude"), nil
	}
	if targetPath == "" {
		return "", fmt.Errorf("project path cannot be empty")
	}
	return filepath.Join(targetPath, ".claude"), nil
}

// InstallTemplateRules installs templates as rules files (supports multiple selection)
// overwrite=false: incremental merge; overwrite=true: force overwrite
func (s *TemplateService) InstallTemplateRules(scope string, targetPath string, templateIDs []string, overwrite bool) error {
	rulesDir, err := getRulesDir(scope, targetPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return fmt.Errorf("failed to create rules directory: %w", err)
	}

	claudeDir, err := getClaudeDir(scope, targetPath)
	if err != nil {
		return err
	}

	for _, id := range templateIDs {
		tmpl := s.GetTemplateByID(id)
		if tmpl == nil {
			continue
		}

		// Write ClaudeMd to rules/tpl-{id}.md
		if tmpl.ClaudeMd != "" {
			header := fmt.Sprintf("<!-- template: %s | %s -->\n\n", tmpl.ID, tmpl.Name)
			content := header + tmpl.ClaudeMd
			filePath := filepath.Join(rulesDir, "tpl-"+tmpl.ID+".md")
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write template %s: %w", tmpl.ID, err)
			}
		}

		// Write accompanying components (agents/commands/skills/rules)
		if err := templatedata.WriteExtensionFiles(claudeDir, "agents", tmpl.Agents, overwrite); err != nil {
			return err
		}
		if err := templatedata.WriteExtensionFiles(claudeDir, "commands", tmpl.Commands, overwrite); err != nil {
			return err
		}
		if err := templatedata.WriteSkillFiles(claudeDir, tmpl.Skills, overwrite); err != nil {
			return err
		}
		if err := templatedata.WriteExtensionFiles(claudeDir, "rules", tmpl.Rules, overwrite); err != nil {
			return err
		}

		// settings.json is always merged, unaffected by overwrite flag
		if tmpl.Settings != nil {
			templatedata.MergeAndWriteJSON(filepath.Join(claudeDir, "settings.json"), tmpl.Settings)
		}
	}

	return nil
}

// UninstallTemplateRules removes template rules files
func (s *TemplateService) UninstallTemplateRules(scope string, targetPath string, templateIDs []string) error {
	rulesDir, err := getRulesDir(scope, targetPath)
	if err != nil {
		return err
	}

	for _, id := range templateIDs {
		filePath := filepath.Join(rulesDir, "tpl-"+id+".md")
		os.Remove(filePath) // ignore file-not-found errors
	}

	return nil
}

// GetInstalledTemplates returns the list of installed templates
func (s *TemplateService) GetInstalledTemplates(scope string, targetPath string) []InstalledTemplateInfo {
	rulesDir, err := getRulesDir(scope, targetPath)
	if err != nil {
		return nil
	}

	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return nil
	}

	var result []InstalledTemplateInfo
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() && strings.HasPrefix(name, "tpl-") && strings.HasSuffix(name, ".md") {
			templateID := strings.TrimSuffix(strings.TrimPrefix(name, "tpl-"), ".md")
			result = append(result, InstalledTemplateInfo{
				TemplateID: templateID,
				Scope:      scope,
				FilePath:   filepath.Join(rulesDir, name),
			})
		}
	}

	return result
}
