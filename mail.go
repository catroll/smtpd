package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

func GenerateID(username, clientIP string) string {
	// Create a unique hash combining time, username and client info
	timestamp := time.Now().UnixNano()
	// Hash the username to avoid exposing it directly
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s-%s-%d", username, clientIP, timestamp)))
	hash := base64.URLEncoding.EncodeToString(h.Sum(nil))
	// Use first 16 chars of hash combined with timestamp
	return fmt.Sprintf("%d-%s", timestamp, hash[:16])
}

func (m *Mail) ToEml() (*os.File, error) {
	// Create a temporary file for writing
	tmpFile, err := os.CreateTemp("", "mail-*.eml")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	metadataJSON, err := json.Marshal(m)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write email headers with metadata
	headers := fmt.Sprintf("Received: from [%s]\r\n"+
		"Date: %s\r\n"+
		"Message-ID: <%s>\r\n"+
		"X-SMTPD-DATA: %s\r\n\r\n",
		m.ClientIP,
		m.ReceivedAt.Format(time.RFC1123Z),
		m.ID,
		string(metadataJSON))

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
