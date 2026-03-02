package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/znlnzi/claude-config-studio/internal/marketplace"
	"github.com/znlnzi/claude-config-studio/internal/templatedata"
)

// setupMarketplaceServer creates a httptest server that serves an index and template.
// Returns the server (caller must defer Close) and the template ID.
func setupMarketplaceServer(t *testing.T) *httptest.Server {
	t.Helper()

	tmpl := templatedata.Template{
		ID:          "community/test-template",
		Name:        "Test Template",
		Category:    "Testing",
		Description: "A test template for marketplace",
		Tags:        []string{"test"},
		ClaudeMd:    "# Test Template\nThis is a test.",
		Agents:      map[string]string{"test-agent": "# Test Agent"},
	}
	tmplData, _ := json.Marshal(tmpl)

	mux := http.NewServeMux()

	// Serve template.json at /templates/community/test-template.json
	mux.HandleFunc("/templates/community/test-template.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(tmplData)
	})

	ts := httptest.NewServer(mux)

	// Build index with template URL pointing to the test server
	idx := marketplace.RegistryIndex{
		Version:   1,
		UpdatedAt: "2026-03-02T10:00:00Z",
		Templates: []marketplace.IndexEntry{
			{
				ID:          "community/test-template",
				Name:        "Test Template",
				Description: "A test template for marketplace",
				Author:      "tester",
				Version:     "1.0.0",
				Category:    "Testing",
				Tags:        []string{"test"},
				URL:         ts.URL + "/templates/community/test-template.json",
				Downloads:   42,
				Stars:       5,
			},
			{
				ID:          "community/frontend-kit",
				Name:        "Frontend Dev Kit",
				Description: "Frontend development best practices",
				Author:      "frontdev",
				Version:     "1.0.0",
				Category:    "Frontend",
				Tags:        []string{"react", "css"},
				URL:         ts.URL + "/templates/community/frontend-kit.json",
				Downloads:   100,
				Stars:       20,
			},
		},
	}
	idxData, _ := json.Marshal(idx)

	// Add index.json handler
	mux.HandleFunc("/index.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(idxData)
	})

	// Update the default mux handler for the root to serve index
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(idxData)
	})

	return ts
}

func TestHandleMarketplaceSearch(t *testing.T) {
	ctx := context.Background()
	ts := setupMarketplaceServer(t)
	defer ts.Close()
	t.Setenv("CLAUDE_MARKETPLACE_REGISTRY_URL", ts.URL)

	t.Run("list all templates", func(t *testing.T) {
		req := newToolRequest(map[string]any{})
		result, err := handleMarketplaceSearch(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		total, _ := m["total"].(float64)
		if total != 2 {
			t.Errorf("total = %v, want 2", total)
		}
	})

	t.Run("search by query", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"query": "test",
		})
		result, err := handleMarketplaceSearch(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		total, _ := m["total"].(float64)
		if total != 1 {
			t.Errorf("total = %v, want 1", total)
		}
	})

	t.Run("filter by category", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"category": "Frontend",
		})
		result, err := handleMarketplaceSearch(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		total, _ := m["total"].(float64)
		if total != 1 {
			t.Errorf("total = %v, want 1", total)
		}
	})

	t.Run("combined query and category", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"query":    "react",
			"category": "Frontend",
		})
		result, err := handleMarketplaceSearch(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		total, _ := m["total"].(float64)
		if total != 1 {
			t.Errorf("total = %v, want 1", total)
		}
	})

	t.Run("no results", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"query": "nonexistent-xyz",
		})
		result, err := handleMarketplaceSearch(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		total, _ := m["total"].(float64)
		if total != 0 {
			t.Errorf("total = %v, want 0", total)
		}
	})
}

func TestHandleMarketplaceSearch_NetworkError(t *testing.T) {
	ctx := context.Background()
	t.Setenv("CLAUDE_MARKETPLACE_REGISTRY_URL", "http://127.0.0.1:1") // unreachable

	req := newToolRequest(map[string]any{})
	result, err := handleMarketplaceSearch(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Error("expected error for network failure")
	}
}

func TestHandleMarketplaceInstall(t *testing.T) {
	ctx := context.Background()
	ts := setupMarketplaceServer(t)
	defer ts.Close()
	t.Setenv("CLAUDE_MARKETPLACE_REGISTRY_URL", ts.URL)

	t.Run("install community template", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"template_id":  "community/test-template",
			"scope":        "project",
			"project_path": dir,
		})

		result, err := handleMarketplaceInstall(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["success"] != true {
			t.Error("expected success=true")
		}
		if m["source"] != "marketplace" {
			t.Errorf("source = %v, want marketplace", m["source"])
		}

		// Verify rules file was created with sanitized name
		rulesFile := filepath.Join(dir, ".claude", "rules", "tpl-community-test-template.md")
		if !fileExists(rulesFile) {
			t.Error("template rules file should exist at tpl-community-test-template.md")
		}

		// Verify agent was installed
		agentFile := filepath.Join(dir, ".claude", "agents", "test-agent.md")
		if !fileExists(agentFile) {
			t.Error("agent file should exist")
		}
	})

	t.Run("template not found in registry", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{
			"template_id":  "community/nonexistent",
			"scope":        "project",
			"project_path": dir,
		})

		result, err := handleMarketplaceInstall(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for nonexistent template")
		}
	})

	t.Run("missing template_id", func(t *testing.T) {
		req := newToolRequest(map[string]any{})
		result, err := handleMarketplaceInstall(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for missing template_id")
		}
	})

	t.Run("project scope without path", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"template_id": "community/test-template",
			"scope":       "project",
		})
		result, err := handleMarketplaceInstall(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for missing project_path")
		}
	})
}

