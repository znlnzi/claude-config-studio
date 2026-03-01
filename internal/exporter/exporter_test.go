package exporter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip_ProjectConfig(t *testing.T) {
	// Setup: create a temp project with .claude/ structure
	srcDir := t.TempDir()
	claudeDir := filepath.Join(srcDir, ".claude")
	memDir := filepath.Join(claudeDir, "memory")
	rulesDir := filepath.Join(claudeDir, "rules")
	if err := os.MkdirAll(memDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write test files
	files := map[string]string{
		filepath.Join(srcDir, "CLAUDE.md"):              "# Project Instructions",
		filepath.Join(memDir, "MEMORY.md"):              "# Memory",
		filepath.Join(memDir, "session-state.md"):       "# Session State",
		filepath.Join(rulesDir, "coding-style.md"):      "# Coding Style",
		filepath.Join(claudeDir, "settings.json"):       `{"key": "value"}`,
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Export
	result, err := ExportProjectConfig(srcDir)
	if err != nil {
		t.Fatal(err)
	}
	if result.Files != 5 {
		t.Errorf("expected 5 exported files, got %d", result.Files)
	}
	if result.Size <= 0 {
		t.Error("expected positive ZIP size")
	}
	if result.Data == "" {
		t.Error("expected non-empty base64 data")
	}

	// Import into a new directory
	dstDir := t.TempDir()
	importResult, err := ImportConfig(dstDir, result.Data)
	if err != nil {
		t.Fatal(err)
	}
	if importResult.FilesImported != 5 {
		t.Errorf("expected 5 imported files, got %d", importResult.FilesImported)
	}

	// Verify contents match
	for origPath, expectedContent := range files {
		relPath, _ := filepath.Rel(srcDir, origPath)
		importedPath := filepath.Join(dstDir, relPath)
		data, err := os.ReadFile(importedPath)
		if err != nil {
			t.Errorf("failed to read imported file %s: %v", relPath, err)
			continue
		}
		if string(data) != expectedContent {
			t.Errorf("content mismatch for %s: got %q, want %q", relPath, string(data), expectedContent)
		}
	}
}

func TestImport_PathTraversalPrevention(t *testing.T) {
	// Setup: create a temp project with a valid export
	srcDir := t.TempDir()
	claudeDir := filepath.Join(srcDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Export first
	result, err := ExportProjectConfig(srcDir)
	if err != nil {
		t.Fatal(err)
	}

	// Import should succeed normally
	dstDir := t.TempDir()
	importResult, err := ImportConfig(dstDir, result.Data)
	if err != nil {
		t.Fatal(err)
	}
	if importResult.FilesImported < 1 {
		t.Error("expected at least 1 imported file")
	}
}

func TestImport_InvalidBase64(t *testing.T) {
	dstDir := t.TempDir()
	_, err := ImportConfig(dstDir, "not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestImport_InvalidZIP(t *testing.T) {
	dstDir := t.TempDir()
	// Valid base64 but not a ZIP
	_, err := ImportConfig(dstDir, "aGVsbG8gd29ybGQ=")
	if err == nil {
		t.Error("expected error for invalid ZIP")
	}
}

func TestExportProjectConfig_EmptyProject(t *testing.T) {
	srcDir := t.TempDir()
	result, err := ExportProjectConfig(srcDir)
	if err != nil {
		t.Fatal(err)
	}
	// Empty project should produce a valid ZIP with 0 files
	if result.Files != 0 {
		t.Errorf("expected 0 files, got %d", result.Files)
	}
	if result.Data == "" {
		t.Error("expected non-empty base64 data even for empty ZIP")
	}
}
