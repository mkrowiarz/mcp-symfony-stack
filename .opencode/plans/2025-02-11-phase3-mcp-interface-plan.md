# Phase 3: MCP Interface - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Expose all Phase 1-2C commands as MCP tools via stdio server.

**Architecture:** MCP server package wraps core commands, mcp-go SDK handles protocol, tool handlers parse args and format results.

**Tech Stack:** Go, github.com/mark3labs/mcp-go, stdio transport

---

## Task 1: Add mcp-go Dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add mcp-go dependency**

```bash
go get github.com/mark3labs/mcp-go
```

**Verification:**
- Run `go mod tidy`
- Run `go build ./...`

---

## Task 2: Create Error Mapping

**Files:**
- Create: `internal/mcp/errors.go`

**Step 1: Create errors.go with error code mapping**

```go
package mcp

import (
	"fmt"

	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
)

const (
	ErrCodeConfigMissing = -32001
	ErrCodeConfigInvalid = -32002
	ErrCodeInvalidName   = -32003
	ErrCodePathTraversal = -32004
	ErrCodeDbNotAllowed  = -32005
	ErrCodeDbIsDefault   = -32006
	ErrCodeFileNotFound  = -32007
)

func toMCPCode(code types.ErrCode) int {
	switch code {
	case types.ErrConfigMissing:
		return ErrCodeConfigMissing
	case types.ErrConfigInvalid:
		return ErrCodeConfigInvalid
	case types.ErrInvalidName:
		return ErrCodeInvalidName
	case types.ErrPathTraversal:
		return ErrCodePathTraversal
	case types.ErrDbNotAllowed:
		return ErrCodeDbNotAllowed
	case types.ErrDbIsDefault:
		return ErrCodeDbIsDefault
	case types.ErrFileNotFound:
		return ErrCodeFileNotFound
	default:
		return -32000
	}
}

func toMCPError(err error) error {
	if cmdErr, ok := err.(*types.CommandError); ok {
		return fmt.Errorf("mcp error %d: %s (data: {\"code\":\"%s\"})", 
			toMCPCode(cmdErr.Code), cmdErr.Message, cmdErr.Code)
	}
	return fmt.Errorf("mcp error -32000: %s", err.Error())
}
```

**Verification:**
- Run `go build ./internal/mcp`

---

## Task 3: Create MCP Server

**Files:**
- Create: `internal/mcp/server.go`

**Step 1: Create server.go with Run() function**

```go
package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Run() error {
	s := server.NewMCPServer(
		"mcp-project-manager",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	registerProjectTools(s)
	registerDatabaseTools(s)
	registerWorktreeTools(s)
	registerWorkflowTools(s)

	return server.ServeStdio(s)
}

func registerProjectTools(s *server.MCPServer) {
	// Registered in tools_project.go
}

func registerDatabaseTools(s *server.MCPServer) {
	// Registered in tools_database.go
}

func registerWorktreeTools(s *server.MCPServer) {
	// Registered in tools_worktree.go
}

func registerWorkflowTools(s *server.MCPServer) {
	// Registered in tools_workflow.go
}
```

**Verification:**
- Run `go build ./internal/mcp`

---

## Task 4: Add MCP Flag to main.go

**Files:**
- Modify: `cmd/pm/main.go`

**Step 1: Read current main.go**

```bash
cat cmd/pm/main.go
```

**Step 2: Add mcp flag and import**

Add to imports:
```go
import (
	// existing imports...
	"github.com/mkrowiarz/mcp-symfony-stack/internal/mcp"
)
```

Add flag and handler:
```go
func main() {
	mcpFlag := flag.Bool("mcp", false, "Run as MCP server")
	flag.Parse()

	if *mcpFlag {
		if err := mcp.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// existing code...
}
```

**Verification:**
- Run `go build -o pm ./cmd/pm`
- Run `./pm --help` to see new flag

---

