package main

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/catroll/smtpd/auth"
	"github.com/catroll/smtpd/config"
	gosmtp "github.com/emersion/go-smtp"
)

// Backend 实现 smtp.Backend 接口
type Backend struct {
	cfg           *config.Config
	dataDir       string
	authenticator *auth.Authenticator
	conn          *gosmtp.Conn
}

// NewBackend 创建新的后端实例
func NewBackend(cfg *config.Config, dataDir string, authenticator *auth.Authenticator) *Backend {
	slog.Info("创建新的 SMTP 后端",
		"data_dir", dataDir,
		"allow_anonymous", cfg.SMTP.AllowAnonymous,
		"timestamp", time.Now().Format(time.RFC3339Nano),
	)
	return &Backend{
		cfg:           cfg,
		dataDir:       dataDir,
		authenticator: authenticator,
	}
}

// NewSession 创建新的会话
func (b *Backend) NewSession(c *gosmtp.Conn) (gosmtp.Session, error) {
	b.conn = c
	sessionID := fmt.Sprintf("%s-%d", time.Now().Format("20060102150405"), c.Conn().RemoteAddr().(*net.TCPAddr).Port)
	remoteAddr := c.Conn().RemoteAddr().String()

	// 记录 TLS 连接信息
	if tlsConn, ok := c.Conn().(*tls.Conn); ok {
		state := tlsConn.ConnectionState()
		slog.Info("新建 TLS 会话",
			"session_id", sessionID,
			"remote_addr", remoteAddr,
			"tls_version", fmt.Sprintf("%X", state.Version),
			"cipher_suite", fmt.Sprintf("%X", state.CipherSuite),
			"server_name", state.ServerName,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
	} else {
		slog.Info("新建普通会话",
			"session_id", sessionID,
			"remote_addr", remoteAddr,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
	}

	return NewSession(b, c, sessionID, remoteAddr), nil
}

// Login 处理登录请求
func (b *Backend) Login(username, password string) (gosmtp.Session, error) {
	if !b.authenticator.Authenticate(username, password) {
		slog.Warn("登录失败",
			"username", username,
			"remote_addr", b.conn.Conn().RemoteAddr().String(),
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		return nil, gosmtp.ErrAuthFailed
	}

	sessionID := fmt.Sprintf("%s-%d", time.Now().Format("20060102150405"), b.conn.Conn().RemoteAddr().(*net.TCPAddr).Port)
	remoteAddr := b.conn.Conn().RemoteAddr().String()

	slog.Info("登录成功",
		"session_id", sessionID,
		"remote_addr", remoteAddr,
		"username", username,
		"timestamp", time.Now().Format(time.RFC3339Nano),
	)

	return NewSession(b, b.conn, sessionID, remoteAddr).WithAuth(username), nil
}

// AnonymousLogin 处理匿名登录请求
func (b *Backend) AnonymousLogin() (gosmtp.Session, error) {
	if !b.cfg.SMTP.AllowAnonymous {
		slog.Warn("匿名登录被拒绝",
			"remote_addr", b.conn.Conn().RemoteAddr().String(),
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		return nil, gosmtp.ErrAuthRequired
	}

	sessionID := fmt.Sprintf("%s-%d", time.Now().Format("20060102150405"), b.conn.Conn().RemoteAddr().(*net.TCPAddr).Port)
	remoteAddr := b.conn.Conn().RemoteAddr().String()

	slog.Info("匿名登录成功",
		"session_id", sessionID,
		"remote_addr", remoteAddr,
		"timestamp", time.Now().Format(time.RFC3339Nano),
	)

	return NewSession(b, b.conn, sessionID, remoteAddr), nil
}