func TestHandleMarketplaceInfo(t *testing.T) {
	ctx := context.Background()
	ts := setupMarketplaceServer(t)
	defer ts.Close()
	t.Setenv("CLAUDE_MARKETPLACE_REGISTRY_URL", ts.URL)

	t.Run("get template info", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"template_id": "community/test-template",
		})

		result, err := handleMarketplaceInfo(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["id"] != "community/test-template" {
			t.Errorf("id = %v", m["id"])
		}
		if m["name"] != "Test Template" {
			t.Errorf("name = %v", m["name"])
		}
		if m["author"] != "tester" {
			t.Errorf("author = %v", m["author"])
		}

		contents, ok := m["contents"].(map[string]any)
		if !ok {
			t.Fatal("expected contents to be a map")
		}
		if contents["has_rules_file"] != true {
			t.Error("expected has_rules_file=true")
		}
		agents, ok := contents["agents"].([]any)
		if !ok || len(agents) == 0 {
			t.Error("expected agents list")
		}
	})

	t.Run("template not found", func(t *testing.T) {
		req := newToolRequest(map[string]any{
			"template_id": "community/nonexistent",
		})

		result, err := handleMarketplaceInfo(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for nonexistent template")
		}
	})

	t.Run("missing template_id", func(t *testing.T) {
		req := newToolRequest(map[string]any{})
		result, err := handleMarketplaceInfo(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error for missing template_id")
		}
	})
}

func TestSanitizeTemplateID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"community/react-testing", "community-react-testing"},
		{"simple", "simple"},
		{"a/b/c", "a-b-c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeTemplateID(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeTemplateID(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestResolveClaudeDir(t *testing.T) {
	t.Run("global scope", func(t *testing.T) {
		dir, errMsg := resolveClaudeDir("global", "")
		if errMsg != "" {
			t.Fatalf("unexpected error: %s", errMsg)
		}
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".claude")
		if dir != expected {
			t.Errorf("got %q, want %q", dir, expected)
		}
	})

	t.Run("project scope", func(t *testing.T) {
		tmpDir := t.TempDir()
		dir, errMsg := resolveClaudeDir("project", tmpDir)
		if errMsg != "" {
			t.Fatalf("unexpected error: %s", errMsg)
		}
		expected := filepath.Join(tmpDir, ".claude")
		if dir != expected {
			t.Errorf("got %q, want %q", dir, expected)
		}
	})

	t.Run("project scope without path", func(t *testing.T) {
		_, errMsg := resolveClaudeDir("project", "")
		if errMsg == "" {
			t.Error("expected error for missing project_path")
		}
	})

	t.Run("nonexistent project path", func(t *testing.T) {
		_, errMsg := resolveClaudeDir("project", "/nonexistent/path/abc123")
		if errMsg == "" {
			t.Error("expected error for nonexistent path")
		}
	})
}

func TestInstallTemplate(t *testing.T) {
	t.Run("install with claudeMd and agents", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), ".claude")
		tmpl := &templatedata.Template{
			ID:       "test-tmpl",
			Name:     "Test",
			ClaudeMd: "# Test",
			Agents:   map[string]string{"agent1": "# Agent 1"},
		}

		files, errMsg := installTemplate(dir, tmpl, "test-tmpl", false)
		if errMsg != "" {
			t.Fatalf("unexpected error: %s", errMsg)
		}
		if len(files) != 2 {
			t.Errorf("installed %d files, want 2", len(files))
		}

		rulesFile := filepath.Join(dir, "rules", "tpl-test-tmpl.md")
		if !fileExists(rulesFile) {
			t.Error("rules file should exist")
		}

		agentFile := filepath.Join(dir, "agents", "agent1.md")
		if !fileExists(agentFile) {
			t.Error("agent file should exist")
		}
	})

	t.Run("install empty template", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), ".claude")
		tmpl := &templatedata.Template{
			ID:   "empty",
			Name: "Empty",
		}

		files, errMsg := installTemplate(dir, tmpl, "empty", false)
		if errMsg != "" {
			t.Fatalf("unexpected error: %s", errMsg)
		}
		if len(files) != 0 {
			t.Errorf("installed %d files, want 0", len(files))
		}
	})
}
