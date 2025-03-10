package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Mail struct {
	ID         string            `json:"id"`
	ReceivedAt time.Time         `json:"received_at"`
	Username   string            `json:"username"`
	MailFrom   string            `json:"mail_from"`
	RcptTo     []string          `json:"rcpt_to"`
	Data       io.Reader         `json:"-"`
	ClientIP   string            `json:"client_ip"`
	Size       int64             `json:"size"`
	Extras     map[string]string `json:"extras,omitempty"`
}

func GenerateRandomString(length int) ([]byte, error) {
	randBytes := make([]byte, length)
	_, err := rand.Read(randBytes)
	if err != nil {
		return nil, err
	}
	return randBytes, nil
}

func GenerateID(instanceName, username string) (string, error) {
	timestamp := time.Now().UnixNano()
	randStr, err := GenerateRandomString(8)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(append([]byte(fmt.Sprintf("%s-%s-", instanceName, username)), randStr...))
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	hash := enc.EncodeToString(h.Sum(nil))
	fmt.Println(hash)

	return fmt.Sprintf("%d-%s", timestamp, hash[:16]), nil
}

func (m *Mail) ToEml() (*os.File, error) {
	// Create a temporary file for writing
	tmpFile, err := os.CreateTemp("", "mail-*.eml")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	// Convert mail metadata to JSON
	metadataJSON, err := json.Marshal(m)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Get JSON length and content
	jsonData := string(metadataJSON)

	// Format current time in PST for Received header
	loc, _ := time.LoadLocation("America/Los_Angeles")
	pstTime := m.ReceivedAt.In(loc)
	timeStr := pstTime.Format("Mon, 2 Jan 2006 15:04:05 -0700 (PST)")

	// Get client hostname if available
	clientHost := m.ClientIP
	if names, err := net.LookupAddr(m.ClientIP); err == nil && len(names) > 0 {
		clientHost = strings.TrimSuffix(names[0], ".")
	}

	// Build TLS info if available
	tlsInfo := ""
	if tlsConn, ok := m.Extras["tls_conn"]; ok {
		tlsInfo = fmt.Sprintf("\n        (version=%s cipher=%s bits=%s)",
			tlsConn,
			m.Extras["tls_cipher"],
			m.Extras["tls_bits"])
	}

	// Write email headers with metadata
	headers := fmt.Sprintf("X-SMTPD-DATA: %s\r\n"+
		"Received: from %s (%s [%s])\r\n"+
		"        by %s with ESMTP%s id %s\r\n"+
		"        for <%s>%s;\r\n"+
		"        %s\r\n",
		jsonData,
		clientHost, clientHost, m.ClientIP,
		m.Extras["server_name"],
		strings.ToUpper(m.Extras["protocol"]),
		m.ID,
		strings.Join(m.RcptTo, ", "),
		tlsInfo,
		timeStr)

	if _, err := tmpFile.WriteString(headers); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to write headers: %w", err)
	}

	return tmpFile, nil
}

func (m *Mail) Save(targetPath string) error {
	// Create temporary file with .eml format
	tmpFile, err := m.ToEml()
	if err != nil {
		return err
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	// Copy email body
	if _, err := io.Copy(tmpFile, m.Data); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	// Ensure the target directory exists
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Sync the temporary file to ensure all data is written
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	// Close the temporary file before moving
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Atomically move the temporary file to the target location
	if err := os.Rename(tmpFile.Name(), targetPath); err != nil {
		return fmt.Errorf("failed to move file to final location: %w", err)
	}

	return nil
}
