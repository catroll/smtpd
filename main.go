package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/catroll/smtpd/config"
	"github.com/emersion/go-smtp"
)

var (
	configFile = "./config.yaml"
)

func init() {
	flag.StringVar(&configFile, "c", configFile, "Configuration file path")
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
