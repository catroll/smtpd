package auth

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"
)

// Authenticator 处理认证相关的功能
type Authenticator struct {
	mu          sync.RWMutex
	credentials map[string]string
}

// New 创建新的认证器实例
func New() *Authenticator {
	return &Authenticator{
		credentials: make(map[string]string),
	}
}

// LoadCredentials 从文件加载认证信息
func (a *Authenticator) LoadCredentials(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("读取认证文件失败",
			"error", err,
			"file", filename,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if err := json.Unmarshal(data, &a.credentials); err != nil {
		slog.Error("解析认证文件失败",
			"error", err,
			"file", filename,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		return err
	}

	slog.Info("加载认证信息成功",
		"file", filename,
		"count", len(a.credentials),
		"timestamp", time.Now().Format(time.RFC3339Nano),
	)
	return nil
}

// Authenticate 验证用户名和密码
func (a *Authenticator) Authenticate(username, password string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	storedPassword, exists := a.credentials[username]
	if !exists {
		slog.Debug("用户不存在",
			"username", username,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		return false
	}

	if storedPassword != password {
		slog.Debug("密码不匹配",
			"username", username,
			"timestamp", time.Now().Format(time.RFC3339Nano),
		)
		return false
	}

	slog.Debug("认证成功",
		"username", username,
		"timestamp", time.Now().Format(time.RFC3339Nano),
	)
	return true
}
