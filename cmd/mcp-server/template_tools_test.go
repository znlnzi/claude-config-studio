package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleListTemplates(t *testing.T) {
	ctx := context.Background()

	t.Run("returns templates", func(t *testing.T) {
		req := newToolRequest(map[string]any{})
		result, err := handleListTemplates(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		totalTemplates, _ := m["total_templates"].(float64)
		if totalTemplates == 0 {
			t.Error("expected at least one template")
		}

		categories, ok := m["categories"].([]any)
		if !ok || len(categories) == 0 {
			t.Error("expected at least one category")
		}
	})
}

func TestHandleInstallTemplate(t *testing.T) {
	ctx := context.Background()

	t.Run("install hackathon-core", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"template_id":  "hackathon-core",
			"scope":        "project",
			"project_path": dir,
		})

		result, err := handleInstallTemplate(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["success"] != true {
			t.Error("expected success=true")
		}
		if m["template_id"] != "hackathon-core" {
			t.Errorf("expected template_id=hackathon-core, got %v", m["template_id"])
		}

		// Verify rules file was created
		rulesFile := filepath.Join(dir, ".claude", "rules", "tpl-hackathon-core.md")
		if !fileExists(rulesFile) {
			t.Error("template rules file should exist")
		}
	})

	t.Run("install cross-session-memory", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"template_id":  "cross-session-memory",
			"scope":        "project",
			"project_path": dir,
		})

		result, err := handleInstallTemplate(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["success"] != true {
			t.Error("expected success=true")
		}
	})

	t.Run("nonexistent template", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"template_id":  "nonexistent-template",
			"scope":        "project",
			"project_path": dir,
		})

		result, err := handleInstallTemplate(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for nonexistent template")
		}
	})

	t.Run("project scope without path", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"template_id": "hackathon-core",
			"scope":       "project",
		})

		result, err := handleInstallTemplate(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for missing project_path")
		}
	})

	t.Run("missing template_id", func(t *testing.T) {
		req := newToolRequest(map[string]any{})
		result, err := handleInstallTemplate(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for missing template_id")
		}
	})
}

func TestHandleUninstallTemplate(t *testing.T) {
	ctx := context.Background()

	t.Run("uninstall existing", func(t *testing.T) {
		dir := t.TempDir()
		rulesDir := filepath.Join(dir, ".claude", "rules")
		os.MkdirAll(rulesDir, 0755)
		os.WriteFile(filepath.Join(rulesDir, "tpl-hackathon-core.md"), []byte("content"), 0644)

		req := newToolRequest(map[string]any{
			"template_id":  "hackathon-core",
			"scope":        "project",
			"project_path": dir,
		})

		result, err := handleUninstallTemplate(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["removed"] != true {
			t.Error("expected removed=true")
		}

		if fileExists(filepath.Join(rulesDir, "tpl-hackathon-core.md")) {
			t.Error("template file should have been deleted")
		}
	})

	t.Run("uninstall nonexistent", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"template_id":  "nonexistent",
			"scope":        "project",
			"project_path": dir,
		})

		result, err := handleUninstallTemplate(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["removed"] != false {
			t.Error("expected removed=false for nonexistent template")
		}
	})
}

func TestHandleGetInstalledTemplates(t *testing.T) {
	ctx := context.Background()

	t.Run("finds installed templates", func(t *testing.T) {
		dir := t.TempDir()
		rulesDir := filepath.Join(dir, ".claude", "rules")
		os.MkdirAll(rulesDir, 0755)
		os.WriteFile(filepath.Join(rulesDir, "tpl-hackathon-core.md"), []byte("template"), 0644)
		os.WriteFile(filepath.Join(rulesDir, "tpl-cross-session-memory.md"), []byte("template"), 0644)
		os.WriteFile(filepath.Join(rulesDir, "not-a-template.md"), []byte("other"), 0644)

		req := newToolRequest(map[string]any{
			"scope":        "project",
			"project_path": dir,
		})

		result, err := handleGetInstalledTemplates(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		total, _ := m["total"].(float64)
		if total != 2 {
			t.Errorf("expected 2 installed templates, got %v", total)
		}
	})

	t.Run("empty rules dir", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"scope":        "project",
			"project_path": dir,
		})

		result, err := handleGetInstalledTemplates(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		installed, ok := m["installed"].([]any)
		if !ok {
			t.Fatal("installed should be an array")
		}
		if len(installed) != 0 {
			t.Errorf("expected 0, got %d", len(installed))
		}
	})
}
