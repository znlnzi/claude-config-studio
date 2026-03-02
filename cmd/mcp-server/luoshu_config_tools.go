package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/znlnzi/claude-config-studio/internal/luoshu"

	"github.com/mark3labs/mcp-go/mcp"
)

// ─── luoshu config tool definitions ─────────────────────────────

func buildLuoshuConfigGetTool() mcp.Tool {
	return mcp.NewTool(
		"luoshu_config_get",
		mcp.WithDescription("Read luoshu memory system configuration. API Keys are automatically masked (first 3 and last 4 characters shown)."),
		mcp.WithString("section",
			mcp.Description("Optional: return only a specific section 'llm', 'embedding', 'memory', 'reminder'; leave empty to return all"),
		),
	)
}

func buildLuoshuConfigSetTool() mcp.Tool {
	return mcp.NewTool(
		"luoshu_config_set",
		mcp.WithDescription("Set a luoshu memory system configuration value. A connection test is automatically performed after setting an API Key."),
		mcp.WithString("key",
			mcp.Required(),
			mcp.Description("Configuration key path, e.g. 'llm.api_key', 'llm.model', 'embedding.api_key', 'memory.auto_extract'"),
		),
		mcp.WithString("value",
			mcp.Required(),
			mcp.Description("Configuration value (as string; numbers and booleans should also be passed as strings)"),
		),
	)
}

func buildLuoshuConfigValidateTool() mcp.Tool {
	return mcp.NewTool(
		"luoshu_config_validate",
		mcp.WithDescription("Validate luoshu configuration completeness and test the LLM API connection."),
	)
}

// ─── luoshu config tool handlers ─────────────────────────────

