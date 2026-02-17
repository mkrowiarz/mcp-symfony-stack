package worktree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/config"
)

func TestUpdateEnvFile(t *testing.T) {
	worktreePath := t.TempDir()

	// Create .env.local
	envContent := `APP_ENV=dev
DATABASE_URL=mysql://user:pass@db:3306/main_db
OTHER_VAR=value
`
	envPath := filepath.Join(worktreePath, ".env.local")
	os.WriteFile(envPath, []byte(envContent), 0644)

	cfg := &config.HaiveConfig{
		Project: config.ProjectConfig{Name: "test"},
		Database: &config.DatabaseConfig{
			DSN: "mysql://user:pass@db:3306/main_db",
		},
		Worktree: &config.WorktreeConfig{
			Env: &config.EnvConfig{
				File:    ".env.local",
				VarName: "DATABASE_URL",
			},
		},
	}

	err := updateEnvFile(worktreePath, cfg, "main_db_feature_x")
	if err != nil {
		t.Fatalf("updateEnvFile failed: %v", err)
	}

	// Read updated file
	updated, _ := os.ReadFile(envPath)
	updatedStr := string(updated)

	if !strings.Contains(updatedStr, "DATABASE_URL=mysql://user:pass@db:3306/main_db_feature_x") {
		t.Errorf("DATABASE_URL not updated correctly:\n%s", updatedStr)
	}

	// Other vars should be preserved
	if !strings.Contains(updatedStr, "APP_ENV=dev") {
		t.Error("APP_ENV was not preserved")
	}
}

func TestUpdateEnvFile_AppendNew(t *testing.T) {
	worktreePath := t.TempDir()

	// Create .env.local without DATABASE_URL
	envContent := `APP_ENV=dev
OTHER_VAR=value
`
	envPath := filepath.Join(worktreePath, ".env.local")
	os.WriteFile(envPath, []byte(envContent), 0644)

	cfg := &config.HaiveConfig{
		Project: config.ProjectConfig{Name: "test"},
		Database: &config.DatabaseConfig{
			DSN: "mysql://user:pass@db:3306/main_db",
		},
		Worktree: &config.WorktreeConfig{
			Env: &config.EnvConfig{
				File:    ".env.local",
				VarName: "DATABASE_URL",
			},
		},
	}

	err := updateEnvFile(worktreePath, cfg, "main_db_feature_x")
	if err != nil {
		t.Fatalf("updateEnvFile failed: %v", err)
	}

	updated, _ := os.ReadFile(envPath)
	if !strings.Contains(string(updated), "DATABASE_URL=mysql://user:pass@db:3306/main_db_feature_x") {
		t.Error("DATABASE_URL not appended")
	}
}

func TestUpdateEnvFile_FileNotFound(t *testing.T) {
	worktreePath := t.TempDir()

	cfg := &config.HaiveConfig{
		Project: config.ProjectConfig{Name: "test"},
		Database: &config.DatabaseConfig{
			DSN: "mysql://user:pass@db:3306/main_db",
		},
		Worktree: &config.WorktreeConfig{
			Env: &config.EnvConfig{
				File:    ".env.local",
				VarName: "DATABASE_URL",
			},
		},
	}

	err := updateEnvFile(worktreePath, cfg, "main_db_feature_x")
	if err == nil {
		t.Error("expected error for missing env file")
	}
}

func TestUpdateEnvFile_NoEnvConfig(t *testing.T) {
	worktreePath := t.TempDir()

	cfg := &config.HaiveConfig{
		Project: config.ProjectConfig{Name: "test"},
		Database: &config.DatabaseConfig{
			DSN: "mysql://user:pass@db:3306/main_db",
		},
		Worktree: &config.WorktreeConfig{
			BasePath: ".worktrees",
			// No Env config
		},
	}

	err := updateEnvFile(worktreePath, cfg, "main_db_feature_x")
	if err != nil {
		t.Errorf("expected no error when env config is nil, got: %v", err)
	}
}

func TestUpdateEnvFile_NoDatabaseConfig(t *testing.T) {
	worktreePath := t.TempDir()

	cfg := &config.HaiveConfig{
		Project: config.ProjectConfig{Name: "test"},
		// No Database config
		Worktree: &config.WorktreeConfig{
			Env: &config.EnvConfig{
				File:    ".env.local",
				VarName: "DATABASE_URL",
			},
		},
	}

	err := updateEnvFile(worktreePath, cfg, "main_db_feature_x")
	if err != nil {
		t.Errorf("expected no error when database config is nil, got: %v", err)
	}
}
