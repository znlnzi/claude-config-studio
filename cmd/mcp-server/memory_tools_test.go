package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// newToolRequest creates a CallToolRequest with the given arguments.
func newToolRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

// parseResultJSON extracts the JSON response from a successful tool result.
func parseResultJSON(t *testing.T, result *mcp.CallToolResult) map[string]any {
	t.Helper()
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Content) == 0 {
		t.Fatal("result has no content")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(tc.Text), &m); err != nil {
		t.Fatalf("failed to parse result JSON: %v\nraw: %s", err, tc.Text)
	}
	return m
}

// isErrorResult checks if the result is a tool error.
func isErrorResult(result *mcp.CallToolResult) bool {
	return result != nil && result.IsError
}

func TestHandleSaveMemory(t *testing.T) {
	ctx := context.Background()

	t.Run("save and overwrite", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"project_path": dir,
			"filename":     "test.md",
			"content":      "hello world",
		})

		result, err := handleSaveMemory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["success"] != true {
			t.Error("expected success=true")
		}

		data, _ := os.ReadFile(filepath.Join(dir, ".claude", "memory", "test.md"))
		if string(data) != "hello world" {
			t.Errorf("file content = %q, want %q", string(data), "hello world")
		}
	})

	t.Run("append mode", func(t *testing.T) {
		dir := t.TempDir()
		memDir := filepath.Join(dir, ".claude", "memory")
		os.MkdirAll(memDir, 0755)
		os.WriteFile(filepath.Join(memDir, "log.md"), []byte("line1\n"), 0644)

		req := newToolRequest(map[string]any{
			"project_path": dir,
			"filename":     "log.md",
			"content":      "line2\n",
			"mode":         "append",
		})

		result, err := handleSaveMemory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["success"] != true {
			t.Error("expected success=true")
		}

		data, _ := os.ReadFile(filepath.Join(memDir, "log.md"))
		if string(data) != "line1\nline2\n" {
			t.Errorf("file content = %q, want %q", string(data), "line1\nline2\n")
		}
	})

	t.Run("missing required params", func(t *testing.T) {
		req := newToolRequest(map[string]any{})
		result, err := handleSaveMemory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error result for missing params")
		}
	})

	t.Run("unsafe filename rejected", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"project_path": dir,
			"filename":     "../etc/passwd",
			"content":      "malicious",
		})

		result, err := handleSaveMemory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error result for unsafe filename")
		}
	})
}

func TestHandleLoadMemory(t *testing.T) {
	ctx := context.Background()

	t.Run("load specific file", func(t *testing.T) {
		dir := t.TempDir()
		memDir := filepath.Join(dir, ".claude", "memory")
		os.MkdirAll(memDir, 0755)
		os.WriteFile(filepath.Join(memDir, "notes.md"), []byte("my notes"), 0644)

		req := newToolRequest(map[string]any{
			"project_path": dir,
			"filename":     "notes.md",
		})

		result, err := handleLoadMemory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["content"] != "my notes" {
			t.Errorf("content = %q, want %q", m["content"], "my notes")
		}
		if m["filename"] != "notes.md" {
			t.Errorf("filename = %q, want %q", m["filename"], "notes.md")
		}
	})

	t.Run("list all files", func(t *testing.T) {
		dir := t.TempDir()
		memDir := filepath.Join(dir, ".claude", "memory")
		os.MkdirAll(memDir, 0755)
		os.WriteFile(filepath.Join(memDir, "a.md"), []byte("aaa"), 0644)
		os.WriteFile(filepath.Join(memDir, "b.md"), []byte("bbb"), 0644)
		os.WriteFile(filepath.Join(memDir, "skip.txt"), []byte("not md"), 0644)

		req := newToolRequest(map[string]any{
			"project_path": dir,
		})

		result, err := handleLoadMemory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		files, ok := m["files"].([]any)
		if !ok {
			t.Fatal("files should be an array")
		}
		if len(files) != 2 {
			t.Errorf("expected 2 md files, got %d", len(files))
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"project_path": dir,
		})

		result, err := handleLoadMemory(ctx, req)
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

	t.Run("nonexistent file returns error", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"project_path": dir,
			"filename":     "nonexistent.md",
		})

		result, err := handleLoadMemory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error result for nonexistent file")
		}
	})

	t.Run("unsafe filename rejected", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"project_path": dir,
			"filename":     "../secrets.md",
		})

		result, err := handleLoadMemory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error result for unsafe filename")
		}
	})
}

func TestHandleSearchMemory(t *testing.T) {
	ctx := context.Background()

	t.Run("finds matching lines", func(t *testing.T) {
		dir := t.TempDir()
		memDir := filepath.Join(dir, ".claude", "memory")
		os.MkdirAll(memDir, 0755)
		os.WriteFile(filepath.Join(memDir, "notes.md"), []byte("line1 hello\nline2 world\nline3 hello again"), 0644)

		req := newToolRequest(map[string]any{
			"project_path": dir,
			"query":        "hello",
		})

		result, err := handleSearchMemory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		totalMatches, _ := m["total_matches"].(float64)
		if totalMatches != 2 {
			t.Errorf("expected 2 matches, got %v", totalMatches)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		dir := t.TempDir()
		memDir := filepath.Join(dir, ".claude", "memory")
		os.MkdirAll(memDir, 0755)
		os.WriteFile(filepath.Join(memDir, "test.md"), []byte("Hello World\nhello world\nHELLO WORLD"), 0644)

		req := newToolRequest(map[string]any{
			"project_path": dir,
			"query":        "hello",
		})

		result, err := handleSearchMemory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		totalMatches, _ := m["total_matches"].(float64)
		if totalMatches != 3 {
			t.Errorf("expected 3 matches, got %v", totalMatches)
		}
	})

	t.Run("respects max_results", func(t *testing.T) {
		dir := t.TempDir()
		memDir := filepath.Join(dir, ".claude", "memory")
		os.MkdirAll(memDir, 0755)
		os.WriteFile(filepath.Join(memDir, "test.md"), []byte("match\nmatch\nmatch\nmatch\nmatch"), 0644)

		req := newToolRequest(map[string]any{
			"project_path": dir,
			"query":        "match",
			"max_results":  float64(2),
		})

		result, err := handleSearchMemory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		totalMatches, _ := m["total_matches"].(float64)
		if totalMatches != 2 {
			t.Errorf("expected 2 matches (limited), got %v", totalMatches)
		}
	})

	t.Run("missing query returns error", func(t *testing.T) {
		req := newToolRequest(map[string]any{})
		result, err := handleSearchMemory(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error result for missing query")
		}
	})
}
