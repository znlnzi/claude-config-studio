package luoshu

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is the configuration structure for the luoshu memory system
type Config struct {
	Version   string          `json:"version"`
	LLM       LLMConfig       `json:"llm"`
	Embedding EmbeddingConfig `json:"embedding"`
	Memory    MemoryConfig    `json:"memory"`
	Reminder  ReminderConfig  `json:"reminder"`
}

// LLMConfig is the LLM provider configuration
type LLMConfig struct {
	Provider    string  `json:"provider"`
	APIKey      string  `json:"api_key"`
	Endpoint    string  `json:"endpoint"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// EmbeddingConfig is the Embedding provider configuration
type EmbeddingConfig struct {
	Provider   string `json:"provider"`
	APIKey     string `json:"api_key"`
	Endpoint   string `json:"endpoint"`
	Model      string `json:"model"`
	Dimensions int    `json:"dimensions"`
}

// MemoryConfig is the memory storage configuration
type MemoryConfig struct {
	AutoExtract      bool `json:"auto_extract"`
	RetentionDays    int  `json:"retention_days"`
	MaxEntries       int  `json:"max_entries"`
	VectorSearchTopK int  `json:"vector_search_top_k"`
}

// ReminderConfig is the reminder configuration
type ReminderConfig struct {
	Dismissed            bool `json:"dismissed"`
	PermanentlyDismissed bool `json:"permanently_dismissed"`
}

const configFileName = "config.json"

// ConfigDir returns the ~/.luoshu/ directory path, creating it if it doesn't exist (permissions 0700)
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".luoshu")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

// configFilePath returns the full path to the configuration file
func configFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: "1",
		LLM: LLMConfig{
			Provider:    "volcengine",
			Endpoint:    "https://ark.cn-beijing.volces.com/api/v3",
			Model:       "doubao-1.5-pro-256k",
			MaxTokens:   4096,
			Temperature: 0.3,
		},
		Embedding: EmbeddingConfig{
			Provider:   "volcengine",
			Endpoint:   "https://ark.cn-beijing.volces.com/api/v3",
			Model:      "doubao-embedding-large",
			Dimensions: 1024,
		},
		Memory: MemoryConfig{
			AutoExtract:      true,
			RetentionDays:    90,
			MaxEntries:       10000,
			VectorSearchTopK: 10,
		},
	}
}

// Load reads the configuration file and merges environment variable overrides
func Load() (*Config, error) {
	cfg := DefaultConfig()

	cfgPath, err := configFilePath()
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			applyEnvOverrides(cfg)
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

// Save writes the configuration to file (permissions 0600)
func Save(cfg *Config) error {
	cfgPath, err := configFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cfgPath, data, 0600)
}

// applyEnvOverrides overrides configuration with environment variables (takes precedence over file)
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("LUOSHU_LLM_API_KEY"); v != "" {
		cfg.LLM.APIKey = v
	}
	if v := os.Getenv("LUOSHU_LLM_MODEL"); v != "" {
		cfg.LLM.Model = v
	}
	if v := os.Getenv("LUOSHU_EMBEDDING_API_KEY"); v != "" {
		cfg.Embedding.APIKey = v
	}
	if v := os.Getenv("LUOSHU_EMBEDDING_MODEL"); v != "" {
		cfg.Embedding.Model = v
	}
}
