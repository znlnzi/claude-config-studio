package luoshu

import (
	"os"
	"testing"
)

func newTestStore(t *testing.T) *MemoryStore {
	t.Helper()
	dir := t.TempDir()
	return &MemoryStore{dir: dir}
}

func TestStore_AppendAndLoadAll(t *testing.T) {
	store := newTestStore(t)

	entry1 := MemoryEntry{ID: "mem-1", Content: "first memory"}
	entry2 := MemoryEntry{ID: "mem-2", Content: "second memory"}

	if err := store.Append(entry1); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(entry2); err != nil {
		t.Fatal(err)
	}

	entries, err := store.LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].ID != "mem-1" || entries[1].ID != "mem-2" {
		t.Errorf("unexpected entries: %v", entries)
	}
}

func TestStore_LoadAll_Empty(t *testing.T) {
	store := newTestStore(t)
	entries, err := store.LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestStore_Count(t *testing.T) {
	store := newTestStore(t)

	count, err := store.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}

	if err := store.Append(MemoryEntry{ID: "mem-1", Content: "test"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(MemoryEntry{ID: "mem-2", Content: "test"}); err != nil {
		t.Fatal(err)
	}

	count, err = store.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestStore_CorruptedLine_Skipped(t *testing.T) {
	store := newTestStore(t)

	// Write a normal record
	if err := store.Append(MemoryEntry{ID: "mem-1", Content: "normal record"}); err != nil {
		t.Fatal(err)
	}

	// Manually append a corrupted line
	f, _ := openAppend(store.entriesPath())
	_, _ = f.WriteString("this is not json\n")
	f.Close()

	// Write another normal record
	if err := store.Append(MemoryEntry{ID: "mem-2", Content: "another normal record"}); err != nil {
		t.Fatal(err)
	}

	entries, err := store.LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 valid entries (skipping corrupted), got %d", len(entries))
	}
}

func TestStore_Tags_Preserved(t *testing.T) {
	store := newTestStore(t)
	entry := MemoryEntry{
		ID:      "mem-1",
		Content: "memory with tags",
		Tags:    []string{"decision", "architecture"},
	}
	if err := store.Append(entry); err != nil {
		t.Fatal(err)
	}

	entries, _ := store.LoadAll()
	if len(entries[0].Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(entries[0].Tags))
	}
	if entries[0].Tags[0] != "decision" {
		t.Errorf("expected 'decision', got %q", entries[0].Tags[0])
	}
}

// openAppend helper function
func openAppend(path string) (*appendWriter, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	return &appendWriter{f}, err
}

type appendWriter struct{ *os.File }
