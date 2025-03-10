package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/catroll/smtpd/config"
	"github.com/emersion/go-smtp"
)

type backend struct {
	cfg     *config.Config
	dataDir string
}

func (bkd *backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &session{
		backend: bkd,
		conn:    c,
	}, nil
}

type session struct {
	backend  *backend
	conn     *smtp.Conn
	from     string
	to       []string
	username string
}

func (s *session) AuthPlain(username, password string) error {
	// TODO: Implement authentication
	s.username = username
	return nil
}

func (s *session) Mail(from string, opts *smtp.MailOptions) error {
	if opts != nil && int64(s.backend.cfg.SMTP.MaxSize) > 0 && opts.Size > int64(s.backend.cfg.SMTP.MaxSize) {
		return fmt.Errorf("message too large, maximum size is %d", s.backend.cfg.SMTP.MaxSize)
	}
	s.from = from
	return nil
}

func (s *session) Rcpt(to string, opts *smtp.RcptOptions) error {
	if len(s.to) >= s.backend.cfg.SMTP.MaxRecipients {
		return fmt.Errorf("too many recipients, maximum is %d", s.backend.cfg.SMTP.MaxRecipients)
	}
	s.to = append(s.to, to)
	return nil
}

func (s *session) Data(r io.Reader) error {
	clientIP := s.conn.Conn().RemoteAddr().String()
	
	mail := &Mail{
		ID:         GenerateID(s.username, clientIP),
		ReceivedAt: time.Now(),
		Username:   s.username,
		MailFrom:   s.from,
		RcptTo:     s.to,
		Data:       r,
		ClientIP:   clientIP,
		Size:       0, // Will be updated after saving
		Extras:     make(map[string]string),
	}

	// Create a unique filename for the message using the mail ID
	filename := filepath.Join(s.backend.dataDir, fmt.Sprintf("%s.eml", mail.ID))

	if err := mail.Save(filename); err != nil {
		return fmt.Errorf("failed to save mail: %w", err)
	}

	// Update mail size after saving
	if fi, err := os.Stat(filename); err == nil {
		mail.Size = fi.Size()
	}

	return nil
}

func (s *session) Reset() {
	s.from = ""
	s.to = nil
}

func (s *session) Logout() error {
	return nil
}
