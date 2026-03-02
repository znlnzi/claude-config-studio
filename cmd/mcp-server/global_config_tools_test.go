package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleSaveProjectConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("save claude_md", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"project_path": dir,
			"field":        "claude_md",
			"content":      "# My Instructions",
		})

		result, err := handleSaveProjectConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["success"] != true {
			t.Error("expected success=true")
		}

		data, _ := os.ReadFile(filepath.Join(dir, ".claude", "CLAUDE.md"))
		if string(data) != "# My Instructions" {
			t.Errorf("content = %q, want %q", string(data), "# My Instructions")
		}
	})

	t.Run("save settings validates json", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"project_path": dir,
			"field":        "settings",
			"content":      `{"key": "value"}`,
		})

		result, err := handleSaveProjectConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["success"] != true {
			t.Error("expected success=true")
		}

		// Verify JSON is formatted
		data, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
		if !strings.Contains(string(data), "  ") {
			t.Error("settings should be formatted with indentation")
		}
	})

	t.Run("invalid json rejected for settings", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"project_path": dir,
			"field":        "settings",
			"content":      "{not valid json}",
		})

		result, err := handleSaveProjectConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("invalid json rejected for mcp", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"project_path": dir,
			"field":        "mcp",
			"content":      "not json",
		})

		result, err := handleSaveProjectConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("invalid field rejected", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"project_path": dir,
			"field":        "unknown_field",
			"content":      "content",
		})

		result, err := handleSaveProjectConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for invalid field")
		}
	})

	t.Run("nonexistent project path", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"project_path": "/nonexistent/path/abc123",
			"field":        "claude_md",
			"content":      "content",
		})

		result, err := handleSaveProjectConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for nonexistent path")
		}
	})
}

func TestHandleSaveGlobalConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("missing field", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"content": "data",
		})

		result, err := handleSaveGlobalConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for missing field")
		}
	})

	t.Run("missing content", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"field": "claude_md",
		})

		result, err := handleSaveGlobalConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for missing content")
		}
	})

	t.Run("invalid field name", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"field":   "invalid",
			"content": "data",
		})

		result, err := handleSaveGlobalConfig(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for invalid field")
		}
	})
}
