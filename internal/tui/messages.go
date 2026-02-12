package tui

type projectLoadedMsg struct {
	name   string
	ptype  string
	status string
}

type worktreesLoadedMsg struct {
	worktrees []worktreeInfo
}

type worktreeInfo struct {
	branch string
	path   string
	isMain bool
}

type databasesLoadedMsg struct {
	databases []databaseInfo
}

type databaseInfo struct {
	name      string
	isDefault bool
}

type dumpsLoadedMsg struct {
	dumps []dumpInfo
}

type dumpInfo struct {
	name string
	size string
	date string
}
