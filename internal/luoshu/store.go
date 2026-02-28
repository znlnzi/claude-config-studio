package luoshu

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
)

// MemoryStore is the JSONL format memory store
type MemoryStore struct {
	dir string // ~/.luoshu/memories/
}

// NewMemoryStore creates a new memory store instance
func NewMemoryStore() (*MemoryStore, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}
	memoriesDir := filepath.Join(dir, "memories")
	if err := os.MkdirAll(memoriesDir, 0700); err != nil {
		return nil, err
	}
	return &MemoryStore{dir: memoriesDir}, nil
}

// entriesPath returns the JSONL file path
func (s *MemoryStore) entriesPath() string {
	return filepath.Join(s.dir, "entries.jsonl")
}

// Append adds a memory entry
func (s *MemoryStore) Append(entry MemoryEntry) error {
	f, err := os.OpenFile(s.entriesPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = f.Write(data)
	return err
}

// LoadAll loads all entries
func (s *MemoryStore) LoadAll() ([]MemoryEntry, error) {
	f, err := os.Open(s.entriesPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []MemoryEntry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB max line
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry MemoryEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue // skip corrupted lines
		}
		entries = append(entries, entry)
	}
	return entries, scanner.Err()
}

// Count performs a fast count (by line count)
func (s *MemoryStore) Count() (int, error) {
	f, err := os.Open(s.entriesPath())
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	defer f.Close()
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if len(scanner.Bytes()) > 0 {
			count++
		}
	}
	return count, scanner.Err()
}
