package templatedata

// GetBuiltinTemplates returns the list of built-in template categories
func GetBuiltinTemplates() []TemplateCategory {
	return []TemplateCategory{
		{
			ID: "general", Name: "General", Icon: "📦",
			Templates: []Template{
				{
					ID: "minimal", Name: "Minimal Config", Category: "General",
					Description: "Minimal CLAUDE.md configuration for a quick start",
					Tags:        []string{"Getting Started", "Simple"},
					ClaudeMd: `# Project Standards

## Code Standards
- Follow the existing code style and patterns in the project
- Keep code concise and avoid over-engineering
- Every change should pass compilation and tests

## Development Workflow
1. Understand the requirements before writing code
2. Iterate in small steps, improve incrementally
3. Study existing code before making changes

## Language
All responses in the user's preferred language.
`,
				},
				{
					ID: "standard", Name: "Standard Config", Category: "General",
					Description: "Standard configuration with common rules and best practices",
					Tags:        []string{"Recommended", "Standard"},
					ClaudeMd: `# Project Standards

## Development Philosophy
- **Incremental Development**: Small commits that always compile and pass tests
- **Learn from Existing Code**: Research and plan before implementing
- **Pragmatic over Dogmatic**: Adapt to the project's actual needs
- **Clear Intent over Clever Code**: Choose simple, straightforward solutions

## Implementation Process
1. **Understand Existing Patterns**: Study 3 similar features in the codebase
2. **Identify Common Patterns**: Discover project conventions
3. **Follow Existing Standards**: Use the same libraries/tools
4. **Implement in Phases**: Break complex work into 3-5 phases

## Quality Standards
- Every commit must compile successfully
- All existing tests must pass
- New features must include tests

## Decision Framework
1. Testability - Is it easy to test?
2. Readability - Will it still make sense in 6 months?
3. Consistency - Does it follow project patterns?
4. Simplicity - Is it the simplest viable solution?

## Error Handling
- Stop after 3 failed attempts at most
- Document failure reasons and specific error messages
- Research 2-3 alternative approaches

## Language
All responses in the user's preferred language.
`,
				},
			},
		},
		{
			ID: "frontend", Name: "Frontend", Icon: "🎨",
			Templates: []Template{
				{
					ID: "react", Name: "React / Next.js", Category: "Frontend",
					Description: "Best practices configuration for the React ecosystem",
					Tags:        []string{"React", "Next.js", "TypeScript"},
					ClaudeMd: `# React Project Standards

## Tech Stack
- React 19+ / Next.js 15+
- TypeScript strict mode
- Tailwind CSS

## Code Standards
- Use functional components + Hooks
- Prefer React built-in state management (useState, useReducer, Context)
- Use PascalCase for component files and camelCase for utility functions
- Use Tailwind utility classes, avoid custom CSS

## Component Standards
- One component per file
- Define Props with interface, not type
- Use React.FC or plain function declarations
- Avoid defining child components inside parent components

## Testing
- Use Vitest + React Testing Library
- Test user behavior, not implementation details
- At least one test file per component

## Performance
- Use React.memo to avoid unnecessary re-renders
- Use virtual scrolling for large lists
- Use next/image for image optimization

## Language
All responses in the user's preferred language.
`,
				},
				{
					ID: "vue", Name: "Vue / Nuxt", Category: "Frontend",
					Description: "Best practices configuration for the Vue 3 ecosystem",
					Tags:        []string{"Vue", "Nuxt", "TypeScript"},
					ClaudeMd: `# Vue Project Standards

## Tech Stack
- Vue 3 Composition API
- TypeScript strict mode
- Pinia state management

## Code Standards
- Use <script setup lang="ts"> syntax
- Use PascalCase for component names
- Use camelCase for utility function names
- Composables must start with use prefix

## Component Standards
- Single File Components (SFC)
- Define Props using defineProps<T>() generic syntax
- Define Emits using defineEmits<T>()
- Avoid using Options API

## Testing
- Use Vitest + Vue Test Utils
- Test component rendering output and interaction behavior

## Language
All responses in the user's preferred language.
`,
				},
			},
		},
		{
			ID: "backend", Name: "Backend", Icon: "⚡",
			Templates: []Template{
				{
					ID: "golang", Name: "Go Project", Category: "Backend",
					Description: "Standard configuration for Go backend projects",
					Tags:        []string{"Go", "Gin", "API"},
					ClaudeMd: `# Go Project Standards

## Code Standards
- Follow Effective Go and Go Code Review Comments
- Format code with gofmt/goimports
- Never use panic for error handling, always return error
- Define interfaces at the call site, not the implementation

## Project Structure
- cmd/ - Application entry points
- internal/ - Internal packages (not exported)
- pkg/ - Reusable public packages
- api/ - API definitions (proto, OpenAPI)

## Testing
- Use table-driven tests
- Mock using interface + hand-written implementations
- Test files in the same directory as source files

## Error Handling
- Wrap errors with fmt.Errorf("context: %w", err)
- Define domain-specific error types
- Unified error response format at the API layer

## Language
All responses in the user's preferred language.
`,
				},
				{
					ID: "python-fastapi", Name: "Python FastAPI", Category: "Backend",
					Description: "Configuration for Python FastAPI projects",
					Tags:        []string{"Python", "FastAPI", "async"},
					ClaudeMd: `# Python FastAPI Project Standards

## Tech Stack
- Python 3.12+
- FastAPI + Pydantic V2
- SQLAlchemy 2.0 async
- uv package manager

## Code Standards
- Type annotations: all function parameters and return values
- Use Pydantic models for data validation
- Async-first: use async/await
- Follow PEP 8, format with ruff

## Project Structure
- app/api/ - Routes
- app/models/ - Data models
- app/schemas/ - Pydantic schemas
- app/services/ - Business logic
- tests/ - Tests

## Testing
- Use pytest + httpx AsyncClient
- Test coverage > 80%

## Language
All responses in the user's preferred language.
`,
				},
				{
					ID: "node-express", Name: "Node.js Express", Category: "Backend",
					Description: "Configuration for Node.js Express/Fastify projects",
					Tags:        []string{"Node.js", "Express", "TypeScript"},
					ClaudeMd: `# Node.js Project Standards

## Tech Stack
- Node.js 22+
- TypeScript strict mode
- Express/Fastify
- Prisma ORM

## Code Standards
- Use ESM (import/export)
- Handle async errors with try/catch
- Use middleware pattern for cross-cutting concerns
- Validate environment variables with dotenv + zod

## Testing
- Use Vitest
- Integration tests with supertest
- Database tests with test containers

## Language
All responses in the user's preferred language.
`,
				},
			},
		},
		{
			ID: "best-practices", Name: "Best Practices", Icon: "🏆",
			Templates: []Template{
				{
					ID: "cross-session-memory", Name: "Cross-Session Memory", Category: "Best Practices",
					Description: "Enhanced cross-session memory: structured state persistence, automatic recovery, and anti-forgetting mechanisms",
					Tags:        []string{"Memory", "Hooks", "Anti-Forgetting", "Recommended"},
					ClaudeMd: `# Cross-Session Memory System

## Core Mechanism

This project has cross-session memory enabled. Your memory is stored in the ` + "`" + `.claude/memory/` + "`" + ` directory. **You must strictly follow the rules below.**

## Session Start (Required)

The **very first thing** in every new session, before responding to the user:

1. Read ` + "`" + `.claude/memory/MEMORY.md` + "`" + ` (main memory index)
2. Read ` + "`" + `.claude/memory/session-state.md` + "`" + ` (previous session state)
3. If session-state.md has unfinished tasks, proactively inform the user: "There is unfinished work from last session: XXX. Would you like to continue?"

## Before Session Ends (Required)

**Save session state immediately** when any of the following occurs:

- The user says goodbye or signals the end of the session
- You sense the context window is nearly full (slower responses, truncated content)
- A significant milestone has been completed
- A complex issue arises that needs to be continued next time

### Save to .claude/memory/session-state.md

Use the following fixed format:

` + "```" + `markdown
# Session State

## Last Updated
[ISO timestamp]

## Current Task
[One-line description of what is being worked on]

## Completed
- [x] Step 1 (brief result)
- [x] Step 2 (brief result)

## Incomplete
- [ ] Step 3 (where it got stuck / what to do next)
- [ ] Step 4

## Key Decisions
- **Chose Option A**: because...
- **Rejected Option B**: because...

## Issues Encountered
- Problem description -> Solution (or unresolved)

## Important Files
- ` + "`" + `path/to/file1` + "`" + ` - What was modified
- ` + "`" + `path/to/file2` + "`" + ` - What was modified

## Next Actions
1. First thing to do
2. Second thing to do
` + "```" + `

### Save to .claude/memory/MEMORY.md

Only update information with **long-term value** (not every session):
- Key architectural decisions
- Recurring problems and their solutions
- Important technical constraints and caveats
- Keep it concise, no more than 200 lines

## During Context Compaction (Required)

When the system performs a compact, you may lose previous conversation details.
If session-state.md has not been updated at that point, **update it immediately**.

## Language
All responses in the user's preferred language.
`,
					Rules: map[string]string{
						"memory-protocol": `# Cross-Session Memory Protocol (Mandatory)

## Session Startup Checks

At the start of every session, you must perform the following checks:

1. Check if the .claude/memory/ directory exists
2. If it does not exist, create the directory and initialize MEMORY.md and session-state.md
3. If it exists, read all .md files
4. Pay special attention to "Incomplete" and "Next Actions" in session-state.md
5. Briefly report the previous session's status to the user (if any)

## Memory Save Triggers

You must immediately save session-state.md in the following situations:

- After completing a feature or fixing a bug
- After making an important technical decision
- After more than 10 consecutive conversation turns
- When the user requests a save or signals the end of the session
- When the context window feels nearly full (don't wait until it's too late)

## Memory File Structure

Files in the ` + "`" + `.claude/memory/` + "`" + ` directory:

| File | Purpose | Update Frequency |
|------|---------|-----------------|
| MEMORY.md | Long-term knowledge (architecture, gotchas, conventions) | When new persistent knowledge is gained |
| session-state.md | Session state (progress, decisions, next steps) | Must update at the end of every session |
| decisions.md | Major decision records (optional) | When major decisions are made |

## Prohibited Actions

- Never start working without reading the memory files first
- Never end a session without saving session-state.md
- Never delete or clear existing memory files
- Never write temporary debug information to MEMORY.md
`,
					},
					Settings: map[string]interface{}{
						"hooks": map[string]interface{}{
							"PreCompact": []map[string]interface{}{
								{
									"hooks": []map[string]interface{}{
										{
											"type":    "command",
											"command": "echo '[auto-save] Context is about to be compacted. Please save session state to .claude/memory/session-state.md immediately, and consider calling memory_extract to capture key memories.'",
										},
									},
								},
							},
							"Stop": []map[string]interface{}{
								{
									"hooks": []map[string]interface{}{
										{
											"type":    "command",
											"command": "echo 'Please confirm that session state has been saved to .claude/memory/session-state.md'",
										},
									},
								},
							},
						},
					},
				},
				{
					ID: "continuous-learning", Name: "Continuous Learning", Category: "Best Practices",
					Description: "Enable Claude Code to continuously learn and update its knowledge base during work",
					Tags:        []string{"Learning", "Knowledge Base", "Advanced"},
					ClaudeMd: `# Project Standards

## Continuous Learning
- When encountering new patterns or solutions, record them in the .claude/learnings/ directory
- Each learning record includes: problem description, solution, and why it works
- Periodically review learning records and update outdated content

## /learn Command
When the user uses the /learn command:
1. Summarize the key learning points from the current session
2. Append the learnings to .claude/learnings/log.md
3. Confirm the save

## Code Standards
- Follow the project's existing code style
- Maintain consistency
- Understand before modifying

## Language
All responses in the user's preferred language.
`,
				},
				{
					ID: "code-review-checkpoint", Name: "Checkpoint Review", Category: "Best Practices",
					Description: "Checkpoint-based code quality evaluation workflow",
					Tags:        []string{"Quality", "Review", "Checkpoint"},
					ClaudeMd: `# Project Standards

## Checkpoint Review Workflow
Perform an evaluation at each significant modification point:

### Checkpoint 1: Before Implementation
- [ ] Requirements are understood
- [ ] Similar existing implementations have been studied
- [ ] Implementation approach is determined

### Checkpoint 2: After Implementation
- [ ] Code compiles successfully
- [ ] No new lint warnings introduced
- [ ] Core logic has test coverage

### Checkpoint 3: Before Commit
- [ ] All tests pass
- [ ] No lingering TODO or FIXME items
- [ ] Code readability is good
- [ ] No redundant code

## Dead Code Cleanup
Periodically check and clean up:
- Unused imports
- Unused variables and functions
- Duplicate code blocks
- Outdated comments

## Language
All responses in the user's preferred language.
`,
				},
				{
					ID: "codemap", Name: "Code Map", Category: "Best Practices",
					Description: "Use a code map to help the AI quickly understand the project layout",
					Tags:        []string{"Code Map", "Entry Point", "Index"},
					ClaudeMd: `# Project Standards

## Code Map
Read this file first for a high-level overview of the project before diving into specific files.

### Project Overview
[Describe the core functionality and goals here]

### Directory Structure
[Describe key directories and their purposes here]

### Core Modules
[Describe core modules and their responsibilities here]

### Data Flow
[Describe the main data flow here]

### Key Files
[List the most important files and their roles here]

## Development Guidelines
- Update this code map after modifying code
- Add new modules to the corresponding section
- Keep the code map in sync with the actual codebase

## Language
All responses in the user's preferred language.
`,
				},
				{
					ID: "context-fields", Name: "Context Fields", Category: "Best Practices",
					Description: "Automatically inject project context (git branch, recent changes, project structure) at session start so Claude understands the project state without extra questions",
					Tags:        []string{"Context", "Hooks", "Automation"},
					Scripts: map[string]string{
						"inject-context.sh": injectContextScript,
					},
					Settings: map[string]interface{}{
						"hooks": map[string]interface{}{
							"SessionStart": []map[string]interface{}{
								{
									"hooks": []map[string]interface{}{
										{
											"type":    "command",
											"command": "bash .claude/scripts/inject-context.sh",
										},
									},
								},
							},
						},
					},
					Rules: map[string]string{
						"tpl-context-fields": contextFieldsRule,
					},
					Skills: map[string]string{
						"context-refresh": contextRefreshSkill,
					},
				},
			},
		},
	}
}

