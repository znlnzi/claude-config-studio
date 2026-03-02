package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleGetProjectConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("full project config", func(t *testing.T) {
		dir := t.TempDir()
		claudeDir := filepath.Join(dir, ".claude")
		os.MkdirAll(filepath.Join(claudeDir, "rules"), 0755)
		os.MkdirAll(filepath.Join(claudeDir, "agents"), 0755)
		os.MkdirAll(filepath.Join(claudeDir, "memory"), 0755)
		os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte("# Instructions"), 0644)
		os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{}"), 0644)
		os.WriteFile(filepath.Join(claudeDir, "rules", "test.md"), []byte("rule"), 0644)
		os.WriteFile(filepath.Join(claudeDir, "agents", "test.md"), []byte("agent"), 0644)
		os.WriteFile(filepath.Join(claudeDir, "memory", "MEMORY.md"), []byte("mem"), 0644)

		req := newToolRequest(map[string]any{"project_path": dir})
		result, err := handleGetProjectConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["has_claude_md"] != true {
			t.Error("expected has_claude_md=true")
		}
		if m["has_settings"] != true {
			t.Error("expected has_settings=true")
		}
		if m["has_rules"] != true {
			t.Error("expected has_rules=true")
		}
		if m["has_agents"] != true {
			t.Error("expected has_agents=true")
		}
		if m["has_memory"] != true {
			t.Error("expected has_memory=true")
		}
		if m["name"] != filepath.Base(dir) {
			t.Errorf("expected name=%q, got %q", filepath.Base(dir), m["name"])
		}
	})

	t.Run("empty project", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{"project_path": dir})
		result, err := handleGetProjectConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["has_claude_md"] != false {
			t.Error("expected has_claude_md=false")
		}
		if m["has_settings"] != false {
			t.Error("expected has_settings=false")
		}
	})

	t.Run("claude_md in project root", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Root instructions"), 0644)

		req := newToolRequest(map[string]any{"project_path": dir})
		result, err := handleGetProjectConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["has_claude_md"] != true {
			t.Error("expected has_claude_md=true for root CLAUDE.md")
		}
	})

	t.Run("claude_md preview truncated", func(t *testing.T) {
		dir := t.TempDir()
		claudeDir := filepath.Join(dir, ".claude")
		os.MkdirAll(claudeDir, 0755)

		longContent := make([]byte, 1000)
		for i := range longContent {
			longContent[i] = 'a'
		}
		os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), longContent, 0644)

		req := newToolRequest(map[string]any{"project_path": dir})
		result, err := handleGetProjectConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		preview, ok := m["claude_md_preview"].(string)
		if !ok {
			t.Fatal("claude_md_preview should be a string")
		}
		if len(preview) > 510 {
			t.Errorf("preview should be truncated, got length %d", len(preview))
		}
	})

	t.Run("nonexistent project path", func(t *testing.T) {
		req := newToolRequest(map[string]any{"project_path": "/nonexistent/path/abc123"})
		result, err := handleGetProjectConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error result for nonexistent path")
		}
	})

	t.Run("missing project_path", func(t *testing.T) {
		req := newToolRequest(map[string]any{})
		result, err := handleGetProjectConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error result for missing project_path")
		}
	})
}
