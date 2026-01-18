package notification

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// logNotificationChannel writes notifications to a log file
type logNotificationChannel struct {
	config *LogNotificationConfig
	file   *os.File
	mu     sync.Mutex
}

// NewLogNotificationChannel creates a new log notification channel
func NewLogNotificationChannel(cfg *LogNotificationConfig) NotificationChannel {
	return &logNotificationChannel{
		config: cfg,
	}
}

// Send writes a notification to the log file
func (c *logNotificationChannel) Send(n Notification) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureFile(); err != nil {
		return err
	}

	// Format: 2026-01-16T10:30:00Z [SYNC_COMPLETE] Message
	typeStr := strings.ToUpper(string(n.Type))
	line := fmt.Sprintf("%s [%s] %s\n", n.Timestamp.UTC().Format("2006-01-02T15:04:05Z"), typeStr, n.Message)

	_, err := c.file.WriteString(line)
	if err != nil {
		return fmt.Errorf("failed to write notification: %w", err)
	}

	return c.file.Sync()
}

// ensureFile ensures the log file is open
func (c *logNotificationChannel) ensureFile() error {
	if c.file != nil {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(c.config.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Check if rotation is needed
	if err := c.rotateIfNeeded(); err != nil {
		return err
	}

	file, err := os.OpenFile(c.config.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	c.file = file
	return nil
}

// rotateIfNeeded checks if the log file exceeds max size and rotates it
func (c *logNotificationChannel) rotateIfNeeded() error {
	info, err := os.Stat(c.config.Path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	maxBytes := int64(c.config.MaxSizeMB) * 1024 * 1024
	if info.Size() < maxBytes {
		return nil
	}

	// Rotate: rename current file with .old extension
	oldPath := c.config.Path + ".old"
	if err := os.Rename(c.config.Path, oldPath); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	return nil
}

// Close closes the log file
func (c *logNotificationChannel) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.file != nil {
		err := c.file.Close()
		c.file = nil
		return err
	}
	return nil
}

// ReadLog reads and returns all entries from the log file
func ReadLog(path string) ([]string, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var entries []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		entries = append(entries, scanner.Text())
	}

	return entries, scanner.Err()
}

// ClearLog clears the log file
func ClearLog(path string) error {
	// Truncate the file
	return os.WriteFile(path, []byte{}, 0644)
}