const injectContextScript = `#!/bin/bash
# inject-context.sh — Automatically inject project context at session start
# Installed by the context-fields template, feel free to customize

echo "[context-fields] Project Context"
echo ""

# Git info
if command -v git &>/dev/null && git rev-parse --is-inside-work-tree &>/dev/null 2>&1; then
    BRANCH=$(git branch --show-current 2>/dev/null || echo "detached")
    echo "## Git"
    echo "- Branch: $BRANCH"

    # Last 5 commits
    COMMITS=$(git log --oneline -5 2>/dev/null)
    if [ -n "$COMMITS" ]; then
        echo "- Recent commits:"
        echo "$COMMITS" | while read -r line; do echo "  - $line"; done
    fi

    # Uncommitted changes
    CHANGES=$(git status --short 2>/dev/null)
    if [ -n "$CHANGES" ]; then
        COUNT=$(echo "$CHANGES" | wc -l | tr -d ' ')
        echo "- Uncommitted changes: ${COUNT} file(s)"
    fi
    echo ""
fi

# Project type detection
echo "## Project"
if [ -f "package.json" ]; then
    NAME=$(grep -o '"name"[[:space:]]*:[[:space:]]*"[^"]*"' package.json | head -1 | cut -d'"' -f4)
    echo "- Type: Node.js ($NAME)"
elif [ -f "go.mod" ]; then
    MOD=$(head -1 go.mod | awk '{print $2}')
    echo "- Type: Go ($MOD)"
elif [ -f "pyproject.toml" ]; then
    echo "- Type: Python"
elif [ -f "Cargo.toml" ]; then
    echo "- Type: Rust"
else
    echo "- Type: Unknown"
fi

# Top-level structure
DIRS=$(ls -d */ 2>/dev/null | head -10 | tr '\n' ' ')
if [ -n "$DIRS" ]; then
    echo "- Directories: $DIRS"
fi
`

