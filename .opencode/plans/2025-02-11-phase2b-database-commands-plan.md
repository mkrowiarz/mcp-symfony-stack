# Phase 2B: Database Commands - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add essential database operations (dump, create, import, drop) for worktree integration, using database-agnostic interface with MySQL/MariaDB first.

**Architecture:** Database-agnostic design with driver pattern. DatabaseExecutor interface separates concerns from git/file operations. Engine interface handles database-specific command building. DSN parsing extracts connection details from database URLs. Guards enforce safety (allowed patterns, default DB protection).

**Tech Stack:** Go standard library (`net/url`, `path/filepath`, `os/exec`), Docker Compose for database operations, Engine driver pattern for extensibility.

---

## Task 1: Extend Types with Database Result Types

**Files:**
- Modify: `internal/core/types/types.go`

**Step 1: Add DumpResult type**

```go
type DumpResult struct {
    Path     string        `json:"path"`
    Size     int64         `json:"size"`
    Database string        `json:"database"`
    Duration time.Duration `json:"duration"`
}
```

**Step 2: Add CreateResult type**

```go
type CreateResult struct {
    Database string `json:"database"`
}
```

**Step 3: Add ImportResult type**

```go
type ImportResult struct {
    Path     string        `json:"path"`
    Database string        `json:"database"`
    Duration time.Duration `json:"duration"`
}
```

**Step 4: Add DropResult type**

```go
type DropResult struct {
    Database string `json:"database"`
}
```

**Step 5: Add DSN type**

```go
type DSN struct {
    Engine        string
    User          string
    Password      string
    Host          string
    Port          string
    Database      string
    ServerVersion string
}
```

**Verification:**
- Run `go build ./internal/core/types` to verify compilation
- Types follow Go naming conventions

---

## Task 2: Implement DSN Parsing

**Files:**
- Create: `internal/core/dsn.go`

**Step 1: Implement ParseDSN function**

```go
package dsn

import (
    "fmt"
    "net/url"
    "strings"
    "github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
)

func ParseDSN(dsnString string) (*types.DSN, error) {
    if dsnString == "" {
        return nil, fmt.Errorf("empty DSN")
    }

    u, err := url.Parse(dsnString)
    if err != nil {
        return nil, fmt.Errorf("invalid DSN format: %w", err)
    }

    if u.Scheme == "" {
        return nil, fmt.Errorf("missing database scheme")
    }

    dsn := &types.DSN{
        User:     u.User.Username(),
        Host:     u.Hostname(),
        Database: strings.TrimPrefix(u.Path, "/"),
    }

    if password, ok := u.User.Password(); ok {
        dsn.Password = password
    }

    if u.Port() != "" {
        dsn.Port = u.Port()
    }

    // Parse query parameters
    query := u.Query()
    if serverVersion := query.Get("serverVersion"); serverVersion != "" {
        dsn.ServerVersion = serverVersion
    }

    // Determine engine from scheme and serverVersion
    dsn.Engine = determineEngine(u.Scheme, dsn.ServerVersion)

    // Set default ports
    if dsn.Port == "" {
        if dsn.Engine == "mysql" || dsn.Engine == "mariadb" {
            dsn.Port = "3306"
        } else if dsn.Engine == "postgresql" {
            dsn.Port = "5432"
        }
    }

    return dsn, nil
}

func determineEngine(scheme, serverVersion string) string {
    if strings.Contains(serverVersion, "mariadb") {
        return "mariadb"
    }

    switch scheme {
    case "mysql":
        return "mysql"
    case "postgresql", "postgres":
        return "postgresql"
    default:
        return scheme
    }
}
```

**Verification:**
- Run `go build ./internal/core` to verify compilation

---

## Task 3: Write DSN Tests

**Files:**
- Create: `internal/core/dsn_test.go`

**Step 1: Write basic DSN parsing test**

```go
func TestParseDSN(t *testing.T) {
    tests := []struct {
        name          string
        input         string
        expected      *types.DSN
        expectedError bool
    }{
        {
            name:  "MySQL DSN",
            input: "mysql://root:secret@database:3306/app",
            expected: &types.DSN{
                Engine:   "mysql",
                User:     "root",
                Password: "secret",
                Host:     "database",
                Port:     "3306",
                Database: "app",
            },
        },
        {
            name:  "MariaDB via serverVersion",
            input: "mysql://root:secret@db:3306/app?serverVersion=mariadb-11.4",
            expected: &types.DSN{
                Engine:        "mariadb",
                User:          "root",
                Password:      "secret",
                Host:          "db",
                Port:          "3306",
                Database:      "app",
                ServerVersion: "mariadb-11.4",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ParseDSN(tt.input)
            
            if tt.expectedError {
                if err == nil {
                    t.Error("expected error, got nil")
                }
                return
            }

            if err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }

            if result.Engine != tt.expected.Engine {
                t.Errorf("expected engine %s, got %s", tt.expected.Engine, result.Engine)
            }
        })
    }
}
```

