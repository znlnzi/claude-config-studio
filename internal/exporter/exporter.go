package exporter

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExportResult contains the result of an export operation
type ExportResult struct {
	Data  string `json:"data"`   // base64-encoded ZIP
	Size  int    `json:"size"`   // raw ZIP size in bytes
	Files int    `json:"files"`  // number of files in the archive
}

// ImportResult contains the result of an import operation
type ImportResult struct {
	FilesImported int      `json:"files_imported"`
	FileNames     []string `json:"file_names"`
}

// globalExportFiles are the individual files to export from ~/.claude/
var globalExportFiles = []string{
	"CLAUDE.md",
	"settings.json",
	"cclsp.json",
	".mcp.json",
}

// globalExportDirs are the directories to recursively export from ~/.claude/
var globalExportDirs = []string{
	"commands",
	"agents",
	"rules",
	"skills",
	"memory",
}

// ExportGlobalConfig exports ~/.claude/ configuration as a base64-encoded ZIP
func ExportGlobalConfig() (*ExportResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	claudeHome := filepath.Join(home, ".claude")
	return exportDir(claudeHome, globalExportFiles, globalExportDirs)
}

// ExportProjectConfig exports a project's .claude/ configuration as a base64-encoded ZIP
func ExportProjectConfig(projectPath string) (*ExportResult, error) {
	claudeDir := filepath.Join(projectPath, ".claude")

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	fileCount := 0

	// Export root CLAUDE.md
	if data, err := os.ReadFile(filepath.Join(projectPath, "CLAUDE.md")); err == nil {
		f, err := w.Create("CLAUDE.md")
		if err == nil {
			_, _ = f.Write(data)
			fileCount++
		}
	}

	// Recursively export .claude/ directory
	_ = filepath.Walk(claudeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(projectPath, path)
		if err != nil {
			return nil
		}
		// Security: skip if path contains ".."
		if strings.Contains(relPath, "..") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		f, createErr := w.Create(relPath)
		if createErr != nil {
			return nil
		}
		_, _ = f.Write(data)
		fileCount++
		return nil
	})

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to create ZIP: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return &ExportResult{
		Data:  encoded,
		Size:  buf.Len(),
		Files: fileCount,
	}, nil
}

// ImportConfig imports a base64-encoded ZIP into the target directory
func ImportConfig(targetPath string, base64Data string) (*ImportResult, error) {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to read ZIP archive: %w", err)
	}

	var imported []string
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		name := filepath.Clean(f.Name)
		// Security: prevent path traversal
		if strings.Contains(name, "..") {
			continue
		}
		// Security: reject absolute paths within ZIP
		if filepath.IsAbs(name) {
			continue
		}

		destPath := filepath.Join(targetPath, name)
		// Double-check the resolved path is within targetPath
		if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(targetPath)) {
			continue
		}

		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			continue
		}

		src, err := f.Open()
		if err != nil {
			continue
		}

		dst, err := os.Create(destPath)
		if err != nil {
			src.Close()
			continue
		}

		_, copyErr := io.Copy(dst, src)
		dst.Close()
		src.Close()

		if copyErr == nil {
			imported = append(imported, name)
		}
	}

	return &ImportResult{
		FilesImported: len(imported),
		FileNames:     imported,
	}, nil
}

// exportDir creates a ZIP from specific files and directories under baseDir
func exportDir(baseDir string, files []string, dirs []string) (*ExportResult, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	fileCount := 0

	for _, name := range files {
		filePath := filepath.Join(baseDir, name)
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		f, err := w.Create(name)
		if err != nil {
			continue
		}
		_, _ = f.Write(data)
		fileCount++
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(baseDir, dir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			filePath := filepath.Join(dirPath, entry.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			f, err := w.Create(filepath.Join(dir, entry.Name()))
			if err != nil {
				continue
			}
			_, _ = f.Write(data)
			fileCount++
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to create ZIP: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return &ExportResult{
		Data:  encoded,
		Size:  buf.Len(),
		Files: fileCount,
	}, nil
}
