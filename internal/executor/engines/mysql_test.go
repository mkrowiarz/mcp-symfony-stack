package engines

import (
	"testing"

	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
)

func TestMySQLEngine_BuildDumpCommand(t *testing.T) {
	tests := []struct {
		name      string
		isMariaDB bool
		dsn       *types.DSN
		tables    []string
		expected  []string
	}{
		{
			name:      "MySQL dump without tables",
			isMariaDB: false,
			dsn: &types.DSN{
				Host:     "database",
				User:     "root",
				Password: "secret",
				Database: "app",
			},
			tables: nil,
			expected: []string{
				"mysqldump", "-h", "database", "-u", "root", "-psecret", "app",
			},
		},
		{
			name:      "MariaDB dump with tables",
			isMariaDB: true,
			dsn: &types.DSN{
				Host:     "db",
				User:     "user",
				Password: "pass",
				Database: "test",
			},
			tables: []string{"users", "posts"},
			expected: []string{
				"mariadb-dump", "-h", "db", "-u", "user", "-ppass", "test", "users", "posts",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewMySQLEngine(tt.isMariaDB)
			result := engine.BuildDumpCommand(tt.dsn, tt.tables)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d args, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("arg[%d] = %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestMySQLEngine_BuildCreateCommand(t *testing.T) {
	engine := NewMySQLEngine(false)
	dsn := &types.DSN{Host: "localhost", User: "root", Password: "secret"}

	result := engine.BuildCreateCommand(dsn, "new_db")

	expected := []string{"mysql", "-h", "localhost", "-u", "root", "-psecret", "-e", "CREATE DATABASE `new_db`"}

	if len(result) != len(expected) {
		t.Errorf("expected %d args, got %d", len(expected), len(result))
		return
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("arg[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestMySQLEngine_BuildImportCommand(t *testing.T) {
	engine := NewMySQLEngine(false)
	dsn := &types.DSN{Host: "db", User: "root", Password: "secret"}

	result := engine.BuildImportCommand(dsn, "app_staging")

	expected := []string{"mysql", "-h", "db", "-u", "root", "-psecret", "app_staging"}

	if len(result) != len(expected) {
		t.Errorf("expected %d args, got %d", len(expected), len(result))
		return
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("arg[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestMySQLEngine_BuildDropCommand(t *testing.T) {
	engine := NewMySQLEngine(false)
	dsn := &types.DSN{Host: "localhost", User: "root", Password: "secret"}

	result := engine.BuildDropCommand(dsn, "old_db")

	expected := []string{"mysql", "-h", "localhost", "-u", "root", "-psecret", "-e", "DROP DATABASE `old_db`"}

	if len(result) != len(expected) {
		t.Errorf("expected %d args, got %d", len(expected), len(result))
		return
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("arg[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestMySQLEngine_Name(t *testing.T) {
	t.Run("MySQL", func(t *testing.T) {
		engine := NewMySQLEngine(false)
		if engine.Name() != "MySQL" {
			t.Errorf("expected MySQL, got %s", engine.Name())
		}
	})

	t.Run("MariaDB", func(t *testing.T) {
		engine := NewMySQLEngine(true)
		if engine.Name() != "MariaDB" {
			t.Errorf("expected MariaDB, got %s", engine.Name())
		}
	})
}
