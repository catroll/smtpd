package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server struct {
		Host           string `yaml:"host"`
		Port           int    `yaml:"port"`
		InstanceName   string `yaml:"instance_name"`
		AllowAnonymous bool   `yaml:"allow_anonymous"`
	} `yaml:"server"`
	SMTP struct {
		Hostname          string `yaml:"hostname"`
		MaxSize           int    `yaml:"max_size"`            // Maximum message size in bytes
		MaxRecipients     int    `yaml:"max_recipients"`      // Maximum number of recipients per message
		AuthFile          string `yaml:"auth_file"`           // Path to auth.txt file
		AllowInsecureAuth bool   `yaml:"allow_insecure_auth"` // Allow authentication without TLS
	} `yaml:"smtp"`
	Storage struct {
		Path string `yaml:"path"` // Path to store mail data
	} `yaml:"storage"`
	TLS struct {
		Enabled  bool   `yaml:"enabled"`
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
	} `yaml:"tls"`
}

// New creates a new Config with default values
func New() *Config {
	cfg := &Config{}

	// Set default values
	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = 25
	cfg.Server.InstanceName = "smtpd"
	cfg.Server.AllowAnonymous = false
	cfg.SMTP.Hostname = "localhost"
	cfg.SMTP.MaxSize = 10 * 1024 * 1024 // 10MB
	cfg.SMTP.MaxRecipients = 100
	cfg.SMTP.AuthFile = "./auth.txt"
	cfg.SMTP.AllowInsecureAuth = false // 默认不允许非 TLS 认证
	cfg.Storage.Path = "./maildata"
	cfg.TLS.Enabled = false

	return cfg
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	config := New()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate server configuration
	if c.Server.Host == "" {
		return fmt.Errorf("server host is required")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	if c.Server.InstanceName == "" {
		return fmt.Errorf("server instance name is required")
	}

	// Validate SMTP configuration
	if c.SMTP.Hostname == "" {
		return fmt.Errorf("SMTP hostname is required")
	}
	if c.SMTP.MaxSize <= 0 {
		return fmt.Errorf("invalid max message size: %d", c.SMTP.MaxSize)
	}
	if c.SMTP.MaxRecipients <= 0 {
		return fmt.Errorf("invalid max recipients: %d", c.SMTP.MaxRecipients)
	}
	if c.SMTP.AuthFile == "" {
		return fmt.Errorf("auth file path is required")
	}

	// Check if auth file exists
	if _, err := os.Stat(c.SMTP.AuthFile); err != nil {
		return fmt.Errorf("auth file not found: %s", c.SMTP.AuthFile)
	}

	// Validate storage configuration
	if c.Storage.Path == "" {
		return fmt.Errorf("storage path is required")
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(c.Storage.Path, 0755); err != nil {
		return fmt.Errorf("creating storage directory: %w", err)
	}

	// Validate TLS configuration if enabled
	if c.TLS.Enabled {
		if c.TLS.CertFile == "" || c.TLS.KeyFile == "" {
			return fmt.Errorf("TLS enabled but cert_file or key_file not specified")
		}

		// Check if TLS files exist
		if _, err := os.Stat(c.TLS.CertFile); err != nil {
			return fmt.Errorf("cert file not found: %s", c.TLS.CertFile)
		}
		if _, err := os.Stat(c.TLS.KeyFile); err != nil {
			return fmt.Errorf("key file not found: %s", c.TLS.KeyFile)
		}
	}

	return nil
}
