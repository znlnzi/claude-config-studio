package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleDetectProject(t *testing.T) {
	ctx := context.Background()

	t.Run("missing project_path", func(t *testing.T) {
		req := newToolRequest(map[string]any{})
		result, err := handleDetectProject(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error result for missing project_path")
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		req := newToolRequest(map[string]any{"project_path": "/nonexistent/abc123"})
		result, err := handleDetectProject(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isErrorResult(result) {
			t.Error("expected error result for nonexistent path")
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()
		req := newToolRequest(map[string]any{"project_path": dir})
		result, err := handleDetectProject(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		if m["project_path"] != dir {
			t.Errorf("expected project_path=%q, got %q", dir, m["project_path"])
		}

		langs, _ := m["languages"].([]any)
		if len(langs) != 0 {
			t.Errorf("expected 0 languages, got %d", len(langs))
		}

		frameworks, _ := m["frameworks"].([]any)
		if len(frameworks) != 0 {
			t.Errorf("expected 0 frameworks, got %d", len(frameworks))
		}

		if m["package_manager"] != nil {
			t.Error("expected nil package_manager for empty dir")
		}
	})

	t.Run("nodejs typescript project", func(t *testing.T) {
		dir := t.TempDir()

		// package.json with React + Jest
		pkg := map[string]any{
			"name": "test-app",
			"dependencies": map[string]string{
				"react":     "^18.2.0",
				"react-dom": "^18.2.0",
				"next":      "14.2.0",
			},
			"devDependencies": map[string]string{
				"jest":       "^29.0.0",
				"typescript": "^5.0.0",
			},
		}
		pkgData, _ := json.Marshal(pkg)
		os.WriteFile(filepath.Join(dir, "package.json"), pkgData, 0644)
		os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte("{}"), 0644)
		os.WriteFile(filepath.Join(dir, "pnpm-lock.yaml"), []byte("lockfileVersion: 9"), 0644)

		req := newToolRequest(map[string]any{"project_path": dir})
		result, err := handleDetectProject(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)

		// Languages: TypeScript + JavaScript
		langs, _ := m["languages"].([]any)
		langNames := extractNames(langs)
		if !sliceContains(langNames, "TypeScript") {
			t.Error("expected TypeScript in languages")
		}
		if !sliceContains(langNames, "JavaScript") {
			t.Error("expected JavaScript in languages")
		}

		// Frameworks: Next.js + React
		frameworks, _ := m["frameworks"].([]any)
		fwNames := extractNames(frameworks)
		if !sliceContains(fwNames, "Next.js") {
			t.Error("expected Next.js in frameworks")
		}
		if !sliceContains(fwNames, "React") {
			t.Error("expected React in frameworks")
		}

		// Check Next.js version
		for _, fw := range frameworks {
			fwMap, _ := fw.(map[string]any)
			if fwMap["name"] == "Next.js" {
				if fwMap["version"] != "14.2.0" {
					t.Errorf("expected Next.js version 14.2.0, got %v", fwMap["version"])
				}
			}
		}

		// Test framework: Jest
		testFws, _ := m["test_frameworks"].([]any)
		tfNames := extractNames(testFws)
		if !sliceContains(tfNames, "Jest") {
			t.Error("expected Jest in test_frameworks")
		}

		// Package manager: pnpm
		pm, _ := m["package_manager"].(map[string]any)
		if pm == nil || pm["name"] != "pnpm" {
			t.Errorf("expected pnpm package manager, got %v", pm)
		}
	})

	t.Run("go project", func(t *testing.T) {
		dir := t.TempDir()

		goMod := `module example.com/myapp

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/mark3labs/mcp-go v0.1.0
)
`
		os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)
		os.WriteFile(filepath.Join(dir, "go.sum"), []byte("hash"), 0644)

		req := newToolRequest(map[string]any{"project_path": dir})
		result, err := handleDetectProject(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)

		// Language: Go
		langs, _ := m["languages"].([]any)
		langNames := extractNames(langs)
		if !sliceContains(langNames, "Go") {
			t.Error("expected Go in languages")
		}

		// Framework: Gin
		frameworks, _ := m["frameworks"].([]any)
		fwNames := extractNames(frameworks)
		if !sliceContains(fwNames, "Gin") {
			t.Error("expected Gin in frameworks")
		}

		// Test framework: go test
		testFws, _ := m["test_frameworks"].([]any)
		tfNames := extractNames(testFws)
		if !sliceContains(tfNames, "go test") {
			t.Error("expected 'go test' in test_frameworks")
		}

		// Package manager: go modules
		pm, _ := m["package_manager"].(map[string]any)
		if pm == nil || pm["name"] != "go modules" {
			t.Errorf("expected 'go modules' package manager, got %v", pm)
		}
	})

	t.Run("python project", func(t *testing.T) {
		dir := t.TempDir()

		pyproject := `[project]
name = "myapp"
version = "0.1.0"

dependencies = [
    "fastapi>=0.100",
    "uvicorn>=0.20",
    "pytest>=7.0",
]
`
		os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(pyproject), 0644)
		os.WriteFile(filepath.Join(dir, "poetry.lock"), []byte("lock"), 0644)

		req := newToolRequest(map[string]any{"project_path": dir})
		result, err := handleDetectProject(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)

		// Language: Python
		langs, _ := m["languages"].([]any)
		langNames := extractNames(langs)
		if !sliceContains(langNames, "Python") {
			t.Error("expected Python in languages")
		}

		// Framework: FastAPI
		frameworks, _ := m["frameworks"].([]any)
		fwNames := extractNames(frameworks)
		if !sliceContains(fwNames, "FastAPI") {
			t.Error("expected FastAPI in frameworks")
		}

		// Test framework: pytest
		testFws, _ := m["test_frameworks"].([]any)
		tfNames := extractNames(testFws)
		if !sliceContains(tfNames, "pytest") {
			t.Error("expected pytest in test_frameworks")
		}

		// Package manager: poetry
		pm, _ := m["package_manager"].(map[string]any)
		if pm == nil || pm["name"] != "poetry" {
			t.Errorf("expected poetry package manager, got %v", pm)
		}
	})

	t.Run("project with claude config", func(t *testing.T) {
		dir := t.TempDir()
		claudeDir := filepath.Join(dir, ".claude")
		os.MkdirAll(claudeDir, 0755)

		meta := `{"setup_version":"0.8.0","setup_date":"2026-01-15"}`
		os.WriteFile(filepath.Join(claudeDir, ".setup-meta.json"), []byte(meta), 0644)

		req := newToolRequest(map[string]any{"project_path": dir})
		result, err := handleDetectProject(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		cc, _ := m["claude_config"].(map[string]any)
		if cc == nil {
			t.Fatal("expected claude_config to be non-nil")
		}
		if cc["has_claude_dir"] != true {
			t.Error("expected has_claude_dir=true")
		}
		if cc["has_setup_meta"] != true {
			t.Error("expected has_setup_meta=true")
		}
		// setup_meta should be parsed as object
		sm, _ := cc["setup_meta"].(map[string]any)
		if sm == nil {
			t.Error("expected setup_meta to be a parsed object")
		}
		if sm["setup_version"] != "0.8.0" {
			t.Errorf("expected setup_version=0.8.0, got %v", sm["setup_version"])
		}
	})

	t.Run("project without claude config", func(t *testing.T) {
		dir := t.TempDir()

		req := newToolRequest(map[string]any{"project_path": dir})
		result, err := handleDetectProject(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		cc, _ := m["claude_config"].(map[string]any)
		if cc == nil {
			t.Fatal("expected claude_config to be non-nil")
		}
		if cc["has_claude_dir"] != false {
			t.Error("expected has_claude_dir=false")
		}
		if cc["has_setup_meta"] != false {
			t.Error("expected has_setup_meta=false")
		}
	})

	t.Run("rust project", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"test\""), 0644)
		os.WriteFile(filepath.Join(dir, "Cargo.lock"), []byte("lock"), 0644)

		req := newToolRequest(map[string]any{"project_path": dir})
		result, err := handleDetectProject(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)

		langs, _ := m["languages"].([]any)
		langNames := extractNames(langs)
		if !sliceContains(langNames, "Rust") {
			t.Error("expected Rust in languages")
		}

		testFws, _ := m["test_frameworks"].([]any)
		tfNames := extractNames(testFws)
		if !sliceContains(tfNames, "cargo test") {
			t.Error("expected 'cargo test' in test_frameworks")
		}

		pm, _ := m["package_manager"].(map[string]any)
		if pm == nil || pm["name"] != "cargo" {
			t.Errorf("expected cargo package manager, got %v", pm)
		}
	})

	t.Run("language dedup", func(t *testing.T) {
		dir := t.TempDir()
		// Both pyproject.toml and requirements.txt exist
		os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]\nname = \"x\""), 0644)
		os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("flask"), 0644)

		req := newToolRequest(map[string]any{"project_path": dir})
		result, err := handleDetectProject(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m := parseResultJSON(t, result)
		langs, _ := m["languages"].([]any)
		count := 0
		for _, l := range langs {
			lm, _ := l.(map[string]any)
			if lm["name"] == "Python" {
				count++
			}
		}
		if count != 1 {
			t.Errorf("expected Python to appear exactly once, got %d", count)
		}
	})
}

