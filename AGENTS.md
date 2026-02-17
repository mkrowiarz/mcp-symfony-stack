# AI Agent Guide for Haive Configuration

This document helps AI agents understand and configure haive projects.

## Project Overview

**Haive** is a development environment manager for Docker Compose-based projects. It provides:
- Git worktree management with per-worktree databases
- Database operations (clone, drop, list)
- Docker container management per worktree
- TUI, CLI, and MCP interfaces

## Configuration File Structure

The main configuration file is `.haive.toml` in the project root.

### Minimal Configuration

```toml
[docker]
compose_files = ["docker-compose.yml"]
```

### Full Configuration Example

```toml
# Docker configuration (required)
[docker]
compose_files = ["compose.yaml", "docker/dev/compose.database.yaml", "docker/dev/compose.app.yaml"]
project_name = "myapp"  # Optional: prefix for docker compose projects

# Database configuration (optional, but required for DB operations)
[database]
service = "database"  # Docker compose service name for DB
dsn = "${DATABASE_URL}"  # Supports env var interpolation
allowed = ["myapp", "myapp_*"]  # Glob patterns for allowed DB names
dumps_path = "var/dumps"  # Optional: default is "var/dumps"

# Worktree configuration (optional, but required for worktree operations)
[worktree]
base_path = ".worktrees"  # Directory for worktrees
db_per_worktree = true    # Auto-create database per worktree

# Files to copy when creating worktree
[worktree.copy]
include = [".env.local", "config/**/*.yaml", "docker/**/*.yaml"]
exclude = ["vendor/", "node_modules/", ".git/"]

# Hooks run during worktree lifecycle
[worktree.hooks]
postCreate = ["composer install", "npm install"]
preRemove = ["./scripts/cleanup.sh"]

# Per-worktree environment configuration
[worktree.env]
file = ".env.local"
var_name = "DATABASE_URL"

# Serve configuration for worktree containers
[serve]
compose_files = ["compose.yaml", "docker/dev/compose.database.yaml", "docker/dev/compose.app.yaml"]

# Worktree-specific serve configuration (optional)
[serve.worktree]
compose_files = ["docker/dev/compose.app.yaml"]

# Database hooks
[database.hooks]
postClone = ["./scripts/seed.sh"]
preDrop = ["./scripts/backup.sh"]
```

## Docker Compose Organization Patterns

### Pattern 1: Monolithic Compose (Simple Projects)

```
project-root/
├── .haive.toml
├── docker-compose.yml          # Everything in one file
└── src/
```

```toml
[docker]
compose_files = ["docker-compose.yml"]

[serve]
compose_files = ["docker-compose.yml"]
```

### Pattern 2: Modular Compose (Recommended)

```
project-root/
├── .haive.toml
├── compose.yaml                # Base services (db, cache)
├── docker/
│   └── dev/
│       ├── compose.database.yaml   # Database overrides
│       ├── compose.app.yaml        # App service
│       └── compose.worktree.yaml   # Worktree-specific overrides
└── src/
```

```toml
[docker]
compose_files = ["compose.yaml", "docker/dev/compose.database.yaml"]

[serve]
compose_files = ["compose.yaml", "docker/dev/compose.database.yaml", "docker/dev/compose.app.yaml"]

[serve.worktree]
compose_files = ["docker/dev/compose.app.yaml"]
```

**Key principle:** Worktrees inherit the network from the main project, so they can connect to the main database.

### Pattern 3: Full Separation (Worktrees Run Everything)

```
project-root/
├── .haive.toml
├── compose.yaml                # Define networks and volumes
├── docker/
│   └── dev/
│       ├── compose.db.yaml     # Database service
│       └── compose.app.yaml    # App service
└── src/
```

```toml
[docker]
compose_files = ["compose.yaml", "docker/dev/compose.db.yaml"]

[serve]
compose_files = ["compose.yaml", "docker/dev/compose.db.yaml", "docker/dev/compose.app.yaml"]

# Worktrees also run full stack (no [serve.worktree] override)
```

