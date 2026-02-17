package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mkrowiarz/mcp-symfony-stack/internal/core/types"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		projectRoot   string
		setupFunc     func() (string, error)
		cleanupFunc   func()
		expectedError types.ErrCode
		checkFunc     func(*testing.T, *Config)
	}{
		{
			name:        "valid config loads successfully",
			projectRoot: "./testdata",
			setupFunc: func() (string, error) {
				configPath := filepath.Join("./testdata", ".haive.json")
				sampleConfig, err := os.ReadFile(filepath.Join("./testdata", "sample-config.json"))
				if err != nil {
					return "", err
				}
				return "", os.WriteFile(configPath, sampleConfig, 0644)
			},
			cleanupFunc: func() {
				os.Remove(filepath.Join("./testdata", ".haive.json"))
			},
			expectedError: "",
		},
		{
			name:        "missing config returns ErrConfigMissing",
			projectRoot: "/nonexistent/path",
			setupFunc: func() (string, error) {
				return "", nil
			},
			cleanupFunc:   func() {},
			expectedError: types.ErrConfigMissing,
		},
		{
			name: "malformed JSON returns ErrConfigInvalid",
			setupFunc: func() (string, error) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".haive.json")
				return tmpDir, os.WriteFile(configPath, []byte("{invalid json"), 0644)
			},
			cleanupFunc:   func() {},
			expectedError: types.ErrConfigInvalid,
		},
		{
			name: "pm namespace in shared .haive.json",
			setupFunc: func() (string, error) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, ".haive.json")
				sharedConfig := `{
					"project": "other-tool",
					"agents": ["claude"],
					"pm": {
						"project": {
							"name": "shared-project",
							"type": "symfony"
						},
						"docker": {
							"compose_files": ["docker-compose.yml"]
						},
						"database": {
							"service": "db",
							"dsn": "mysql://root:pass@db/app",
							"allowed": ["app", "app_*"]
						}
					}
				}`
				return tmpDir, os.WriteFile(configPath, []byte(sharedConfig), 0644)
			},
			cleanupFunc:   func() {},
			expectedError: "",
			checkFunc: func(t *testing.T, cfg *Config) {
				if cfg.Database == nil {
					t.Error("expected Database section, got nil")
					return
				}
				if cfg.Database.Service != "db" {
					t.Errorf("expected database service 'db', got '%s'", cfg.Database.Service)
				}
			},
		},
		{
			name: "pm namespace in .haive/config.json",
			setupFunc: func() (string, error) {
				tmpDir := t.TempDir()
				haiveDir := filepath.Join(tmpDir, ".haive")
				if err := os.MkdirAll(haiveDir, 0755); err != nil {
					return "", err
				}
				configPath := filepath.Join(haiveDir, "config.json")
				sharedConfig := `{
					"project": "other-tool",
					"pm": {
						"docker": {
							"compose_files": ["compose.yml"]
						}
					}
				}`
				return tmpDir, os.WriteFile(configPath, []byte(sharedConfig), 0644)
			},
			cleanupFunc:   func() {},
			expectedError: "",
			checkFunc: func(t *testing.T, cfg *Config) {
				if cfg.Docker.ComposeFiles == nil {
					t.Error("expected Docker ComposeFiles, got nil")
					return
				}
				if len(cfg.Docker.ComposeFiles) != 1 || cfg.Docker.ComposeFiles[0] != "compose.yml" {
					t.Errorf("expected compose_files ['compose.yml'], got %v", cfg.Docker.ComposeFiles)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				root, err := tt.setupFunc()
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				if root != "" {
					tt.projectRoot = root
				}
			}
			if tt.cleanupFunc != nil {
				defer tt.cleanupFunc()
			}

			cfg, err := Load(tt.projectRoot)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %s, got nil", tt.expectedError)
					return
				}
				cmdErr, ok := err.(*types.CommandError)
				if !ok {
					t.Errorf("expected CommandError, got %T", err)
					return
				}
				if cmdErr.Code != tt.expectedError {
					t.Errorf("expected error code %s, got %s", tt.expectedError, cmdErr.Code)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if cfg == nil {
				t.Error("expected config, got nil")
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, cfg)
				return
			}

			if cfg.Docker == nil {
				t.Error("expected Docker section, got nil")
			} else {
				if len(cfg.Docker.ComposeFiles) != 1 || cfg.Docker.ComposeFiles[0] != "docker-compose.yaml" {
					t.Errorf("expected compose_files ['docker-compose.yaml'], got '%v'", cfg.Docker.ComposeFiles)
				}
			}
		})
	}
}
