# Phase 2C: Database List/Clone/Dumps + Worktree Integration - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add db.list, db.clone, db.dumps commands and integrate database cloning into worktree orchestrator.

**Architecture:** Extend existing database engine pattern with list command. Clone operation uses temp file for dump+import. Orchestrator updated to auto-clone when db_per_worktree enabled.

**Tech Stack:** Go standard library (os, io, filepath, time), existing executor pattern, glob matching for dump files.

---

## Task 1: Add New Types

**Files:**
- Modify: `internal/core/types/types.go`

**Step 1: Add DatabaseInfo and DatabaseListResult**

```go
type DatabaseInfo struct {
    Name      string `json:"name"`
    IsDefault bool   `json:"is_default"`
}

type DatabaseListResult struct {
    Databases []DatabaseInfo `json:"databases"`
}
```

**Step 2: Add CloneResult**

```go
type CloneResult struct {
    Source   string        `json:"source"`
    Target   string        `json:"target"`
    Size     int64         `json:"size"`
    Duration time.Duration `json:"duration"`
}
```

**Step 3: Add DumpFileInfo and DumpsListResult**

```go
type DumpFileInfo struct {
    Name     string `json:"name"`
    Database string `json:"database"`
    Size     int64  `json:"size"`
    Modified string `json:"modified"`
}

type DumpsListResult struct {
    Dumps []DumpFileInfo `json:"dumps"`
}
```

**Step 4: Update WorkflowCreateResult**

```go
type WorkflowCreateResult struct {
    WorktreePath   string `json:"worktree_path"`
    WorktreeBranch string `json:"worktree_branch"`
    DatabaseName   string `json:"database_name,omitempty"`
    ClonedFrom     string `json:"cloned_from,omitempty"`
}
```

**Step 5: Add WorkflowRemoveResult**

```go
type WorkflowRemoveResult struct {
    WorktreePath string `json:"worktree_path"`
    DatabaseName string `json:"database_name,omitempty"`
}
```

**Verification:**
- Run `go build ./internal/core/types`

---

## Task 2: Add BuildListCommand to Engine Interface

**Files:**
- Modify: `internal/executor/engines/engine.go`

**Step 1: Add method to interface**

```go
type DatabaseEngine interface {
    BuildDumpCommand(dsn *types.DSN, tables []string) []string
    BuildCreateCommand(dsn *types.DSN, dbName string) []string
    BuildImportCommand(dsn *types.DSN, dbName string) []string
    BuildDropCommand(dsn *types.DSN, dbName string) []string
    BuildListCommand() []string
    Name() string
}
```

