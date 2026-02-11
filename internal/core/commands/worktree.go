package commands

import (
	"fmt"
	"os"
	"path/filepath"

	pmcore "github.com/mkrowiarz/mcp-symfony-stack/internal/core"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/config"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/executor"
)

func List(projectRoot string) ([]types.WorktreeInfo, error) {
	_, err := config.Load(projectRoot)
	if err != nil {
		if cmdErr, ok := err.(*types.CommandError); ok && cmdErr.Code == types.ErrConfigMissing {
			return nil, fmt.Errorf("no worktrees config, returning empty list")
		}
		return nil, err
	}

	convertWorktreeInfo := func(wt types.WorktreeInfo) types.WorktreeInfo {
		return types.WorktreeInfo{
			Path:   wt.Path,
			Branch: wt.Branch,
			IsMain: wt.IsMain,
		}
	}

	externalWorktrees, err := executor.GitWorktreeList()
	if err != nil {
		return nil, err
	}

	result := make([]types.WorktreeInfo, len(externalWorktrees))
	for i, wt := range externalWorktrees {
		result[i] = convertWorktreeInfo(wt)
	}

	return result, nil
}

func Create(projectRoot string, branch string, newBranch bool) (*types.WorktreeCreateResult, error) {
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, err
	}

	if err := pmcore.ValidateBranchName(branch); err != nil {
		return nil, err
	}

	dirName, _ := pmcore.SanitizeWorktreeName(branch)
	worktreePath := filepath.Join(cfg.Worktrees.BasePath, dirName)

	if err := pmcore.CheckPathTraversal(worktreePath, cfg.Worktrees.BasePath); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create worktree directory: %w", err)
	}

	if err := executor.GitWorktreeAdd(worktreePath, branch, newBranch); err != nil {
		return nil, err
	}

	return &types.WorktreeCreateResult{
		Path:   worktreePath,
		Branch: branch,
	}, nil
}

func Remove(projectRoot string, branch string) (*types.WorktreeRemoveResult, error) {
	cfg, err := config.Load(projectRoot)
	if err != nil {
		return nil, err
	}

	if err := pmcore.ValidateBranchName(branch); err != nil {
		return nil, err
	}

	dirName, _ := pmcore.SanitizeWorktreeName(branch)
	worktreePath := filepath.Join(cfg.Worktrees.BasePath, dirName)

	if err := pmcore.CheckPathTraversal(worktreePath, cfg.Worktrees.BasePath); err != nil {
		return nil, err
	}

	if err := executor.GitWorktreeRemove(worktreePath); err != nil {
		return nil, err
	}

	return &types.WorktreeRemoveResult{
		Path: worktreePath,
	}, nil
}
