package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/config"
	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/dsn"
)

// SetWorktreeDatabase configures the database for a worktree
// - Sets git config haive.database
// - Updates .env.local with new DATABASE_URL
func SetWorktreeDatabase(worktreePath string, cfg *config.HaiveConfig, dbName string) error {
	// Set git config
	cmd := exec.Command("git", "config", "--local", "haive.database", dbName)
	cmd.Dir = worktreePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set git config: %w", err)
	}

	// Update .env.local if configured
	if cfg.Worktree.Env != nil {
		if err := updateEnvFile(worktreePath, cfg, dbName); err != nil {
			return fmt.Errorf("failed to update env file: %w", err)
		}
	}

	return nil
}

// updateEnvFile updates the DATABASE_URL in the env file
func updateEnvFile(worktreePath string, cfg *config.HaiveConfig, dbName string) error {
	if cfg.Worktree.Env == nil || cfg.Database == nil {
		return nil
	}

	envFile := cfg.Worktree.Env.File
	varName := cfg.Worktree.Env.VarName

	envPath := filepath.Join(worktreePath, envFile)

	// Check if file exists
	_, err := os.Stat(envPath)
	if err != nil {
		return fmt.Errorf("env file not found: %s", envPath)
	}

	// Parse original DSN to get connection details
	parsedDSN, err := dsn.ParseDSN(cfg.Database.DSN)
	if err != nil {
		return fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Build new DSN with worktree database name
	parsedDSN.Database = dbName
	newDSN := parsedDSN.String()

	// Read existing file
	content, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read env file: %w", err)
	}

	// Replace the variable
	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, varName+"=") {
			lines[i] = fmt.Sprintf("%s=%s", varName, newDSN)
			found = true
			break
		}
	}

	if !found {
		// Variable not found, append it
		lines = append(lines, fmt.Sprintf("%s=%s", varName, newDSN))
	}

	// Write back
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(envPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	return nil
}

// GetWorktreeDatabase reads the configured database for a worktree
func GetWorktreeDatabase(worktreePath string) (string, error) {
	cmd := exec.Command("git", "config", "--local", "haive.database")
	cmd.Dir = worktreePath
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("no database configured for this worktree")
	}
	return strings.TrimSpace(string(out)), nil
}
