---
name: luoshu.setup
description: Project initialization guide. Automatically detects project characteristics, completes configuration through 2-3 questions, and guides intelligent memory setup.
argument-hint: "[--reset reconfigure] [--upgrade check for updates]"
---

# /luoshu.setup -- Project Initialization & Upgrade

Automatically recommends configuration based on project characteristics. New projects go through the full initialization flow; already-configured projects go through the incremental upgrade flow.

---

## Mode Selection

Automatically selects the mode based on project status:

| Condition | Mode | Behavior |
|-----------|------|----------|
| No `.claude/` directory | **Fresh initialization** | Detect project → 3 questions → Install templates → LLM config → Write meta |
| Has `.claude/` + no `.setup-meta.json` | **Register existing config** | Scan existing config → Ask user preferences → LLM config → Write meta |
| Has `.setup-meta.json` or `--upgrade` flag | **Incremental upgrade** | Compare installed vs available → Show new features → Check LLM config |
| `--reset` flag passed | **Reconfigure** | Backup → Re-run full flow |

---

## Mode A: Fresh Initialization (no .claude/)

### Step 1. Detect Project Characteristics (automatic, zero interaction)

Scan the current project directory and collect the following information:
- **Language/Framework**: Check package.json / pyproject.toml / go.mod / Cargo.toml
- **Test Framework**: Jest/Vitest/pytest/go test, etc.
- **Package Manager**: npm/pnpm/yarn/pip/poetry/go
- **Git Status**: Whether initialized, recent commit authors (determine solo/team)

Present detection results concisely:

```
Quick scan results:
- Project type: Next.js 14 + TypeScript
- Test framework: Jest + React Testing Library
- Package manager: pnpm
- Git: Initialized, last 3 commits all from the same author
```

### Step 2. Quick Configuration (strictly limited to 3 questions)

Use the AskUserQuestion tool to ask 3 questions at once:

**Question 1: Development Style Preference**
- (a) Strict -- TDD, code review, strict type checking (recommended for production projects)
- (b) Fast -- Prioritize speed over process (recommended for prototypes/experiments/hackathons)
- (c) Balanced -- Strict on critical paths, flexible elsewhere

**Question 2: Team Size**
- (a) Solo development (mark [detected] if git log shows only one author)
- (b) Team collaboration

**Question 3: Cross-session Memory**
- (a) Enable (recommended) -- Claude remembers where you left off, your preferences, and project knowledge
- (b) Not needed

### Step 3. Configuration Mapping & Installation

Call MCP tools to install templates based on selections:

| Selection Combination | Installed Templates | Core Capabilities |
|----------------------|--------------------|--------------------|
| Strict + Solo + Memory | `hackathon-core` + `cross-session-memory` + `code-review-checkpoint` | TDD, code review, memory, build diagnostics |
| Strict + Team + Memory | `hackathon-core` + `cross-session-memory` + `code-review-checkpoint` | Same as above + architecture review, security review |
| Fast + Solo + No memory | `hackathon-core` | Basic validation, build diagnostics |
| Fast + Solo + Memory | `hackathon-core` + `cross-session-memory` | Basic validation + memory |
| Balanced + Any + Memory | `hackathon-core` + `cross-session-memory` + `code-review-checkpoint` | Planning, validation, code review, memory |
| Balanced + Any + No memory | `hackathon-core` + `code-review-checkpoint` | Planning, validation, code review |

Installation steps:
1. Call `template_install` to install each template (scope is the current project path)
2. If memory is enabled, call `save_memory` to create the initial `session-state.md` and `MEMORY.md`

### Step 4. Configure Intelligent Memory AI Engine (strongly recommended)

After template installation, guide the user to configure the luoshu intelligent memory LLM service.

#### 4.1 Check Current Configuration Status

Call `luoshu_config_get` to check if LLM is configured:
- **Configured and valid** → Call `luoshu_config_validate` to verify connection; skip silently if successful
- **Configured but invalid** → Guide key update
- **Not configured** → Enter configuration guide

#### 4.2 Configuration Guide Script

```
Project configuration has been installed.

Last step -- Configure the AI engine for intelligent memory.

The core capability of this product is cross-session memory:
- Automatically remembers your preferences and context
- Seamlessly resumes in the next conversation
- Search any previous content using natural language

This requires connecting to an AI service to power it.
Currently supports Volcengine Doubao (low latency and cost for users in China).

Do you have a Volcengine API Key?
```

#### 4.3 User Has a Key

After the user inputs a Key:
1. Call `luoshu_config_set` (key=`llm.api_key`, value=user input)
   - The tool automatically performs PreValidateKey format check
   - If recognized as a key from another platform (OpenAI/Anthropic, etc.), display specific diagnostics
2. Call `luoshu_config_validate` to verify the connection
3. On success → Display configuration info (Key masked) + security statement

```
Connection successful. Configuration:
- Provider: Volcengine
- API Key: sk-****xxxx
- Model: doubao-1.5-pro-256k (default)

Key is saved locally at ~/.luoshu/config.json
It will not be sent to Anthropic or any other external service.
```