func TestParsePackageJSONDeps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")

	pkg := `{
		"dependencies": {"react": "^18.0.0", "next": "14.0.0"},
		"devDependencies": {"jest": "^29.0.0", "vitest": "^1.0.0"}
	}`
	os.WriteFile(path, []byte(pkg), 0644)

	deps, devDeps := parsePackageJSONDeps(path)

	if deps["react"] != "^18.0.0" {
		t.Errorf("expected react=^18.0.0, got %q", deps["react"])
	}
	if deps["next"] != "14.0.0" {
		t.Errorf("expected next=14.0.0, got %q", deps["next"])
	}
	if devDeps["jest"] != "^29.0.0" {
		t.Errorf("expected jest=^29.0.0, got %q", devDeps["jest"])
	}
	if devDeps["vitest"] != "^1.0.0" {
		t.Errorf("expected vitest=^1.0.0, got %q", devDeps["vitest"])
	}
}

func TestParseGoModDeps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")

	content := `module example.com/test

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/labstack/echo/v4 v4.11.0
)

require github.com/stretchr/testify v1.8.0
`
	os.WriteFile(path, []byte(content), 0644)

	deps := parseGoModDeps(path)

	if !sliceContains(deps, "github.com/gin-gonic/gin") {
		t.Error("expected gin in go.mod deps")
	}
	if !sliceContains(deps, "github.com/labstack/echo/v4") {
		t.Error("expected echo in go.mod deps")
	}
	if !sliceContains(deps, "github.com/stretchr/testify") {
		t.Error("expected testify in go.mod deps (single-line require)")
	}
}

