---
name: luoshu.config
description: 查看和修改 luoshu 智能记忆的配置（LLM 服务、API Key 等）。
---

# /luoshu.config -- 配置管理

展示当前 luoshu 配置状态，引导用户配置或修改 LLM 服务。

---

## 流程

### 1. 获取当前配置

调用 `luoshu_config_get` 获取完整配置（API Key 已自动脱敏）。

### 2. 验证连接状态

如果 LLM 已配置（api_key 非空），调用 `luoshu_config_validate` 检查连接。

### 3. 按状态展示

根据配置和连接状态，展示对应界面。

---

## 展示格式

### 未配置时

```
当前 luoshu 配置：

  服务商:  未配置
  API Key: 未配置
  状态:    未连接

以下功能因未配置而不可用：
  - 跨会话智能记忆
  - 语义搜索
  - 自动提取会话要点

配置 LLM 服务？

  [1] 配置火山引擎（推荐，国内用户延迟低）
  [2] 稍后再说
```

选择 [1] 后，进入配置流程：
1. 提问"你有火山引擎的 API Key 吗？"
2. **有 Key** → 调用 `luoshu_config_set`(key=`llm.api_key`, value=用户输入) → `luoshu_config_validate` 验证
3. **没有 Key** → 分步引导注册火山引擎 + 获取 Key（同 /luoshu.setup Step 4.4）
4. **跳过** → 结束

### 已配置且连接正常时

```
当前 luoshu 配置：

  服务商:  火山引擎
  API Key: sk-****xxxx
  Endpoint: ark.cn-beijing.volces.com
  LLM Model: doubao-1.5-pro-256k
  Embedding Model: doubao-embedding-large
  状态:    已连接

  记忆策略:
  - 自动提取: 已启用
  - 保留天数: 90 天

输入编号修改：
  [1] 更新 API Key
  [2] 修改 LLM Model
  [3] 修改 Embedding Model
  [4] 修改记忆策略
  [5] 测试连接
  [6] 查看系统状态
  [7] 重建向量索引
```

各选项操作：
- **[1] 更新 API Key** → 要求输入新 Key → `luoshu_config_set`(key=`llm.api_key`) → `luoshu_config_validate`
- **[2] 修改 LLM Model** → 展示可选模型 → `luoshu_config_set`(key=`llm.model`)
- **[3] 修改 Embedding Model** → 展示可选模型 → `luoshu_config_set`(key=`embedding.model`)
- **[4] 修改记忆策略** → 展示当前值 → `luoshu_config_set`(key=`memory.auto_extract` 或 `memory.retention_days`)
- **[5] 测试连接** → `luoshu_config_validate`
- **[6] 查看系统状态** → `luoshu_status`（展示记忆条目数、向量索引数、缓存大小）
- **[7] 重建向量索引** → `luoshu_reindex`

### 已配置但连接失败时

```
当前 luoshu 配置：

  服务商:  火山引擎
  API Key: sk-****xxxx
  状态:    连接失败（认证被拒绝）

API Key 可能已过期或被撤销。

  [1] 更新 API Key
  [2] 帮我获取新的 Key
  [3] 测试连接（重试）
```

---

## Key 验证失败的诊断

当用户输入的 Key 验证失败时，根据 `luoshu_config_set` 返回的预检结果展示诊断：

| 预检结果 | 展示信息 |
|---------|---------|
| OpenAI Key (sk-proj-) | "这看起来是 OpenAI 的 Key，请确认使用火山引擎的 Key" |
| Anthropic Key (sk-ant-) | "这看起来是 Anthropic 的 Key" |
| AWS Key (AKIA) | "这看起来是 AWS 的 Key" |
| GitHub Token (ghp_) | "这看起来是 GitHub Token" |
| Key 过短 (<20字符) | "Key 只有 N 个字符，可能复制不完整" |

诊断后提供选项：
```
[1] 重新输入
[2] 帮我获取正确的 Key
[3] 先跳过
```

---

## 安全声明

配置完成后始终提示：
```
Key 保存在本地 ~/.luoshu/config.json
不会发送到 Anthropic 或其他外部服务。
```

---

## 关键原则

- 始终脱敏展示 API Key（只显示前 3 + 后 4 位）
- 连接测试失败时不恐慌，给出具体诊断和解决方案
- 修改后自动验证，确保配置有效
- 操作完成后展示配置全貌，满足可见性原则