#### 4.4 User Does Not Have a Key

Step-by-step guide to obtain one:

**Step 1: Registration**
```
No problem, I'll guide you step by step. It takes about 3 minutes.

Volcengine is ByteDance's cloud service platform.

Please visit in your browser:
https://console.volcengine.com/auth/signup

Let me know when you're done, and I'll guide you to the next step.
```

**Step 2: Obtain Key**
```
Visit the API Key management page:
https://console.volcengine.com/ark/region:ark+cn-beijing/apiKey

Click "Create API Key" and copy the created Key (starts with sk-).

Note: You won't be able to see the full Key after closing the page, so make sure you've copied it.
```

After the user pastes the Key, follow the same verification flow as 4.3.

#### 4.5 User Chooses to Skip

```
After skipping LLM configuration, the following core features will be unavailable:
- Cross-session intelligent memory
- Semantic search
- Automatic session highlights extraction

Basic configuration management features are not affected.

[1] Skip for now, use basic features
[2] I'll get a Key (about 3 minutes, I'll guide you)
[3] I already have a Key
```

Use AskUserQuestion to display options. If the user selects [1], mark as skipped and continue to Step 5.

### Step 5. Write setup-meta.json

After installation, use the Write tool to create `.claude/.setup-meta.json`:

```json
{
  "setup_version": "0.6.0",
  "setup_date": "YYYY-MM-DD",
  "style": "strict|fast|balanced",
  "team": "solo|team",
  "memory": true|false,
  "luoshu_configured": true|false,
  "installed_templates": ["hackathon-core", "cross-session-memory"],
  "installed_rules": ["tpl-hackathon-core.md"],
  "installed_agents": ["architect.md"],
  "installed_commands": ["plan.md", "tdd.md"],
  "installed_skills": ["luoshu.setup", "luoshu.config"]
}
```

### Step 6. Confirmation & Guidance

Display results in two layers:

**Enabled** (takes effect automatically):
- Describe in user-friendly language, e.g., "Code quality guardian" instead of "code-reviewer agent"
- Display differently based on luoshu configuration status

**When luoshu is configured:**
```
Setup complete! The following capabilities are configured for this project:

[Enabled] Configuration management, template system, intelligent memory

Try saying "What do you remember about me" or "Help me continue where I left off".
```

**When luoshu is not configured:**
```
Setup complete. Current status:

[Enabled] Configuration management, template system
[Not enabled] Intelligent memory (requires LLM configuration)

Run /luoshu.config anytime to configure it.
You can start coding now. What would you like to do?
```

---

## Mode B: Register Existing Configuration (has .claude/ but no meta)

### 1. Scan Existing Configuration

Call MCP tools to get the current state:
- `template_installed` — Check which templates are installed
- `extension_list(type="rules")` — Check which rules exist
- `extension_list(type="agents")` — Check which agents exist
- `extension_list(type="commands")` — Check which commands exist
- `hooks_list` — Check which hooks exist

### 2. Present Current State + Quick Supplement

Ask only one question (skip if it can be inferred from existing configuration).

### 3. Check LLM Configuration

Call `luoshu_config_get` to check status. If not configured, follow the Step 4 configuration guide (same as Mode A).

### 4. Write Meta File

Write `.claude/.setup-meta.json` based on scan results without modifying any existing configuration.

---

## Mode C: Incremental Upgrade (has meta or --upgrade)

### 1. Read Meta + Compare

Read `.claude/.setup-meta.json`, compare installed vs available templates, and display new features.

### 2. Incremental Installation

Install new features selected by the user one by one. **Do not modify or overwrite** existing configuration.

### 3. Check LLM Configuration

If `luoshu_configured` in meta is false or the field does not exist:

```
Note: You have not configured an LLM service yet.
Intelligent memory is the core feature of this product and requires LLM to power it.

Configure now? (about 3 minutes)

[1] Configure now (recommended)
[2] Later
```

If already configured, call `luoshu_config_validate` for silent verification; guide key update if expired.

### 4. Update Meta File

Merge newly installed content into meta, update `setup_date` and `setup_version`.

---

## Mode D: Reconfigure (--reset)

Read existing meta to understand user preferences, then re-run the full Mode A flow. Use overwrite=true during installation.

---

## Key Principles

- **Detect instead of asking**: Don't ask the user for things that can be auto-detected
- **Incremental, no overwrite**: Upgrade mode only adds new content, never modifies existing configuration
- **Defaults instead of choices**: Every option has a recommended default
- **Result-oriented descriptions**: Use "Code quality guardian" instead of "code-reviewer agent"
- **3 questions is the maximum**: 3 for fresh initialization, 1 for registration, 0 for upgrade
- **LLM configuration is strongly recommended but skippable**: Ensure users can complete setup in any case
- **Meta file is critical**: Must update `.claude/.setup-meta.json` after every operation
