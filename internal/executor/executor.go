package executor

import "github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"

type Executor interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	FileExists(path string) bool

	GitWorktreeList() ([]types.WorktreeInfo, error)
	GitWorktreeAdd(path, branch string, newBranch bool) error
	GitWorktreeRemove(path string) error
}
