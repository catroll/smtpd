package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/catroll/smtpd/config"
	"github.com/emersion/go-smtp"
)

var (
	configFile = "./config.yaml"
)

func init() {
	flag.StringVar(&configFile, "c", configFile, "Configuration file path")
}

type backend struct {
	cfg     *config.Config
	dataDir string
}

func (bkd *backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &session{
		backend: bkd,
	}, nil
}

type session struct {
	backend *backend
	from    string
	to      []string
}

func (s *session) AuthPlain(username, password string) error {
	// TODO: Implement authentication
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
	// Create a unique filename for the message
	filename := filepath.Join(s.backend.dataDir, fmt.Sprintf("%d.eml", os.Getpid()))

	// Create the file
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create mail file: %w", err)
	}
	defer f.Close()

	// Write message headers
	fmt.Fprintf(f, "From: %s\n", s.from)
	fmt.Fprintf(f, "To: %s\n", s.to)
	fmt.Fprintf(f, "\n")

	// Copy the message body
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("failed to write mail content: %w", err)
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

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create mail storage directory if it doesn't exist
	if err := os.MkdirAll(cfg.Storage.Path, 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	// Initialize backend
	bkd := &backend{
		cfg:     cfg,
		dataDir: cfg.Storage.Path,
	}

	// Create SMTP server
	s := smtp.NewServer(bkd)

	// Configure server
	s.Addr = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	s.Domain = cfg.SMTP.Hostname
	s.MaxMessageBytes = int64(cfg.SMTP.MaxSize)
	s.MaxRecipients = cfg.SMTP.MaxRecipients
	s.AllowInsecureAuth = !cfg.TLS.Enabled
	s.Debug = os.Stdout

	// Configure TLS if enabled
	if cfg.TLS.Enabled {
		cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
		if err != nil {
			log.Fatalf("Failed to load TLS certificates: %v", err)
		}
		s.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
	}

	log.Printf("Starting SMTP server at %s", s.Addr)
	if cfg.TLS.Enabled {
		log.Printf("TLS is enabled")
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
