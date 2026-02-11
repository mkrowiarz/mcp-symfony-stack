# Phase 2A: Worktree Commands - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extend core library with worktree management commands (list, create, remove) and git operations, providing dual-level API (low-level commands + high-level orchestrator) for flexible MCP and TUI usage.

**Architecture:** Implement worktree commands with clean separation between low-level operations (worktree list/create/remove) and high-level orchestration (full feature setup). Add git executor to handle shell operations, guard functions for safety, and porcelain parsing for reliable worktree listing. Test with mock executors for isolation, integration tests for real git operations verification.

**Tech Stack:** Go standard library (`os/exec`, `path/filepath`, `regexp`, `strings`), Git operations via executor interface (wrapper around `git worktree` commands), Porcelain format parsing (`git worktree list --porcelain`), Error handling with typed codes from phase 1.

---

## Task 1: Extend Types for Worktree

**Files:**
- Modify: `internal/core/types/types.go`

**Step 1: Add WorktreeInfo type**

```go
type WorktreeInfo struct {
    Path    string `json:"path"`
    Branch   string `json:"branch"`
    IsMain   bool   `json:"is_main"`
}
```

**Step 2: Add WorktreeCreateResult type**

```go
type WorktreeCreateResult struct {
    Path   string `json:"path"`
    Branch string `json:"branch"`
}
```

**Step 3: Add WorktreeRemoveResult type**

```go
type WorktreeRemoveResult struct {
    Path string `json:"path"`
}
```

**Verification:**
- Run `go build ./internal/core/types` to verify compilation
- Types follow Go naming conventions (PascalCase for exported, camelCase for fields)

---

## Task 2: Add Guard Functions

**Files:**
- Create: `internal/core/guard.go`

**Step 1: Implement ValidateBranchName**

```go
func ValidateBranchName(name string) error {
    if name == "" {
        return &CommandError{Code: ErrInvalidName, Message: "branch name cannot be empty"}
    }

    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_\-\/]+$`, name)
    if !matched {
        return &CommandError{Code: ErrInvalidName, Message: "branch name contains invalid characters"}
    }

    return nil
}
```

**Step 2: Implement CheckPathTraversal**

```go
func CheckPathTraversal(resolvedPath, basePath string) error {
    absResolved, err := filepath.Abs(resolvedPath)
    if err != nil {
        return &CommandError{Code: ErrPathTraversal, Message: "failed to resolve path: " + err.Error()}
    }

    absBase, err := filepath.Abs(basePath)
    if err != nil {
        return &CommandError{Code: ErrPathTraversal, Message: "failed to resolve base path: " + err.Error()}
    }

    rel, err := filepath.Rel(absBase, absResolved)
    if err != nil || strings.Contains(rel, "..") {
        return &CommandError{Code: ErrPathTraversal, Message: "path traversal attempt detected"}
    }

    return nil
}
```

**Step 3: Implement SanitizeWorktreeName**

```go
func SanitizeWorktreeName(branchName string) (dirName, dbName string) {
    dirName = strings.ReplaceAll(branchName, "/", "-")

    dbName = strings.ReplaceAll(branchName, "/", "_")
    dbName = strings.ReplaceAll(dbName, "-", "_")

    return
}
```

**Verification:**
- Run `go test ./internal/core/guard/...` to verify guard functions
- Test cases cover: empty names, invalid characters, path traversal, normal names with slashes

---

## Task 3: Implement Git Executor

**Files:**
- Modify: `internal/executor/executor.go`
- Modify: `internal/executor/git.go`

**Step 1: Update Executor Interface**

```go
type Executor interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte) error
    FileExists(path string) bool

    GitWorktreeList() ([]WorktreeInfo, error)
    GitWorktreeAdd(path, branch string, newBranch bool) error
    GitWorktreeRemove(path string) error
}
```

**Step 2: Implement GitWorktreeList**

```go
func (g *GitExecutor) GitWorktreeList() ([]WorktreeInfo, error) {
    cmd := exec.Command("git", "worktree", "list", "--porcelain")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("git worktree list failed: %w", err)
    }

    return parseWorktreeListOutput(output)
}
```

**Step 3: Implement GitWorktreeAdd**

```go
func (g *GitExecutor) GitWorktreeAdd(path, branch string, newBranch bool) error {
    args := []string{"worktree", "add", path, branch}
    if newBranch {
        args = append(args, "-b")
    }

    cmd := exec.Command("git", args...)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("git worktree add failed: %w", err)
    }

    return nil
}
```

**Step 4: Implement GitWorktreeRemove**

```go
func (g *GitExecutor) GitWorktreeRemove(path string) error {
    cmd := exec.Command("git", "worktree", "remove", path, "--force")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("git worktree remove failed: %w", err)
    }

    return nil
}
```

**Step 5: Implement Porcelain Parsing**

```go
type worktreeOutput struct {
    Path   string
    Head   string
    State  string
}

