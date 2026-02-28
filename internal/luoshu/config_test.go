package luoshu

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Version != "1" {
		t.Errorf("expected version '1', got %q", cfg.Version)
	}
	if cfg.LLM.Provider != "volcengine" {
		t.Errorf("expected volcengine, got %q", cfg.LLM.Provider)
	}
	if cfg.LLM.MaxTokens != 4096 {
		t.Errorf("expected 4096, got %d", cfg.LLM.MaxTokens)
	}
	if cfg.Embedding.Dimensions != 1024 {
		t.Errorf("expected 1024, got %d", cfg.Embedding.Dimensions)
	}
	if !cfg.Memory.AutoExtract {
		t.Error("expected auto_extract true by default")
	}
	if cfg.Memory.RetentionDays != 90 {
		t.Errorf("expected 90, got %d", cfg.Memory.RetentionDays)
	}
}

func TestApplyEnvOverrides_LLMKey(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("LUOSHU_LLM_API_KEY", "test-key-123")
	applyEnvOverrides(cfg)

	if cfg.LLM.APIKey != "test-key-123" {
		t.Errorf("expected 'test-key-123', got %q", cfg.LLM.APIKey)
	}
}

func TestApplyEnvOverrides_EmbeddingKey(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("LUOSHU_EMBEDDING_API_KEY", "embed-key-456")
	applyEnvOverrides(cfg)

	if cfg.Embedding.APIKey != "embed-key-456" {
		t.Errorf("expected 'embed-key-456', got %q", cfg.Embedding.APIKey)
	}
}

func TestApplyEnvOverrides_Models(t *testing.T) {
	cfg := DefaultConfig()

	t.Setenv("LUOSHU_LLM_MODEL", "custom-llm")
	t.Setenv("LUOSHU_EMBEDDING_MODEL", "custom-embed")
	applyEnvOverrides(cfg)

	if cfg.LLM.Model != "custom-llm" {
		t.Errorf("expected 'custom-llm', got %q", cfg.LLM.Model)
	}
	if cfg.Embedding.Model != "custom-embed" {
		t.Errorf("expected 'custom-embed', got %q", cfg.Embedding.Model)
	}
}

func TestApplyEnvOverrides_NoOverride(t *testing.T) {
	cfg := DefaultConfig()
	originalModel := cfg.LLM.Model

	// Do not set any environment variables
	applyEnvOverrides(cfg)

	if cfg.LLM.Model != originalModel {
		t.Errorf("model should not change without env var")
	}
}

func TestNewProviders_NoAPIKey(t *testing.T) {
	cfg := DefaultConfig()
	// Do not set API Key

	llm, embed := NewProviders(cfg)
	if llm.Name() != "noop" {
		t.Errorf("expected noop LLM, got %s", llm.Name())
	}
	if embed.Name() != "noop" {
		t.Errorf("expected noop embedder, got %s", embed.Name())
	}
}

func TestNewProviders_WithAPIKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LLM.APIKey = "test-key"
	cfg.Embedding.APIKey = "test-key"

	llm, embed := NewProviders(cfg)
	if llm.Name() == "noop" {
		t.Error("expected non-noop LLM with API key")
	}
	if embed.Name() == "noop" {
		t.Error("expected non-noop embedder with API key")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	// Use temp directory to simulate ~/.luoshu
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := DefaultConfig()
	cfg.LLM.APIKey = "round-trip-test"

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.LLM.APIKey != "round-trip-test" {
		t.Errorf("expected 'round-trip-test', got %q", loaded.LLM.APIKey)
	}
}
