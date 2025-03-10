package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	SMTP struct {
		Hostname      string `yaml:"hostname"`
		MaxSize      int    `yaml:"max_size"`       // Maximum message size in bytes
		MaxRecipients int   `yaml:"max_recipients"` // Maximum number of recipients per message
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
	cfg.SMTP.Hostname = "localhost"
	cfg.SMTP.MaxSize = 10 * 1024 * 1024 // 10MB
	cfg.SMTP.MaxRecipients = 100
	cfg.Storage.Path = "./maildata"
	cfg.TLS.Enabled = false
	
	return cfg
}

// Load reads the configuration from the specified file
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
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Server.Port)
	}

	if c.SMTP.MaxSize < 1 {
		return fmt.Errorf("max_size must be positive")
	}

	if c.SMTP.MaxRecipients < 1 {
		return fmt.Errorf("max_recipients must be positive")
	}

	if c.Storage.Path == "" {
		return fmt.Errorf("storage path cannot be empty")
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(c.Storage.Path, 0755); err != nil {
		return fmt.Errorf("creating storage directory: %w", err)
	}

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
