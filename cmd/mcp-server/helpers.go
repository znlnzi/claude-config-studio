package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// resolveMemoryDir resolves the memory directory based on the project_path parameter
func resolveMemoryDir(projectPath string) (string, error) {
	if projectPath == "" || projectPath == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".claude", "memory"), nil
	}
	return filepath.Join(projectPath, ".claude", "memory"), nil
}

// resolveClaudeDir resolves the .claude directory based on the project_path parameter
func resolveClaudeDir(projectPath string) (string, error) {
	if projectPath == "" || projectPath == "global" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".claude"), nil
	}
	return filepath.Join(projectPath, ".claude"), nil
}

// isSafeFilename checks whether the filename is safe (no path traversal)
func isSafeFilename(name string) bool {
	if name == "" {
		return false
	}
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	return true
}

// fileExists checks whether a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// dirHasFiles checks whether a directory contains files
func dirHasFiles(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	return len(entries) > 0
}

// getProjectsDir returns the ~/.claude/projects/ directory path
func getProjectsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "projects"), nil
}

// scanAllProjects scans all project paths under ~/.claude/projects/
func scanAllProjects() ([]string, error) {
	projDir, err := getProjectsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(projDir)
	if err != nil {
		return nil, nil // return empty if directory doesn't exist
	}

	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		realPath := decodeProjectPath(entry.Name())
		if realPath == "" {
			continue
		}
		if _, err := os.Stat(realPath); os.IsNotExist(err) {
			continue
		}
		paths = append(paths, realPath)
	}
	return paths, nil
}

// --- The following functions are copied from services/project_service.go ---

// decodeProjectPath converts an encoded directory name under ~/.claude/projects/ to a real path.
// Claude Code encoding rule: replaces /, _, spaces, and non-ASCII characters with -
func decodeProjectPath(encoded string) string {
	if encoded == "" || !strings.HasPrefix(encoded, "-") {
		return ""
	}
	return resolveEncodedPath("/", encoded[1:])
}

// resolveEncodedPath recursively matches the filesystem to decode a path
func resolveEncodedPath(dir, remaining string) string {
	if remaining == "" {
		return dir
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	type candidate struct {
		name    string
		encLen  int
		isExact bool
	}

	var candidates []candidate
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		enc := encodePathSegment(e.Name())
		if remaining == enc {
			candidates = append(candidates, candidate{e.Name(), len(enc), true})
		} else if strings.HasPrefix(remaining, enc+"-") {
			candidates = append(candidates, candidate{e.Name(), len(enc) + 1, false})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].encLen > candidates[j].encLen
	})

	for _, c := range candidates {
		nextDir := filepath.Join(dir, c.name)
		if c.isExact {
			if _, err := os.Stat(nextDir); err == nil {
				return nextDir
			}
		} else {
			result := resolveEncodedPath(nextDir, remaining[c.encLen:])
			if result != "" {
				return result
			}
		}
	}

	return ""
}

// encodePathSegment encodes a directory name according to Claude Code's encoding rules
func encodePathSegment(name string) string {
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '.' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	return b.String()
}
