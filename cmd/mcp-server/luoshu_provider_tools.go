package main

import (
	"context"
	"encoding/json"

	"github.com/znlnzi/claude-config-studio/internal/luoshu"

	"github.com/mark3labs/mcp-go/mcp"
)

// ─── luoshu_provider_list tool ─────────────────────────────

func buildLuoshuProviderListTool() mcp.Tool {
	return mcp.NewTool(
		"luoshu_provider_list",
		mcp.WithDescription("List all available LLM/Embedding provider presets. Each preset includes recommended endpoint, model, and embedding configuration. Use with luoshu_config_set to quickly configure a provider."),
	)
}

func handleLuoshuProviderList(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	presets := luoshu.AllPresets()

	// Build a sorted list for consistent output
	type presetInfo struct {
		Name           string `json:"name"`
		DisplayName    string `json:"display_name"`
		LLMEndpoint    string `json:"llm_endpoint,omitempty"`
		LLMModel       string `json:"llm_model,omitempty"`
		EmbedEndpoint  string `json:"embed_endpoint,omitempty"`
		EmbedModel     string `json:"embed_model,omitempty"`
		EmbedDimension int    `json:"embed_dimension,omitempty"`
	}

	var list []presetInfo
	// Define display order
	order := []string{"openai", "deepseek", "moonshot", "zhipu", "siliconflow", "volcengine", "custom"}
	for _, name := range order {
		if p, ok := presets[name]; ok {
			list = append(list, presetInfo{
				Name:           p.Name,
				DisplayName:    p.DisplayName,
				LLMEndpoint:    p.LLMEndpoint,
				LLMModel:       p.LLMModel,
				EmbedEndpoint:  p.EmbedEndpoint,
				EmbedModel:     p.EmbedModel,
				EmbedDimension: p.EmbedDimension,
			})
		}
	}

	result := map[string]interface{}{
		"providers": list,
		"usage":     "Set provider with: luoshu_config_set key='llm.provider' value='<name>', then set API key with: luoshu_config_set key='llm.api_key' value='<key>'",
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
