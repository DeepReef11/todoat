// Package daemon provides a background daemon process for async sync operations.
// The daemon runs as a separate process that handles sync operations asynchronously,
// allowing the CLI to return immediately after local operations.
package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Config holds daemon configuration.
type Config struct {
	PIDPath           string        // Path to PID file
	SocketPath        string        // Path to Unix socket
	LogPath           string        // Path to log file
	Interval          time.Duration // Sync interval
	IdleTimeout       time.Duration // Timeout before daemon exits when idle
	HeartbeatInterval time.Duration // Interval for heartbeat updates
	ConfigPath        string        // Path to app config file
	DBPath            string        // Path to database
	CachePath         string        // Path to cache
	Executable        string        // Optional: explicit path to executable (for testing)
}

// Message represents an IPC message between CLI and daemon.
type Message struct {
	Type string `json:"type"` // "notify", "status", "stop"
	Data string `json:"data,omitempty"`
}

// Response represents a daemon response to CLI.
type Response struct {
	Status    string `json:"status"` // "ok", "error"
	Message   string `json:"message,omitempty"`
	SyncCount int    `json:"sync_count,omitempty"`
	LastSync  string `json:"last_sync,omitempty"`
	Running   bool   `json:"running"`
}

// Daemon represents a running daemon process.
type Daemon struct {
	cfg       *Config
	syncCount int
	lastSync  time.Time
	mu        sync.RWMutex
	stopChan  chan struct{}
	listener  net.Listener
	syncFunc  func() error // Function to call for sync operations
}

// New creates a new Daemon instance.
func New(cfg *Config) *Daemon {
	return &Daemon{
		cfg:      cfg,
		stopChan: make(chan struct{}),
	}
}

// SetSyncFunc sets the function to call for sync operations.
func (d *Daemon) SetSyncFunc(f func() error) {
	d.syncFunc = f
}