## Task 5: Create Project Tools

**Files:**
- Create: `internal/mcp/tools_project.go`

**Step 1: Create tools_project.go**

```go
package mcp

import (
	"encoding/json"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/commands"
)

func registerProjectTools(s *server.MCPServer) {
	// project.info
	s.AddTool(mcp.NewTool("project.info",
		mcp.WithDescription("Get project configuration and status"),
	), handleProjectInfo)

	// project.init
	s.AddTool(mcp.NewTool("project.init",
		mcp.WithDescription("Generate suggested project configuration"),
	), handleProjectInit)
}

func handleProjectInfo(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectRoot, _ := os.Getwd()
	result, err := commands.Info(projectRoot)
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleProjectInit(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectRoot, _ := os.Getwd()
	result, err := commands.Init(projectRoot)
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
```

**Verification:**
- Run `go build ./internal/mcp`

---

## Task 6: Create Database Tools

**Files:**
- Create: `internal/mcp/tools_database.go`

**Step 1: Create tools_database.go**

```go
package mcp

import (
	"encoding/json"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/commands"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
)

func registerDatabaseTools(s *server.MCPServer) {
	// db.list
	s.AddTool(mcp.NewTool("db.list",
		mcp.WithDescription("List all databases in the container"),
	), handleDbList)

	// db.dump
	s.AddTool(mcp.NewTool("db.dump",
		mcp.WithDescription("Dump a database to a SQL file"),
		mcp.WithString("database", mcp.Description("Database name (optional, defaults to DSN database)")),
		mcp.WithArray("tables", mcp.Description("Specific tables to dump (optional)")),
	), handleDbDump)

	// db.import
	s.AddTool(mcp.NewTool("db.import",
		mcp.WithDescription("Import a SQL file into a database"),
		mcp.WithString("database", mcp.Required(), mcp.Description("Target database name")),
		mcp.WithString("sql_path", mcp.Required(), mcp.Description("Path to SQL file")),
	), handleDbImport)

	// db.create
	s.AddTool(mcp.NewTool("db.create",
		mcp.WithDescription("Create a new empty database"),
		mcp.WithString("database", mcp.Required(), mcp.Description("Database name")),
	), handleDbCreate)

	// db.drop
	s.AddTool(mcp.NewTool("db.drop",
		mcp.WithDescription("Drop a database (destructive)"),
		mcp.WithString("database", mcp.Required(), mcp.Description("Database name")),
		mcp.WithBoolean("confirm", mcp.Required(), mcp.Description("Must be true to confirm destructive operation")),
	), handleDbDrop)

	// db.clone
	s.AddTool(mcp.NewTool("db.clone",
		mcp.WithDescription("Clone a database (dump + create + import)"),
		mcp.WithString("source", mcp.Description("Source database (optional, defaults to DSN database)")),
		mcp.WithString("target", mcp.Required(), mcp.Description("Target database name")),
	), handleDbClone)

	// db.dumps
	s.AddTool(mcp.NewTool("db.dumps",
		mcp.WithDescription("List available SQL dump files"),
	), handleDbDumps)
}

func handleDbList(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectRoot, _ := os.Getwd()
	result, err := commands.ListDBs(projectRoot)
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleDbDump(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectRoot, _ := os.Getwd()
	
	database := ""
	if v, ok := arguments["database"].(string); ok {
		database = v
	}

	var tables []string
	if v, ok := arguments["tables"].([]interface{}); ok {
		for _, t := range v {
			if s, ok := t.(string); ok {
				tables = append(tables, s)
			}
		}
	}

	result, err := commands.Dump(projectRoot, database, tables)
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleDbImport(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectRoot, _ := os.Getwd()
	
	database := arguments["database"].(string)
	sqlPath := arguments["sql_path"].(string)

	result, err := commands.ImportDB(projectRoot, database, sqlPath)
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleDbCreate(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectRoot, _ := os.Getwd()
	database := arguments["database"].(string)

	result, err := commands.CreateDB(projectRoot, database)
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleDbDrop(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	confirm, _ := arguments["confirm"].(bool)
	if !confirm {
		return nil, toMCPError(&types.CommandError{
			Code:    types.ErrConfigInvalid,
			Message: "confirm must be true to drop database",
		})
	}

	projectRoot, _ := os.Getwd()
	database := arguments["database"].(string)

	result, err := commands.DropDB(projectRoot, database)
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleDbClone(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectRoot, _ := os.Getwd()
	
	source := ""
	if v, ok := arguments["source"].(string); ok {
		source = v
	}
	target := arguments["target"].(string)

	result, err := commands.CloneDB(projectRoot, source, target)
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleDbDumps(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectRoot, _ := os.Getwd()
	result, err := commands.ListDumps(projectRoot)
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
```

