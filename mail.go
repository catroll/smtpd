package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Mail struct {
	ID         string
	ReceivedAt time.Time
	Username   string
	MailFrom   string
	RcptTo     []string
	Data       io.Reader
	ClientIP   string
	Size       int64
	Extras     map[string]string
}

func GenerateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (m *Mail) ToEml() (*os.File, error) {
	// Create a temporary file for writing
	tmpFile, err := os.CreateTemp("", "mail-*.eml")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	// Write email headers
	headers := fmt.Sprintf("Received: from [%s]\r\n"+
		"Date: %s\r\n"+
		"From: %s\r\n"+
		"To: %s\r\n"+
		"Message-ID: <%s>\r\n\r\n",
		m.ClientIP,
		m.ReceivedAt.Format(time.RFC1123Z),
		m.MailFrom,
		m.RcptTo[0],
		m.ID)

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
