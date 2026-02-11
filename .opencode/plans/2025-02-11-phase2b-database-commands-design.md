# Phase 2B: Database Commands - Design Document

**Goal:** Add essential database operations (dump, create, import, drop) for worktree integration, using database-agnostic interface with MySQL/MariaDB first.

**Architecture:** Database-agnostic design with driver pattern. DatabaseExecutor interface separates concerns from git/file operations. Engine interface handles database-specific command building. DSN parsing extracts connection details from database URLs. Guards enforce safety (allowed patterns, default DB protection).

**Tech Stack:** Go standard library (`net/url`, `path/filepath`, `os/exec`), Docker Compose for database operations, Engine driver pattern for extensibility.

---

## Architecture Overview

Phase 2B adds essential database operations for worktree integration with a database-agnostic design. The architecture extends Phase 1 with three new layers:

**DatabaseExecutor Interface** (`internal/executor/database.go`):
Separate interface for database operations, keeping concerns isolated from git/file operations. Methods: `Dump()`, `Create()`, `Import()`, `Drop()`. Implementation uses Docker internally via `docker compose exec -T`.

**DatabaseEngine Driver Pattern** (`internal/executor/engines/`):
Defines `DatabaseEngine` interface with engine-specific command builders. Initial implementation: `MySQLEngine` (handles both MySQL and MariaDB via serverVersion parameter). Each engine knows how to build dump, create, import, and drop commands for its database type.

**DSN Parser** (`internal/core/dsn.go`):
Parses database URLs to extract: engine (mysql/postgresql), user, password, host (Docker service name), port, database name, and serverVersion query parameter. Returns typed `DSN` struct with all connection details.

**Command Flow:**
1. Config loads `.claude/project.json` with database section
2. DSN parsed on startup, engine selected based on scheme + serverVersion
3. Database commands receive typed DSN, call DatabaseExecutor
4. Executor delegates to appropriate engine for command building
5. Commands executed via Docker Compose

---

## Types and DSN Parsing

**DSN Struct** (`internal/core/dsn.go`):
```go
type DSN struct {
    Engine        string  // "mysql" or "mariadb" or "postgresql"
    User          string
    Password      string
    Host          string  // Docker service name
    Port          string
    Database      string
    ServerVersion string  // From ?serverVersion=mariadb-11.4
}

func ParseDSN(dsnString string) (*DSN, error)
```