**Verification:**
- Run `go build ./internal/executor/engines` (will fail - MySQL engine doesn't implement)

---

## Task 3: Implement MySQL BuildListCommand

**Files:**
- Modify: `internal/executor/engines/mysql.go`

**Step 1: Add BuildListCommand method**

```go
func (e *MySQLEngine) BuildListCommand() []string {
    return []string{
        "mysql",
        "-h", dsn.Host,
        "-u", dsn.User,
        fmt.Sprintf("-p%s", dsn.Password),
        "-e", "SHOW DATABASES",
    }
}
```

Wait - this needs DSN. Let me check the pattern.

Actually, looking at the existing methods, they all take `dsn *types.DSN` as parameter. But ListCommand doesn't need database-specific info, just connection. Let me adjust:

```go
func (e *MySQLEngine) BuildListCommand(dsn *types.DSN) []string {
    return []string{
        "mysql",
        "-h", dsn.Host,
        "-u", dsn.User,
        fmt.Sprintf("-p%s", dsn.Password),
        "-e", "SHOW DATABASES",
    }
}
```

Also need to update interface in Task 2.

**Step 2: Update engine.go interface**

```go
BuildListCommand(dsn *types.DSN) []string
```

**Verification:**
- Run `go build ./internal/executor/engines`

---

## Task 4: Add MySQL Engine Tests for BuildListCommand

**Files:**
- Modify: `internal/executor/engines/mysql_test.go`

**Step 1: Add test for BuildListCommand**

```go
func TestMySQLEngine_BuildListCommand(t *testing.T) {
    engine := NewMySQLEngine(false)
    dsn := &types.DSN{Host: "localhost", User: "root", Password: "secret"}

    result := engine.BuildListCommand(dsn)

    expected := []string{"mysql", "-h", "localhost", "-u", "root", "-psecret", "-e", "SHOW DATABASES"}

    if len(result) != len(expected) {
        t.Errorf("expected %d args, got %d", len(expected), len(result))
        return
    }

    for i, v := range result {
        if v != expected[i] {
            t.Errorf("arg[%d] = %q, want %q", i, v, expected[i])
        }
    }
}
```

**Verification:**
- Run `go test ./internal/executor/engines -v -run TestMySQLEngine_BuildListCommand`

---

## Task 5: Add List Method to DatabaseExecutor Interface

**Files:**
- Modify: `internal/executor/database.go`

**Step 1: Add List to interface**

```go
type DatabaseExecutor interface {
    Dump(service string, dsn *types.DSN, destPath string, tables []string) (*types.DumpResult, error)
    Create(service string, dsn *types.DSN, dbName string) (*types.CreateResult, error)
    Import(service string, dsn *types.DSN, sourcePath string, dbName string) (*types.ImportResult, error)
    Drop(service string, dsn *types.DSN, dbName string) (*types.DropResult, error)
    List(service string, dsn *types.DSN, defaultDB string) (*types.DatabaseListResult, error)
}
```

**Step 2: Implement List in DockerDatabaseExecutor**

```go
func (d *DockerDatabaseExecutor) List(service string, dsn *types.DSN, defaultDB string) (*types.DatabaseListResult, error) {
    cmd := d.engine.BuildListCommand(dsn)
    args := append([]string{"compose", "-f", d.composeFile, "exec", "-T", service}, cmd...)

    execCmd := exec.Command("docker", args...)
    output, err := execCmd.Output()
    if err != nil {
        return nil, fmt.Errorf("list databases failed: %w", err)
    }

    return parseDatabaseList(string(output), defaultDB)
}

func parseDatabaseList(output, defaultDB string) (*types.DatabaseListResult, error) {
    lines := strings.Split(strings.TrimSpace(output), "\n")
    var databases []types.DatabaseInfo

    systemDBs := map[string]bool{
        "information_schema": true,
        "mysql":              true,
        "performance_schema": true,
        "sys":                true,
    }

    for _, line := range lines {
        name := strings.TrimSpace(line)
        if name == "" || name == "Database" {
            continue
        }
        if systemDBs[name] {
            continue
        }

        databases = append(databases, types.DatabaseInfo{
            Name:      name,
            IsDefault: name == defaultDB,
        })
    }

    return &types.DatabaseListResult{Databases: databases}, nil
}
```

**Step 3: Add strings import if needed**

Check if `strings` is already imported. If not, add it.

**Verification:**
- Run `go build ./internal/executor`

---

## Task 6: Implement ListDBs Command

**Files:**
- Modify: `internal/core/commands/database.go`

**Step 1: Add ListDBs function**

```go
func ListDBs(projectRoot string) (*types.DatabaseListResult, error) {
    cfg, err := config.Load(projectRoot)
    if err != nil {
        return nil, err
    }

    if cfg.Database == nil {
        return nil, &types.CommandError{
            Code:    types.ErrConfigMissing,
            Message: "database configuration is required for list operations",
        }
    }

    parsedDSN, err := dsn.ParseDSN(cfg.Database.DSN)
    if err != nil {
        return nil, err
    }

    engine := getEngine(parsedDSN.Engine)

    dbExecutor := executor.NewDockerDatabaseExecutor(engine, cfg.Docker.ComposeFile)

    return dbExecutor.List(cfg.Database.Service, parsedDSN, parsedDSN.Database)
}
```

**Verification:**
- Run `go build ./internal/core/commands`

---

## Task 7: Add ListDBs Tests

**Files:**
- Modify: `internal/core/commands/database_test.go`

**Step 1: Add test for missing config**

```go
func TestListDBsMissingConfig(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "db-test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    cfgPath := filepath.Join(tmpDir, ".claude", "project.json")
    cfgDir := filepath.Dir(cfgPath)
    if err := os.MkdirAll(cfgDir, 0755); err != nil {
        t.Fatal(err)
    }

    cfgContent := `{
        "project": {"name": "test", "type": "symfony"},
        "docker": {"compose_file": "docker-compose.yaml"}
    }`
    if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
        t.Fatal(err)
    }

    _, err = ListDBs(tmpDir)
    if err == nil {
        t.Error("expected error when database config is missing")
    }

    cmdErr, ok := err.(*types.CommandError)
    if !ok {
        t.Errorf("expected CommandError, got %T", err)
        return
    }

    if cmdErr.Code != types.ErrConfigMissing {
        t.Errorf("expected ErrConfigMissing, got %s", cmdErr.Code)
    }
}
```

**Verification:**
- Run `go test ./internal/core/commands -v -run TestListDBs`

---

## Task 8: Implement CloneDB Command

**Files:**
- Modify: `internal/core/commands/database.go`

**Step 1: Add CloneDB function**

```go
func CloneDB(projectRoot, sourceDB, targetDB string) (*types.CloneResult, error) {
    start := time.Now()

    cfg, err := config.Load(projectRoot)
    if err != nil {
        return nil, err
    }

    if cfg.Database == nil {
        return nil, &types.CommandError{
            Code:    types.ErrConfigMissing,
            Message: "database configuration is required for clone operations",
        }
    }

    if sourceDB == "" {
        parsedDSN, _ := dsn.ParseDSN(cfg.Database.DSN)
        sourceDB = parsedDSN.Database
    }

    if err := core.IsDatabaseAllowed(sourceDB, cfg.Database.Allowed); err != nil {
        return nil, err
    }

    if err := core.IsDatabaseAllowed(targetDB, cfg.Database.Allowed); err != nil {
        return nil, err
    }

    parsedDSN, err := dsn.ParseDSN(cfg.Database.DSN)
    if err != nil {
        return nil, err
    }

    engine := getEngine(parsedDSN.Engine)

    dbExecutor := executor.NewDockerDatabaseExecutor(engine, cfg.Docker.ComposeFile)

    _, err = dbExecutor.Create(cfg.Database.Service, parsedDSN, targetDB)
    if err != nil {
        return nil, fmt.Errorf("failed to create target database: %w", err)
    }

    tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("clone_%s_%d.sql", targetDB, time.Now().UnixNano()))

    dumpResult, err := dbExecutor.Dump(cfg.Database.Service, parsedDSN, tmpFile, nil)
    if err != nil {
        os.Remove(tmpFile)
        return nil, fmt.Errorf("failed to dump source database: %w", err)
    }

    _, err = dbExecutor.Import(cfg.Database.Service, parsedDSN, tmpFile, targetDB)
    if err != nil {
        os.Remove(tmpFile)
        return nil, fmt.Errorf("failed to import into target database: %w", err)
    }

    os.Remove(tmpFile)

    return &types.CloneResult{
        Source:   sourceDB,
        Target:   targetDB,
        Size:     dumpResult.Size,
        Duration: time.Since(start),
    }, nil
}
```

**Step 2: Add os import if needed**

Check imports, add `os` if not present.

**Verification:**
- Run `go build ./internal/core/commands`

---

## Task 9: Add CloneDB Tests

**Files:**
- Modify: `internal/core/commands/database_test.go`

**Step 1: Add test for disallowed source**

```go
func TestCloneDBDisallowedSource(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "db-test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    cfgPath := filepath.Join(tmpDir, ".claude", "project.json")
    cfgDir := filepath.Dir(cfgPath)
    if err := os.MkdirAll(cfgDir, 0755); err != nil {
        t.Fatal(err)
    }

    cfgContent := `{
        "project": {"name": "test", "type": "symfony"},
        "docker": {"compose_file": "docker-compose.yaml"},
        "database": {
            "service": "database",
            "dsn": "mysql://root:secret@database:3306/app",
            "allowed": ["app", "app_*"],
            "dumps_path": "var/dumps"
        }
    }`
    if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
        t.Fatal(err)
    }

    _, err = CloneDB(tmpDir, "other_db", "app_test")
    if err == nil {
        t.Error("expected error for disallowed source database")
    }

    cmdErr, ok := err.(*types.CommandError)
    if !ok {
        t.Errorf("expected CommandError, got %T", err)
        return
    }

    if cmdErr.Code != types.ErrDbNotAllowed {
        t.Errorf("expected ErrDbNotAllowed, got %s", cmdErr.Code)
    }
}
```

**Step 2: Add test for disallowed target**

```go
func TestCloneDBDisallowedTarget(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "db-test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    cfgPath := filepath.Join(tmpDir, ".claude", "project.json")
    cfgDir := filepath.Dir(cfgPath)
    if err := os.MkdirAll(cfgDir, 0755); err != nil {
        t.Fatal(err)
    }

    cfgContent := `{
        "project": {"name": "test", "type": "symfony"},
        "docker": {"compose_file": "docker-compose.yaml"},
        "database": {
            "service": "database",
            "dsn": "mysql://root:secret@database:3306/app",
            "allowed": ["app", "app_*"],
            "dumps_path": "var/dumps"
        }
    }`
    if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
        t.Fatal(err)
    }

    _, err = CloneDB(tmpDir, "app", "other_db")
    if err == nil {
        t.Error("expected error for disallowed target database")
    }

    cmdErr, ok := err.(*types.CommandError)
    if !ok {
        t.Errorf("expected CommandError, got %T", err)
        return
    }

    if cmdErr.Code != types.ErrDbNotAllowed {
        t.Errorf("expected ErrDbNotAllowed, got %s", cmdErr.Code)
    }
}
```

**Verification:**
- Run `go test ./internal/core/commands -v -run TestCloneDB`

---

## Task 10: Implement ListDumps Command

**Files:**
- Modify: `internal/core/commands/database.go`

**Step 1: Add ListDumps function**

```go
func ListDumps(projectRoot string) (*types.DumpsListResult, error) {
    cfg, err := config.Load(projectRoot)
    if err != nil {
        return nil, err
    }

    if cfg.Database == nil {
        return nil, &types.CommandError{
            Code:    types.ErrConfigMissing,
            Message: "database configuration is required for dumps list operations",
        }
    }

    dumpsPath := cfg.Database.DumpsPath
    if dumpsPath == "" {
        dumpsPath = "var/dumps"
    }

    if _, err := os.Stat(dumpsPath); os.IsNotExist(err) {
        return &types.DumpsListResult{Dumps: []types.DumpFileInfo{}}, nil
    }

    files, err := filepath.Glob(filepath.Join(dumpsPath, "*.sql"))
    if err != nil {
        return nil, fmt.Errorf("failed to read dumps directory: %w", err)
    }

    var dumps []types.DumpFileInfo
    for _, file := range files {
        info, err := os.Stat(file)
        if err != nil {
            continue
        }

        filename := filepath.Base(file)
        dbName, timestamp := parseDumpFilename(filename)
        if dbName == "" {
            continue
        }

        dumps = append(dumps, types.DumpFileInfo{
            Name:     filename,
            Database: dbName,
            Size:     info.Size(),
            Modified: info.ModTime().Format(time.RFC3339),
        })
    }

    sort.Slice(dumps, func(i, j int) bool {
        return dumps[i].Modified > dumps[j].Modified
    })

    return &types.DumpsListResult{Dumps: dumps}, nil
}

func parseDumpFilename(filename string) (dbName, timestamp string) {
    ext := filepath.Ext(filename)
    if ext != ".sql" {
        return "", ""
    }

    name := strings.TrimSuffix(filename, ext)

    idx := strings.LastIndex(name, "_")
    if idx == -1 {
        return "", ""
    }

    dbName = name[:idx]
    timestamp = name[idx+1:]

    return dbName, timestamp
}
```

**Step 2: Add sort import if needed**

Check imports, add `sort` if not present.

**Verification:**
- Run `go build ./internal/core/commands`

---

## Task 11: Add ListDumps Tests

**Files:**
- Modify: `internal/core/commands/database_test.go`

**Step 1: Add test for empty directory**

```go
func TestListDumpsEmptyDirectory(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "db-test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    cfgPath := filepath.Join(tmpDir, ".claude", "project.json")
    cfgDir := filepath.Dir(cfgPath)
    if err := os.MkdirAll(cfgDir, 0755); err != nil {
        t.Fatal(err)
    }

    dumpsDir := filepath.Join(tmpDir, "var", "dumps")
    if err := os.MkdirAll(dumpsDir, 0755); err != nil {
        t.Fatal(err)
    }

    cfgContent := `{
        "project": {"name": "test", "type": "symfony"},
        "docker": {"compose_file": "docker-compose.yaml"},
        "database": {
            "service": "database",
            "dsn": "mysql://root:secret@database:3306/app",
            "allowed": ["app"],
            "dumps_path": "` + dumpsDir + `"
        }
    }`
    if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
        t.Fatal(err)
    }

    result, err := ListDumps(tmpDir)
    if err != nil {
        t.Errorf("unexpected error: %v", err)
        return
    }

    if len(result.Dumps) != 0 {
        t.Errorf("expected empty list, got %d dumps", len(result.Dumps))
    }
}
```

**Step 2: Add test for files with correct pattern**

```go
func TestListDumpsValidFiles(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "db-test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    cfgPath := filepath.Join(tmpDir, ".claude", "project.json")
    cfgDir := filepath.Dir(cfgPath)
    if err := os.MkdirAll(cfgDir, 0755); err != nil {
        t.Fatal(err)
    }

    dumpsDir := filepath.Join(tmpDir, "var", "dumps")
    if err := os.MkdirAll(dumpsDir, 0755); err != nil {
        t.Fatal(err)
    }

    os.WriteFile(filepath.Join(dumpsDir, "app_2025-02-11T10-30.sql"), []byte("test"), 0644)
    os.WriteFile(filepath.Join(dumpsDir, "app_2025-02-10T18-00.sql"), []byte("test"), 0644)
    os.WriteFile(filepath.Join(dumpsDir, "invalid.txt"), []byte("test"), 0644)

    cfgContent := `{
        "project": {"name": "test", "type": "symfony"},
        "docker": {"compose_file": "docker-compose.yaml"},
        "database": {
            "service": "database",
            "dsn": "mysql://root:secret@database:3306/app",
            "allowed": ["app"],
            "dumps_path": "` + dumpsDir + `"
        }
    }`
    if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
        t.Fatal(err)
    }

    result, err := ListDumps(tmpDir)
    if err != nil {
        t.Errorf("unexpected error: %v", err)
        return
    }

    if len(result.Dumps) != 2 {
        t.Errorf("expected 2 dumps, got %d", len(result.Dumps))
        return
    }

    if result.Dumps[0].Database != "app" {
        t.Errorf("expected database 'app', got %s", result.Dumps[0].Database)
    }
}
```

**Verification:**
- Run `go test ./internal/core/commands -v -run TestListDumps`

---

## Task 12: Add db_prefix Default to Config

**Files:**
- Modify: `internal/core/config/config.go`

**Step 1: Add default db_prefix in Load function**

After the worktrees validation block, add:

```go
if cfg.Worktrees != nil && cfg.Worktrees.DBPrefix == "" && cfg.Database != nil {
    parsedDSN, err := parseDSNForPrefix(cfg.Database.DSN)
    if err == nil && parsedDSN.Database != "" {
        cfg.Worktrees.DBPrefix = parsedDSN.Database + "_wt_"
    }
}
```

Wait, this creates a circular dependency. The dsn package is in internal/core/dsn. Let me think about this differently.

Actually, we can just extract the database name from the DSN string directly:

```go
if cfg.Worktrees != nil && cfg.Worktrees.DBPrefix == "" && cfg.Database != nil {
    dbName := extractDatabaseFromDSN(cfg.Database.DSN)
    if dbName != "" {
        cfg.Worktrees.DBPrefix = dbName + "_wt_"
    }
}
```

**Step 2: Add helper function**

```go
func extractDatabaseFromDSN(dsnString string) string {
    if dsnString == "" {
        return ""
    }

    idx := strings.LastIndex(dsnString, "/")
    if idx == -1 {
        return ""
    }

    dbPart := dsnString[idx+1:]

    queryIdx := strings.Index(dbPart, "?")
    if queryIdx != -1 {
        dbPart = dbPart[:queryIdx]
    }

    return dbPart
}
```

Actually, this is getting complicated. Let me use the existing dsn package properly. Import it at the top of config.go.

**Step 2: Import dsn package**

```go
import (
    // ... existing imports
    "github.com/mkrowiarz/mcp-symfony-stack/internal/core/dsn"
)
```

**Step 3: Use ParseDSN for db_prefix default**

```go
if cfg.Worktrees != nil && cfg.Worktrees.DBPrefix == "" && cfg.Database != nil {
    parsedDSN, err := dsn.ParseDSN(cfg.Database.DSN)
    if err == nil && parsedDSN.Database != "" {
        cfg.Worktrees.DBPrefix = parsedDSN.Database + "_wt_"
    }
}
```

**Verification:**
- Run `go test ./internal/core/config -v`

---

## Task 13: Update CreateIsolatedWorktree Orchestrator

**Files:**
- Modify: `internal/core/commands/workflow.go`

**Step 1: Update CreateIsolatedWorktree function**

```go
func CreateIsolatedWorktree(projectRoot, branch, newBranch, newDB string) (*WorkflowCreateResult, error) {
    // TODO: newDB parameter is reserved for Phase 2C when database cloning
    // will be integrated into the worktree creation workflow.
    // Currently, only worktree operations are performed.
    _ = newDB // unused - controlled by config.Worktrees.DBPerWorktree

    newBranchBool, _ := strconv.ParseBool(newBranch)
    result, err := Create(projectRoot, branch, newBranchBool)
    if err != nil {
        return nil, err
    }

    workflowResult := &WorkflowCreateResult{
        WorktreePath:   result.Path,
        WorktreeBranch: result.Branch,
    }

    cfg, err := config.Load(projectRoot)
    if err != nil {
        return workflowResult, nil
    }

    if cfg.Worktrees == nil || !cfg.Worktrees.DBPerWorktree || cfg.Database == nil {
        return workflowResult, nil
    }

    _, dbName := core.SanitizeWorktreeName(branch)
    targetDB := cfg.Worktrees.DBPrefix + dbName

    cloneResult, err := CloneDB(projectRoot, "", targetDB)
    if err != nil {
        return workflowResult, fmt.Errorf("worktree created but database clone failed: %w", err)
    }

    envPath := filepath.Join(result.Path, ".env.local")
    newDSN := strings.Replace(cfg.Database.DSN, cloneResult.Source, cloneResult.Target, 1)

    if err := os.WriteFile(envPath, []byte("DATABASE_URL="+newDSN+"\n"), 0644); err != nil {
        return workflowResult, fmt.Errorf("worktree and DB created but .env.local patch failed: %w", err)
    }

    workflowResult.DatabaseName = cloneResult.Target
    workflowResult.ClonedFrom = cloneResult.Source

    return workflowResult, nil
}
```

**Step 2: Add missing imports**

Add `config` and `filepath` imports if not present.

**Verification:**
- Run `go build ./internal/core/commands`

---

## Task 14: Add RemoveIsolatedWorktree Orchestrator

**Files:**
- Modify: `internal/core/commands/workflow.go`

**Step 1: Add RemoveIsolatedWorktree function**

```go
func RemoveIsolatedWorktree(projectRoot, branch string, dropDB bool) (*WorkflowRemoveResult, error) {
    result, err := Remove(projectRoot, branch)
    if err != nil {
        return nil, err
    }

    workflowResult := &WorkflowRemoveResult{
        WorktreePath: result.Path,
    }

    if !dropDB {
        return workflowResult, nil
    }

    cfg, err := config.Load(projectRoot)
    if err != nil {
        return workflowResult, nil
    }

    if cfg.Worktrees == nil || cfg.Database == nil {
        return workflowResult, nil
    }

    _, dbName := core.SanitizeWorktreeName(branch)
    targetDB := cfg.Worktrees.DBPrefix + dbName

    _, err = DropDB(projectRoot, targetDB)
    if err != nil {
        if cmdErr, ok := err.(*types.CommandError); ok && cmdErr.Code == types.ErrDbNotAllowed {
            return workflowResult, nil
        }
        return workflowResult, fmt.Errorf("worktree removed but database drop failed: %w", err)
    }

    workflowResult.DatabaseName = targetDB

    return workflowResult, nil
}
```

**Verification:**
- Run `go build ./internal/core/commands`

---

## Task 15: Add Orchestrator Tests

**Files:**
- Create: `internal/core/commands/workflow_test.go`

**Step 1: Create test file with basic tests**

```go
package commands

import (
    "os"
    "path/filepath"
    "testing"
)

func TestCreateIsolatedWorktreeNoDB(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "workflow-test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    cfgPath := filepath.Join(tmpDir, ".claude", "project.json")
    cfgDir := filepath.Dir(cfgPath)
    if err := os.MkdirAll(cfgDir, 0755); err != nil {
        t.Fatal(err)
    }

    cfgContent := `{
        "project": {"name": "test", "type": "symfony"},
        "docker": {"compose_file": "docker-compose.yaml"},
        "worktrees": {"base_path": "` + tmpDir + `/wt"}
    }`
    if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
        t.Fatal(err)
    }

    _, err = CreateIsolatedWorktree(tmpDir, "feature/test", "true", "")
    if err == nil {
        t.Error("expected error (git not available)")
    }
}

func TestRemoveIsolatedWorktreeNoDB(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "workflow-test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    cfgPath := filepath.Join(tmpDir, ".claude", "project.json")
    cfgDir := filepath.Dir(cfgPath)
    if err := os.MkdirAll(cfgDir, 0755); err != nil {
        t.Fatal(err)
    }

    cfgContent := `{
        "project": {"name": "test", "type": "symfony"},
        "docker": {"compose_file": "docker-compose.yaml"},
        "worktrees": {"base_path": "` + tmpDir + `/wt"}
    }`
    if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
        t.Fatal(err)
    }

    _, err = RemoveIsolatedWorktree(tmpDir, "feature/test", false)
    if err == nil {
        t.Error("expected error (git not available)")
    }
}
```

**Verification:**
- Run `go test ./internal/core/commands -v -run TestCreateIsolated|TestRemoveIsolated`

---

## Task 16: Final Verification and Commit

**Step 1: Run all tests**

```bash
go test ./... -v
```

**Step 2: Run build**

```bash
go build ./...
```

**Step 3: Commit all changes**

```bash
git add -A
git commit -m "feat: add Phase 2C database commands and worktree integration

- db.list: List databases in container (excludes system DBs)
- db.clone: Clone database via dump+create+import
- db.dumps: List SQL dump files from dumps directory
- CreateIsolatedWorktree: Auto-clone DB when db_per_worktree enabled
- RemoveIsolatedWorktree: Drop DB when removing worktree
- Config: Add default db_prefix (<default_db>_wt_)

Tests: 10+ new test cases for commands and orchestrators"
```

---

## Success Criteria

Phase 2C is complete when:
- ✅ All 16 tasks completed
- ✅ `db.list` returns databases with default flag
- ✅ `db.clone` creates target database with cloned data
- ✅ `db.dumps` lists SQL files from dumps directory
- ✅ `CreateIsolatedWorktree` clones DB when `db_per_worktree: true`
- ✅ `RemoveIsolatedWorktree` drops DB when `drop_db: true`
- ✅ All tests passing
- ✅ No breaking changes to existing code
