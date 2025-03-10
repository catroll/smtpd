package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"time"

	"github.com/catroll/smtpd/auth"
	"github.com/catroll/smtpd/config"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

type backend struct {
	cfg           *config.Config
	dataDir       string
	authenticator *auth.Authenticator
}

func (bkd *backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &session{
		backend: bkd,
		conn:    c,
	}, nil
}

func (bkd *backend) AnonymousLogin() (smtp.Session, error) {
	return nil, smtp.ErrAuthRequired
}

func (bkd *backend) Login(username, password string) (smtp.Session, error) {
	if !bkd.authenticator.Authenticate(username, password) {
		return nil, smtp.ErrAuthFailed
	}

	return &session{
		backend:  bkd,
		conn:     nil, // Will be set by NewSession
		username: username,
	}, nil
}

type session struct {
	backend  *backend
	conn     *smtp.Conn
	from     string
	to       []string
	username string
	mailID   string
}

func (s *session) AuthMechanisms() []string {
	return []string{"PLAIN", "LOGIN"}
}

func (s *session) Auth(mech string) (sasl.Server, error) {
	switch mech {
	case "PLAIN":
		return sasl.NewPlainServer(func(identity, username, password string) error {
			if !s.backend.authenticator.Authenticate(username, password) {
				return smtp.ErrAuthFailed
			}
			s.username = username
			return nil
		}), nil
	case "LOGIN":
		return sasl.NewLoginServer(func(username, password string) error {
			if !s.backend.authenticator.Authenticate(username, password) {
				return smtp.ErrAuthFailed
			}
			s.username = username
			return nil
		}), nil
	default:
		return nil, smtp.ErrAuthUnsupported
	}
}

func (s *session) Mail(from string, opts *smtp.MailOptions) error {
	if !s.backend.cfg.Server.AllowAnonymous && s.username == "" {
		return smtp.ErrAuthRequired
	}
	fmt.Println("Mail from:", from)
	fmt.Println("Mail options:", opts)
	fmt.Println("Username:", s.username)
	fmt.Println("AllowAnonymous:", s.backend.cfg.Server.AllowAnonymous)

	if opts != nil && int64(s.backend.cfg.SMTP.MaxSize) > 0 && opts.Size > int64(s.backend.cfg.SMTP.MaxSize) {
		return fmt.Errorf("message too large, maximum size is %d", s.backend.cfg.SMTP.MaxSize)
	}
	s.from = from
	return nil
}

func (s *session) Rcpt(to string, opts *smtp.RcptOptions) error {
	if !s.backend.cfg.Server.AllowAnonymous && s.username == "" {
		return smtp.ErrAuthRequired
	}

	if len(s.to) >= s.backend.cfg.SMTP.MaxRecipients {
		return fmt.Errorf("too many recipients, maximum is %d", s.backend.cfg.SMTP.MaxRecipients)
	}
	s.to = append(s.to, to)
	return nil
}

func (s *session) Data(r io.Reader) error {
	if !s.backend.cfg.Server.AllowAnonymous && s.username == "" {
		return smtp.ErrAuthRequired
	}

	clientIP := ""
	if addr, ok := s.conn.Conn().RemoteAddr().(*net.TCPAddr); ok {
		clientIP = addr.IP.String()
	}

	mailID, err := GenerateID(s.backend.cfg.Server.InstanceName, s.username)
	if err != nil {
		return fmt.Errorf("failed to generate ID: %w", err)
	}

	// Get TLS connection details if available
	extras := make(map[string]string)
	extras["server_name"] = s.backend.cfg.SMTP.Hostname
	extras["protocol"] = "s" // Default to ESMTPS

	if tlsConn, ok := s.conn.Conn().(*tls.Conn); ok {
		state := tlsConn.ConnectionState()
		switch state.Version {
		case tls.VersionTLS10:
			extras["tls_conn"] = "TLS1_0"
		case tls.VersionTLS11:
			extras["tls_conn"] = "TLS1_1"
		case tls.VersionTLS12:
			extras["tls_conn"] = "TLS1_2"
		case tls.VersionTLS13:
			extras["tls_conn"] = "TLS1_3"
		}

		// Map cipher suite to string
		switch state.CipherSuite {
		case tls.TLS_AES_128_GCM_SHA256:
			extras["tls_cipher"] = "TLS_AES_128_GCM_SHA256"
			extras["tls_bits"] = "128/128"
		case tls.TLS_AES_256_GCM_SHA384:
			extras["tls_cipher"] = "TLS_AES_256_GCM_SHA384"
			extras["tls_bits"] = "256/256"
		case tls.TLS_CHACHA20_POLY1305_SHA256:
			extras["tls_cipher"] = "TLS_CHACHA20_POLY1305_SHA256"
			extras["tls_bits"] = "256/256"
		default:
			extras["tls_cipher"] = fmt.Sprintf("0x%04x", state.CipherSuite)
			extras["tls_bits"] = "128/128" // Default to most common
		}
	} else {
		extras["protocol"] = "" // Plain ESMTP
	}

	mail := &Mail{
		ID:         mailID,
		ReceivedAt: time.Now(),
		Username:   s.username,
		MailFrom:   s.from,
		RcptTo:     s.to,
		Data:       r,
		ClientIP:   clientIP,
		Size:       0, // Will be updated after saving
		Extras:     extras,
	}

	// Create a unique filename for the message using the mail ID
	filename := filepath.Join(s.backend.dataDir, fmt.Sprintf("%s.eml", mail.ID))

	if err := mail.Save(filename); err != nil {
		return fmt.Errorf("failed to save mail: %w", err)
	}

	s.mailID = mail.ID
	return &smtp.SMTPError{Code: 250, EnhancedCode: smtp.EnhancedCode{2, 0, 0}, Message: fmt.Sprintf("Message %s accepted for delivery", mail.ID)}
}

func (s *session) Reset() {
	s.from = ""
	s.to = nil
}

func (s *session) Logout() error {
	return nil
}
