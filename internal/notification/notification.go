// Package notification provides notification services for background sync events.
package notification

import (
	"time"
)

// NotificationType identifies the type of notification
type NotificationType string

const (
	NotifySyncComplete NotificationType = "sync_complete"
	NotifySyncError    NotificationType = "sync_error"
	NotifyConflict     NotificationType = "conflict"
	NotifyReminder     NotificationType = "reminder"
	NotifyTest         NotificationType = "test"
)

// Notification represents a notification to be sent
type Notification struct {
	Type      NotificationType
	Title     string
	Message   string
	Timestamp time.Time
	Metadata  map[string]string
}

// NotificationManager is the interface for managing notifications
type NotificationManager interface {
	Send(n Notification) error
	SendAsync(n Notification)
	Close() error
	ChannelCount() int
}

// NotificationChannel is the interface for a notification channel
type NotificationChannel interface {
	Send(n Notification) error
	Close() error
}

// Config holds the notification configuration
type Config struct {
	Enabled         bool
	OSNotification  OSNotificationConfig
	LogNotification LogNotificationConfig
}

// OSNotificationConfig holds OS notification configuration
type OSNotificationConfig struct {
	Enabled        bool
	OnSyncComplete bool
	OnSyncError    bool
	OnConflict     bool
}

// LogNotificationConfig holds log notification configuration
type LogNotificationConfig struct {
	Enabled       bool
	Path          string
	MaxSizeMB     int
	RetentionDays int
}

// CommandExecutor is the interface for executing system commands
type CommandExecutor interface {
	Execute(cmd string, args ...string) error
}

// MockCommandExecutor is a mock implementation of CommandExecutor for testing
type MockCommandExecutor struct {
	ExecuteFunc func(cmd string, args ...string) error
}

// Execute implements CommandExecutor
func (m *MockCommandExecutor) Execute(cmd string, args ...string) error {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(cmd, args...)
	}
	return nil
}

// Option is a functional option for configuring notification channels
type Option func(interface{})

// WithCommandExecutor sets a custom command executor
func WithCommandExecutor(executor CommandExecutor) Option {
	return func(c interface{}) {
		if ch, ok := c.(*osNotificationChannel); ok {
			ch.executor = executor
		}
		if mgr, ok := c.(*manager); ok {
			mgr.commandExecutor = executor
		}
	}
}

// WithPlatform sets the platform for OS notifications
func WithPlatform(platform string) Option {
	return func(c interface{}) {
		if ch, ok := c.(*osNotificationChannel); ok {
			ch.platform = platform
		}
	}
}

// WithSendCallback sets a callback to be called when a notification is sent
func WithSendCallback(callback func(Notification)) Option {
	return func(c interface{}) {
		if ch, ok := c.(*osNotificationChannel); ok {
			ch.sendCallback = callback
		}
		if mgr, ok := c.(*manager); ok {
			mgr.sendCallback = callback
		}
	}
}
