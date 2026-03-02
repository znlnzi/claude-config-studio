package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleListExtensions(t *testing.T) {
	ctx := context.Background()

	t.Run("list rules files", func(t *testing.T) {
		dir := t.TempDir()
		rulesDir := filepath.Join(dir, ".claude", "rules")
		os.MkdirAll(rulesDir, 0755)
		os.WriteFile(filepath.Join(rulesDir, "coding.md"), []byte("# Coding"), 0644)
		os.WriteFile(filepath.Join(rulesDir, "testing.md"), []byte("# Testing"), 0644)

		req := newToolRequest(map[string]any{"type": "rules", "scope": dir})
		result, err := handleListExtensions(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		total, _ := m["total"].(float64)
		if total != 2 {
			t.Errorf("expected 2 files, got %v", total)
		}
	})

	t.Run("list skills with directory format", func(t *testing.T) {
		dir := t.TempDir()
		skillsDir := filepath.Join(dir, ".claude", "skills")
		os.MkdirAll(filepath.Join(skillsDir, "myskill"), 0755)
		os.WriteFile(filepath.Join(skillsDir, "myskill", "SKILL.md"), []byte("# Skill"), 0644)

		req := newToolRequest(map[string]any{"type": "skills", "scope": dir})
		result, err := handleListExtensions(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		files, ok := m["files"].([]any)
		if !ok {
			t.Fatal("files should be an array")
		}
		if len(files) != 1 {
			t.Errorf("expected 1 skill, got %d", len(files))
		}
		if len(files) > 0 {
			f := files[0].(map[string]any)
			if f["is_dir"] != true {
				t.Error("expected skill to be marked as directory")
			}
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{"type": "agents", "scope": dir})
		result, err := handleListExtensions(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		files, ok := m["files"].([]any)
		if !ok {
			t.Fatal("files should be an array")
		}
		if len(files) != 0 {
			t.Errorf("expected 0 files, got %d", len(files))
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		req := newToolRequest(map[string]any{"type": "invalid", "scope": "global"})
		result, err := handleListExtensions(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for invalid type")
		}
	})

	t.Run("missing type", func(t *testing.T) {
		req := newToolRequest(map[string]any{})
		result, err := handleListExtensions(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for missing type")
		}
	})
}

func TestHandleReadExtension(t *testing.T) {
	ctx := context.Background()

	t.Run("read flat file", func(t *testing.T) {
		dir := t.TempDir()
		rulesDir := filepath.Join(dir, ".claude", "rules")
		os.MkdirAll(rulesDir, 0755)
		os.WriteFile(filepath.Join(rulesDir, "coding.md"), []byte("# Coding Rules"), 0644)

		req := newToolRequest(map[string]any{
			"type":  "rules",
			"name":  "coding",
			"scope": dir,
		})

		result, err := handleReadExtension(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["content"] != "# Coding Rules" {
			t.Errorf("content = %q, want %q", m["content"], "# Coding Rules")
		}
	})

	t.Run("read skill directory", func(t *testing.T) {
		dir := t.TempDir()
		skillDir := filepath.Join(dir, ".claude", "skills", "tdd")
		os.MkdirAll(skillDir, 0755)
		os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# TDD Skill"), 0644)

		req := newToolRequest(map[string]any{
			"type":  "skills",
			"name":  "tdd",
			"scope": dir,
		})

		result, err := handleReadExtension(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["content"] != "# TDD Skill" {
			t.Errorf("content = %q, want %q", m["content"], "# TDD Skill")
		}
	})

	t.Run("not found", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"type":  "rules",
			"name":  "nonexistent",
			"scope": dir,
		})

		result, err := handleReadExtension(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for nonexistent extension")
		}
	})

	t.Run("unsafe name rejected", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"type":  "rules",
			"name":  "../secret",
			"scope": "global",
		})

		result, err := handleReadExtension(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for unsafe name")
		}
	})
}

func TestHandleSaveExtension(t *testing.T) {
	ctx := context.Background()

	t.Run("save rule file", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"type":    "rules",
			"name":    "coding",
			"content": "# New Rule",
			"scope":   dir,
		})

		result, err := handleSaveExtension(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["success"] != true {
			t.Error("expected success=true")
		}

		data, _ := os.ReadFile(filepath.Join(dir, ".claude", "rules", "coding.md"))
		if string(data) != "# New Rule" {
			t.Errorf("file content = %q, want %q", string(data), "# New Rule")
		}
	})

	t.Run("save skill creates directory", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"type":    "skills",
			"name":    "debug",
			"content": "# Debug Skill",
			"scope":   dir,
		})

		result, err := handleSaveExtension(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["success"] != true {
			t.Error("expected success=true")
		}

		data, _ := os.ReadFile(filepath.Join(dir, ".claude", "skills", "debug", "SKILL.md"))
		if string(data) != "# Debug Skill" {
			t.Errorf("file content = %q, want %q", string(data), "# Debug Skill")
		}
	})

	t.Run("unsafe name rejected", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"type":    "rules",
			"name":    "../hack",
			"content": "malicious",
			"scope":   "global",
		})

		result, err := handleSaveExtension(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for unsafe name")
		}
	})
}

func TestHandleDeleteExtension(t *testing.T) {
	ctx := context.Background()

	t.Run("delete flat file", func(t *testing.T) {
		dir := t.TempDir()
		rulesDir := filepath.Join(dir, ".claude", "rules")
		os.MkdirAll(rulesDir, 0755)
		os.WriteFile(filepath.Join(rulesDir, "old.md"), []byte("old"), 0644)

		req := newToolRequest(map[string]any{
			"type":  "rules",
			"name":  "old",
			"scope": dir,
		})

		result, err := handleDeleteExtension(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["success"] != true {
			t.Error("expected success=true")
		}

		if fileExists(filepath.Join(rulesDir, "old.md")) {
			t.Error("file should have been deleted")
		}
	})

	t.Run("delete skill directory", func(t *testing.T) {
		dir := t.TempDir()
		skillDir := filepath.Join(dir, ".claude", "skills", "old-skill")
		os.MkdirAll(skillDir, 0755)
		os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("old"), 0644)

		req := newToolRequest(map[string]any{
			"type":  "skills",
			"name":  "old-skill",
			"scope": dir,
		})

		result, err := handleDeleteExtension(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["success"] != true {
			t.Error("expected success=true")
		}

		if fileExists(skillDir) {
			t.Error("skill directory should have been deleted")
		}
	})

	t.Run("nonexistent returns error", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"type":  "rules",
			"name":  "nonexistent",
			"scope": dir,
		})

		result, err := handleDeleteExtension(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for nonexistent extension")
		}
	})
}