func TestParsePyprojectDeps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pyproject.toml")

	content := `[project]
name = "myapp"
version = "0.1.0"

dependencies = [
    "fastapi>=0.100",
    "uvicorn>=0.20",
    "pydantic~=2.0",
]

[tool.pytest.ini_options]
testpaths = ["tests"]
`
	os.WriteFile(path, []byte(content), 0644)

	deps := parsePyprojectDeps(path)

	if !sliceContains(deps, "fastapi") {
		t.Error("expected fastapi in pyproject deps")
	}
	if !sliceContains(deps, "uvicorn") {
		t.Error("expected uvicorn in pyproject deps")
	}
	if !sliceContains(deps, "pydantic") {
		t.Error("expected pydantic in pyproject deps")
	}
}

func TestDetectPackageManager(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		expected string
	}{
		{"pnpm", "pnpm-lock.yaml", "pnpm"},
		{"yarn", "yarn.lock", "yarn"},
		{"npm", "package-lock.json", "npm"},
		{"bun lockb", "bun.lockb", "bun"},
		{"bun lock", "bun.lock", "bun"},
		{"poetry", "poetry.lock", "poetry"},
		{"uv", "uv.lock", "uv"},
		{"go modules", "go.sum", "go modules"},
		{"cargo", "Cargo.lock", "cargo"},
		{"bundler", "Gemfile.lock", "bundler"},
		{"pipenv", "Pipfile.lock", "pipenv"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, tt.file), []byte("lock"), 0644)

			pm := detectPackageManager(dir)
			if pm == nil {
				t.Fatalf("expected package manager %q, got nil", tt.expected)
			}
			if pm.Name != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, pm.Name)
			}
		})
	}

	t.Run("no lock file", func(t *testing.T) {
		dir := t.TempDir()
		pm := detectPackageManager(dir)
		if pm != nil {
			t.Errorf("expected nil package manager, got %v", pm)
		}
	})
}

func TestDetectClaudeConfig(t *testing.T) {
	t.Run("no claude dir", func(t *testing.T) {
		dir := t.TempDir()
		cc := detectClaudeConfig(dir)
		if cc.HasClaudeDir {
			t.Error("expected HasClaudeDir=false")
		}
		if cc.HasSetupMeta {
			t.Error("expected HasSetupMeta=false")
		}
	})

	t.Run("claude dir without meta", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".claude"), 0755)

		cc := detectClaudeConfig(dir)
		if !cc.HasClaudeDir {
			t.Error("expected HasClaudeDir=true")
		}
		if cc.HasSetupMeta {
			t.Error("expected HasSetupMeta=false")
		}
	})

	t.Run("claude dir with meta", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, ".claude"), 0755)
		meta := `{"setup_version":"0.8.0"}`
		os.WriteFile(filepath.Join(dir, ".claude", ".setup-meta.json"), []byte(meta), 0644)

		cc := detectClaudeConfig(dir)
		if !cc.HasClaudeDir {
			t.Error("expected HasClaudeDir=true")
		}
		if !cc.HasSetupMeta {
			t.Error("expected HasSetupMeta=true")
		}
		if cc.SetupMeta == nil {
			t.Error("expected SetupMeta to be non-nil")
		}
	})
}

// ─── Test Helpers ───────────────────────────────────────────

func extractNames(items []any) []string {
	var names []string
	for _, item := range items {
		m, _ := item.(map[string]any)
		if name, ok := m["name"].(string); ok {
			names = append(names, name)
		}
	}
	return names
}

func sliceContains(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
