package luoshu

// ProviderPreset defines a pre-configured LLM/Embedding provider
type ProviderPreset struct {
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	LLMEndpoint    string `json:"llm_endpoint"`
	LLMModel       string `json:"llm_model"`
	EmbedEndpoint  string `json:"embed_endpoint"`
	EmbedModel     string `json:"embed_model"`
	EmbedDimension int    `json:"embed_dimension"`
	KeyPrefix      string `json:"key_prefix,omitempty"` // expected API key prefix
}

// AllPresets returns all available provider presets
func AllPresets() map[string]ProviderPreset {
	return map[string]ProviderPreset{
		"volcengine": {
			Name:           "volcengine",
			DisplayName:    "Volcengine (Doubao)",
			LLMEndpoint:    "https://ark.cn-beijing.volces.com/api/v3",
			LLMModel:       "doubao-1.5-pro-256k",
			EmbedEndpoint:  "https://ark.cn-beijing.volces.com/api/v3",
			EmbedModel:     "doubao-embedding-large",
			EmbedDimension: 1024,
		},
		"openai": {
			Name:           "openai",
			DisplayName:    "OpenAI",
			LLMEndpoint:    "https://api.openai.com/v1",
			LLMModel:       "gpt-4o-mini",
			EmbedEndpoint:  "https://api.openai.com/v1",
			EmbedModel:     "text-embedding-3-small",
			EmbedDimension: 1536,
			KeyPrefix:      "sk-",
		},
		"deepseek": {
			Name:           "deepseek",
			DisplayName:    "DeepSeek",
			LLMEndpoint:    "https://api.deepseek.com/v1",
			LLMModel:       "deepseek-chat",
			EmbedEndpoint:  "https://api.deepseek.com/v1",
			EmbedModel:     "deepseek-chat",
			EmbedDimension: 1024,
			KeyPrefix:      "sk-",
		},
		"moonshot": {
			Name:           "moonshot",
			DisplayName:    "Moonshot AI (Kimi)",
			LLMEndpoint:    "https://api.moonshot.cn/v1",
			LLMModel:       "moonshot-v1-8k",
			EmbedEndpoint:  "https://api.moonshot.cn/v1",
			EmbedModel:     "moonshot-v1-8k",
			EmbedDimension: 1024,
			KeyPrefix:      "sk-",
		},
		"zhipu": {
			Name:           "zhipu",
			DisplayName:    "Zhipu AI (GLM)",
			LLMEndpoint:    "https://open.bigmodel.cn/api/paas/v4",
			LLMModel:       "glm-4-flash",
			EmbedEndpoint:  "https://open.bigmodel.cn/api/paas/v4",
			EmbedModel:     "embedding-3",
			EmbedDimension: 2048,
		},
		"siliconflow": {
			Name:           "siliconflow",
			DisplayName:    "SiliconFlow",
			LLMEndpoint:    "https://api.siliconflow.cn/v1",
			LLMModel:       "Qwen/Qwen2.5-7B-Instruct",
			EmbedEndpoint:  "https://api.siliconflow.cn/v1",
			EmbedModel:     "BAAI/bge-m3",
			EmbedDimension: 1024,
			KeyPrefix:      "sk-",
		},
		"custom": {
			Name:        "custom",
			DisplayName: "Custom (OpenAI-compatible)",
		},
	}
}

// GetPreset returns a provider preset by name, or nil if not found
func GetPreset(name string) *ProviderPreset {
	presets := AllPresets()
	p, ok := presets[name]
	if !ok {
		return nil
	}
	return &p
}
