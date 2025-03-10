package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load 从文件加载配置
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return cfg, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证服务器配置
	if c.Server.Host == "" {
		return fmt.Errorf("server host is required")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port")
	}
	if c.Server.InstanceName == "" {
		return fmt.Errorf("server instance name is required")
	}

	// 验证 SMTP 配置
	if c.SMTP.Hostname == "" {
		return fmt.Errorf("smtp hostname is required")
	}
	if c.SMTP.MaxSize <= 0 {
		return fmt.Errorf("invalid smtp max size")
	}
	if c.SMTP.MaxRecipients <= 0 {
		return fmt.Errorf("invalid smtp max recipients")
	}
	if !c.SMTP.AllowAnonymous && c.SMTP.AuthFile == "" {
		return fmt.Errorf("auth file is required when anonymous access is disabled")
	}
	if c.SMTP.AuthFile != "" {
		if _, err := os.Stat(c.SMTP.AuthFile); err != nil {
			return fmt.Errorf("auth file not found: %w", err)
		}
	}

	// 验证存储配置
	if c.Storage.Path == "" {
		return fmt.Errorf("storage path is required")
	}

	// 创建存储目录
	if err := os.MkdirAll(c.Storage.Path, 0755); err != nil {
		return fmt.Errorf("creating storage directory: %w", err)
	}

	// 验证 TLS 配置
	if c.TLS.Enabled {
		if c.TLS.CertFile == "" || c.TLS.KeyFile == "" {
			return fmt.Errorf("cert file and key file are required when TLS is enabled")
		}
		if _, err := os.Stat(c.TLS.CertFile); err != nil {
			return fmt.Errorf("cert file not found: %w", err)
		}
		if _, err := os.Stat(c.TLS.KeyFile); err != nil {
			return fmt.Errorf("key file not found: %w", err)
		}
	}

	// 验证日志配置
	if c.Log.Level != "" {
		switch c.Log.Level {
		case "debug", "info", "warn", "error":
		default:
			return fmt.Errorf("invalid log level: %s", c.Log.Level)
		}
	}
	if c.Log.Format != "" {
		switch c.Log.Format {
		case "text", "json":
		default:
			return fmt.Errorf("invalid log format: %s", c.Log.Format)
		}
	}
	if c.Log.File != "" {
		logDir := filepath.Dir(c.Log.File)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("creating log directory: %w", err)
		}
	}

	return nil
}