func parseWorktreeListOutput(output string) ([]WorktreeInfo, error) {
    lines := strings.Split(strings.TrimSpace(output), "\n")
    var worktrees []WorktreeInfo

    for i := 0; i < len(lines); i += 3 {
        if i >= len(lines) {
            break
        }

        output := worktreeOutput{
            Path:  strings.TrimSpace(strings.TrimPrefix(lines[i], "worktree ")),
            Head:  strings.TrimSpace(lines[i+1]),
            State: strings.TrimSpace(lines[i+2]),
        }

        worktrees = append(worktrees, WorktreeInfo{
            Path:   output.Path,
            Branch: strings.TrimPrefix(output.Head, "HEAD "),
            IsMain: output.State == "main",
        })
    }

    return worktrees, nil
}
```

**Verification:**
- Create temporary git repo, run `git worktree list --porcelain`, verify parsing matches real git output
- Test edge cases: empty output, single worktree, detached HEAD state

---

## Task 4: Implement Worktree Commands

**Files:**
- Create: `internal/core/commands/worktree.go`

**Step 1: Implement worktree.list**

```go
func List(projectRoot string) ([]WorktreeInfo, error) {
    cfg, err := config.Load(projectRoot)
    if err != nil {
        if cmdErr, ok := err.(*CommandError); ok && cmdErr.Code == ErrConfigMissing {
            return nil, fmt.Errorf("no worktrees config, returning empty list")
        }
        return nil, err
    }

    return executor.GitWorktreeList()
}
```

**Step 2: Implement worktree.create**

```go
func Create(projectRoot string, branch string, newBranch bool) (*WorktreeCreateResult, error) {
    cfg, err := config.Load(projectRoot)
    if err != nil {
        return nil, err
    }

    if err := ValidateBranchName(branch); err != nil {
        return nil, err
    }

    worktreePath := filepath.Join(cfg.Worktrees.BasePath, SanitizeWorktreeName(branch).dirName)

    if err := CheckPathTraversal(worktreePath, cfg.Worktrees.BasePath); err != nil {
        return nil, err
    }

    dirName := SanitizeWorktreeName(branch).dirName
    if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
        return nil, fmt.Errorf("failed to create worktree directory: %w", err)
    }

    if err := executor.GitWorktreeAdd(worktreePath, branch, newBranch); err != nil {
        return nil, err
    }

    return &WorktreeCreateResult{
        Path:   worktreePath,
        Branch: branch,
    }, nil
}
```

**Step 3: Implement worktree.remove**

```go
func Remove(projectRoot string, branch string) (*WorktreeRemoveResult, error) {
    cfg, err := config.Load(projectRoot)
    if err != nil {
        return nil, err
    }

    if err := ValidateBranchName(branch); err != nil {
        return nil, err
    }

    worktreePath := filepath.Join(cfg.Worktrees.BasePath, SanitizeWorktreeName(branch).dirName)

    if err := CheckPathTraversal(worktreePath, cfg.Worktrees.BasePath); err != nil {
        return nil, err
    }

    if err := executor.GitWorktreeRemove(worktreePath); err != nil {
        return nil, err
    }

    return &WorktreeRemoveResult{
        Path: worktreePath,
    }, nil
}
```

**Verification:**
- Run `go test ./internal/core/commands/... -v` to verify all commands work
- Test with mock executor to verify correct git commands called
- Test error paths: missing config, invalid names, path traversal

---

## Task 5: Implement Orchestrator

**Files:**
- Create: `internal/core/commands/workflow.go`

**Step 1: Define WorkflowCreateResult**

```go
type WorkflowCreateResult struct {
    WorktreePath   string `json:"worktree_path"`
    WorktreeBranch string `json:"worktree_branch"`
}
```

**Step 2: Implement CreateIsolatedWorktree**

```go
func CreateIsolatedWorktree(projectRoot, branch, newBranch bool, newDB string) (*WorkflowCreateResult, error) {
    // 1. Create worktree (calls worktree.Create)
    // 2. Note: Phase 2B will add database integration
    // 3. Return result

    result, err := Create(projectRoot, branch, newBranch)
    if err != nil {
        return nil, err
    }

    return &WorkflowCreateResult{
        WorktreePath:   result.Path,
        WorktreeBranch: result.Branch,
    }, nil
}
```

**Step 3: Update Worktree Command Documentation**

In `README.md`, add usage example for orchestrator:

```go
// Quick workflow
result, _ := workflow.CreateIsolatedWorktree(".", "feature/abc", true, "")
fmt.Printf("Worktree: %s (branch: %s)\n", result.WorktreePath, result.WorktreeBranch)
```

**Verification:**
- Run `go test ./internal/core/commands/... -run TestWorkflow` to verify orchestrator
- Test with missing config, invalid names

---

## Task 6: Write Tests

**Files:**
- Create: `internal/core/commands/worktree_test.go`
- Create: `internal/executor/executor_test.go`

**Step 1: Guard Tests**

```go
func TestValidateBranchName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected error
    }{
        {"valid normal name", "feature/test", nil},
        {"empty name", "", true},
        {"invalid chars", "test;rm -rf", true},
        {"path traversal", "../etc/passwd", true},
        {"slashes", "feature/test", nil},
        {"hyphens", "fix-auth-bug", nil},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateBranchName(tt.input)
            if (err == nil) != (tt.expected == nil) {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

**Step 2: Worktree Command Tests**

```go
func TestWorktreeList(t *testing.T) {
    mockExec := &MockExecutor{
        gitListOutput: "worktree /path/to/wt1\nHEAD abc123\nmain\nworktree /path/to/wt2\nHEAD def456\ndetached",
    }

    worktrees, err := List(".", mockExec)
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}

func TestWorktreeCreate(t *testing.T) {
    mockExec := &MockExecutor{}
    result, err := Create(".", mockExec, "feature/test", true)
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}
```

**Step 3: Git Executor Tests**

```go
func TestGitExecutor(t *testing.T) {
    // Create temporary git repo
    tmpDir := t.TempDir()

    // Test git worktree add
    exec.Command("git", "-C", tmpDir, "init").Run()
    exec.Command("git", "-C", tmpDir, "commit", "--allow-empty", "-m", "initial").Run()
    exec.Command("git", "-C", tmpDir, "branch", "main").Run()

    executor := &GitExecutor{}
    err := executor.GitWorktreeAdd(filepath.Join(tmpDir, "wt1"), "feature/1", true)
    if err != nil {
        t.Errorf("git worktree add failed: %v", err)
    }
}
```

**Verification:**
- Run `go test ./internal/core/... -cover` - Target 85%+ coverage for worktree package
- Run `go build ./internal/core/...` to verify compilation

---

## Task 7: Update Config Types and Documentation

**Files:**
- Modify: `internal/core/config/config.go`
- Modify: `README.md`

**Step 1: Add Worktrees Validation to Config Load**

Update `Load()` function to validate Worktrees section:

```go
if cfg.Worktrees != nil {
    if cfg.Worktrees.BasePath == "" {
        return nil, &CommandError{
            Code:    ErrConfigInvalid,
            Message: "worktrees.base_path is required when worktrees section is present",
        }
    }
}
```

**Step 2: Update README.md**

Add Phase 2A documentation:

```markdown
## Phase 2A: Worktree Commands

### Worktree Commands

**`worktree.list`**: List all git worktrees

```go
info, _ := commands.Info(".")
worktrees, _ := worktree.List(".")
for _, wt := range worktrees {
    fmt.Printf("%s: %s\n", wt.Branch, wt.Path)
}
```

**`worktree.create`**: Create a git worktree

```go
result, _ := worktree.Create(".", "feature/my-feature", true)
fmt.Printf("Created worktree: %s\n", result.Path)
```

**`workflow.create_isolated_worktree`**: Quick one-click workflow

```go
result, _ := workflow.CreateIsolatedWorktree(".", "feature/abc", true, "")
fmt.Printf("Worktree: %s (branch: %s)\n", result.WorktreePath, result.WorktreeBranch)
```

**Orchestrator Notes:**
- Phase 2A implements worktree commands only
- Database integration (clone/drop) will be added in Phase 2B
- Orchestrator provides one-click workflows for common cases
- Low-level commands available for granular control
```

**Verification:**
- Build binary: `go build -o pm ./cmd/pm`
- Test with sample project
- Verify README examples compile

---

## Testing Strategy

**Unit Tests (85%+ coverage goal):**
- Guard functions: All validation paths
- Worktree commands: Mock executor, verify correct operations
- Git executor: Integration tests with real git repo (CI only)
- Porcelain parsing: Various output formats

**Integration Tests (CI only):**
- Create temporary git repository
- Test worktree operations on real git
- Verify porcelain parsing matches actual git output

**Test Execution:**
```bash
# Unit tests
go test ./internal/core/... -cover

# Full test suite
go test ./...
```

---

## Success Criteria

Phase 2A is complete when:
- ✅ All 7 tasks implemented and tested
- ✅ Test coverage 85%+ for worktree package
- ✅ Binary builds successfully
- ✅ README.md updated with Phase 2A examples
- ✅ No breaking changes to Phase 1 code
- ✅ All tests passing

---

## Notes for Phase 2B

Phase 2B will add database integration:
- `db_prefix` parameter in workflow commands
- Database clone/drop operations integrated into orchestrator
- Update WorktreeInfo to include associated database
