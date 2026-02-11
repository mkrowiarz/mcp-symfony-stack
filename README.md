# mcp-symfony-stack

A standalone, reusable tool for managing Docker Compose-based development projects. Provides both an interactive TUI (Terminal UI) and an MCP server for Claude Code, enabling database operations and git worktree management.

## Purpose

- **TUI Mode**: Interactive terminal interface for database dumps/imports, worktree creation/removal, and project status
- **MCP Mode**: Stdio-based MCP server for Claude Code to manage infrastructure through `.claude/project.json`
- **CLI Mode**: One-shot commands for scripting and automation

Initially targeting Symfony (7/8, PHP 8.3+) but designed to be framework-agnostic where possible.

## Main Assumptions

- **Docker Compose is the runtime** — all database interactions happen via `docker compose exec`
- **Config-driven** — project-specific knowledge lives in `.claude/project.json`; the tool is stateless and generic
- **Env var resolution** — config values can reference `.env`/`.env.local` variables using `${VAR_NAME}` syntax
- **Safety by default** — database operations restricted to an explicit allowlist; default database cannot be dropped
- **Refuse without config** — any infrastructure operation requires valid project configuration

## Technology Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Language | **Go** | Single binary, fast startup, excellent exec/process handling |
| TUI | **Bubble Tea** + **Lip Gloss** | Industry standard for Go TUIs (lazygit, lazydocker) |
| MCP | **mcp-go** (`mark3labs/mcp-go`) | Go MCP SDK with stdio transport support |
| Config | **JSON + JSON Schema** | Native parsing, editor autocompletion via `$schema` |
| CLI | **cobra** or bare `os.Args` | One-shot commands for scripting |

## Phase 1: Core Library (Project Commands)

Current implementation supports project information and initialization commands.

### Workflow Commands

**`workflow.create_isolated_worktree`**: Create isolated worktree for feature work

```go
package main

import (
	"fmt"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/commands"
)

func main() {
	result, err := commands.CreateIsolatedWorktree(".", "feature/abc", "true", "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Worktree: %s (branch: %s)\n", result.WorktreePath, result.WorktreeBranch)
}
```

### Project Commands

**`project.info`**: Get project configuration and status

```go
package main

import (
    "fmt"
    "github.com/mkrowiarz/mcp-symfony-stack/internal/core/commands"
)

func main() {
    info, err := commands.Info(".")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Project: %s (%s)\n", info.ConfigSummary.Name, info.ConfigSummary.Type)
    fmt.Printf("Docker Compose: %v\n", info.DockerComposeExists)
    fmt.Printf("Env files: %v\n", info.EnvFiles)
}
```

**`project.init`**: Generate suggested project configuration

```go
package main

import (
    "fmt"
    "github.com/mkrowiarz/mcp-symfony-stack/internal/core/commands"
)

func main() {
    suggestion, err := commands.Init(".")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Println("Suggested configuration:")
    fmt.Println(suggestion.SuggestedConfig)
    fmt.Printf("\nDetected services: %v\n", suggestion.DetectedServices)
    fmt.Printf("Detected env vars: %v\n", suggestion.DetectedEnvVars)
}
```

### Configuration File

Project configuration lives in `.claude/project.json`:

```json
{
  "project": {
    "name": "your-project",
    "type": "symfony"
  },
  "docker": {
    "compose_file": "docker-compose.yaml"
  }
}
```

*Note: Phase 1 supports minimal configuration only. Database and worktrees sections will be added in phase 2.*

## Phase 2A: Worktree Commands

### Low-Level Commands

**`worktree.list`**: List all git worktrees

```go
import "fmt"
import "github.com/mkrowiarz/mcp-symfony-stack/internal/core/commands"

func main() {
    worktrees, _ := commands.worktree.List(".")
    for _, wt := range worktrees {
        marker := " "
        if wt.IsMain {
            marker = "*"
        }
        fmt.Printf("%s %s: %s (%s)\n", marker, wt.Branch, wt.Path, "main", marker)
    }
}
```

**`worktree.create`**: Create a git worktree

```go
result, _ := commands.worktree.Create(".", "feature/my-feature", true, "")
fmt.Printf("Created worktree: %s\n", result.Path)
```

**`worktree.remove`**: Remove a git worktree

```go
_, _ := commands.worktree.Remove(".", "feature/my-feature")
fmt.Println("Worktree removed")
```

### High-Level Orchestrator

**`workflow.create_isolated_worktree`**: Quick one-click workflow

```go
result, _ := workflow.CreateIsolatedWorktree(".", "feature/abc", true, "")
fmt.Printf("Worktree: %s (branch: %s)\n", result.WorktreePath, result.WorktreeBranch)
```

**Notes for Phase 2B**:
- Phase 2A implements worktree commands only
- Database integration (clone/drop) will be added in Phase 2B
- Orchestrator provides one-click workflows for common cases
- Low-level commands remain available for granular control

## Phase 2B: Database Commands

Essential database operations for worktree integration.

### Database Commands

**`db.dump`**: Dump database to SQL file

```go
import (
    "fmt"
    "github.com/mkrowiarz/mcp-symfony-stack/internal/core/commands"
)

func main() {
    result, err := commands.Dump(".", "app_db", nil)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Dumped %s (%d bytes) to %s\n", result.Database, result.Size, result.Path)
}
```

**`db.create`**: Create empty database

```go
_, err := commands.CreateDB(".", "new_db")
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}
fmt.Println("Database created")
```

**`db.import`**: Import SQL file into database

```go
result, err := commands.ImportDB(".", "app_db", "var/dumps/app_db.sql")
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}
fmt.Printf("Imported %s into %s\n", result.Path, result.Database)
```

**`db.drop`**: Drop database (protected: can't drop default)

```go
_, err := commands.DropDB(".", "old_db")
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}
fmt.Println("Database dropped")
```

### Configuration

Database operations require the `database` section in `.claude/project.json`:

```json
{
  "project": {
    "name": "your-project",
    "type": "symfony"
  },
  "docker": {
    "compose_file": "docker-compose.yaml"
  },
  "database": {
    "service": "database",
    "dsn": "mysql://root:${DATABASE_PASSWORD}@database:3306/app",
    "allowed": ["app", "app_*"],
    "dumps_path": "var/dumps"
  }
}
```

**Configuration fields:**
- `database.service`: Docker Compose service name for database
- `database.dsn`: Database URL (supports `mysql://`, `postgresql://`, env var interpolation)
- `database.allowed`: Glob patterns for allowed database names (e.g., `["app", "app_*"]`)
- `database.dumps_path`: Directory for SQL dumps (default: `var/dumps`)

**Safety guards:**
- Only databases matching `allowed` patterns can be operated on
- The default database (from DSN) cannot be dropped
- Database config is required for all database operations

**Supported engines:**
- MySQL (default port: 3306)
- MariaDB (detected via `serverVersion` query param)
- PostgreSQL (planned, default port: 5432)

### Notes

- Phase 2B implements essential database operations only
- `db.list`, `db.clone`, `db.dumps` will be added in future phases
- MySQL/MariaDB fully supported, PostgreSQL support planned