**Verification:**
- Run `go test ./internal/core -run TestParseDSN -v`

---

## Task 4: Add Database Guards

**Files:**
- Modify: `internal/core/guard.go`

**Step 1: Add ErrDbNotAllowed error code**

```go
const (
    ErrConfigMissing ErrCode = "CONFIG_MISSING"
    ErrConfigInvalid ErrCode = "CONFIG_INVALID"
    ErrInvalidName   ErrCode = "INVALID_NAME"
    ErrPathTraversal ErrCode = "PATH_TRAVERSAL"
    ErrDbNotAllowed  ErrCode = "DB_NOT_ALLOWED"
    ErrDbIsDefault   ErrCode = "DB_IS_DEFAULT"
    ErrFileNotFound  ErrCode = "FILE_NOT_FOUND"
)
```

**Step 2: Implement IsDatabaseAllowed**

```go
func IsDatabaseAllowed(dbName string, allowed []string) error {
    for _, pattern := range allowed {
        matched, err := filepath.Match(pattern, dbName)
        if err == nil && matched {
            return nil
        }
    }
    return &CommandError{
        Code:    ErrDbNotAllowed,
        Message: fmt.Sprintf("database '%s' is not in allowed list", dbName),
    }
}
```

**Step 3: Implement IsNotDefaultDB**

```go
func IsNotDefaultDB(dbName string, defaultDB string) error {
    if dbName == defaultDB {
        return &CommandError{
            Code:    ErrDbIsDefault,
            Message: "cannot drop the default database",
        }
    }
    return nil
}
```

**Verification:**
- Run `go test ./internal/core -run TestIsDatabaseAllowed -v`
- Run `go test ./internal/core -run TestIsNotDefaultDB -v`

---

## Task 5: Create Engine Interface

**Files:**
- Create: `internal/executor/engines/engine.go`

**Step 1: Define DatabaseEngine interface**

```go
package engines

import "github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"

type DatabaseEngine interface {
    BuildDumpCommand(dsn *types.DSN, tables []string) []string
    BuildCreateCommand(dsn *types.DSN, dbName string) []string
    BuildImportCommand(dsn *types.DSN, dbName string) []string
    BuildDropCommand(dsn *types.DSN, dbName string) []string
    Name() string
}
```

**Verification:**
- Run `go build ./internal/executor/engines`

---

## Task 6: Implement MySQL Engine

**Files:**
- Create: `internal/executor/engines/mysql.go`

**Step 1: Implement MySQLEngine struct**

```go
package engines

import (
    "fmt"
    "github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
)

type MySQLEngine struct {
    isMariaDB bool
}

func NewMySQLEngine(isMariaDB bool) *MySQLEngine {
    return &MySQLEngine{isMariaDB: isMariaDB}
}
```

**Step 2: Implement BuildDumpCommand**

```go
func (e *MySQLEngine) BuildDumpCommand(dsn *types.DSN, tables []string) []string {
    dumpCmd := "mysqldump"
    if e.isMariaDB {
        dumpCmd = "mariadb-dump"
    }

    cmd := []string{
        dumpCmd,
        "-h", dsn.Host,
        "-u", dsn.User,
        fmt.Sprintf("-p%s", dsn.Password),
        dsn.Database,
    }

    if len(tables) > 0 {
        cmd = append(cmd, tables...)
    }

    return cmd
}
```

**Step 3: Implement BuildCreateCommand**

```go
func (e *MySQLEngine) BuildCreateCommand(dsn *types.DSN, dbName string) []string {
    return []string{
        "mysql",
        "-h", dsn.Host,
        "-u", dsn.User,
        fmt.Sprintf("-p%s", dsn.Password),
        "-e", fmt.Sprintf("CREATE DATABASE `%s`", dbName),
    }
}
```

**Step 4: Implement BuildImportCommand**

```go
func (e *MySQLEngine) BuildImportCommand(dsn *types.DSN, dbName string) []string {
    return []string{
        "mysql",
        "-h", dsn.Host,
        "-u", dsn.User,
        fmt.Sprintf("-p%s", dsn.Password),
        dbName,
    }
}
```

**Step 5: Implement BuildDropCommand**

