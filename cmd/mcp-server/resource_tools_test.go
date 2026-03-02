package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func newResourceRequest(uri string) mcp.ReadResourceRequest {
	return mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: uri,
		},
	}
}

func TestHandleGlobalMemoryFile(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid filename rejected", func(t *testing.T) {
		req := newResourceRequest("claude://global/memory/../secret.md")
		_, err := handleGlobalMemoryFile(ctx, req)
		if err == nil {
			t.Error("expected error for path traversal")
		}
	})

	t.Run("non-md file rejected", func(t *testing.T) {
		req := newResourceRequest("claude://global/memory/file.txt")
		_, err := handleGlobalMemoryFile(ctx, req)
		if err == nil {
			t.Error("expected error for non-md file")
		}
	})

	t.Run("empty filename rejected", func(t *testing.T) {
		req := newResourceRequest("claude://global/memory/")
		_, err := handleGlobalMemoryFile(ctx, req)
		if err == nil {
			t.Error("expected error for empty filename")
		}
	})
}

func TestHandleProjectMemoryFile(t *testing.T) {
	ctx := context.Background()

	t.Run("reads project memory file", func(t *testing.T) {
		dir := t.TempDir()
		memDir := filepath.Join(dir, ".claude", "memory")
		os.MkdirAll(memDir, 0755)
		os.WriteFile(filepath.Join(memDir, "test.md"), []byte("project memory"), 0644)

		uri := "claude://project/" + dir + "/memory/test.md"
		req := newResourceRequest(uri)
		contents, err := handleProjectMemoryFile(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(contents) != 1 {
			t.Fatalf("expected 1 content, got %d", len(contents))
		}

		tc, ok := contents[0].(mcp.TextResourceContents)
		if !ok {
			t.Fatal("expected TextResourceContents")
		}
		if tc.Text != "project memory" {
			t.Errorf("text = %q, want %q", tc.Text, "project memory")
		}
	})

	t.Run("invalid URI format", func(t *testing.T) {
		req := newResourceRequest("claude://invalid/uri")
		_, err := handleProjectMemoryFile(ctx, req)
		if err == nil {
			t.Error("expected error for invalid URI")
		}
	})

	t.Run("missing memory segment", func(t *testing.T) {
		req := newResourceRequest("claude://project/some/path")
		_, err := handleProjectMemoryFile(ctx, req)
		if err == nil {
			t.Error("expected error for missing /memory/ segment")
		}
	})

	t.Run("path traversal in filename rejected", func(t *testing.T) {
		req := newResourceRequest("claude://project//tmp/proj/memory/../secret.md")
		_, err := handleProjectMemoryFile(ctx, req)
		if err == nil {
			t.Error("expected error for path traversal")
		}
	})

	t.Run("relative project path rejected", func(t *testing.T) {
		req := newResourceRequest("claude://project/relative/path/memory/test.md")
		_, err := handleProjectMemoryFile(ctx, req)
		if err == nil {
			t.Error("expected error for relative project path")
		}
	})
}
