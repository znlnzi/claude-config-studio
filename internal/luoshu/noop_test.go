package luoshu

import (
	"context"
	"errors"
	"testing"
)

func TestNoopLLMProvider(t *testing.T) {
	noop := &NoopLLMProvider{}

	if noop.Name() != "noop" {
		t.Errorf("expected 'noop', got %q", noop.Name())
	}

	_, err := noop.Chat(context.Background(), ChatRequest{})
	if !errors.Is(err, ErrNotConfigured) {
		t.Errorf("expected ErrNotConfigured, got %v", err)
	}
}

func TestNoopEmbeddingProvider(t *testing.T) {
	noop := &NoopEmbeddingProvider{}

	if noop.Name() != "noop" {
		t.Errorf("expected 'noop', got %q", noop.Name())
	}

	if noop.Dimensions() != 0 {
		t.Errorf("expected 0, got %d", noop.Dimensions())
	}

	_, err := noop.Embed(context.Background(), []string{"test"})
	if !errors.Is(err, ErrNotConfigured) {
		t.Errorf("expected ErrNotConfigured, got %v", err)
	}
}