```go
func (e *MySQLEngine) BuildDropCommand(dsn *types.DSN, dbName string) []string {
    return []string{
        "mysql",
        "-h", dsn.Host,
        "-u", dsn.User,
        fmt.Sprintf("-p%s", dsn.Password),
        "-e", fmt.Sprintf("DROP DATABASE `%s`", dbName),
    }
}
```

**Step 6: Implement Name method**

```go
func (e *MySQLEngine) Name() string {
    if e.isMariaDB {
        return "MariaDB"
    }
    return "MySQL"
}
```

**Verification:**
- Run `go build ./internal/executor/engines`

---

## Task 7: Implement DatabaseExecutor Interface

**Files:**
- Create: `internal/executor/database.go`

**Step 1: Define DatabaseExecutor interface**

```go
package executor

import (
    "github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
    "github.com/mkrowiarz/mcp-symfony-stack/internal/executor/engines"
)

type DatabaseExecutor interface {
    Dump(service string, dsn *types.DSN, destPath string, tables []string) (*types.DumpResult, error)
    Create(service string, dsn *types.DSN, dbName string) (*types.CreateResult, error)
    Import(service string, dsn *types.DSN, sourcePath string, dbName string) (*types.ImportResult, error)
    Drop(service string, dsn *types.DSN, dbName string) (*types.DropResult, error)
}
```

**Step 2: Implement DockerDatabaseExecutor**

```go
type DockerDatabaseExecutor struct {
    engine      engines.DatabaseEngine
    composeFile string
}

func NewDockerDatabaseExecutor(engine engines.DatabaseEngine, composeFile string) *DockerDatabaseExecutor {
    return &DockerDatabaseExecutor{
        engine:      engine,
        composeFile: composeFile,
    }
}
```

**Step 3: Implement Dump method**

```go
func (d *DockerDatabaseExecutor) Dump(service string, dsn *types.DSN, destPath string, tables []string) (*types.DumpResult, error) {
    start := time.Now()
    
    cmd := d.engine.BuildDumpCommand(dsn, tables)
    args := append([]string{"compose", "-f", d.composeFile, "exec", "-T", service}, cmd...)
    
    execCmd := exec.Command("docker", args...)
    output, err := execCmd.Output()
    if err != nil {
        return nil, fmt.Errorf("dump failed: %w", err)
    }
    
    if err := os.WriteFile(destPath, output, 0644); err != nil {
        return nil, fmt.Errorf("failed to write dump file: %w", err)
    }
    
    stat, _ := os.Stat(destPath)
    
    return &types.DumpResult{
        Path:     destPath,
        Size:     stat.Size(),
        Database: dsn.Database,
        Duration: time.Since(start),
    }, nil
}
```

**Step 4: Implement Create method**

```go
func (d *DockerDatabaseExecutor) Create(service string, dsn *types.DSN, dbName string) (*types.CreateResult, error) {
    cmd := d.engine.BuildCreateCommand(dsn, dbName)
    args := append([]string{"compose", "-f", d.composeFile, "exec", "-T", service}, cmd...)
    
    execCmd := exec.Command("docker", args...)
    if err := execCmd.Run(); err != nil {
        return nil, fmt.Errorf("create database failed: %w", err)
    }
    
    return &types.CreateResult{Database: dbName}, nil
}
```

**Step 5: Implement Import method**

```go
func (d *DockerDatabaseExecutor) Import(service string, dsn *types.DSN, sourcePath string, dbName string) (*types.ImportResult, error) {
    start := time.Now()
    
    sqlData, err := os.ReadFile(sourcePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read SQL file: %w", err)
    }
    
    cmd := d.engine.BuildImportCommand(dsn, dbName)
    args := append([]string{"compose", "-f", d.composeFile, "exec", "-T", service}, cmd...)
    
    execCmd := exec.Command("docker", args...)
    execCmd.Stdin = bytes.NewReader(sqlData)
    
    if err := execCmd.Run(); err != nil {
        return nil, fmt.Errorf("import failed: %w", err)
    }
    
    return &types.ImportResult{
        Path:     sourcePath,
        Database: dbName,
        Duration: time.Since(start),
    }, nil
}
```

**Step 6: Implement Drop method**

```go
func (d *DockerDatabaseExecutor) Drop(service string, dsn *types.DSN, dbName string) (*types.DropResult, error) {
    cmd := d.engine.BuildDropCommand(dsn, dbName)
    args := append([]string{"compose", "-f", d.composeFile, "exec", "-T", service}, cmd...)
    
    execCmd := exec.Command("docker", args...)
    if err := execCmd.Run(); err != nil {
        return nil, fmt.Errorf("drop database failed: %w", err)
    }
    
    return &types.DropResult{Database: dbName}, nil
}
```