**Verification:**
- Run `go build ./internal/mcp`

---

## Task 7: Create Worktree Tools

**Files:**
- Create: `internal/mcp/tools_worktree.go`

**Step 1: Create tools_worktree.go**

```go
package mcp

import (
	"encoding/json"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/commands"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
)

func registerWorktreeTools(s *server.MCPServer) {
	// worktree.list
	s.AddTool(mcp.NewTool("worktree.list",
		mcp.WithDescription("List all git worktrees"),
	), handleWorktreeList)

	// worktree.create
	s.AddTool(mcp.NewTool("worktree.create",
		mcp.WithDescription("Create a new git worktree"),
		mcp.WithString("branch", mcp.Required(), mcp.Description("Branch name")),
		mcp.WithBoolean("new_branch", mcp.Description("Create new branch (default false)")),
	), handleWorktreeCreate)

	// worktree.remove
	s.AddTool(mcp.NewTool("worktree.remove",
		mcp.WithDescription("Remove a git worktree (destructive)"),
		mcp.WithString("branch", mcp.Required(), mcp.Description("Branch name")),
		mcp.WithBoolean("confirm", mcp.Required(), mcp.Description("Must be true to confirm destructive operation")),
	), handleWorktreeRemove)
}

func handleWorktreeList(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectRoot, _ := os.Getwd()
	result, err := commands.List(projectRoot)
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleWorktreeCreate(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectRoot, _ := os.Getwd()
	branch := arguments["branch"].(string)
	
	newBranch := false
	if v, ok := arguments["new_branch"].(bool); ok {
		newBranch = v
	}

	result, err := commands.Create(projectRoot, branch, newBranch)
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleWorktreeRemove(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	confirm, _ := arguments["confirm"].(bool)
	if !confirm {
		return nil, toMCPError(&types.CommandError{
			Code:    types.ErrConfigInvalid,
			Message: "confirm must be true to remove worktree",
		})
	}

	projectRoot, _ := os.Getwd()
	branch := arguments["branch"].(string)

	result, err := commands.Remove(projectRoot, branch)
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}
```

**Verification:**
- Run `go build ./internal/mcp`

---

## Task 8: Create Workflow Tools

**Files:**
- Create: `internal/mcp/tools_workflow.go`

**Step 1: Create tools_workflow.go**

