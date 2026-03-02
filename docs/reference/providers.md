# Providers

Luoshu supports multiple LLM providers through an OpenAI-compatible API interface. Each provider has a built-in preset that auto-fills endpoint and model defaults.

## Built-in Presets

| Provider | Preset Name | LLM Model | Embedding Model |
|----------|------------|-----------|-----------------|
| OpenAI | `openai` | gpt-4o-mini | text-embedding-3-small |
| DeepSeek | `deepseek` | deepseek-chat | — |
| Moonshot (Kimi) | `moonshot` | moonshot-v1-8k | — |
| Zhipu (GLM) | `zhipu` | glm-4-flash | embedding-3 |
| SiliconFlow | `siliconflow` | Qwen/Qwen2.5-7B-Instruct | BAAI/bge-m3 |
| Volcengine (Doubao) | `volcengine` | doubao-1-5-lite-32k | doubao-embedding-large |
| Custom | `custom` | (user-defined) | (user-defined) |

Use `luoshu_provider_list` to see the full details including endpoints and embedding dimensions.

## Selecting a Provider

Set the provider preset to auto-fill defaults:

```
Tool: luoshu_config_set
Parameters:
  key: "llm.provider"
  value: "deepseek"
```

This automatically sets `llm.endpoint` and `llm.model` to the preset defaults. You can override individual fields afterward.

## Setting API Keys

```
Tool: luoshu_config_set
Parameters:
  key: "llm.api_key"
  value: "sk-your-api-key"
```

A connection test runs automatically after setting an API key.

## LLM vs Embedding

Luoshu uses two separate services:

| Service | Purpose | Required For |
|---------|---------|-------------|
| LLM | Text generation (recall synthesis, memory extraction) | `luoshu_recall`, `memory_extract` |
| Embedding | Vector embeddings (similarity search) | `memory_semantic_search`, `ov_search` |

You can use different providers for LLM and embedding. For example, use DeepSeek for LLM and OpenAI for embeddings.

### Configuring Embedding Separately

```
Tool: luoshu_config_set
Parameters:
  key: "embedding.provider"
  value: "openai"
```

```
Tool: luoshu_config_set
Parameters:
  key: "embedding.api_key"
  value: "sk-your-openai-key"
```

## Custom Provider

For providers not in the preset list, use the `custom` preset and set all fields manually:

```
Tool: luoshu_config_set
  key: "llm.provider", value: "custom"

Tool: luoshu_config_set
  key: "llm.endpoint", value: "https://your-api.example.com/v1"

Tool: luoshu_config_set
  key: "llm.model", value: "your-model-name"

Tool: luoshu_config_set
  key: "llm.api_key", value: "your-api-key"
```

Any OpenAI-compatible API endpoint works as a custom provider.

## Environment Variables

Override config file values with environment variables:

| Variable | Description |
|----------|-------------|
| `LUOSHU_LLM_API_KEY` | LLM service API key |
| `LUOSHU_LLM_MODEL` | LLM model name |
| `LUOSHU_EMBEDDING_API_KEY` | Embedding service API key |
| `LUOSHU_EMBEDDING_MODEL` | Embedding model name |

Environment variables take precedence over `~/.luoshu/config.json` values.

## Advanced Configuration

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `llm.max_tokens` | number | (provider default) | Maximum tokens for LLM responses |
| `llm.temperature` | number | (provider default) | LLM temperature setting |
| `embedding.dimensions` | number | (provider default) | Embedding vector dimensions |
| `memory.auto_extract` | boolean | true | Auto-extract memories from conversations |
| `memory.retention_days` | number | 90 | Days to retain memories |
| `memory.max_entries` | number | 1000 | Maximum memory entries |
| `memory.vector_search_top_k` | number | 10 | Top-K results for vector search |