**Parsing Logic:**
- Parse using Go's `net/url` package
- Scheme determines base engine (mysql://, postgresql://, postgres://)
- Query param `serverVersion` distinguishes MariaDB from MySQL
- Port defaults: 3306 (MySQL/MariaDB), 5432 (PostgreSQL)
- URL-decode password (handles special characters)

**Database Result Types** (`internal/core/types/types.go`):
```go
type DumpResult struct {
    Path     string        `json:"path"`
    Size     int64         `json:"size"`
    Database string        `json:"database"`
    Duration time.Duration `json:"duration"`
}

type CreateResult struct {
    Database string `json:"database"`
}

type ImportResult struct {
    Path     string        `json:"path"`
    Database string        `json:"database"`
    Duration time.Duration `json:"duration"`
}

type DropResult struct {
    Database string `json:"database"`
}
```

---

## Engine Interface

**Engine Interface** (`internal/executor/engines/engine.go`):
```go
type DatabaseEngine interface {
    // Build command to dump database to stdout
    BuildDumpCommand(dsn *DSN, tables []string) []string
    
    // Build command to create database
    BuildCreateCommand(dsn *DSN, dbName string) []string
    
    // Build command to import SQL from stdin
    BuildImportCommand(dsn *DSN, dbName string) []string
    
    // Build command to drop database
    BuildDropCommand(dsn *DSN, dbName string) []string
    
    // Return engine name for logging/errors
    Name() string
}
```

**MySQL Engine Implementation** (`internal/executor/engines/mysql.go`):
```go
type MySQLEngine struct {
    isMariaDB bool
}

func (e *MySQLEngine) BuildDumpCommand(dsn *DSN, tables []string) []string {
    // mysqldump or mariadb-dump based on isMariaDB
    cmd := []string{"mysqldump", "-h", dsn.Host, "-u", dsn.User, 
                    "-p" + dsn.Password, dsn.Database}
    if len(tables) > 0 {
        cmd = append(cmd, tables...)
    }
    return cmd
}

func (e *MySQLEngine) BuildCreateCommand(dsn *DSN, dbName string) []string {
    return []string{"mysql", "-h", dsn.Host, "-u", dsn.User,
                    "-p" + dsn.Password, "-e", 
                    fmt.Sprintf("CREATE DATABASE `%s`", dbName)}
}
```

**Engine Selection** (in `DatabaseExecutor`):
- Check DSN.Engine: "mariadb" → MySQLEngine{isMariaDB: true}
- Check DSN.Engine: "mysql" → MySQLEngine{isMariaDB: false}
- Future: "postgresql" → PostgresEngine{}

---

## DatabaseExecutor Interface

**Interface Definition** (`internal/executor/database.go`):
```go
type DatabaseExecutor interface {
    Dump(service string, dsn *DSN, destPath string, tables []string) (*DumpResult, error)
    Create(service string, dsn *DSN, dbName string) (*CreateResult, error)
    Import(service string, dsn *DSN, sourcePath string, dbName string) (*ImportResult, error)
    Drop(service string, dsn *DSN, dbName string) (*DropResult, error)
}
```

**Docker Implementation** (`internal/executor/database.go`):
```go
type DockerDatabaseExecutor struct {
    engine      DatabaseEngine
    composeFile string
}

func (d *DockerDatabaseExecutor) Dump(service string, dsn *DSN, destPath string, tables []string) (*DumpResult, error) {
    // 1. Build engine-specific dump command
    cmd := d.engine.BuildDumpCommand(dsn, tables)
    
    // 2. Execute: docker compose exec -T <service> mysqldump ... > destPath
    fullCmd := append([]string{"exec", "-T", service}, cmd...)
    exec.Command("docker", "compose", fullCmd...)
    
    // 3. Write stdout to destPath
    // 4. Return DumpResult with path, size, duration
}
```

---

## Guards & Commands

**Database Guards** (`internal/core/guard.go`):
```go
// Check if database name matches allowed patterns
func IsDatabaseAllowed(dbName string, allowed []string) error {
    for _, pattern := range allowed {
        if matched, _ := filepath.Match(pattern, dbName); matched {
            return nil
        }
    }
    return &CommandError{Code: ErrDbNotAllowed, Message: "database not in allowed list"}
}

// Prevent dropping the default database
func IsNotDefaultDB(dbName string, defaultDB string) error {
    if dbName == defaultDB {
        return &CommandError{Code: ErrDbIsDefault, Message: "cannot drop default database"}
    }
    return nil
}
```

**Database Commands** (`internal/core/commands/database.go`):
```go
func Dump(projectRoot, dbName string, tables []string) (*DumpResult, error) {
    cfg, _ := config.Load(projectRoot)
    
    // Guards
    if err := IsDatabaseAllowed(dbName, cfg.Database.Allowed); err != nil {
        return nil, err
    }
    
    dsn, _ := ParseDSN(cfg.Database.DSN)
    executor := NewDockerDatabaseExecutor(dsn.Engine, cfg.Docker.ComposeFile)
    
    destPath := filepath.Join(cfg.Database.DumpsPath, 
                               fmt.Sprintf("%s_%s.sql", dbName, time.Now().Format("2006-01-02T15-04")))
    
    return executor.Dump(cfg.Database.Service, dsn, destPath, tables)
}

func Create(projectRoot, dbName string) (*CreateResult, error)
func Import(projectRoot, dbName, sourcePath string) (*ImportResult, error)
func Drop(projectRoot, dbName string) (*DropResult, error)
```

---

## Testing Strategy

**DSN Parsing Tests** (`internal/core/dsn_test.go`):
- Valid MySQL/MariaDB DSNs with serverVersion
- PostgreSQL DSNs
- URL-encoded passwords
- Missing components (port, user)
- Invalid DSNs (not a URL, missing database)

**Guard Tests** (`internal/core/guard_test.go`):
- Allowed patterns: exact match, wildcards, multiple patterns
- Default DB protection: refuse to drop default

**Database Command Tests** (`internal/core/commands/database_test.go`):
- Mock DatabaseExecutor, verify correct commands called
- Guards checked before executor calls
- Error handling: disallowed DB, file not found, etc.

---

## File Structure

```
internal/core/
├── dsn.go                    # NEW: DSN parsing
├── dsn_test.go               # NEW: DSN tests
├── guard.go                  # UPDATE: add database guards
├── guard_test.go             # UPDATE: add guard tests
├── types/types.go            # UPDATE: add DB result types
└── commands/
    └── database.go           # NEW: Dump, Create, Import, Drop

internal/executor/
├── database.go               # NEW: DatabaseExecutor interface + Docker impl
├── database_test.go          # NEW: Executor tests
└── engines/
    ├── engine.go             # NEW: DatabaseEngine interface
    └── mysql.go              # NEW: MySQL/MariaDB implementation
```

---

## Notes for Future Phases

**Phase 2C (or later):**
- `db.list` - List databases in container
- `db.clone` - Clone database in one operation (dump + create + import)
- `db.dumps` - List available SQL dump files
- `PostgresEngine` - PostgreSQL support
- Worktree integration - Add database clone/drop to worktree commands
