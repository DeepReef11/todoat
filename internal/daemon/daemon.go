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
	Status        string                    `json:"status"` // "ok", "error"
	Message       string                    `json:"message,omitempty"`
	SyncCount     int                       `json:"sync_count,omitempty"`
	LastSync      string                    `json:"last_sync,omitempty"`
	Running       bool                      `json:"running"`
	BackendStates map[string]*BackendStatus `json:"backend_states,omitempty"` // Per-backend status (Issue #40)
}

// BackendStatus represents the status of a backend for API responses.
type BackendStatus struct {
	SyncCount  int    `json:"sync_count"`
	ErrorCount int    `json:"error_count"`
	LastSync   string `json:"last_sync,omitempty"`
	LastError  string `json:"last_error,omitempty"`
	Healthy    bool   `json:"healthy"`
}

// BackendState holds the per-backend sync state (Issue #40).
type BackendState struct {
	Name       string    // Backend name
	SyncCount  int       // Number of syncs performed
	ErrorCount int       // Number of consecutive errors
	LastSync   time.Time // Last successful sync time
	LastError  string    // Last error message (if any)
}

// backendEntry holds a backend's sync function and configuration.
type backendEntry struct {
	name     string
	syncFunc func() error
	interval time.Duration // Per-backend interval (0 = use global interval)
	lastSync time.Time     // When this backend was last synced
}

// Daemon represents a running daemon process.
type Daemon struct {
	cfg       *Config
	syncCount int
	lastSync  time.Time
	mu        sync.RWMutex
	syncMu    sync.Mutex // Serializes performSync calls (Issue #52)
	stopChan  chan struct{}
	listener  net.Listener
	syncFunc  func() error // Function to call for sync operations (legacy single-backend)

	// Multi-backend support (Issue #40)
	backends      []*backendEntry          // List of backends with their sync functions
	backendStates map[string]*BackendState // Per-backend state tracking
	backendsMu    sync.RWMutex             // Protects backends and backendStates
}

// New creates a new Daemon instance.
func New(cfg *Config) *Daemon {
	return &Daemon{
		cfg:           cfg,
		stopChan:      make(chan struct{}),
		backends:      make([]*backendEntry, 0),
		backendStates: make(map[string]*BackendState),
	}
}

// SetSyncFunc sets the function to call for sync operations.
// This is the legacy single-backend API. For multi-backend support,
// use AddBackendSyncFunc instead.
func (d *Daemon) SetSyncFunc(f func() error) {
	d.syncFunc = f
}

// AddBackendSyncFunc adds a backend sync function using the global interval.
// This enables multi-backend sync support (Issue #40).
func (d *Daemon) AddBackendSyncFunc(name string, syncFunc func() error) {
	d.AddBackendSyncFuncWithInterval(name, 0, syncFunc)
}

// AddBackendSyncFuncWithInterval adds a backend sync function with a custom interval.
// If interval is 0, the global daemon interval is used.
func (d *Daemon) AddBackendSyncFuncWithInterval(name string, interval time.Duration, syncFunc func() error) {
	d.backendsMu.Lock()
	defer d.backendsMu.Unlock()

	entry := &backendEntry{
		name:     name,
		syncFunc: syncFunc,
		interval: interval,
	}
	d.backends = append(d.backends, entry)

	// Initialize state for this backend
	d.backendStates[name] = &BackendState{
		Name: name,
	}
}

// GetBackendState returns the state for a specific backend.
// Returns nil if the backend is not registered.
func (d *Daemon) GetBackendState(name string) *BackendState {
	d.backendsMu.RLock()
	defer d.backendsMu.RUnlock()
	return d.backendStates[name]
}

// GetAllBackendStates returns a copy of all backend states.
func (d *Daemon) GetAllBackendStates() map[string]*BackendState {
	d.backendsMu.RLock()
	defer d.backendsMu.RUnlock()

	result := make(map[string]*BackendState, len(d.backendStates))
	for k, v := range d.backendStates {
		// Return a copy to avoid race conditions
		stateCopy := *v
		result[k] = &stateCopy
	}
	return result
}

// getMinTickInterval returns the minimum tick interval needed to support all backends.
// This is the smallest of the global interval and any backend-specific intervals.
func (d *Daemon) getMinTickInterval() time.Duration {
	d.backendsMu.RLock()
	defer d.backendsMu.RUnlock()

	minInterval := d.cfg.Interval
	if minInterval == 0 {
		minInterval = 5 * time.Minute // Default global interval
	}

	for _, be := range d.backends {
		if be.interval > 0 && be.interval < minInterval {
			minInterval = be.interval
		}
	}

	return minInterval
}

// getBackendStatuses returns the current status of all backends for API responses.
func (d *Daemon) getBackendStatuses() map[string]*BackendStatus {
	d.backendsMu.RLock()
	defer d.backendsMu.RUnlock()

	if len(d.backendStates) == 0 {
		return nil
	}

	statuses := make(map[string]*BackendStatus, len(d.backendStates))
	for name, state := range d.backendStates {
		lastSync := ""
		if !state.LastSync.IsZero() {
			lastSync = state.LastSync.Format(time.RFC3339)
		}
		statuses[name] = &BackendStatus{
			SyncCount:  state.SyncCount,
			ErrorCount: state.ErrorCount,
			LastSync:   lastSync,
			LastError:  state.LastError,
			Healthy:    state.ErrorCount == 0,
		}
	}
	return statuses
}

