# Templates

Templates are pre-built configuration packs that install rules, agents, skills, and commands into your project or global scope.

## Listing Templates

Use the `template_list` tool to see all available templates grouped by category.

## Built-in Templates

### Core

| Template ID | Name | Description |
|------------|------|-------------|
| `hackathon-core` | Hackathon Core | Development toolkit with TDD, code review, build fix, and verification workflows |
| `cross-session-memory` | Cross-Session Memory | Automatic session state persistence and cross-session memory protocol |
| `codemap` | Code Map | Project structure overview and code navigation guide |
| `continuous-learning` | Continuous Learning | Extract and accumulate knowledge from sessions |
| `code-review-checkpoint` | Code Review Checkpoint | Checkpoint-based evaluation at implementation milestones |

### Teams

| Template ID | Name | Description |
|------------|------|-------------|
| `solo-full` | Solo Full Team | One-person company with 10 AI agent roles (CEO, CTO, Product, UI, Interaction, Fullstack, QA, Marketing, Operations, Sales) |

## Installing a Template

```
Tool: template_install
Parameters:
  template_id: "hackathon-core"
  scope: "project"           # or "global"
  project_path: "/path/to/project"
  overwrite: "false"         # set "true" to replace existing files
```

Each template installs:
- A rules file at `.claude/rules/tpl-{id}.md`
- Optionally: agents, skills, and commands files

## Checking Installed Templates

```
Tool: template_installed
Parameters:
  scope: "project"
  project_path: "/path/to/project"
```

Returns a list of installed templates detected by scanning for `tpl-*.md` files in the rules directory.

## Uninstalling a Template

```
Tool: template_uninstall
Parameters:
  template_id: "hackathon-core"
  scope: "project"
  project_path: "/path/to/project"
```

This removes the template's rules file. Agents, skills, and commands installed by the template are not automatically removed.

## Template Structure

Each template is defined in `internal/templatedata/` and contains:

```go
type Template struct {
    ID          string
    Name        string
    Category    string
    Description string
    Tags        []string
    Files       []TemplateFile
}

type TemplateFile struct {
    Type    string // "rules", "agents", "skills", "commands"
    Name    string // filename without .md
    Content string // file content
}
```

## Adding Custom Templates

To contribute a new template, see [Contributing](/contributing). Templates are defined in Go source files under `internal/templatedata/` and registered in `catalog.go`.
