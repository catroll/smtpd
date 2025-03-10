package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/emersion/go-sasl"
	gosmtp "github.com/emersion/go-smtp"
)

// Session 会话结构体
type Session struct {
	backend       *Backend
	conn          *gosmtp.Conn
	from          string
	to            []string
	sessionID     string
	remoteAddr    string
	authenticated bool
	username      string
}

// NewSession 创建新的会话实例
func NewSession(backend *Backend, conn *gosmtp.Conn, sessionID, remoteAddr string) *Session {
	return &Session{
		backend:    backend,
		conn:       conn,
		sessionID:  sessionID,
		remoteAddr: remoteAddr,
	}
}

// WithAuth 设置认证信息
func (s *Session) WithAuth(username string) *Session {
	s.authenticated = true
	s.username = username
	return s
}

// AuthMechanisms 返回支持的认证机制
func (s *Session) AuthMechanisms() []string {
	return []string{"PLAIN", "LOGIN"}
}

// Auth 实现认证机制
func (s *Session) Auth(mech string, fr sasl.Server) error {
	switch mech {
	case "PLAIN", "LOGIN":
		// 第一次调用 Next，获取用户名
		username, done, err := fr.Next(nil)
		if err != nil {
			return err
		}
		if done {
			return fmt.Errorf("认证未完成就收到了完成信号")
		}

		// 第二次调用 Next，获取密码
		password, done, err := fr.Next(nil)
		if err != nil {
			return err
		}
		if !done {
			return fmt.Errorf("认证完成但未收到完成信号")
		}

		// 确保 username 和 password 正确转换为 string 类型
		usernameStr := string(username)
		passwordStr := string(password)

		if !s.backend.authenticator.Authenticate(usernameStr, passwordStr) {
			slog.Warn("SMTP 认证失败",
				"session_id", s.sessionID,
				"remote_addr", s.remoteAddr,
				"username", usernameStr,
				"auth_method", mech,
				"timestamp", time.Now().Format(time.RFC3339Nano),
			)
			return gosmtp.ErrAuthFailed
		}

		s.authenticated = true
		s.username = usernameStr
		slog.Info("SMTP 认证成功",
			"session_id", s.sessionID,
			"remote_addr", s.remoteAddr,
			"username", usernameStr,
			"auth_method", mech,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		return nil
	default:
		return gosmtp.ErrAuthUnsupported
	}
}

// Mail 设置发件人
func (s *Session) Mail(from string, opts *gosmtp.MailOptions) error {
	if !s.backend.cfg.SMTP.AllowAnonymous && !s.authenticated {
		slog.Warn("未认证的发送尝试",
			"session_id", s.sessionID,
			"remote_addr", s.remoteAddr,
			"from", from,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		return gosmtp.ErrAuthRequired
	}

	s.from = from
	slog.Info("设置发件人",
		"session_id", s.sessionID,
		"remote_addr", s.remoteAddr,
		"from", from,
		"timestamp", time.Now().Format(time.RFC3339Nano),
	)
	return nil
}

// Rcpt 添加收件人
func (s *Session) Rcpt(to string, opts *gosmtp.RcptOptions) error {
	if !s.backend.cfg.SMTP.AllowAnonymous && !s.authenticated {
		slog.Warn("未认证的收件人添加尝试",
			"session_id", s.sessionID,
			"remote_addr", s.remoteAddr,
			"to", to,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		return gosmtp.ErrAuthRequired
	}

	if len(s.to) >= s.backend.cfg.SMTP.MaxRecipients {
		slog.Warn("超出最大收件人数量限制",
			"session_id", s.sessionID,
			"remote_addr", s.remoteAddr,
			"max_recipients", s.backend.cfg.SMTP.MaxRecipients,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		return fmt.Errorf("too many recipients")
	}

	s.to = append(s.to, to)
	slog.Info("添加收件人",
		"session_id", s.sessionID,
		"remote_addr", s.remoteAddr,
		"to", to,
		"recipient_count", len(s.to),
		"timestamp", time.Now().Format(time.RFC3339Nano),
	)
	return nil
}

// Data 处理邮件内容
func (s *Session) Data(r io.Reader) error {
	if !s.backend.cfg.SMTP.AllowAnonymous && !s.authenticated {
		slog.Warn("未认证的数据发送尝试",
			"session_id", s.sessionID,
			"remote_addr", s.remoteAddr,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		return gosmtp.ErrAuthRequired
	}

	// 生成邮件文件名
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s-%s.eml", timestamp, s.sessionID)
	filepath := filepath.Join(s.backend.dataDir, filename)

	// 创建邮件文件
	f, err := os.Create(filepath)
	if err != nil {
		slog.Error("创建邮件文件失败",
			"session_id", s.sessionID,
			"remote_addr", s.remoteAddr,
			"filepath", filepath,
			"error", err.Error(),
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		return err
	}
	defer f.Close()

	// 写入邮件内容
	n, err := io.Copy(f, r)
	if err != nil {
		slog.Error("写入邮件内容失败",
			"session_id", s.sessionID,
			"remote_addr", s.remoteAddr,
			"filepath", filepath,
			"error", err.Error(),
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		return err
	}

	slog.Info("邮件保存成功",
		"session_id", s.sessionID,
		"remote_addr", s.remoteAddr,
		"filepath", filepath,
		"size", n,
		"from", s.from,
		"to", s.to,
		"timestamp", time.Now().Format(time.RFC3339Nano),
	)
	return nil
}

// Reset 重置会话状态
func (s *Session) Reset() {
	s.from = ""
	s.to = nil
	slog.Info("重置会话状态",
		"session_id", s.sessionID,
		"remote_addr", s.remoteAddr,
		"timestamp", time.Now().Format(time.RFC3339Nano),
	)
}

// Logout 处理客户端断开连接
func (s *Session) Logout() error {
	slog.Info("客户端断开连接",
		"session_id", s.sessionID,
		"remote_addr", s.remoteAddr,
		"timestamp", time.Now().Format(time.RFC3339Nano),
	)
	return nil
}