```go
package mcp

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/commands"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
)

func registerWorkflowTools(s *server.MCPServer) {
	// workflow.create
	s.AddTool(mcp.NewTool("workflow.create",
		mcp.WithDescription("Create isolated worktree with database (if db_per_worktree enabled)"),
		mcp.WithString("branch", mcp.Required(), mcp.Description("Branch name")),
		mcp.WithBoolean("new_branch", mcp.Description("Create new branch (default false)")),
	), handleWorkflowCreate)

	// workflow.remove
	s.AddTool(mcp.NewTool("workflow.remove",
		mcp.WithDescription("Remove worktree and optionally drop database (destructive)"),
		mcp.WithString("branch", mcp.Required(), mcp.Description("Branch name")),
		mcp.WithBoolean("drop_db", mcp.Description("Drop associated database (default true)")),
		mcp.WithBoolean("confirm", mcp.Required(), mcp.Description("Must be true to confirm destructive operation")),
	), handleWorkflowRemove)
}

func handleWorkflowCreate(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	projectRoot, _ := os.Getwd()
	branch := arguments["branch"].(string)
	
	newBranch := "false"
	if v, ok := arguments["new_branch"].(bool); ok && v {
		newBranch = "true"
	}

	result, err := commands.CreateIsolatedWorktree(projectRoot, branch, newBranch, "")
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func handleWorkflowRemove(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	confirm, _ := arguments["confirm"].(bool)
	if !confirm {
		return nil, toMCPError(&types.CommandError{
			Code:    types.ErrConfigInvalid,
			Message: "confirm must be true to remove worktree",
		})
	}

	projectRoot, _ := os.Getwd()
	branch := arguments["branch"].(string)
	
	dropDB := true
	if v, ok := arguments["drop_db"].(bool); ok {
		dropDB = v
	}

	result, err := commands.RemoveIsolatedWorktree(projectRoot, branch, dropDB)
	if err != nil {
		return nil, toMCPError(err)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// Helper for boolean parsing
func parseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}
```

**Verification:**
- Run `go build ./internal/mcp`

---

## Task 9: Add MCP Server Tests

**Files:**
- Create: `internal/mcp/errors_test.go`

**Step 1: Create errors_test.go**

```go
package mcp

import (
	"testing"

	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
)

func TestToMCPCode(t *testing.T) {
	tests := []struct {
		input    types.ErrCode
		expected int
	}{
		{types.ErrConfigMissing, ErrCodeConfigMissing},
		{types.ErrConfigInvalid, ErrCodeConfigInvalid},
		{types.ErrInvalidName, ErrCodeInvalidName},
		{types.ErrPathTraversal, ErrCodePathTraversal},
		{types.ErrDbNotAllowed, ErrCodeDbNotAllowed},
		{types.ErrDbIsDefault, ErrCodeDbIsDefault},
		{types.ErrFileNotFound, ErrCodeFileNotFound},
		{types.ErrCode("UNKNOWN"), -32000},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := toMCPCode(tt.input)
			if result != tt.expected {
				t.Errorf("toMCPCode(%s) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToMCPError(t *testing.T) {
	cmdErr := &types.CommandError{
		Code:    types.ErrDbNotAllowed,
		Message: "database 'other' not allowed",
	}

	err := toMCPError(cmdErr)
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Verify error message contains code and message
	errStr := err.Error()
	if errStr == "" {
		t.Error("error message is empty")
	}
}
```

**Verification:**
- Run `go test ./internal/mcp -v -run TestToMCP`

---

## Task 10: Final Verification

**Step 1: Build everything**

```bash
go build ./...
go build -o pm ./cmd/pm
```

**Step 2: Run all tests**

```bash
go test ./... -v
```

**Step 3: Test MCP server manually**

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./pm --mcp | head -20
```

**Step 4: Commit**

```bash
jj describe -m "feat: add Phase 3 MCP interface

- Add mcp-go dependency
- Create MCP server with stdio transport
- Expose 14 tools: project.info/init, db.list/dump/import/create/drop/clone/dumps, worktree.list/create/remove, workflow.create/remove
- Map error codes to JSON-RPC -32000 range
- Require confirm parameter for destructive operations

Tools: 14 total (2 project, 7 database, 3 worktree, 2 workflow)"
```

---

## Success Criteria

Phase 3 is complete when:
- ✅ `pm --mcp` starts MCP server
- ✅ All 14 tools registered
- ✅ Tools/list returns tool definitions
- ✅ Error codes in -32000 range
- ✅ Destructive tools require confirm=true
- ✅ All tests passing
