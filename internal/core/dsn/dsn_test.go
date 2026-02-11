package dsn

import (
	"testing"

	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
)

func TestParseDSN(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      *types.DSN
		expectedError bool
	}{
		{
			name:  "MySQL DSN",
			input: "mysql://root:secret@database:3306/app",
			expected: &types.DSN{
				Engine:   "mysql",
				User:     "root",
				Password: "secret",
				Host:     "database",
				Port:     "3306",
				Database: "app",
			},
		},
		{
			name:  "MariaDB via serverVersion",
			input: "mysql://root:secret@db:3306/app?serverVersion=mariadb-11.4",
			expected: &types.DSN{
				Engine:        "mariadb",
				User:          "root",
				Password:      "secret",
				Host:          "db",
				Port:          "3306",
				Database:      "app",
				ServerVersion: "mariadb-11.4",
			},
		},
		{
			name:  "MySQL without port",
			input: "mysql://root:secret@localhost/app",
			expected: &types.DSN{
				Engine:   "mysql",
				User:     "root",
				Password: "secret",
				Host:     "localhost",
				Port:     "3306",
				Database: "app",
			},
		},
		{
			name:  "PostgreSQL DSN",
			input: "postgresql://user:pass@host:5432/mydb",
			expected: &types.DSN{
				Engine:   "postgresql",
				User:     "user",
				Password: "pass",
				Host:     "host",
				Port:     "5432",
				Database: "mydb",
			},
		},
		{
			name:  "PostgreSQL without port",
			input: "postgres://user:pass@localhost/mydb",
			expected: &types.DSN{
				Engine:   "postgresql",
				User:     "user",
				Password: "pass",
				Host:     "localhost",
				Port:     "5432",
				Database: "mydb",
			},
		},
		{
			name:          "Empty DSN",
			input:         "",
			expectedError: true,
		},
		{
			name:  "MySQL without password",
			input: "mysql://root@localhost/app",
			expected: &types.DSN{
				Engine:   "mysql",
				User:     "root",
				Password: "",
				Host:     "localhost",
				Port:     "3306",
				Database: "app",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDSN(tt.input)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Engine != tt.expected.Engine {
				t.Errorf("expected engine %s, got %s", tt.expected.Engine, result.Engine)
			}
			if result.User != tt.expected.User {
				t.Errorf("expected user %s, got %s", tt.expected.User, result.User)
			}
			if result.Password != tt.expected.Password {
				t.Errorf("expected password %s, got %s", tt.expected.Password, result.Password)
			}
			if result.Host != tt.expected.Host {
				t.Errorf("expected host %s, got %s", tt.expected.Host, result.Host)
			}
			if result.Port != tt.expected.Port {
				t.Errorf("expected port %s, got %s", tt.expected.Port, result.Port)
			}
			if result.Database != tt.expected.Database {
				t.Errorf("expected database %s, got %s", tt.expected.Database, result.Database)
			}
			if result.ServerVersion != tt.expected.ServerVersion {
				t.Errorf("expected serverVersion %s, got %s", tt.expected.ServerVersion, result.ServerVersion)
			}
		})
	}
}

func TestDetermineEngine(t *testing.T) {
	tests := []struct {
		scheme        string
		serverVersion string
		expected      string
	}{
		{"mysql", "", "mysql"},
		{"mysql", "mariadb-10.5", "mariadb"},
		{"mysql", "8.0", "mysql"},
		{"postgresql", "", "postgresql"},
		{"postgres", "", "postgresql"},
		{"unknown", "", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.scheme+"_"+tt.serverVersion, func(t *testing.T) {
			result := determineEngine(tt.scheme, tt.serverVersion)
			if result != tt.expected {
				t.Errorf("determineEngine(%s, %s) = %s, want %s", tt.scheme, tt.serverVersion, result, tt.expected)
			}
		})
	}
}
