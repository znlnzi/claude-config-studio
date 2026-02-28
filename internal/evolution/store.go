package evolution

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Store manages persistence of evolution data
type Store struct {
	baseDir string // ~/.claude/evolution/
}

// NewStore creates a Store instance
func NewStore(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

// EnsureDir ensures the directory exists
func (s *Store) EnsureDir() error {
	if err := os.MkdirAll(s.baseDir, 0755); err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(s.baseDir, "backups"), 0755)
}

// suggestionsPath returns the path to suggestions.json
func (s *Store) suggestionsPath() string {
	return filepath.Join(s.baseDir, "suggestions.json")
}

// historyPath returns the path to history.json
func (s *Store) historyPath() string {
	return filepath.Join(s.baseDir, "history.json")
}

// LoadSuggestions loads all suggestions
func (s *Store) LoadSuggestions() ([]Suggestion, error) {
	data, err := os.ReadFile(s.suggestionsPath())
	if err != nil {
		return []Suggestion{}, nil
	}
	var suggestions []Suggestion
	if err := json.Unmarshal(data, &suggestions); err != nil {
		return []Suggestion{}, nil
	}
	return suggestions, nil
}

// SaveSuggestions saves all suggestions
func (s *Store) SaveSuggestions(suggestions []Suggestion) error {
	if err := s.EnsureDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(suggestions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.suggestionsPath(), data, 0644)
}

// AddSuggestions appends new suggestions (auto-dedup: suggestions with the same type + title are not added again)
func (s *Store) AddSuggestions(newSuggs []Suggestion) error {
	existing, _ := s.LoadSuggestions()

	for _, ns := range newSuggs {
		isDuplicate := false
		for _, es := range existing {
			if es.Status == "pending" && es.Type == ns.Type && es.Title == ns.Title {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			existing = append(existing, ns)
		}
	}

	return s.SaveSuggestions(existing)
}

// GetPendingSuggestions returns suggestions awaiting approval
func (s *Store) GetPendingSuggestions() ([]Suggestion, error) {
	all, err := s.LoadSuggestions()
	if err != nil {
		return nil, err
	}
	var pending []Suggestion
	for _, sg := range all {
		if sg.Status == "pending" {
			pending = append(pending, sg)
		}
	}
	return pending, nil
}

// UpdateSuggestionStatus updates the status of a suggestion
func (s *Store) UpdateSuggestionStatus(id, status string) error {
	suggestions, err := s.LoadSuggestions()
	if err != nil {
		return err
	}

	found := false
	for i := range suggestions {
		if suggestions[i].ID == id {
			suggestions[i].Status = status
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("suggestion not found: %s", id)
	}

	return s.SaveSuggestions(suggestions)
}

// LoadHistory loads the analysis history
func (s *Store) LoadHistory() ([]AnalysisRecord, error) {
	data, err := os.ReadFile(s.historyPath())
	if err != nil {
		return []AnalysisRecord{}, nil
	}
	var history []AnalysisRecord
	if err := json.Unmarshal(data, &history); err != nil {
		return []AnalysisRecord{}, nil
	}
	return history, nil
}

// AddHistory adds an analysis history record
func (s *Store) AddHistory(record AnalysisRecord) error {
	if err := s.EnsureDir(); err != nil {
		return err
	}
	history, _ := s.LoadHistory()
	history = append(history, record)

	// Keep only the most recent 100 records
	if len(history) > 100 {
		history = history[len(history)-100:]
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.historyPath(), data, 0644)
}

// BackupFile backs up a file before applying a suggestion
func (s *Store) BackupFile(originalPath string) (string, error) {
	if err := s.EnsureDir(); err != nil {
		return "", err
	}

	data, err := os.ReadFile(originalPath)
	if err != nil {
		return "", err
	}

	baseName := filepath.Base(originalPath)
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s.%s.bak", baseName, timestamp)
	backupPath := filepath.Join(s.baseDir, "backups", backupName)

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", err
	}

	return backupPath, nil
}

// GetStatus returns the current status of the evolution system
func (s *Store) GetStatus() (EvolveStatus, error) {
	suggestions, _ := s.LoadSuggestions()
	history, _ := s.LoadHistory()

	status := EvolveStatus{
		Initialized:   fileExists(s.suggestionsPath()) || fileExists(s.historyPath()),
		TotalAnalyses: len(history),
	}

	for _, sg := range suggestions {
		switch sg.Status {
		case "pending":
			status.PendingSuggestions++
		case "approved":
			status.ApprovedTotal++
		case "rejected":
			status.RejectedTotal++
		}
	}

	if len(history) > 0 {
		status.LastAnalysis = history[len(history)-1].Timestamp
	}

	return status, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