**Verification:**
- Run `go build ./internal/executor`

---

## Task 8: Implement Database Commands

**Files:**
- Create: `internal/core/commands/database.go`

**Step 1: Implement Dump command**

```go
package commands

import (
    "fmt"
    "path/filepath"
    "time"
    
    "github.com/mkrowiarz/mcp-symfony-stack/internal/core/config"
    "github.com/mkrowiarz/mcp-symfony-stack/internal/core/dsn"
    "github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
    "github.com/mkrowiarz/mcp-symfony-stack/internal/executor"
    "github.com/mkrowiarz/mcp-symfony-stack/internal/executor/engines"
)

func Dump(projectRoot, dbName string, tables []string) (*types.DumpResult, error) {
    cfg, err := config.Load(projectRoot)
    if err != nil {
        return nil, err
    }
    
    if err := IsDatabaseAllowed(dbName, cfg.Database.Allowed); err != nil {
        return nil, err
    }
    
    parsedDSN, err := dsn.ParseDSN(cfg.Database.DSN)
    if err != nil {
        return nil, err
    }
    
    var engine engines.DatabaseEngine
    if parsedDSN.Engine == "mariadb" {
        engine = engines.NewMySQLEngine(true)
    } else {
        engine = engines.NewMySQLEngine(false)
    }
    
    dbExecutor := executor.NewDockerDatabaseExecutor(engine, cfg.Docker.ComposeFile)
    
    destPath := filepath.Join(cfg.Database.DumpsPath, 
        fmt.Sprintf("%s_%s.sql", dbName, time.Now().Format("2006-01-02T15-04")))
    
    return dbExecutor.Dump(cfg.Database.Service, parsedDSN, destPath, tables)
}
```

**Step 2: Implement Create command**

```go
func Create(projectRoot, dbName string) (*types.CreateResult, error) {
    cfg, err := config.Load(projectRoot)
    if err != nil {
        return nil, err
    }
    
    if err := IsDatabaseAllowed(dbName, cfg.Database.Allowed); err != nil {
        return nil, err
    }
    
    parsedDSN, err := dsn.ParseDSN(cfg.Database.DSN)
    if err != nil {
        return nil, err
    }
    
    var engine engines.DatabaseEngine
    if parsedDSN.Engine == "mariadb" {
        engine = engines.NewMySQLEngine(true)
    } else {
        engine = engines.NewMySQLEngine(false)
    }
    
    dbExecutor := executor.NewDockerDatabaseExecutor(engine, cfg.Docker.ComposeFile)
    
    return dbExecutor.Create(cfg.Database.Service, parsedDSN, dbName)
}
```

**Step 3: Implement Import command**

```go
func Import(projectRoot, dbName, sourcePath string) (*types.ImportResult, error) {
    cfg, err := config.Load(projectRoot)
    if err != nil {
        return nil, err
    }
    
    if err := IsDatabaseAllowed(dbName, cfg.Database.Allowed); err != nil {
        return nil, err
    }
    
    if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
        return nil, &types.CommandError{
            Code:    types.ErrFileNotFound,
            Message: fmt.Sprintf("SQL file not found: %s", sourcePath),
        }
    }
    
    parsedDSN, err := dsn.ParseDSN(cfg.Database.DSN)
    if err != nil {
        return nil, err
    }
    
    var engine engines.DatabaseEngine
    if parsedDSN.Engine == "mariadb" {
        engine = engines.NewMySQLEngine(true)
    } else {
        engine = engines.NewMySQLEngine(false)
    }
    
    dbExecutor := executor.NewDockerDatabaseExecutor(engine, cfg.Docker.ComposeFile)
    
    return dbExecutor.Import(cfg.Database.Service, parsedDSN, sourcePath, dbName)
}
```

**Step 4: Implement Drop command**

```go
func Drop(projectRoot, dbName string) (*types.DropResult, error) {
    cfg, err := config.Load(projectRoot)
    if err != nil {
        return nil, err
    }
    
    if err := IsDatabaseAllowed(dbName, cfg.Database.Allowed); err != nil {
        return nil, err
    }
    
    parsedDSN, err := dsn.ParseDSN(cfg.Database.DSN)
    if err != nil {
        return nil, err
    }
    
    if err := IsNotDefaultDB(dbName, parsedDSN.Database); err != nil {
        return nil, err
    }
    
    var engine engines.DatabaseEngine
    if parsedDSN.Engine == "mariadb" {
        engine = engines.NewMySQLEngine(true)
    } else {
        engine = engines.NewMySQLEngine(false)
    }
    
    dbExecutor := executor.NewDockerDatabaseExecutor(engine, cfg.Docker.ComposeFile)
    
    return dbExecutor.Drop(cfg.Database.Service, parsedDSN, dbName)
}
```

