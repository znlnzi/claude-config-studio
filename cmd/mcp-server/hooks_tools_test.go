package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleHooksList(t *testing.T) {
	ctx := context.Background()

	t.Run("reads hooks from settings", func(t *testing.T) {
		dir := t.TempDir()
		claudeDir := filepath.Join(dir, ".claude")
		os.MkdirAll(claudeDir, 0755)
		os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(`{
			"hooks": {
				"PreToolUse": [{"matcher": "Write", "hooks": [{"type": "command", "command": "echo check"}]}],
				"Stop": [{"matcher": "", "hooks": [{"type": "command", "command": "echo done"}]}]
			}
		}`), 0644)

		req := newToolRequest(map[string]any{"scope": dir})
		result, err := handleHooksList(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		totalEvents, _ := m["total_events"].(float64)
		if totalEvents != 2 {
			t.Errorf("expected 2 events, got %v", totalEvents)
		}
	})

	t.Run("no settings file returns empty", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{"scope": dir})
		result, err := handleHooksList(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		totalEvents, _ := m["total_events"].(float64)
		if totalEvents != 0 {
			t.Errorf("expected 0 events, got %v", totalEvents)
		}
	})

	t.Run("settings without hooks returns empty", func(t *testing.T) {
		dir := t.TempDir()
		claudeDir := filepath.Join(dir, ".claude")
		os.MkdirAll(claudeDir, 0755)
		os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(`{"other": "value"}`), 0644)

		req := newToolRequest(map[string]any{"scope": dir})
		result, err := handleHooksList(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		totalEvents, _ := m["total_events"].(float64)
		if totalEvents != 0 {
			t.Errorf("expected 0 events, got %v", totalEvents)
		}
	})
}

func TestHandleHooksSave(t *testing.T) {
	ctx := context.Background()

	t.Run("save valid hooks", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"hooks": `{"PreToolUse": [{"matcher": "Write", "hooks": [{"type": "command", "command": "echo check"}]}]}`,
			"scope": dir,
		})

		result, err := handleHooksSave(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["success"] != true {
			t.Error("expected success=true")
		}
		eventsSaved, _ := m["events_saved"].(float64)
		if eventsSaved != 1 {
			t.Errorf("expected 1 event saved, got %v", eventsSaved)
		}

		// Verify file was created
		data, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
		if err != nil {
			t.Fatalf("failed to read settings: %v", err)
		}
		if len(data) == 0 {
			t.Error("settings file should not be empty")
		}
	})

	t.Run("preserves existing settings", func(t *testing.T) {
		dir := t.TempDir()
		claudeDir := filepath.Join(dir, ".claude")
		os.MkdirAll(claudeDir, 0755)
		os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(`{"allowedTools": ["Read"]}`), 0644)

		req := newToolRequest(map[string]any{
			"hooks": `{"Stop": []}`,
			"scope": dir,
		})

		result, err := handleHooksSave(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["success"] != true {
			t.Error("expected success=true")
		}

		data, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
		content := string(data)
		if !contains(content, "allowedTools") {
			t.Error("existing settings should be preserved")
		}
		if !contains(content, "hooks") {
			t.Error("hooks should be added")
		}
	})

	t.Run("invalid json rejected", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"hooks": `{invalid}`,
			"scope": dir,
		})

		result, err := handleHooksSave(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("unknown event rejected", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"hooks": `{"UnknownEvent": []}`,
			"scope": dir,
		})

		result, err := handleHooksSave(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for unknown event")
		}
	})

	t.Run("missing hooks param", func(t *testing.T) {
		req := newToolRequest(map[string]any{})
		result, err := handleHooksSave(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for missing hooks")
		}
	})
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsString(s, substr))
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
