---
name: luoshu.config
description: View and modify luoshu intelligent memory configuration (LLM service, API Key, etc.).
---

# /luoshu.config -- Configuration Management

Display current luoshu configuration status and guide users to configure or modify the LLM service.

---

## Flow

### 1. Get Current Configuration

Call `luoshu_config_get` to retrieve the full configuration (API Key is automatically masked).

### 2. Verify Connection Status

If LLM is configured (api_key is not empty), call `luoshu_config_validate` to check the connection.

### 3. Display Based on Status

Display the corresponding interface based on configuration and connection status.

---

## Display Format

### When Not Configured

```
Current luoshu configuration:

  Provider:  Not configured
  API Key:   Not configured
  Status:    Not connected

The following features are unavailable due to missing configuration:
  - Cross-session intelligent memory
  - Semantic search
  - Automatic session highlights extraction

Configure LLM service?

  [1] Configure Volcengine (recommended, low latency for users in China)
  [2] Later
```

After selecting [1], enter the configuration flow:
1. Ask "Do you have a Volcengine API Key?"
2. **Has Key** → Call `luoshu_config_set`(key=`llm.api_key`, value=user input) → `luoshu_config_validate` to verify
3. **No Key** → Step-by-step guide to register Volcengine + obtain Key (same as /luoshu.setup Step 4.4)
4. **Skip** → End

### When Configured and Connected Successfully

```
Current luoshu configuration:

  Provider:  Volcengine
  API Key:   sk-****xxxx
  Endpoint:  ark.cn-beijing.volces.com
  LLM Model: doubao-1.5-pro-256k
  Embedding Model: doubao-embedding-large
  Status:    Connected

  Memory strategy:
  - Auto extraction: Enabled
  - Retention days: 90 days

Enter a number to modify:
  [1] Update API Key
  [2] Change LLM Model
  [3] Change Embedding Model
  [4] Modify memory strategy
  [5] Test connection
  [6] View system status
  [7] Rebuild vector index
```

Operations for each option:
- **[1] Update API Key** → Request new Key input → `luoshu_config_set`(key=`llm.api_key`) → `luoshu_config_validate`
- **[2] Change LLM Model** → Display available models → `luoshu_config_set`(key=`llm.model`)
- **[3] Change Embedding Model** → Display available models → `luoshu_config_set`(key=`embedding.model`)
- **[4] Modify memory strategy** → Display current values → `luoshu_config_set`(key=`memory.auto_extract` or `memory.retention_days`)
- **[5] Test connection** → `luoshu_config_validate`
- **[6] View system status** → `luoshu_status` (display memory entry count, vector index count, cache size)
- **[7] Rebuild vector index** → `luoshu_reindex`

### When Configured but Connection Failed

```
Current luoshu configuration:

  Provider:  Volcengine
  API Key:   sk-****xxxx
  Status:    Connection failed (authentication rejected)

The API Key may have expired or been revoked.

  [1] Update API Key
  [2] Help me get a new Key
  [3] Test connection (retry)
```

---

## Key Validation Failure Diagnostics

When the Key entered by the user fails validation, display diagnostics based on the pre-check result from `luoshu_config_set`:

| Pre-check Result | Display Message |
|-----------------|-----------------|
| OpenAI Key (sk-proj-) | "This appears to be an OpenAI Key. Please confirm you're using a Volcengine Key" |
| Anthropic Key (sk-ant-) | "This appears to be an Anthropic Key" |
| AWS Key (AKIA) | "This appears to be an AWS Key" |
| GitHub Token (ghp_) | "This appears to be a GitHub Token" |
| Key too short (<20 characters) | "The Key is only N characters long, it may have been copied incompletely" |

After diagnostics, provide options:
```
[1] Re-enter
[2] Help me get the correct Key
[3] Skip for now
```

---

## Security Statement

Always display after configuration is complete:
```
Key is saved locally at ~/.luoshu/config.json
It will not be sent to Anthropic or any other external service.
```

---

## Key Principles

- Always display API Key in masked form (show only first 3 + last 4 characters)
- Don't panic on connection test failure; provide specific diagnostics and solutions
- Automatically verify after modification to ensure the configuration is valid
- Display the full configuration overview after each operation, satisfying the visibility principle