**Verification:**
- Run `go build ./internal/core/commands`

---

## Task 9: Update Config Validation

**Files:**
- Modify: `internal/core/config/config.go`

**Step 1: Add Database section validation**

```go
if cfg.Database != nil {
    if cfg.Database.Service == "" {
        return nil, &types.CommandError{
            Code:    types.ErrConfigInvalid,
            Message: "database.service is required when database section is present",
        }
    }
    if cfg.Database.DSN == "" {
        return nil, &types.CommandError{
            Code:    types.ErrConfigInvalid,
            Message: "database.dsn is required when database section is present",
        }
    }
    if len(cfg.Database.Allowed) == 0 {
        return nil, &types.CommandError{
            Code:    types.ErrConfigInvalid,
            Message: "database.allowed must have at least one pattern",
        }
    }
}
```

**Verification:**
- Run `go test ./internal/core/config -v`

---

## Task 10: Write Database Command Tests

**Files:**
- Create: `internal/core/commands/database_test.go`

**Step 1: Write mock DatabaseExecutor**

```go
type MockDatabaseExecutor struct {
    dumpResult   *types.DumpResult
    createResult *types.CreateResult
    importResult *types.ImportResult
    dropResult   *types.DropResult
    err          error
}

func (m *MockDatabaseExecutor) Dump(service string, dsn *types.DSN, destPath string, tables []string) (*types.DumpResult, error) {
    return m.dumpResult, m.err
}

// Similar for Create, Import, Drop...
```

**Step 2: Write tests for guard enforcement**

```go
func TestDumpDisallowedDB(t *testing.T) {
    // Setup config with allowed = ["app"]
    // Call Dump with dbName = "other_db"
    // Verify ErrDbNotAllowed returned before executor called
}

func TestDropDefaultDB(t *testing.T) {
    // Setup config with DSN database = "app"
    // Call Drop with dbName = "app"
    // Verify ErrDbIsDefault returned
}
```

**Verification:**
- Run `go test ./internal/core/commands -run TestDatabase -v`

---

## Task 11: Update Documentation

**Files:**
- Modify: `README.md`

**Step 1: Add Phase 2B documentation**

```markdown
## Phase 2B: Database Commands

### Database Commands

**`db.dump`**: Dump database to SQL file

```go
result, _ := database.Dump(".", "app_db", nil)
fmt.Printf("Dumped %s (%d bytes) to %s\n", result.Database, result.Size, result.Path)
```

**`db.create`**: Create empty database

```go
_, _ := database.Create(".", "new_db")
fmt.Println("Database created")
```

**`db.import`**: Import SQL file into database

```go
result, _ := database.Import(".", "app_db", "var/dumps/app_db.sql")
fmt.Printf("Imported %s into %s\n", result.Path, result.Database)
```

**`db.drop`**: Drop database (protected: can't drop default)

```go
_, _ := database.Drop(".", "old_db")
fmt.Println("Database dropped")
```

**Configuration:**
- `database.dsn`: Database URL (supports mysql://, postgresql://)
- `database.allowed`: List of allowed database patterns (supports glob wildcards)
- `database.dumps_path`: Directory for SQL dumps (default: var/dumps)

**Notes:**
- Phase 2B implements essential database operations only
- `db.list`, `db.clone`, `db.dumps` will be added in future phases
- MySQL/MariaDB supported, PostgreSQL support planned
```

**Verification:**
- Build binary: `go build -o pm ./cmd/pm`
- Verify README examples compile

---

## Testing Strategy

**Unit Tests:**
- DSN parsing with various formats
- Guard functions: allowed patterns, default DB protection
- Database commands: mock executor, verify guard enforcement

**Integration Tests (CI only):**
- Requires running Docker Compose with database service
- Test actual dump/create/import/drop operations
- Verify SQL files are created correctly

**Coverage Target:** 80%+ for database-related packages

---

## Success Criteria

Phase 2B is complete when:
- ✅ All 11 tasks implemented and tested
- ✅ Test coverage 80%+ for database packages
- ✅ Binary builds successfully
- ✅ README.md updated with Phase 2B examples
- ✅ No breaking changes to Phase 1/2A code
- ✅ All tests passing

---

## Notes for Future Phases

**Phase 2C (or later):**
- `db.list` - List databases in container
- `db.clone` - Clone database in one operation (dump + create + import)
- `db.dumps` - List available SQL dump files
- `PostgresEngine` - PostgreSQL support
- Worktree integration - Add database clone/drop to worktree commands
