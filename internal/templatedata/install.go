package templatedata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// WriteExtensionFiles writes map[name]content as .md files into the specified subdirectory.
// When overwrite=false, existing files are skipped; when overwrite=true, files are forcefully overwritten.
func WriteExtensionFiles(claudeDir, subDir string, files map[string]string, overwrite bool) error {
	if len(files) == 0 {
		return nil
	}
	dir := filepath.Join(claudeDir, subDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	for name, content := range files {
		filePath := filepath.Join(dir, name+".md")
		if !overwrite {
			if _, err := os.Stat(filePath); err == nil {
				continue
			}
			dirPath := filepath.Join(dir, name)
			if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
				continue
			}
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

// WriteSkillFiles writes skills in directory format: skills/{name}/SKILL.md
// Automatically migrates existing flat files to directory format before writing.
// When overwrite=false, existing entries are skipped; when overwrite=true, files are forcefully overwritten.
func WriteSkillFiles(claudeDir string, files map[string]string, overwrite bool) error {
	if len(files) == 0 {
		return nil
	}
	skillsDir := filepath.Join(claudeDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return err
	}
	MigrateFlatSkills(skillsDir)
	for name, content := range files {
		skillDir := filepath.Join(skillsDir, name)
		if !overwrite {
			if _, err := os.Stat(skillDir); err == nil {
				continue
			}
		}
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

// MergeAndWriteJSON reads an existing JSON file and deep-merges it with incoming data before writing back.
func MergeAndWriteJSON(path string, incoming interface{}) error {
	if incoming == nil {
		return nil
	}
	// Normalize types: Go native types → JSON standard types
	incomingMap, ok := normalizeToJSONTypes(incoming)
	if !ok {
		data, err := json.MarshalIndent(incoming, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(path, data, 0644)
	}

	existing := make(map[string]interface{})
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &existing)
	}

	merged := deepMergeMap(existing, incomingMap)
	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

// MigrateFlatSkills migrates all flat .md format skills in the specified directory to directory format {name}/SKILL.md.
func MigrateFlatSkills(skillsDir string) int {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".md")
		flatPath := filepath.Join(skillsDir, entry.Name())
		dirPath := filepath.Join(skillsDir, name)
		if _, err := os.Stat(dirPath); err == nil {
			continue // Directory with the same name already exists, skip
		}
		content, err := os.ReadFile(flatPath)
		if err != nil {
			continue
		}
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(dirPath, "SKILL.md"), content, 0644); err != nil {
			continue
		}
		os.Remove(flatPath)
		count++
	}
	return count
}

// deepMergeMap merges src into dst, preserving existing values without overwriting, only adding new keys.
func deepMergeMap(dst, src map[string]interface{}) map[string]interface{} {
	for key, srcVal := range src {
		dstVal, exists := dst[key]
		if !exists {
			dst[key] = srcVal
			continue
		}
		srcMap, srcIsMap := srcVal.(map[string]interface{})
		dstMap, dstIsMap := dstVal.(map[string]interface{})
		if srcIsMap && dstIsMap {
			dst[key] = deepMergeMap(dstMap, srcMap)
			continue
		}
		srcArr, srcIsArr := srcVal.([]interface{})
		dstArr, dstIsArr := dstVal.([]interface{})
		if srcIsArr && dstIsArr {
			dst[key] = append(dstArr, srcArr...)
			continue
		}
		// Different types or both scalars, keep existing value
	}
	return dst
}

// normalizeToJSONTypes normalizes Go types via JSON marshal/unmarshal round-trip.
// Resolves type mismatch between []map[string]interface{} and []interface{}.
func normalizeToJSONTypes(v interface{}) (map[string]interface{}, bool) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false
	}
	return result, true
}
