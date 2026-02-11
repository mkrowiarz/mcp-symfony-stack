package executor

import "github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"

func GitWorktreeList() ([]types.WorktreeInfo, error) {
	executor := NewGitExecutor()
	return executor.GitWorktreeList()
}

func GitWorktreeAdd(path, branch string, newBranch bool) error {
	executor := NewGitExecutor()
	return executor.GitWorktreeAdd(path, branch, newBranch)
}

func GitWorktreeRemove(path string) error {
	executor := NewGitExecutor()
	return executor.GitWorktreeRemove(path)
}
