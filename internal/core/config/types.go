package config

import (
	"fmt"
)

// Module is the interface all modules implement
type Module interface {
	Name() string
	Validate() error
}

// WorktreeConfig holds worktree module configuration
type WorktreeConfig struct {
	BasePath      string         `toml:"base_path" yaml:"base_path" json:"base_path"`
	DBPerWorktree bool           `toml:"db_per_worktree,omitempty" yaml:"db_per_worktree,omitempty" json:"db_per_worktree,omitempty"`
	DBPrefix      string         `toml:"db_prefix,omitempty" yaml:"db_prefix,omitempty" json:"db_prefix,omitempty"`
	Copy          *CopyConfig    `toml:"copy,omitempty" yaml:"copy,omitempty" json:"copy,omitempty"`
	Hooks         *WorktreeHooks `toml:"hooks,omitempty" yaml:"hooks,omitempty" json:"hooks,omitempty"`
	Env           *EnvConfig     `toml:"env,omitempty" yaml:"env,omitempty" json:"env,omitempty"`
}

func (w *WorktreeConfig) Name() string { return "worktree" }

func (w *WorktreeConfig) Validate() error {
	if w.BasePath == "" {
		return fmt.Errorf("worktree.base_path is required")
	}
	return nil
}

// CopyConfig holds file copy patterns
type CopyConfig struct {
	Include []string `toml:"include,omitempty" yaml:"include,omitempty" json:"include,omitempty"`
	Exclude []string `toml:"exclude,omitempty" yaml:"exclude,omitempty" json:"exclude,omitempty"`
}

// WorktreeHooks holds worktree lifecycle hooks
type WorktreeHooks struct {
	PostCreate []string `toml:"postCreate,omitempty" yaml:"postCreate,omitempty" json:"postCreate,omitempty"`
	PreRemove  []string `toml:"preRemove,omitempty" yaml:"preRemove,omitempty" json:"preRemove,omitempty"`
	PostRemove []string `toml:"postRemove,omitempty" yaml:"postRemove,omitempty" json:"postRemove,omitempty"`
}

// EnvConfig holds per-worktree environment configuration
type EnvConfig struct {
	File    string `toml:"file" yaml:"file" json:"file"`
	VarName string `toml:"var_name" yaml:"var_name" json:"var_name"`
}

// DatabaseConfig holds database module configuration
type DatabaseConfig struct {
	Service   string         `toml:"service" yaml:"service" json:"service"`
	DSN       string         `toml:"dsn" yaml:"dsn" json:"dsn"`
	Allowed   []string       `toml:"allowed" yaml:"allowed" json:"allowed"`
	DumpsPath string         `toml:"dumps_path,omitempty" yaml:"dumps_path,omitempty" json:"dumps_path,omitempty"`
	Hooks     *DatabaseHooks `toml:"hooks,omitempty" yaml:"hooks,omitempty" json:"hooks,omitempty"`
}

func (d *DatabaseConfig) Name() string { return "database" }

func (d *DatabaseConfig) Validate() error {
	if d.Service == "" {
		return fmt.Errorf("database.service is required")
	}
	if d.DSN == "" {
		return fmt.Errorf("database.dsn is required")
	}
	if len(d.Allowed) == 0 {
		return fmt.Errorf("database.allowed is required")
	}
	return nil
}

// DatabaseHooks holds database lifecycle hooks
type DatabaseHooks struct {
	PostClone []string `toml:"postClone,omitempty" yaml:"postClone,omitempty" json:"postClone,omitempty"`
	PreDrop   []string `toml:"preDrop,omitempty" yaml:"preDrop,omitempty" json:"preDrop,omitempty"`
}

// DockerConfig holds Docker settings
type DockerConfig struct {
	ComposeFiles []string `toml:"compose_files,omitempty" yaml:"compose_files,omitempty" json:"compose_files,omitempty"`
	ProjectName  string   `toml:"project_name,omitempty" yaml:"project_name,omitempty" json:"project_name,omitempty"`
}

// HaiveConfig is the top-level configuration structure
type HaiveConfig struct {
	Docker      DockerConfig    `toml:"docker" yaml:"docker" json:"docker"`
	Worktree    *WorktreeConfig `toml:"worktree,omitempty" yaml:"worktree,omitempty" json:"worktree,omitempty"`
	Database    *DatabaseConfig `toml:"database,omitempty" yaml:"database,omitempty" json:"database,omitempty"`
	ProjectRoot string          `toml:"-" yaml:"-" json:"-"` // Set at runtime
}

// Validate validates the entire configuration
func (c *HaiveConfig) Validate() error {
	if c.Worktree != nil {
		if err := c.Worktree.Validate(); err != nil {
			return err
		}
	}

	if c.Database != nil {
		if err := c.Database.Validate(); err != nil {
			return err
		}
	}

	// Validate worktree.env requires database
	if c.Worktree != nil && c.Worktree.Env != nil && c.Database == nil {
		return fmt.Errorf("worktree.env requires database configuration")
	}

	return nil
}
