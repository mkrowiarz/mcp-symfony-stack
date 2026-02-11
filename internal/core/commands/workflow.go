package commands

import (
	"strconv"
)

type WorkflowCreateResult struct {
	WorktreePath   string `json:"worktree_path"`
	WorktreeBranch string `json:"worktree_branch"`
}

func CreateIsolatedWorktree(projectRoot, branch, newBranch, newDB string) (*WorkflowCreateResult, error) {
	// TODO: newDB parameter is reserved for Phase 2C when database cloning
	// will be integrated into the worktree creation workflow.
	// Currently, only worktree operations are performed.
	_ = newDB // unused until Phase 2C

	newBranchBool, _ := strconv.ParseBool(newBranch)
	result, err := Create(projectRoot, branch, newBranchBool)
	if err != nil {
		return nil, err
	}

	return &WorkflowCreateResult{
		WorktreePath:   result.Path,
		WorktreeBranch: result.Branch,
	}, nil
}
