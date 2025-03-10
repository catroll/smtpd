package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Authenticator handles SMTP authentication
type Authenticator struct {
	credentials map[string]string
	mu         sync.RWMutex
}

// New creates a new Authenticator instance
func New() *Authenticator {
	return &Authenticator{
		credentials: make(map[string]string),
	}
}

// LoadCredentials loads credentials from a file in username:password format
func (a *Authenticator) LoadCredentials(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("opening auth file: %w", err)
	}
	defer file.Close()

	a.mu.Lock()
	defer a.mu.Unlock()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid format at line %d: expected username:password", lineNum)
		}

		username := strings.TrimSpace(parts[0])
		password := strings.TrimSpace(parts[1])

		if username == "" || password == "" {
			return fmt.Errorf("invalid format at line %d: username and password cannot be empty", lineNum)
		}

		a.credentials[username] = password
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading auth file: %w", err)
	}

	return nil
}

// Authenticate checks if the provided credentials are valid
func (a *Authenticator) Authenticate(username, password string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	storedPassword, exists := a.credentials[username]
	return exists && storedPassword == password
}
