package services

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ExportService handles import/export of configuration
type ExportService struct {
	ctx context.Context
}

func NewExportService() *ExportService {
	return &ExportService{}
}

func (s *ExportService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// ExportGlobalConfig exports global configuration as a zip file
func (s *ExportService) ExportGlobalConfig() (string, error) {
	savePath, err := runtime.SaveFileDialog(s.ctx, runtime.SaveDialogOptions{
		Title:           "Export Global Configuration",
		DefaultFilename: "claude-global-config.zip",
		Filters: []runtime.FileFilter{
			{DisplayName: "ZIP Files", Pattern: "*.zip"},
		},
	})
	if err != nil || savePath == "" {
		return "", err
	}

	home, _ := os.UserHomeDir()
	claudeHome := filepath.Join(home, ".claude")

	// List of files to export
	exportFiles := []string{
		"CLAUDE.md",
		"settings.json",
		"cclsp.json",
		".mcp.json",
	}
	exportDirs := []string{
		"commands",
		"agents",
	}

	zipFile, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	// Export files
	for _, name := range exportFiles {
		filePath := filepath.Join(claudeHome, name)
		if data, err := os.ReadFile(filePath); err == nil {
			f, _ := w.Create(name)
			_, _ = f.Write(data)
		}
	}

	// Export directories
	for _, dir := range exportDirs {
		dirPath := filepath.Join(claudeHome, dir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			filePath := filepath.Join(dirPath, entry.Name())
			if data, err := os.ReadFile(filePath); err == nil {
				f, _ := w.Create(filepath.Join(dir, entry.Name()))
				_, _ = f.Write(data)
			}
		}
	}

	return savePath, nil
}

// ExportProjectConfig exports project configuration as a zip file
func (s *ExportService) ExportProjectConfig(projectPath string) (string, error) {
	projectName := filepath.Base(projectPath)
	savePath, err := runtime.SaveFileDialog(s.ctx, runtime.SaveDialogOptions{
		Title:           "Export Project Configuration",
		DefaultFilename: projectName + "-claude-config.zip",
		Filters: []runtime.FileFilter{
			{DisplayName: "ZIP Files", Pattern: "*.zip"},
		},
	})
	if err != nil || savePath == "" {
		return "", err
	}

	claudeDir := filepath.Join(projectPath, ".claude")

	zipFile, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	// Export root directory CLAUDE.md
	if data, err := os.ReadFile(filepath.Join(projectPath, "CLAUDE.md")); err == nil {
		f, _ := w.Create("CLAUDE.md")
		_, _ = f.Write(data)
	}

	// Recursively export .claude directory
	_ = filepath.Walk(claudeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(projectPath, path)
		if data, readErr := os.ReadFile(path); readErr == nil {
			f, _ := w.Create(relPath)
			_, _ = f.Write(data)
		}
		return nil
	})

	return savePath, nil
}

// ImportConfig imports configuration from a zip file
func (s *ExportService) ImportConfig(targetPath string) (int, error) {
	openPath, err := runtime.OpenFileDialog(s.ctx, runtime.OpenDialogOptions{
		Title: "Select Configuration ZIP File",
		Filters: []runtime.FileFilter{
			{DisplayName: "ZIP Files", Pattern: "*.zip"},
		},
	})
	if err != nil || openPath == "" {
		return 0, err
	}

	r, err := zip.OpenReader(openPath)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	count := 0
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		// Security check: prevent path traversal
		name := filepath.Clean(f.Name)
		if strings.Contains(name, "..") {
			continue
		}

		destPath := filepath.Join(targetPath, name)
		destDir := filepath.Dir(destPath)
		_ = os.MkdirAll(destDir, 0755)

		src, err := f.Open()
		if err != nil {
			continue
		}

		dst, err := os.Create(destPath)
		if err != nil {
			src.Close()
			continue
		}

		_, _ = io.Copy(dst, src)
		dst.Close()
		src.Close()
		count++
	}

	return count, nil
}

// ImportGlobalConfig imports to global configuration
func (s *ExportService) ImportGlobalConfig() (int, error) {
	home, _ := os.UserHomeDir()
	return s.ImportConfig(filepath.Join(home, ".claude"))
}

// ImportProjectConfig imports to project configuration
func (s *ExportService) ImportProjectConfig(projectPath string) (int, error) {
	return s.ImportConfig(projectPath)
}