## Environment Variable Resolution

Haive resolves `${VAR_NAME}` syntax in this order:
1. OS environment variables
2. `.env.local` file
3. `.env` file

Common pattern:
```toml
[database]
dsn = "${DATABASE_URL}"
```

## Worktree Flow

1. **Main project** runs full stack (database + app)
2. **Worktrees** created via `haive worktree create <branch>`
3. **Files copied** based on `[worktree.copy]` patterns
4. **Environment updated** with worktree-specific DB URL
5. **Serve command** starts only app service in worktree

## Common Configuration Tasks

### Task: Add Database Support

1. Check for existing docker-compose with database service
2. Add `[database]` section to `.haive.toml`
3. Ensure `allowed` patterns include main DB and worktree DBs
4. Verify DATABASE_URL in `.env.local`

### Task: Setup Worktrees with Shared Database

1. Configure `[worktree]` section with `base_path`
2. Setup `[worktree.copy]` to include necessary files
3. Configure `[serve]` with full compose files
4. Add `[serve.worktree]` with only app compose file
5. Ensure main project compose defines an external network

Example compose for main project:
```yaml
# compose.yaml
services:
  database:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root
    networks:
      - local

networks:
  local:
    name: ${PROJECT_NAME:-app}_local  # Named network
```

Example compose for worktree app:
```yaml
# docker/dev/compose.app.yaml
services:
  app:
    build: .
    environment:
      DATABASE_URL: "mysql://root:root@database:3306/myapp_wt_branch"
    networks:
      - local

networks:
  local:
    external: true
    name: ${PROJECT_NAME:-app}_local  # Connect to main project's network
```

### Task: Configure Database Hooks

Add to `.haive.toml`:
```toml
[database.hooks]
postClone = ["./scripts/seed-after-clone.sh"]
preDrop = ["./scripts/backup-before-drop.sh"]
```

Hooks receive environment variables:
- `REPO_ROOT` - Path to main repository
- `DATABASE_NAME` - Database being operated on
- `DATABASE_URL` - Full connection URL
- `SOURCE_DATABASE` - For clone operations
- `TARGET_DATABASE` - For clone operations

## Troubleshooting

### "[serve] section not configured"
Add `[serve]` with `compose_files` to `.haive.toml`

### "not in a worktree directory"
`haive serve` only works inside worktrees, not the main project.

### Database connection fails from worktree
1. Check that main project defines a named network
2. Ensure worktree compose uses `external: true` network
3. Verify network name matches between main and worktree

### Worktree files not copied
Check `[worktree.copy]` patterns:
- `include` uses glob patterns with `**` for recursive
- `exclude` patterns skip files even if matched by include

## Quick Reference: Configuration Validation

Run these commands to validate configuration:

```bash
# Check config is valid
haive init

# List worktrees
haive worktree list

# Test database operations
haive checkout feature/test --create

# Test serve (from inside a worktree)
cd .worktrees/feature-test
haive serve
```

## Migration from Other Tools

### From plain docker-compose:
1. Create `.haive.toml` with `[docker]` section
2. Add `[database]` if using database features
3. Add `[worktree]` if using worktree features
4. Organize compose files if needed for worktree support

### From Laravel Sail:
1. Extract compose files from Sail
2. Configure `[docker]` with compose files
3. Update `docker` service references to match Sail services
4. Configure `[database]` with Sail's DB service name

## AI Agent Checklist

When helping configure haive:

- [ ] Identify project type and existing docker-compose setup
- [ ] Check if database is used and what service name it has
- [ ] Determine if worktrees need shared or separate databases
- [ ] Organize compose files if worktree support is needed
- [ ] Configure `[docker]` with correct compose files
- [ ] Configure `[database]` with service name and DSN pattern
- [ ] Set up `[worktree]` with appropriate file copy patterns
- [ ] Configure `[serve]` and `[serve.worktree]` for container management
- [ ] Add hooks if custom scripts are needed
- [ ] Test configuration with `haive init` and `haive worktree list`
