//go:build linux || darwin || windows

package notification

import (
	"fmt"
	"os/exec"
	"runtime"
)

// osNotificationChannel sends notifications via OS-native notification systems
type osNotificationChannel struct {
	config       *OSNotificationConfig
	executor     CommandExecutor
	platform     string
	sendCallback func(Notification)
}

// NewOSNotificationChannel creates a new OS notification channel
func NewOSNotificationChannel(cfg *OSNotificationConfig, opts ...Option) NotificationChannel {
	ch := &osNotificationChannel{
		config:   cfg,
		platform: runtime.GOOS,
	}

	for _, opt := range opts {
		opt(ch)
	}

	if ch.executor == nil {
		ch.executor = &realCommandExecutor{}
	}

	return ch
}

// Send sends a notification via the OS notification system
func (c *osNotificationChannel) Send(n Notification) error {
	// Check if this notification type is enabled
	if !c.shouldSend(n.Type) {
		return nil
	}

	// Call send callback if set (for testing)
	if c.sendCallback != nil {
		c.sendCallback(n)
	}

	switch c.platform {
	case "linux":
		return c.sendLinux(n)
	case "darwin":
		return c.sendDarwin(n)
	case "windows":
		return c.sendWindows(n)
	default:
		return fmt.Errorf("unsupported platform: %s", c.platform)
	}
}

// shouldSend checks if the notification type should be sent based on config
func (c *osNotificationChannel) shouldSend(t NotificationType) bool {
	switch t {
	case NotifySyncComplete:
		return c.config.OnSyncComplete
	case NotifySyncError:
		return c.config.OnSyncError
	case NotifyConflict:
		return c.config.OnConflict
	case NotifyTest:
		return true // Always send test notifications
	default:
		return true
	}
}

// sendLinux sends notification using notify-send
func (c *osNotificationChannel) sendLinux(n Notification) error {
	return c.executor.Execute("notify-send", n.Title, n.Message)
}

// sendDarwin sends notification using osascript
func (c *osNotificationChannel) sendDarwin(n Notification) error {
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, n.Message, n.Title)
	return c.executor.Execute("osascript", "-e", script)
}

// sendWindows sends notification using PowerShell
func (c *osNotificationChannel) sendWindows(n Notification) error {
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$notification = New-Object System.Windows.Forms.NotifyIcon
$notification.Icon = [System.Drawing.SystemIcons]::Information
$notification.BalloonTipTitle = "%s"
$notification.BalloonTipText = "%s"
$notification.Visible = $true
$notification.ShowBalloonTip(5000)
`, n.Title, n.Message)
	return c.executor.Execute("powershell", "-Command", script)
}

// Close cleans up resources
func (c *osNotificationChannel) Close() error {
	return nil
}

// realCommandExecutor executes real system commands
type realCommandExecutor struct{}

// Execute runs a command
func (e *realCommandExecutor) Execute(cmd string, args ...string) error {
	return exec.Command(cmd, args...).Run()
}
