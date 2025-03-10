package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/catroll/smtpd/auth"
	"github.com/catroll/smtpd/config"

	"github.com/emersion/go-smtp"
)

var (
	configFile = flag.String("config", "config.yaml", "Path to configuration file")
)

func init() {
	flag.StringVar(configFile, "c", "config.yaml", "Configuration file path")
}

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize authenticator
	authenticator := auth.New()
	if err := authenticator.LoadCredentials(cfg.SMTP.AuthFile); err != nil {
		log.Fatalf("Failed to load auth credentials: %v", err)
	}

	// Create mail storage directory structure
	mailDataPath := cfg.Storage.Path
	if !filepath.IsAbs(mailDataPath) {
		mailDataPath = filepath.Join(".", mailDataPath)
	}
	if err := os.MkdirAll(mailDataPath, 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	// Initialize backend
	bkd := &backend{
		cfg:           cfg,
		dataDir:       mailDataPath,
		authenticator: authenticator,
	}

	// Create SMTP server
	s := smtp.NewServer(bkd)

	// Configure server
	s.Addr = fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	s.Domain = cfg.SMTP.Hostname
	s.MaxMessageBytes = int64(cfg.SMTP.MaxSize)
	s.MaxRecipients = cfg.SMTP.MaxRecipients
	s.AllowInsecureAuth = !cfg.TLS.Enabled
	s.EnableSMTPUTF8 = true // 支持 UTF8，以便正确处理中文
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
	log.Printf("Mail data directory: %s", mailDataPath)
	if cfg.TLS.Enabled {
		log.Printf("TLS is enabled")
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
