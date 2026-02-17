package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// ConfigFile represents a discovered config file
type ConfigFile struct {
	Path     string
	Format   string // "toml", "yaml", "json"
	Priority int
}

// Loader handles config file discovery and parsing
type Loader struct {
	searchPaths []ConfigFile
}

// NewLoader creates a new config loader with default search paths
func NewLoader() *Loader {
	return &Loader{
		searchPaths: []ConfigFile{
			{Path: "haive.toml", Format: "toml", Priority: 1},
			{Path: ".haive/config.toml", Format: "toml", Priority: 2},
			{Path: "haive.yaml", Format: "yaml", Priority: 3},
			{Path: ".haive/config.yaml", Format: "yaml", Priority: 4},
			{Path: "haive.json", Format: "json", Priority: 5},
			{Path: ".haive/config.json", Format: "json", Priority: 6},
			{Path: ".claude/project.json", Format: "json", Priority: 7},
		},
	}
}

// Load discovers and loads config from project root or parent directories
func (l *Loader) Load(startDir string) (*HaiveConfig, error) {
	searchDir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	for {
		for _, cf := range l.searchPaths {
			configPath := filepath.Join(searchDir, cf.Path)

			if _, err := os.Stat(configPath); err != nil {
				continue // File doesn't exist
			}

			cfg, err := l.parseFile(configPath, cf.Format)
			if err != nil {
				// File exists but is invalid - continue searching
				continue
			}

			if cfg == nil || !l.hasContent(cfg) {
				continue // File exists but has no haive content
			}

			cfg.ProjectRoot = searchDir
			return cfg, nil
		}

		// Move to parent directory
		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			break // Reached root
		}
		searchDir = parent
	}

	return nil, fmt.Errorf("config file not found in %s or parent directories", startDir)
}

func (l *Loader) parseFile(path string, format string) (*HaiveConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg HaiveConfig

	switch format {
	case "toml":
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse TOML: %w", err)
		}
	case "yaml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	case "json":
		// Try namespaced format first (legacy .haive.json)
		var wrapper struct {
			PM *HaiveConfig `json:"pm"`
		}
		if err := json.Unmarshal(data, &wrapper); err == nil && wrapper.PM != nil && l.hasContent(wrapper.PM) {
			return wrapper.PM, nil
		}

		// Try direct format
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}

	return &cfg, nil
}

func (l *Loader) hasContent(cfg *HaiveConfig) bool {
	return cfg.Project.Name != "" ||
		cfg.Docker.ComposeFiles != nil ||
		cfg.Worktree != nil ||
		cfg.Database != nil
}

// LoadHaive is the convenience function for loading HaiveConfig
func LoadHaive(projectRoot string) (*HaiveConfig, error) {
	loader := NewLoader()
	cfg, err := loader.Load(projectRoot)
	if err != nil {
		return nil, err
	}

	// Resolve environment variables in DSN
	if cfg.Database != nil {
		cfg.Database.DSN = ResolveEnvVars(cfg.Database.DSN, cfg.ProjectRoot)
	}

	// Set default dumps path
	if cfg.Database != nil && cfg.Database.DumpsPath == "" {
		cfg.Database.DumpsPath = "var/dumps"
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
