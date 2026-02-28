package luoshu

import "context"

// NoopLLMProvider is the LLM fallback implementation when not configured
type NoopLLMProvider struct{}

// Chat returns an error indicating not configured
func (n *NoopLLMProvider) Chat(_ context.Context, _ ChatRequest) (*ChatResponse, error) {
	return nil, ErrNotConfigured
}

// Name returns the Provider name
func (n *NoopLLMProvider) Name() string { return "noop" }

// NoopEmbeddingProvider is the Embedding fallback implementation when not configured
type NoopEmbeddingProvider struct{}

// Embed returns an error indicating not configured
func (n *NoopEmbeddingProvider) Embed(_ context.Context, _ []string) ([][]float32, error) {
	return nil, ErrNotConfigured
}

// Dimensions returns zero dimensions
func (n *NoopEmbeddingProvider) Dimensions() int { return 0 }

// Name returns the Provider name
func (n *NoopEmbeddingProvider) Name() string { return "noop" }