// Start starts the daemon process. This should be called in the forked process.
func (d *Daemon) Start() error {
	// Write PID file
	if err := os.MkdirAll(filepath.Dir(d.cfg.PIDPath), 0755); err != nil {
		return fmt.Errorf("failed to create PID directory: %w", err)
	}
	if err := os.WriteFile(d.cfg.PIDPath, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Create Unix socket
	if err := os.MkdirAll(filepath.Dir(d.cfg.SocketPath), 0755); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}
	// Remove existing socket file if present
	_ = os.Remove(d.cfg.SocketPath)

	listener, err := net.Listen("unix", d.cfg.SocketPath)
	if err != nil {
		return fmt.Errorf("failed to create Unix socket: %w", err)
	}
	d.listener = listener

	// Set up signal handlers
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Create log file directory
	if err := os.MkdirAll(filepath.Dir(d.cfg.LogPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	d.log("Daemon started (PID: %d, interval: %v)", os.Getpid(), d.cfg.Interval)

	// Start IPC listener
	go d.handleConnections()

	// Start sync loop
	ticker := time.NewTicker(d.cfg.Interval)
	defer ticker.Stop()

	// Idle timer for auto-shutdown
	var idleTimer *time.Timer
	if d.cfg.IdleTimeout > 0 {
		idleTimer = time.NewTimer(d.cfg.IdleTimeout)
	}

	for {
		select {
		case <-sigChan:
			d.log("Received shutdown signal")
			d.cleanup()
			return nil

		case <-d.stopChan:
			d.log("Stop requested via IPC")
			d.cleanup()
			return nil

		case <-ticker.C:
			d.performSync()
			// Reset idle timer
			if idleTimer != nil {
				if !idleTimer.Stop() {
					select {
					case <-idleTimer.C:
					default:
					}
				}
				idleTimer.Reset(d.cfg.IdleTimeout)
			}

		case <-func() <-chan time.Time {
			if idleTimer != nil {
				return idleTimer.C
			}
			return make(chan time.Time) // Never fires
		}():
			d.log("Idle timeout reached, shutting down")
			d.cleanup()
			return nil
		}
	}
}

// Stop signals the daemon to stop.
func (d *Daemon) Stop() {
	close(d.stopChan)
}

func (d *Daemon) handleConnections() {
	for {
		conn, err := d.listener.Accept()
		if err != nil {
			// Check if listener was closed
			select {
			case <-d.stopChan:
				return
			default:
			}
			d.log("Accept error: %v", err)
			continue
		}
		go d.handleConnection(conn)
	}
}

func (d *Daemon) handleConnection(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	// Set read deadline
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var msg Message
	if err := decoder.Decode(&msg); err != nil {
		return
	}

	var resp Response
	switch msg.Type {
	case "notify":
		// Trigger immediate sync
		go d.performSync()
		resp = Response{Status: "ok", Running: true}

	case "status":
		d.mu.RLock()
		resp = Response{
			Status:    "ok",
			Running:   true,
			SyncCount: d.syncCount,
			LastSync:  d.lastSync.Format(time.RFC3339),
		}
		d.mu.RUnlock()

	case "stop":
		resp = Response{Status: "ok", Running: false}
		_ = encoder.Encode(resp)
		d.Stop()
		return

	default:
		resp = Response{Status: "error", Message: "unknown message type"}
	}

	_ = encoder.Encode(resp)
}

func (d *Daemon) performSync() {
	d.mu.Lock()
	d.syncCount++
	count := d.syncCount
	d.mu.Unlock()

	d.log("Starting sync (count: %d)", count)

	if d.syncFunc != nil {
		if err := d.syncFunc(); err != nil {
			d.log("Sync error (count: %d): %v", count, err)
		} else {
			d.log("Sync completed (count: %d)", count)
		}
	}

	d.mu.Lock()
	d.lastSync = time.Now()
	d.mu.Unlock()
}

func (d *Daemon) cleanup() {
	if d.listener != nil {
		_ = d.listener.Close()
	}
	_ = os.Remove(d.cfg.PIDPath)
	_ = os.Remove(d.cfg.SocketPath)
	d.log("Daemon stopped")
}

func (d *Daemon) log(format string, args ...interface{}) {
	entry := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
	f, err := os.OpenFile(d.cfg.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = f.WriteString(entry)
}

// Client provides methods to communicate with a running daemon.
type Client struct {
	socketPath string
}

// NewClient creates a new daemon client.
func NewClient(socketPath string) *Client {
	return &Client{socketPath: socketPath}
}

// Notify sends a notification to the daemon to trigger a sync.
func (c *Client) Notify() error {
	return c.send(Message{Type: "notify"})
}

// Status gets the daemon status.
func (c *Client) Status() (*Response, error) {
	return c.sendAndReceive(Message{Type: "status"})
}

// Stop requests the daemon to stop and waits for confirmation.
func (c *Client) Stop() error {
	_, err := c.sendAndReceive(Message{Type: "stop"})
	return err
}

func (c *Client) send(msg Message) error {
	conn, err := net.DialTimeout("unix", c.socketPath, 500*time.Millisecond)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return json.NewEncoder(conn).Encode(msg)
}

func (c *Client) sendAndReceive(msg Message) (*Response, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, 500*time.Millisecond)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := json.NewEncoder(conn).Encode(msg); err != nil {
		return nil, err
	}

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Fork spawns a new daemon process.
func Fork(cfg *Config) error {
	// Get the path to the executable
	executable := cfg.Executable
	if executable == "" {
		var err error
		executable, err = os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}
	}

	// Build arguments for daemon mode
	args := []string{
		"--daemon-mode",
		"--daemon-pid-path", cfg.PIDPath,
		"--daemon-socket-path", cfg.SocketPath,
		"--daemon-log-path", cfg.LogPath,
		"--daemon-interval", strconv.FormatInt(int64(cfg.Interval.Seconds()), 10),
	}
	if cfg.IdleTimeout > 0 {
		args = append(args, "--daemon-idle-timeout", strconv.FormatInt(int64(cfg.IdleTimeout.Seconds()), 10))
	}
	if cfg.ConfigPath != "" {
		args = append(args, "--config-path", cfg.ConfigPath)
	}
	if cfg.DBPath != "" {
		args = append(args, "--db-path", cfg.DBPath)
	}
	if cfg.CachePath != "" {
		args = append(args, "--cache-path", cfg.CachePath)
	}

	// Create the command
	cmd := exec.Command(executable, args...)

	// Detach from terminal
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Create new session
	}

	// Set environment
	cmd.Env = os.Environ()

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon process: %w", err)
	}

	// Release the process so it can run independently
	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("failed to release daemon process: %w", err)
	}

	return nil
}

// IsRunning checks if a daemon is running by checking the PID file and socket.
func IsRunning(pidPath, socketPath string) bool {
	// Check if PID file exists
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return false
	}

	// Parse PID
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, FindProcess always succeeds, so we need to send signal 0
	// to check if process exists
	if err := process.Signal(syscall.Signal(0)); err != nil {
		// Process doesn't exist, clean up stale PID file
		_ = os.Remove(pidPath)
		_ = os.Remove(socketPath)
		return false
	}

	// Try to connect to socket
	conn, err := net.DialTimeout("unix", socketPath, 100*time.Millisecond)
	if err != nil {
		// Socket not available, process might be hung
		return false
	}
	_ = conn.Close()

	return true
}

// GetSocketPath returns the default socket path.
func GetSocketPath() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir != "" {
		return filepath.Join(runtimeDir, "todoat", "daemon.sock")
	}
	return "/tmp/todoat-daemon.sock"
}

// RunDaemonMode is called when the executable is invoked with --daemon-mode.
// This function runs the daemon and never returns (exits the process).
func RunDaemonMode(ctx context.Context, cfg *Config, syncFunc func() error) {
	d := New(cfg)
	d.SetSyncFunc(syncFunc)
	if err := d.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "daemon error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
