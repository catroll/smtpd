package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary test config file
	content := []byte(`
server:
  host: "127.0.0.1"
  port: 2525

smtp:
  hostname: "test.local"
  max_size: 5242880
  max_recipients: 50

storage:
  path: "./testdata"

tls:
  enabled: false
`)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test loaded values
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Expected host 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 2525 {
		t.Errorf("Expected port 2525, got %d", cfg.Server.Port)
	}
	if cfg.SMTP.Hostname != "test.local" {
		t.Errorf("Expected hostname test.local, got %s", cfg.SMTP.Hostname)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "Valid default config",
			config:  New(),
			wantErr: false,
		},
		{
			name: "Invalid port",
			config: func() *Config {
				cfg := New()
				cfg.Server.Port = 70000
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "Invalid max size",
			config: func() *Config {
				cfg := New()
				cfg.SMTP.MaxSize = 0
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "Invalid max recipients",
			config: func() *Config {
				cfg := New()
				cfg.SMTP.MaxRecipients = -1
				return cfg
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
