package executor

import (
	"os/exec"
	"testing"
)

func TestParseWorktreeListOutput(t *testing.T) {
	t.Run("single worktree", func(t *testing.T) {
		output := `worktree /private/tmp/test-git-worktree
HEAD 5c3d2446196ac3b38ed8f6ea94635ce0e4f6b7cb
branch refs/heads/main
`
		toplevelPath := "/private/tmp/test-git-worktree"

		worktrees, err := parseWorktreeListOutput(output, toplevelPath)
		if err != nil {
			t.Fatalf("parseWorktreeListOutput() error = %v", err)
		}

		if len(worktrees) != 1 {
			t.Fatalf("expected 1 worktree, got %d", len(worktrees))
		}

		if worktrees[0].Path != "/private/tmp/test-git-worktree" {
			t.Errorf("expected path /private/tmp/test-git-worktree, got %s", worktrees[0].Path)
		}

		if worktrees[0].Branch != "main" {
			t.Errorf("expected branch main, got %s", worktrees[0].Branch)
		}

		if !worktrees[0].IsMain {
			t.Errorf("expected IsMain to be true for main worktree")
		}
	})

	t.Run("multiple worktrees", func(t *testing.T) {
		output := `worktree /private/tmp/test-git-worktree
HEAD 5c3d2446196ac3b38ed8f6ea94635ce0e4f6b7cb
branch refs/heads/main

worktree /private/tmp/test-wt-1
HEAD 5c3d2446196ac3b38ed8f6ea94635ce0e4f6b7cb
branch refs/heads/feature1

worktree /private/tmp/test-wt-2
HEAD 5c3d2446196ac3b38ed8f6ea94635ce0e4f6b7cb
branch refs/heads/feature2
`
		toplevelPath := "/private/tmp/test-git-worktree"

		worktrees, err := parseWorktreeListOutput(output, toplevelPath)
		if err != nil {
			t.Fatalf("parseWorktreeListOutput() error = %v", err)
		}

		if len(worktrees) != 3 {
			t.Fatalf("expected 3 worktrees, got %d", len(worktrees))
		}

		if !worktrees[0].IsMain {
			t.Errorf("expected first worktree to be main")
		}

		if worktrees[1].IsMain {
			t.Errorf("expected second worktree not to be main")
		}

		if worktrees[2].Branch != "feature2" {
			t.Errorf("expected third worktree branch to be feature2, got %s", worktrees[2].Branch)
		}
	})

	t.Run("detached HEAD", func(t *testing.T) {
		output := `worktree /private/tmp/test-git-worktree
HEAD 5c3d2446196ac3b38ed8f6ea94635ce0e4f6b7cb
detached
`
		toplevelPath := "/private/tmp/test-git-worktree"

		worktrees, err := parseWorktreeListOutput(output, toplevelPath)
		if err != nil {
			t.Fatalf("parseWorktreeListOutput() error = %v", err)
		}

		if len(worktrees) != 1 {
			t.Fatalf("expected 1 worktree, got %d", len(worktrees))
		}

		if worktrees[0].Branch != "detached" {
			t.Errorf("expected branch 'detached', got %s", worktrees[0].Branch)
		}
	})

	t.Run("empty output", func(t *testing.T) {
		output := ""
		toplevelPath := "/private/tmp/test-git-worktree"

		worktrees, err := parseWorktreeListOutput(output, toplevelPath)
		if err != nil {
			t.Fatalf("parseWorktreeListOutput() error = %v", err)
		}

		if len(worktrees) != 0 {
			t.Errorf("expected 0 worktrees for empty output, got %d", len(worktrees))
		}
	})
}

func TestGitWorktreeList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("real git command", func(t *testing.T) {
		g := &GitExecutor{}

		cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
		if err := cmd.Run(); err != nil {
			t.Skip("not in a git repository")
		}

		worktrees, err := g.GitWorktreeList()
		if err != nil {
			t.Fatalf("GitWorktreeList() error = %v", err)
		}

		if len(worktrees) == 0 {
			t.Error("expected at least one worktree")
		}

		for _, wt := range worktrees {
			if wt.Path == "" {
				t.Error("worktree path should not be empty")
			}
			if wt.Branch == "" {
				t.Error("worktree branch should not be empty")
			}
		}
	})
}
