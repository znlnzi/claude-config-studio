package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsSafeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid filename", "MEMORY.md", true},
		{"valid with dash", "session-state.md", true},
		{"empty string", "", false},
		{"path traversal", "../etc/passwd", false},
		{"contains slash", "path/file.md", false},
		{"contains backslash", "path\\file.md", false},
		{"double dot only", "..", false},
		{"double dot in name", "foo..bar", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSafeFilename(tt.input)
			if got != tt.expected {
				t.Errorf("isSafeFilename(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()

	existingFile := filepath.Join(dir, "exists.txt")
	os.WriteFile(existingFile, []byte("hello"), 0644)

	if !fileExists(existingFile) {
		t.Error("fileExists should return true for existing file")
	}
	if fileExists(filepath.Join(dir, "nonexistent.txt")) {
		t.Error("fileExists should return false for non-existing file")
	}
}

func TestDirHasFiles(t *testing.T) {
	emptyDir := t.TempDir()
	if dirHasFiles(emptyDir) {
		t.Error("dirHasFiles should return false for empty directory")
	}

	populatedDir := t.TempDir()
	os.WriteFile(filepath.Join(populatedDir, "file.md"), []byte("content"), 0644)
	if !dirHasFiles(populatedDir) {
		t.Error("dirHasFiles should return true for directory with files")
	}

	if dirHasFiles(filepath.Join(emptyDir, "nonexistent")) {
		t.Error("dirHasFiles should return false for non-existing directory")
	}
}

func TestResolveMemoryDir(t *testing.T) {
	t.Run("global scope", func(t *testing.T) {
		result, err := resolveMemoryDir("global")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".claude", "memory")
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("empty string defaults to global", func(t *testing.T) {
		result, err := resolveMemoryDir("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".claude", "memory")
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("project path", func(t *testing.T) {
		result, err := resolveMemoryDir("/tmp/myproject")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "/tmp/myproject/.claude/memory"
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})
}

func TestEncodePathSegment(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with space", "with-space"},
		{"with_underscore", "with-underscore"},
		{"with/slash", "with-slash"},
		{"CamelCase", "CamelCase"},
		{"file.txt", "file.txt"},
		{"a-b-c", "a-b-c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := encodePathSegment(tt.input)
			if got != tt.expected {
				t.Errorf("encodePathSegment(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDecodeProjectPath(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		if got := decodeProjectPath(""); got != "" {
			t.Errorf("decodeProjectPath(\"\") = %q, want \"\"", got)
		}
	})

	t.Run("no leading dash", func(t *testing.T) {
		if got := decodeProjectPath("noleadingdash"); got != "" {
			t.Errorf("decodeProjectPath(\"noleadingdash\") = %q, want \"\"", got)
		}
	})
}

func TestResolveExtensionDir(t *testing.T) {
	tests := []struct {
		name    string
		extType string
		scope   string
		wantErr bool
	}{
		{"agents global", "agents", "global", false},
		{"rules global", "rules", "", false},
		{"skills project", "skills", "/tmp/proj", false},
		{"commands global", "commands", "global", false},
		{"invalid type", "invalid", "global", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveExtensionDir(tt.extType, tt.scope)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == "" {
				t.Error("expected non-empty path")
			}
		})
	}
}

func TestIsValidJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid object", `{"key":"value"}`, true},
		{"valid array", `[1,2,3]`, true},
		{"valid string", `"hello"`, true},
		{"empty string", "", false},
		{"invalid json", `{invalid}`, false},
		{"whitespace only", "   ", false},
		{"valid with whitespace", `  {"key": "value"}  `, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidJSON(tt.input)
			if got != tt.expected {
				t.Errorf("isValidJSON(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	t.Run("formats compact json", func(t *testing.T) {
		input := `{"key":"value"}`
		got := formatJSON(input)
		expected := "{\n  \"key\": \"value\"\n}\n"
		if got != expected {
			t.Errorf("formatJSON(%q) = %q, want %q", input, got, expected)
		}
	})

	t.Run("invalid json returns input", func(t *testing.T) {
		input := `{invalid}`
		got := formatJSON(input)
		if got != input {
			t.Errorf("formatJSON(%q) = %q, want %q", input, got, input)
		}
	})
}

func TestResolveSettingsPath(t *testing.T) {
	t.Run("global scope", func(t *testing.T) {
		path, err := resolveSettingsPath("global")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".claude", "settings.json")
		if path != expected {
			t.Errorf("got %q, want %q", path, expected)
		}
	})

	t.Run("empty defaults to global", func(t *testing.T) {
		path, err := resolveSettingsPath("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".claude", "settings.json")
		if path != expected {
			t.Errorf("got %q, want %q", path, expected)
		}
	})

	t.Run("nonexistent project path", func(t *testing.T) {
		_, err := resolveSettingsPath("/nonexistent/path/abc123")
		if err == nil {
			t.Error("expected error for nonexistent path")
		}
	})

	t.Run("valid project path", func(t *testing.T) {
		dir := t.TempDir()
		path, err := resolveSettingsPath(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Join(dir, ".claude", "settings.json")
		if path != expected {
			t.Errorf("got %q, want %q", path, expected)
		}
	})
}

func TestIsResourceSafeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid md file", "MEMORY.md", true},
		{"empty", "", false},
		{"path traversal", "../file.md", false},
		{"contains slash", "dir/file.md", false},
		{"not md extension", "file.txt", false},
		{"no extension", "file", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isResourceSafeFilename(tt.input)
			if got != tt.expected {
				t.Errorf("isResourceSafeFilename(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsResourceSafePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid absolute path", "/Users/test/project", true},
		{"empty", "", false},
		{"path traversal", "/Users/../etc", false},
		{"relative path", "relative/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isResourceSafePath(tt.input)
			if got != tt.expected {
				t.Errorf("isResourceSafePath(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractURIParam(t *testing.T) {
	tests := []struct {
		uri      string
		prefix   string
		expected string
	}{
		{"claude://global/memory/test.md", "claude://global/memory/", "test.md"},
		{"claude://global/memory/", "claude://global/memory/", ""},
		{"other://uri", "claude://global/memory/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			got := extractURIParam(tt.uri, tt.prefix)
			if got != tt.expected {
				t.Errorf("extractURIParam(%q, %q) = %q, want %q", tt.uri, tt.prefix, got, tt.expected)
			}
		})
	}
}
