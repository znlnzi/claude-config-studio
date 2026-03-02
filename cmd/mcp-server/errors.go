package main

import "fmt"

// Error formatting helpers that attach actionable fix suggestions to error messages.

func errReadFailed(path string, err error) string {
	return fmt.Sprintf("failed to read %s: %v. Check that the file exists and has correct permissions", path, err)
}

func errWriteFailed(path string, err error) string {
	return fmt.Sprintf("failed to write %s: %v. Check file permissions and disk space", path, err)
}

func errCreateDir(path string, err error) string {
	return fmt.Sprintf("failed to create directory %s: %v. Check parent directory permissions", path, err)
}

func errDeleteFailed(path string, err error) string {
	return fmt.Sprintf("failed to delete %s: %v. Check file permissions", path, err)
}

func errHomeDir(err error) string {
	return fmt.Sprintf("failed to resolve home directory: %v. Check $HOME environment variable", err)
}

func errPathNotFound(path string) string {
	return fmt.Sprintf("project path does not exist: %s. Verify the absolute path is correct", path)
}

func errInvalidJSON(field string, err error) string {
	return fmt.Sprintf("invalid JSON for %s: %v. Ensure the content is valid JSON", field, err)
}

func errConfigLoad(err error) string {
	return fmt.Sprintf("failed to load luoshu config (~/.luoshu/config.json): %v. Run /luoshu.config to initialize", err)
}

func errConfigSave(err error) string {
	return fmt.Sprintf("failed to save luoshu config: %v. Check ~/.luoshu/ directory permissions", err)
}

func errInitFailed(component string, err error) string {
	return fmt.Sprintf("failed to initialize %s: %v. Check ~/.luoshu/ directory permissions", component, err)
}

func errTemplateNotFound(id string) string {
	return fmt.Sprintf("template not found: %s. Run template_list to see available templates", id)
}

func errExtensionNotFound(extType, name, scope string) string {
	return fmt.Sprintf("extension not found: %s/%s in %s scope. Check the name with extension_list", extType, name, scope)
}

func errSuggestionNotFound(id string) string {
	return fmt.Sprintf("suggestion not found: %s. Run evolve_status to see pending suggestions", id)
}

func errSearchFailed(err error) string {
	return fmt.Sprintf("semantic search failed: %v. Check embedding configuration with luoshu_config_validate", err)
}

func errInvalidFilename(name string) string {
	return fmt.Sprintf("invalid filename: %s. Filename must end in .md and cannot contain '..' or path separators", name)
}
