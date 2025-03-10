package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/catroll/smtpd/auth"
	"github.com/catroll/smtpd/config"
	gosmtp "github.com/emersion/go-smtp"
)

var (
	configFile = flag.String("config", "config.yaml", "Path to configuration file")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configFile)
	if err != nil {
		slog.Error("加载配置失败",
			"error", err,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		os.Exit(1)
	}

	// 设置日志
	if err := cfg.SetupLogger(); err != nil {
		slog.Error("设置日志失败",
			"error", err,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		os.Exit(1)
	}

	// 初始化认证器
	authenticator := auth.New()
	if err := authenticator.LoadCredentials(cfg.SMTP.AuthFile); err != nil {
		slog.Error("加载认证信息失败",
			"error", err,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		os.Exit(1)
	}

	// 创建邮件存储目录
	mailDataPath := cfg.Storage.Path
	if !filepath.IsAbs(mailDataPath) {
		mailDataPath = filepath.Join(".", mailDataPath)
	}
	if err := os.MkdirAll(mailDataPath, 0755); err != nil {
		slog.Error("创建存储目录失败",
			"error", err,
			"path", mailDataPath,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		os.Exit(1)
	}

	// 初始化后端
	bkd := NewBackend(cfg, mailDataPath, authenticator)

	// 创建 SMTP 服务器
	s := gosmtp.NewServer(bkd)

	s.Addr = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	s.Domain = cfg.SMTP.Hostname
	s.MaxMessageBytes = int64(cfg.SMTP.MaxSize)
	s.MaxRecipients = cfg.SMTP.MaxRecipients
	s.AllowInsecureAuth = cfg.SMTP.AllowInsecureAuth
	s.EnableSMTPUTF8 = true // 支持 UTF8，以便正确处理中文
	s.Debug = os.Stdout

	if cfg.TLS.Enabled {
		s.TLSConfig = &tls.Config{
			Certificates: make([]tls.Certificate, 1),
		}
		s.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if err != nil {
			slog.Error("加载 TLS 证书失败",
				"error", err,
				"timestamp", time.Now().Format(time.RFC3339Nano),
			)
			os.Exit(1)
		}
	}

	// 记录服务器状态
	slog.Info("SMTP 服务器启动",
		"addr", s.Addr,
		"data_dir", mailDataPath,
		"tls", map[bool]string{true: "已启用", false: "已禁用"}[cfg.TLS.Enabled],
		"insecure_auth", map[bool]string{true: "允许", false: "禁止"}[cfg.SMTP.AllowInsecureAuth],
		"max_size", cfg.SMTP.MaxSize,
		"max_recipients", cfg.SMTP.MaxRecipients,
		"timestamp", time.Now().Format(time.RFC3339Nano),
	)

	if err := s.ListenAndServe(); err != nil {
		slog.Error("服务器启动失败",
			"error", err,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		os.Exit(1)
	}
}
