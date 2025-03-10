package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// SetupLogger 设置全局日志配置
func (c *Config) SetupLogger() error {
	// 设置默认值
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "text"
	}

	// 解析日志级别
	var level slog.Level
	switch c.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return fmt.Errorf("invalid log level: %s", c.Log.Level)
	}

	// 创建日志处理器选项
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: c.Log.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 自定义时间格式
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.String(slog.TimeKey, t.Format("2006-01-02 15:04:05.000"))
				}
			}
			return a
		},
	}

	// 创建日志处理器
	var handler slog.Handler
	if c.Log.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// 如果指定了日志文件，创建文件输出
	if c.Log.File != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(c.Log.File)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("creating log directory: %w", err)
		}

		// 打开日志文件
		f, err := os.OpenFile(c.Log.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("opening log file: %w", err)
		}

		// 根据格式创建文件处理器
		if c.Log.Format == "json" {
			handler = slog.NewJSONHandler(f, opts)
		} else {
			handler = slog.NewTextHandler(f, opts)
		}
	}

	// 设置全局日志记录器
	slog.SetDefault(slog.New(handler))
	return nil
}