// Start starts the daemon process. This should be called in the forked process.
func (d *Daemon) Start() error {
	// Write PID file
	if err := os.MkdirAll(filepath.Dir(d.cfg.PIDPath), 0700); err != nil {
		return fmt.Errorf("failed to create PID directory: %w", err)
	}
	if err := os.WriteFile(d.cfg.PIDPath, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Create Unix socket
	if err := os.MkdirAll(filepath.Dir(d.cfg.SocketPath), 0700); err != nil {
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
	if err := os.MkdirAll(filepath.Dir(d.cfg.LogPath), 0700); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	d.log("Daemon started (PID: %d, interval: %v)", os.Getpid(), d.cfg.Interval)

	// Start IPC listener
	go d.handleConnections()

	// Determine the tick interval - use the minimum of global interval and any backend intervals
	tickInterval := d.getMinTickInterval()
	d.log("Using tick interval: %v (backends: %d)", tickInterval, len(d.backends))

	// Start sync loop
	ticker := time.NewTicker(tickInterval)
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
			// Check if listener was closed - use a non-blocking check first,
			// then return on any error if stop was requested to avoid race
			// conditions with cleanup() removing the log file
			select {
			case <-d.stopChan:
				return
			default:
				// Double-check with a small timeout to handle race conditions
				select {
				case <-d.stopChan:
					return
				case <-time.After(1 * time.Millisecond):
					// Not stopping, log the error
					d.log("Accept error: %v", err)
				}
			}
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

		// Add per-backend status (Issue #40)
		resp.BackendStates = d.getBackendStatuses()

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
	// Serialize concurrent performSync calls (Issue #52).
	// The ticker and IPC notify handler can both trigger performSync.
	d.syncMu.Lock()
	defer d.syncMu.Unlock()

	d.mu.Lock()
	d.syncCount++
	count := d.syncCount
	d.mu.Unlock()

	d.log("Starting sync (count: %d)", count)

	// Check if we have multi-backend configuration
	d.backendsMu.RLock()
	hasBackends := len(d.backends) > 0
	d.backendsMu.RUnlock()

	if hasBackends {
		// Multi-backend sync (Issue #40)
		d.performMultiBackendSync(count)
	} else if d.syncFunc != nil {
		// Legacy single-backend sync
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

// performMultiBackendSync iterates through all backends and syncs each one.
// Failure in one backend does not affect others (failure isolation).
func (d *Daemon) performMultiBackendSync(globalCount int) {
	d.backendsMu.Lock()
	backends := make([]*backendEntry, len(d.backends))
	copy(backends, d.backends)
	d.backendsMu.Unlock()

	now := time.Now()
	globalInterval := d.cfg.Interval

	for _, be := range backends {
		// Check if this backend should sync based on its interval.
		// Read be.lastSync under lock to avoid data race (Issue #52).
		interval := be.interval
		if interval == 0 {
			interval = globalInterval
		}

		d.backendsMu.RLock()
		lastSync := be.lastSync
		d.backendsMu.RUnlock()

		if !lastSync.IsZero() && now.Sub(lastSync) < interval {
			continue // Skip this backend - not yet time to sync
		}

		d.log("Syncing backend: %s", be.name)

		// Perform sync with failure isolation
		err := be.syncFunc()

		// Update backend state
		d.backendsMu.Lock()
		state := d.backendStates[be.name]
		if state == nil {
			state = &BackendState{Name: be.name}
			d.backendStates[be.name] = state
		}

		if err != nil {
			// Record error but continue with other backends (failure isolation)
			state.ErrorCount++
			state.LastError = err.Error()
			d.log("Backend %s sync error: %v (error count: %d)", be.name, err, state.ErrorCount)
		} else {
			// Success - reset error count
			state.SyncCount++
			state.ErrorCount = 0
			state.LastError = ""
			state.LastSync = now
			d.log("Backend %s sync completed (sync count: %d)", be.name, state.SyncCount)
		}

		// Update backend's last sync time (regardless of success/failure)
		be.lastSync = now
		d.backendsMu.Unlock()
	}
}

func (d *Daemon) cleanup() {
	// Signal handleConnections to exit before closing the listener
	// Use a non-blocking close to avoid panic if already closed
	select {
	case <-d.stopChan:
		// Already closed
	default:
		close(d.stopChan)
	}

	if d.listener != nil {
		_ = d.listener.Close()
	}

	// Log before removing any files to avoid race conditions where
	// handleConnections() might recreate the log file after removal
	d.log("Daemon stopped")

	// Small delay to ensure handleConnections() has exited after listener.Close()
	// This prevents the race where log() is called after files are removed
	time.Sleep(10 * time.Millisecond)

	_ = os.Remove(d.cfg.PIDPath)
	_ = os.Remove(d.cfg.SocketPath)
	_ = os.Remove(d.cfg.LogPath)
}

func (d *Daemon) log(format string, args ...interface{}) {
	entry := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
	f, err := os.OpenFile(d.cfg.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
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
	return fmt.Sprintf("/tmp/todoat-daemon-%d.sock", os.Getuid())
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
