package luoshu

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// LLMProvider is the LLM invocation interface
type LLMProvider interface {
	// Chat sends a chat request
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	// Name returns the Provider name
	Name() string
}

// EmbeddingProvider is the vector embedding interface
type EmbeddingProvider interface {
	// Embed converts text into vectors
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	// Dimensions returns the vector dimensions
	Dimensions() int
	// Name returns the Provider name
	Name() string
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Messages    []Message
	MaxTokens   int
	Temperature float64
}

// Message represents a chat message
type Message struct {
	Role    string // "system", "user", "assistant"
	Content string
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Content    string
	TokensUsed int
}

// ErrNotConfigured is the error returned when LLM/Embedding is not configured
var ErrNotConfigured = errors.New("luoshu: LLM/Embedding not configured")

// NewProviders creates LLM and Embedding Providers based on the configuration
// Returns Noop fallback implementations when api_key is not configured
func NewProviders(cfg *Config) (LLMProvider, EmbeddingProvider) {
	var llm LLMProvider = &NoopLLMProvider{}
	var embed EmbeddingProvider = &NoopEmbeddingProvider{}

	client := &http.Client{Timeout: 30 * time.Second}

	if cfg.LLM.APIKey != "" {
		llm = &VolcengineProvider{
			apiKey:     cfg.LLM.APIKey,
			endpoint:   cfg.LLM.Endpoint,
			llmModel:   cfg.LLM.Model,
			embedModel: cfg.Embedding.Model,
			dimensions: cfg.Embedding.Dimensions,
			httpClient: client,
		}
	}

	if cfg.Embedding.APIKey != "" {
		embed = &VolcengineProvider{
			apiKey:     cfg.Embedding.APIKey,
			endpoint:   cfg.Embedding.Endpoint,
			llmModel:   cfg.LLM.Model,
			embedModel: cfg.Embedding.Model,
			dimensions: cfg.Embedding.Dimensions,
			httpClient: client,
		}
	}

	return llm, embed
}
