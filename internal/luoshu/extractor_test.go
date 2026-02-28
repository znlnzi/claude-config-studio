package luoshu

import (
	"context"
	"testing"
)

func TestExtractJSON_FromCodeBlock(t *testing.T) {
	text := "Some text\n```json\n[{\"content\":\"test\"}]\n```\nMore text"
	result := extractJSON(text)
	if result != "[{\"content\":\"test\"}]" {
		t.Errorf("unexpected: %q", result)
	}
}

func TestExtractJSON_Plain(t *testing.T) {
	text := `[{"content":"test"}]`
	result := extractJSON(text)
	if result != text {
		t.Errorf("unexpected: %q", result)
	}
}

func TestExtractJSON_NoArray(t *testing.T) {
	text := "no json here"
	result := extractJSON(text)
	if result != text {
		t.Errorf("expected original text when no JSON found, got %q", result)
	}
}

func TestExtractor_Extract(t *testing.T) {
	store := newTestStore(t)
	llm := &mockLLM{
		response: `[{"content":"Use JWT for authentication","tags":["decision"],"importance":4}]`,
	}

	extractor := NewExtractor(llm, store)
	entries, err := extractor.Extract(context.Background(), "Discussed authentication approach today, decided to use JWT", "/my-project", []string{"session"})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 extracted entry, got %d", len(entries))
	}
	if entries[0].Content != "Use JWT for authentication" {
		t.Errorf("unexpected content: %s", entries[0].Content)
	}
	// Verify tags are merged
	found := false
	for _, tag := range entries[0].Tags {
		if tag == "session" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'session' tag to be merged")
	}

	// Verify persistence
	stored, _ := store.LoadAll()
	if len(stored) != 1 {
		t.Fatalf("expected 1 stored entry, got %d", len(stored))
	}
}

func TestExtractor_LLMError(t *testing.T) {
	store := newTestStore(t)
	llm := &mockLLM{err: ErrNotConfigured}

	extractor := NewExtractor(llm, store)
	_, err := extractor.Extract(context.Background(), "test", "/project", nil)
	if err == nil {
		t.Fatal("expected error when LLM fails")
	}
}

func TestExtractor_InvalidJSON(t *testing.T) {
	store := newTestStore(t)
	llm := &mockLLM{response: "This is not JSON"}

	extractor := NewExtractor(llm, store)
	_, err := extractor.Extract(context.Background(), "test", "/project", nil)
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}