func handleLuoshuConfigGet(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg, err := luoshu.Load()
	if err != nil {
		return mcp.NewToolResultError(errConfigLoad(err)), nil
	}

	section := req.GetString("section", "")

	// Mask sensitive fields: create a copy
	masked := maskConfig(cfg)

	var output interface{}
	switch section {
	case "llm":
		output = masked.LLM
	case "embedding":
		output = masked.Embedding
	case "memory":
		output = masked.Memory
	case "reminder":
		output = masked.Reminder
	case "":
		output = masked
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unknown section: %s. Valid options: llm, embedding, memory, reminder", section)), nil
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleLuoshuConfigSet(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	key, err := req.RequireString("key")
	if err != nil {
		return mcp.NewToolResultError("key parameter is required"), nil
	}
	value, err := req.RequireString("value")
	if err != nil {
		return mcp.NewToolResultError("value parameter is required"), nil
	}

	cfg, err := luoshu.Load()
	if err != nil {
		return mcp.NewToolResultError(errConfigLoad(err)), nil
	}

	if err := applyConfigValue(cfg, key, value); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := luoshu.Save(cfg); err != nil {
		return mcp.NewToolResultError(errConfigSave(err)), nil
	}

	result := map[string]interface{}{
		"success": true,
		"key":     key,
	}

	// Automatically perform connection test after setting api_key
	if strings.HasSuffix(key, ".api_key") {
		connected, status, testErr := luoshu.TestConnection(cfg)
		result["connection_test"] = map[string]interface{}{
			"connected": connected,
			"status":    status,
		}
		if testErr != nil {
			result["connection_test"].(map[string]interface{})["error"] = testErr.Error()
		}
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleLuoshuConfigValidate(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg, err := luoshu.Load()
	if err != nil {
		return mcp.NewToolResultError(errConfigLoad(err)), nil
	}

	issues := validateConfig(cfg)

	// Connection test
	connected, status, testErr := luoshu.TestConnection(cfg)
	connResult := map[string]interface{}{
		"connected": connected,
		"status":    status,
	}
	if testErr != nil {
		connResult["error"] = testErr.Error()
	}

	result := map[string]interface{}{
		"valid":           len(issues) == 0,
		"issues":          issues,
		"connection_test": connResult,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// ─── Helper functions ─────────────────────────────

// maskConfig returns a masked copy of the config
func maskConfig(cfg *luoshu.Config) *luoshu.Config {
	masked := *cfg
	masked.LLM.APIKey = luoshu.MaskKey(cfg.LLM.APIKey)
	masked.Embedding.APIKey = luoshu.MaskKey(cfg.Embedding.APIKey)
	return &masked
}

// applyConfigValue applies a string value to the specified config path
func applyConfigValue(cfg *luoshu.Config, key, value string) error {
	switch key {
	// LLM
	case "llm.provider":
		cfg.LLM.Provider = value
		// Auto-fill defaults from preset when provider is set
		if preset := luoshu.GetPreset(value); preset != nil && preset.LLMEndpoint != "" {
			if cfg.LLM.Endpoint == "" || cfg.LLM.Endpoint == luoshu.DefaultConfig().LLM.Endpoint {
				cfg.LLM.Endpoint = preset.LLMEndpoint
			}
			if cfg.LLM.Model == "" || cfg.LLM.Model == luoshu.DefaultConfig().LLM.Model {
				cfg.LLM.Model = preset.LLMModel
			}
			// Also set embedding defaults if not independently configured
			if cfg.Embedding.Provider == "" || cfg.Embedding.Provider == cfg.LLM.Provider {
				cfg.Embedding.Provider = value
				if cfg.Embedding.Endpoint == "" || cfg.Embedding.Endpoint == luoshu.DefaultConfig().Embedding.Endpoint {
					cfg.Embedding.Endpoint = preset.EmbedEndpoint
				}
				if cfg.Embedding.Model == "" || cfg.Embedding.Model == luoshu.DefaultConfig().Embedding.Model {
					cfg.Embedding.Model = preset.EmbedModel
				}
				if preset.EmbedDimension > 0 && cfg.Embedding.Dimensions == luoshu.DefaultConfig().Embedding.Dimensions {
					cfg.Embedding.Dimensions = preset.EmbedDimension
				}
			}
		}
	case "llm.api_key":
		if _, err := luoshu.PreValidateKey(value); err != nil {
			return fmt.Errorf("API Key pre-validation failed: %v", err)
		}
		cfg.LLM.APIKey = value
	case "llm.endpoint":
		cfg.LLM.Endpoint = value
	case "llm.model":
		cfg.LLM.Model = value
	case "llm.max_tokens":
		n, err := parsePositiveInt(value)
		if err != nil {
			return fmt.Errorf("max_tokens must be a positive integer: %v", err)
		}
		cfg.LLM.MaxTokens = n
	case "llm.temperature":
		f, err := parseFloat(value)
		if err != nil {
			return fmt.Errorf("temperature must be a number: %v", err)
		}
		cfg.LLM.Temperature = f
	// Embedding
	case "embedding.provider":
		cfg.Embedding.Provider = value
	case "embedding.api_key":
		if _, err := luoshu.PreValidateKey(value); err != nil {
			return fmt.Errorf("API Key pre-validation failed: %v", err)
		}
		cfg.Embedding.APIKey = value
	case "embedding.endpoint":
		cfg.Embedding.Endpoint = value
	case "embedding.model":
		cfg.Embedding.Model = value
	case "embedding.dimensions":
		n, err := parsePositiveInt(value)
		if err != nil {
			return fmt.Errorf("dimensions must be a positive integer: %v", err)
		}
		cfg.Embedding.Dimensions = n
	// Memory
	case "memory.auto_extract":
		cfg.Memory.AutoExtract = parseBool(value)
	case "memory.retention_days":
		n, err := parsePositiveInt(value)
		if err != nil {
			return fmt.Errorf("retention_days must be a positive integer: %v", err)
		}
		cfg.Memory.RetentionDays = n
	case "memory.max_entries":
		n, err := parsePositiveInt(value)
		if err != nil {
			return fmt.Errorf("max_entries must be a positive integer: %v", err)
		}
		cfg.Memory.MaxEntries = n
	case "memory.vector_search_top_k":
		n, err := parsePositiveInt(value)
		if err != nil {
			return fmt.Errorf("vector_search_top_k must be a positive integer: %v", err)
		}
		cfg.Memory.VectorSearchTopK = n
	// Reminder
	case "reminder.dismissed":
		cfg.Reminder.Dismissed = parseBool(value)
	case "reminder.permanently_dismissed":
		cfg.Reminder.PermanentlyDismissed = parseBool(value)
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

// validateConfig checks config completeness and returns a list of issues
func validateConfig(cfg *luoshu.Config) []string {
	var issues []string
	if cfg.LLM.APIKey == "" {
		issues = append(issues, "llm.api_key is not set")
	}
	if cfg.LLM.Endpoint == "" {
		issues = append(issues, "llm.endpoint is not set")
	}
	if cfg.LLM.Model == "" {
		issues = append(issues, "llm.model is not set")
	}
	if cfg.Embedding.APIKey == "" {
		issues = append(issues, "embedding.api_key is not set")
	}
	if cfg.Embedding.Endpoint == "" {
		issues = append(issues, "embedding.endpoint is not set")
	}
	if cfg.Embedding.Model == "" {
		issues = append(issues, "embedding.model is not set")
	}
	return issues
}

// parsePositiveInt parses a positive integer
func parsePositiveInt(s string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, fmt.Errorf("value must be greater than 0")
	}
	return n, nil
}

// parseFloat parses a floating-point number
func parseFloat(s string) (float64, error) {
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return 0, err
	}
	return f, nil
}

// parseBool parses a boolean value
func parseBool(s string) bool {
	return s == "true" || s == "1" || s == "yes"
}