const contextFieldsRule = `<!-- template: context-fields | Context Fields -->

# Context Fields

## Mechanism

On every session start, the SessionStart hook automatically runs ` + "`" + `.claude/scripts/inject-context.sh` + "`" + `, injecting the following project information into the context:

- **Git Status**: Current branch, last 5 commits, uncommitted changes
- **Project Structure**: Top-level directory layout
- **Dependency Info**: Extracted from package.json / go.mod / pyproject.toml etc.

## Usage

- Automatic injection: Takes effect automatically on every new session, no manual action needed
- Manual refresh: Use the ` + "`" + `/context-refresh` + "`" + ` command to re-inject the latest context

## Customization

Edit ` + "`" + `.claude/scripts/inject-context.sh` + "`" + ` to add the context fields you need. Anything the script outputs to stdout will be injected into the session context.

## Language
All responses in the user's preferred language.
`

const contextRefreshSkill = `---
name: context-refresh
description: Manually refresh project context (git status, project structure, dependency info)
---

# /context-refresh — Refresh Project Context

Manually re-collect and display the current project's context information.

## Steps

1. Run ` + "`" + `bash .claude/scripts/inject-context.sh` + "`" + ` (using the Bash tool)
2. Present the output to the user
3. Briefly summarize the current project state

## Output Format

Present the script output concisely, highlighting:
- Current branch and recent changes
- Uncommitted modifications
- Project type and key dependencies
`
