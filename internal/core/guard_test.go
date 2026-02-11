package core

import (
	"testing"

	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
)

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name      string
		branch    string
		wantError bool
		code      types.ErrCode
	}{
		{
			name:      "empty name",
			branch:    "",
			wantError: true,
			code:      types.ErrInvalidName,
		},
		{
			name:      "valid simple name",
			branch:    "main",
			wantError: false,
		},
		{
			name:      "valid name with slash",
			branch:    "feature/new-auth",
			wantError: false,
		},
		{
			name:      "valid name with dash",
			branch:    "feature-new-auth",
			wantError: false,
		},
		{
			name:      "valid name with underscore",
			branch:    "feature_new_auth",
			wantError: false,
		},
		{
			name:      "invalid name with space",
			branch:    "feature new auth",
			wantError: true,
			code:      types.ErrInvalidName,
		},
		{
			name:      "invalid name with special chars",
			branch:    "feature@new",
			wantError: true,
			code:      types.ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBranchName(tt.branch)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateBranchName() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError && err.(*types.CommandError).Code != tt.code {
				t.Errorf("ValidateBranchName() error code = %v, want %v", err.(*types.CommandError).Code, tt.code)
			}
		})
	}
}

func TestCheckPathTraversal(t *testing.T) {
	tests := []struct {
		name         string
		resolvedPath string
		basePath     string
		wantError    bool
		code         types.ErrCode
	}{
		{
			name:         "valid subdirectory",
			resolvedPath: "/project/subdir",
			basePath:     "/project",
			wantError:    false,
		},
		{
			name:         "path traversal with dot dot",
			resolvedPath: "/project/../other",
			basePath:     "/project",
			wantError:    true,
			code:         types.ErrPathTraversal,
		},
		{
			name:         "path traversal starting with dot dot",
			resolvedPath: "../etc",
			basePath:     "/project",
			wantError:    true,
			code:         types.ErrPathTraversal,
		},
		{
			name:         "same directory",
			resolvedPath: "/project",
			basePath:     "/project",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPathTraversal(tt.resolvedPath, tt.basePath)
			if (err != nil) != tt.wantError {
				t.Errorf("CheckPathTraversal() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError && err.(*types.CommandError).Code != tt.code {
				t.Errorf("CheckPathTraversal() error code = %v, want %v", err.(*types.CommandError).Code, tt.code)
			}
		})
	}
}

func TestSanitizeWorktreeName(t *testing.T) {
	tests := []struct {
		name        string
		branchName  string
		wantDirName string
		wantDbName  string
	}{
		{
			name:        "simple name",
			branchName:  "feature",
			wantDirName: "feature",
			wantDbName:  "feature",
		},
		{
			name:        "name with slashes",
			branchName:  "feature/new-auth",
			wantDirName: "feature-new-auth",
			wantDbName:  "feature_new_auth",
		},
		{
			name:        "name with dashes and underscores",
			branchName:  "feature/new_auth-test",
			wantDirName: "feature-new_auth-test",
			wantDbName:  "feature_new_auth_test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirName, dbName := SanitizeWorktreeName(tt.branchName)
			if dirName != tt.wantDirName {
				t.Errorf("SanitizeWorktreeName() dirName = %v, want %v", dirName, tt.wantDirName)
			}
			if dbName != tt.wantDbName {
				t.Errorf("SanitizeWorktreeName() dbName = %v, want %v", dbName, tt.wantDbName)
			}
		})
	}
}

func TestIsDatabaseAllowed(t *testing.T) {
	tests := []struct {
		name      string
		dbName    string
		allowed   []string
		wantError bool
		code      types.ErrCode
	}{
		{
			name:      "exact match",
			dbName:    "app",
			allowed:   []string{"app"},
			wantError: false,
		},
		{
			name:      "wildcard match",
			dbName:    "app_test",
			allowed:   []string{"app_*"},
			wantError: false,
		},
		{
			name:      "not in list",
			dbName:    "other_db",
			allowed:   []string{"app", "app_*"},
			wantError: true,
			code:      types.ErrDbNotAllowed,
		},
		{
			name:      "multiple patterns match first",
			dbName:    "app_staging",
			allowed:   []string{"app_*", "staging_*"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsDatabaseAllowed(tt.dbName, tt.allowed)
			if (err != nil) != tt.wantError {
				t.Errorf("IsDatabaseAllowed() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError && err.(*types.CommandError).Code != tt.code {
				t.Errorf("IsDatabaseAllowed() error code = %v, want %v", err.(*types.CommandError).Code, tt.code)
			}
		})
	}
}

func TestIsNotDefaultDB(t *testing.T) {
	tests := []struct {
		name      string
		dbName    string
		defaultDB string
		wantError bool
		code      types.ErrCode
	}{
		{
			name:      "different database",
			dbName:    "app_staging",
			defaultDB: "app",
			wantError: false,
		},
		{
			name:      "same database",
			dbName:    "app",
			defaultDB: "app",
			wantError: true,
			code:      types.ErrDbIsDefault,
		},
		{
			name:      "empty default",
			dbName:    "app",
			defaultDB: "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsNotDefaultDB(tt.dbName, tt.defaultDB)
			if (err != nil) != tt.wantError {
				t.Errorf("IsNotDefaultDB() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError && err.(*types.CommandError).Code != tt.code {
				t.Errorf("IsNotDefaultDB() error code = %v, want %v", err.(*types.CommandError).Code, tt.code)
			}
		})
	}
}
