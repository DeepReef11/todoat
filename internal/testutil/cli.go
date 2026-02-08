// Package testutil provides shared test utilities for CLI testing across packages.
// This enables co-located CLI tests while maintaining consistent test infrastructure.
package testutil

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"todoat/cmd/todoat/cmd"
)

// defaultTestConfig is the minimal config used by most test constructors to ensure isolation.
const defaultTestConfig = "# test config\ndefault_backend: sqlite\n"

// CLITest provides a test helper for running CLI commands in isolation.
type CLITest struct {
	t             *testing.T
	cfg           *cmd.Config
	tmpDir        string
	configPath    string // Optional path to config file for SetConfigValue
	customDBPath  string // Custom DB path for testing config-based path
	defaultDBPath string // Default DB path (XDG-based) for testing config-based path
}

// NewCLITest creates a new CLI test helper with an isolated in-memory database.
func NewCLITest(t *testing.T) *CLITest {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")
	configPath := tmpDir + "/config.yaml"

	// Write a minimal default config to ensure isolation
	if err := os.WriteFile(configPath, []byte(defaultTestConfig), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg := &cmd.Config{
		NoPrompt:   true,
		DBPath:     dbPath,
		CachePath:  cachePath,  // Use test-specific cache path
		ConfigPath: configPath, // Use test-specific config path for isolation
	}

	return &CLITest{
		t:      t,
		cfg:    cfg,
		tmpDir: tmpDir,
	}
}

// NewCLITestWithViews creates a new CLI test helper with views directory support.
// Returns the CLITest and the viewsDir (for placing view YAML files).
func NewCLITestWithViews(t *testing.T) (*CLITest, string) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	viewsDir := tmpDir + "/views"
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")
	configPath := tmpDir + "/config.yaml"

	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatalf("failed to create views directory: %v", err)
	}

	// Write a minimal default config to ensure isolation
	if err := os.WriteFile(configPath, []byte(defaultTestConfig), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg := &cmd.Config{
		NoPrompt:   true,
		DBPath:     dbPath,
		ViewsPath:  viewsDir,
		CachePath:  cachePath,
		ConfigPath: configPath,
	}

	return &CLITest{
		t:      t,
		cfg:    cfg,
		tmpDir: tmpDir,
	}, viewsDir
}

// NewCLITestWithViewsAndTmpDir creates a new CLI test helper with views directory support.
// Returns the CLITest, viewsDir, and tmpDir (base directory for config.yaml).
// This function also sets XDG_CONFIG_HOME to isolate the test from user's real config
// and enables proper plugin directory resolution.
func NewCLITestWithViewsAndTmpDir(t *testing.T) (*CLITest, string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	viewsDir := tmpDir + "/views"
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")
	configPath := tmpDir + "/config.yaml"

	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatalf("failed to create views directory: %v", err)
	}

	// Write a minimal default config to ensure isolation
	if err := os.WriteFile(configPath, []byte(defaultTestConfig), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Set XDG_CONFIG_HOME to isolate from user's real config
	// This ensures config.GetConfigDir() returns a test-specific path
	// which is needed for plugin directory resolution (security fix issue #73)
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}

	t.Cleanup(func() {
		if oldConfigHome == "" {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			_ = os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		}
	})

	cfg := &cmd.Config{
		NoPrompt:   true,
		DBPath:     dbPath,
		ViewsPath:  viewsDir,
		CachePath:  cachePath,
		ConfigPath: configPath,
	}

	return &CLITest{
		t:      t,
		cfg:    cfg,
		tmpDir: tmpDir,
	}, viewsDir, tmpDir
}

// NewCLITestWithViewsPath creates a new CLI test helper with custom paths.
// This is used for testing views folder setup behavior.
// The views directory is NOT created - the test can control whether it exists.
func NewCLITestWithViewsPath(t *testing.T, dbPath, viewsDir, configPath, cachePath string) *CLITest {
	t.Helper()

	cfg := &cmd.Config{
		NoPrompt:   true,
		DBPath:     dbPath,
		ViewsPath:  viewsDir,
		CachePath:  cachePath,
		ConfigPath: configPath,
	}

	return &CLITest{
		t:      t,
		cfg:    cfg,
		tmpDir: filepath.Dir(dbPath),
	}
}

// NewCLITestWithConfig creates a new CLI test helper with config file support.
// This is used for testing configuration CLI commands.
func NewCLITestWithConfig(t *testing.T) *CLITest {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	configPath := tmpDir + "/config.yaml"
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")

	// Write initial config file
	initialConfig := "# test config\n"
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg := &cmd.Config{
		NoPrompt:   true,
		DBPath:     dbPath,
		ConfigPath: configPath,
		CachePath:  cachePath,
	}

	return &CLITest{
		t:          t,
		cfg:        cfg,
		tmpDir:     tmpDir,
		configPath: configPath,
	}
}

// NewCLITestWithCustomDBPath creates a new CLI test helper that tests custom database path from config.
// Unlike other helpers, this does NOT set cfg.DBPath, so the path from config file is used.
// This is used for testing issue #068 (config path should be used when DBPath flag is not set).
func NewCLITestWithCustomDBPath(t *testing.T) *CLITest {
	t.Helper()

	tmpDir := t.TempDir()

	// Custom DB path from config - different from default
	customDBPath := filepath.Join(tmpDir, "custom", "my-tasks.db")

	// What would be the default DB path (in the temp XDG data dir)
	xdgDataDir := filepath.Join(tmpDir, "xdg-data", "todoat")
	defaultDBPath := filepath.Join(xdgDataDir, "tasks.db")

	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")

	// Write config with custom database path
	configContent := `backends:
  sqlite:
    enabled: true
    path: "` + customDBPath + `"
default_backend: sqlite
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Create directories for the XDG paths (to isolate from real user data)
	xdgConfigDir := filepath.Join(tmpDir, "xdg-config", "todoat")
	if err := os.MkdirAll(xdgConfigDir, 0755); err != nil {
		t.Fatalf("failed to create XDG config dir: %v", err)
	}
	if err := os.MkdirAll(xdgDataDir, 0755); err != nil {
		t.Fatalf("failed to create XDG data dir: %v", err)
	}

	// Set XDG environment variables to isolate from user's real data
	// Save old values and restore them at test cleanup
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldCacheHome := os.Getenv("XDG_CACHE_HOME")

	xdgCacheDir := filepath.Join(tmpDir, "xdg-cache")
	if err := os.MkdirAll(xdgCacheDir, 0755); err != nil {
		t.Fatalf("failed to create XDG cache dir: %v", err)
	}

	if err := os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "xdg-data")); err != nil {
		t.Fatalf("failed to set XDG_DATA_HOME: %v", err)
	}
	if err := os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "xdg-config")); err != nil {
		t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
	}
	if err := os.Setenv("XDG_CACHE_HOME", xdgCacheDir); err != nil {
		t.Fatalf("failed to set XDG_CACHE_HOME: %v", err)
	}

	t.Cleanup(func() {
		if oldDataHome == "" {
			_ = os.Unsetenv("XDG_DATA_HOME")
		} else {
			_ = os.Setenv("XDG_DATA_HOME", oldDataHome)
		}
		if oldConfigHome == "" {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			_ = os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		}
		if oldCacheHome == "" {
			_ = os.Unsetenv("XDG_CACHE_HOME")
		} else {
			_ = os.Setenv("XDG_CACHE_HOME", oldCacheHome)
		}
	})

	cfg := &cmd.Config{
		NoPrompt:   true,
		DBPath:     "", // Intentionally empty - should use config file path
		CachePath:  cachePath,
		ConfigPath: configPath,
	}

	return &CLITest{
		t:             t,
		cfg:           cfg,
		tmpDir:        tmpDir,
		configPath:    configPath,
		customDBPath:  customDBPath,
		defaultDBPath: defaultDBPath,
	}
}

// CustomDBPath returns the custom database path configured in the config file.
// This is used for testing issue #068.
func (c *CLITest) CustomDBPath() string {
	return c.customDBPath
}

// DefaultDBPath returns what would be the default database path.
// This is used for testing issue #068.
func (c *CLITest) DefaultDBPath() string {
	return c.defaultDBPath
}

// NewCLITestWithViewsAndConfig creates a new CLI test helper with views directory and config file support.
// This is used for testing features that depend on configuration (like default_view).
func NewCLITestWithViewsAndConfig(t *testing.T) (*CLITest, string) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	viewsDir := tmpDir + "/views"
	configPath := tmpDir + "/config.yaml"
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")

	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatalf("failed to create views directory: %v", err)
	}

	// Write initial config file
	initialConfig := "# test config\n"
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg := &cmd.Config{
		NoPrompt:   true,
		DBPath:     dbPath,
		ViewsPath:  viewsDir,
		ConfigPath: configPath,
		CachePath:  cachePath,
	}

	return &CLITest{
		t:          t,
		cfg:        cfg,
		tmpDir:     tmpDir,
		configPath: configPath,
	}, viewsDir
}

// NewCLITestWithNotification creates a new CLI test helper with notification support.
func NewCLITestWithNotification(t *testing.T) *CLITest {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	notificationLogPath := tmpDir + "/notifications.log"
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")
	configPath := tmpDir + "/config.yaml"

	// Write a minimal default config to ensure isolation
	if err := os.WriteFile(configPath, []byte(defaultTestConfig), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg := &cmd.Config{
		NoPrompt:            true,
		DBPath:              dbPath,
		NotificationLogPath: notificationLogPath,
		NotificationMock:    true, // Use mock executor for OS notifications
		CachePath:           cachePath,
		ConfigPath:          configPath,
	}

	return &CLITest{
		t:      t,
		cfg:    cfg,
		tmpDir: tmpDir,
	}
}

// Config returns the test configuration.
func (c *CLITest) Config() *cmd.Config {
	return c.cfg
}

// TmpDir returns the temporary directory for the test.
func (c *CLITest) TmpDir() string {
	return c.tmpDir
}

// SetConfigValue sets a configuration key-value pair in the test config file.
// This is used for testing configuration-based features like default_view.
func (c *CLITest) SetConfigValue(key, value string) {
	c.t.Helper()

	if c.configPath == "" {
		c.t.Fatalf("SetConfigValue requires a CLITest created with NewCLITestWithViewsAndConfig or NewCLITestWithConfig")
	}

	// Read existing config
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		c.t.Fatalf("failed to read config file: %v", err)
	}

	// Append the new key-value pair
	newConfig := string(data) + key + ": " + value + "\n"

	if err := os.WriteFile(c.configPath, []byte(newConfig), 0644); err != nil {
		c.t.Fatalf("failed to write config file: %v", err)
	}
}

// SetFullConfig replaces the entire config file with the given YAML content.
// This is used for testing configuration CLI commands.
func (c *CLITest) SetFullConfig(yamlContent string) {
	c.t.Helper()

	if c.configPath == "" {
		c.t.Fatalf("SetFullConfig requires a CLITest created with NewCLITestWithConfig")
	}

	if err := os.WriteFile(c.configPath, []byte(yamlContent), 0644); err != nil {
		c.t.Fatalf("failed to write config file: %v", err)
	}
}

// ConfigPath returns the path to the config file.
func (c *CLITest) ConfigPath() string {
	return c.configPath
}

// Execute runs a CLI command with the given arguments and returns stdout, stderr, and exit code.
func (c *CLITest) Execute(args ...string) (stdout, stderr string, exitCode int) {
	c.t.Helper()

	var stdoutBuf, stderrBuf bytes.Buffer
	// Set Stderr on cfg to capture warnings from getBackend
	c.cfg.Stderr = &stderrBuf
	exitCode = cmd.Execute(args, &stdoutBuf, &stderrBuf, c.cfg)
	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

// ExecuteWithStdin runs a CLI command with simulated stdin input.
// This is used for testing interactive prompts.
func (c *CLITest) ExecuteWithStdin(stdinInput string, args ...string) (stdout, stderr string, exitCode int) {
	c.t.Helper()

	var stdoutBuf, stderrBuf bytes.Buffer
	c.cfg.Stderr = &stderrBuf
	c.cfg.Stdin = strings.NewReader(stdinInput)
	defer func() { c.cfg.Stdin = nil }()
	exitCode = cmd.Execute(args, &stdoutBuf, &stderrBuf, c.cfg)
	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

// MustExecute runs a CLI command and fails the test if exit code is non-zero.
func (c *CLITest) MustExecute(args ...string) string {
	c.t.Helper()

	stdout, stderr, exitCode := c.Execute(args...)
	if exitCode != 0 {
		c.t.Fatalf("expected exit code 0, got %d: stdout=%s stderr=%s", exitCode, stdout, stderr)
	}
	return stdout
}

// ExecuteAndFail runs a CLI command and fails the test if exit code is zero.
func (c *CLITest) ExecuteAndFail(args ...string) (stdout, stderr string) {
	c.t.Helper()

	stdout, stderr, exitCode := c.Execute(args...)
	if exitCode == 0 {
		c.t.Fatalf("expected non-zero exit code, got 0: stdout=%s", stdout)
	}
	return stdout, stderr
}

// AssertContains fails the test if output doesn't contain expected string.
func AssertContains(t *testing.T, output, expected string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, output)
	}
}

// AssertNotContains fails the test if output contains unexpected string.
func AssertNotContains(t *testing.T, output, unexpected string) {
	t.Helper()
	if strings.Contains(output, unexpected) {
		t.Errorf("expected output NOT to contain %q, got:\n%s", unexpected, output)
	}
}

// AssertExitCode fails the test if exit code doesn't match expected.
func AssertExitCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("expected exit code %d, got %d", want, got)
	}
}

// AssertResultCode verifies that the output ends with the expected result code.
func AssertResultCode(t *testing.T, output, expectedCode string) {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Errorf("expected result code %q but output is empty", expectedCode)
		return
	}
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	if lastLine != expectedCode {
		t.Errorf("expected result code %q, got %q\nFull output:\n%s", expectedCode, lastLine, output)
	}
}

// Result code constants for convenience.
const (
	ResultActionCompleted = cmd.ResultActionCompleted
	ResultInfoOnly        = cmd.ResultInfoOnly
	ResultError           = cmd.ResultError
)

// DaemonCLITest extends CLITest with daemon-specific helpers for testing sync daemon.
type DaemonCLITest struct {
	*CLITest
	pidFile        string
	socketFile     string
	logFile        string
	configPath     string
	daemonInterval time.Duration
	offlineMode    bool
	forkedMode     bool // True if testing forked daemon (not in-process)
}

// NewCLITestWithDaemon creates a new CLI test helper with daemon support.
func NewCLITestWithDaemon(t *testing.T) *DaemonCLITest {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	notificationLogPath := filepath.Join(tmpDir, "notifications.log")
	pidFile := filepath.Join(tmpDir, "daemon.pid")
	socketFile := filepath.Join(tmpDir, "daemon.sock")
	logFile := filepath.Join(tmpDir, "daemon.log")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")

	// Write a minimal default config to ensure isolation
	if err := os.WriteFile(configPath, []byte(defaultTestConfig), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg := &cmd.Config{
		NoPrompt:            true,
		DBPath:              dbPath,
		ConfigPath:          configPath,
		NotificationLogPath: notificationLogPath,
		NotificationMock:    true,
		DaemonPIDPath:       pidFile,
		DaemonSocketPath:    socketFile,
		DaemonLogPath:       logFile,
		DaemonTestMode:      true, // Use in-process daemon for testing
		CachePath:           cachePath,
	}

	return &DaemonCLITest{
		CLITest: &CLITest{
			t:      t,
			cfg:    cfg,
			tmpDir: tmpDir,
		},
		pidFile:        pidFile,
		socketFile:     socketFile,
		logFile:        logFile,
		configPath:     configPath,
		daemonInterval: 5 * time.Minute,
		offlineMode:    false,
		forkedMode:     false,
	}
}

// NewCLITestWithForkedDaemon creates a new CLI test helper for testing forked daemon.
// Unlike NewCLITestWithDaemon, this does NOT use DaemonTestMode, so it tests the
// actual forked process behavior.
//
// IMPORTANT: This test requires a pre-built todoat binary. If the binary is not
// available, the test will be skipped. Set TODOAT_BINARY environment variable
// to specify the binary path, or run `go build -o ./bin/todoat ./cmd/todoat` first.
func NewCLITestWithForkedDaemon(t *testing.T) *DaemonCLITest {
	t.Helper()

	// Check for todoat binary - tests need a real binary to fork
	binaryPath := os.Getenv("TODOAT_BINARY")
	if binaryPath == "" {
		// Try common locations
		candidates := []string{
			"./bin/todoat",
			"./todoat",
			"../bin/todoat",
			"../../bin/todoat",
		}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				binaryPath = candidate
				break
			}
		}
	}

	if binaryPath == "" {
		t.Skip("Skipping forked daemon test: no todoat binary found. Build with: go build -o ./bin/todoat ./cmd/todoat")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		t.Fatalf("failed to get absolute path for binary: %v", err)
	}
	binaryPath = absPath

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	notificationLogPath := filepath.Join(tmpDir, "notifications.log")
	pidFile := filepath.Join(tmpDir, "daemon.pid")
	socketFile := filepath.Join(tmpDir, "daemon.sock")
	logFile := filepath.Join(tmpDir, "daemon.log")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")

	// Write a minimal default config to ensure isolation
	if err := os.WriteFile(configPath, []byte(defaultTestConfig), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg := &cmd.Config{
		NoPrompt:            true,
		DBPath:              dbPath,
		ConfigPath:          configPath,
		NotificationLogPath: notificationLogPath,
		NotificationMock:    true,
		DaemonPIDPath:       pidFile,
		DaemonSocketPath:    socketFile,
		DaemonLogPath:       logFile,
		DaemonTestMode:      false, // Use real forked daemon
		DaemonEnabled:       true,  // Enable forked daemon feature
		DaemonBinaryPath:    binaryPath,
		CachePath:           cachePath,
	}

	return &DaemonCLITest{
		CLITest: &CLITest{
			t:      t,
			cfg:    cfg,
			tmpDir: tmpDir,
		},
		pidFile:        pidFile,
		socketFile:     socketFile,
		logFile:        logFile,
		configPath:     configPath,
		daemonInterval: 5 * time.Minute,
		offlineMode:    false,
		forkedMode:     true,
	}
}

// PIDFilePath returns the path to the daemon PID file.
func (d *DaemonCLITest) PIDFilePath() string {
	return d.pidFile
}

// SocketPath returns the path to the daemon Unix socket file.
func (d *DaemonCLITest) SocketPath() string {
	return d.socketFile
}

// DaemonLogPath returns the path to the daemon log file.
func (d *DaemonCLITest) DaemonLogPath() string {
	return d.logFile
}

// ConfigPath returns the path to the config file.
func (d *DaemonCLITest) ConfigPath() string {
	return d.configPath
}

// SetDaemonInterval sets the sync interval for testing.
func (d *DaemonCLITest) SetDaemonInterval(interval time.Duration) {
	d.daemonInterval = interval
	d.cfg.DaemonInterval = interval
}

// SetDaemonOffline sets the daemon to offline mode for testing.
func (d *DaemonCLITest) SetDaemonOffline(offline bool) {
	d.offlineMode = offline
	d.cfg.DaemonOfflineMode = offline
}

// MigrateCLITest extends CLITest with migration-specific helpers.
type MigrateCLITest struct {
	*CLITest
	mockNextcloudTasks map[string][]string
}

// NewCLITestWithMigrate creates a new CLI test helper with migration support.
func NewCLITestWithMigrate(t *testing.T) *MigrateCLITest {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	migrateTargetDir := filepath.Join(tmpDir, "migrate-target")
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.MkdirAll(migrateTargetDir, 0755); err != nil {
		t.Fatalf("failed to create migrate target directory: %v", err)
	}

	// Write a minimal default config to ensure isolation
	if err := os.WriteFile(configPath, []byte(defaultTestConfig), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg := &cmd.Config{
		NoPrompt:         true,
		DBPath:           dbPath,
		MigrateTargetDir: migrateTargetDir,
		MigrateMockMode:  true, // Enable mock backends for testing
		CachePath:        cachePath,
		ConfigPath:       configPath,
	}

	return &MigrateCLITest{
		CLITest: &CLITest{
			t:      t,
			cfg:    cfg,
			tmpDir: tmpDir,
		},
		mockNextcloudTasks: make(map[string][]string),
	}
}

// SetupMockNextcloudTasks sets up mock tasks for the nextcloud-mock backend.
func (m *MigrateCLITest) SetupMockNextcloudTasks(listName string, tasks []string) {
	m.mockNextcloudTasks[listName] = tasks
	// Write mock data to temp file for the mock backend to read
	mockDataPath := filepath.Join(m.tmpDir, "mock-nextcloud-data.json")
	data := make(map[string][]string)
	for k, v := range m.mockNextcloudTasks {
		data[k] = v
	}
	jsonData, _ := json.Marshal(data)
	_ = os.WriteFile(mockDataPath, jsonData, 0644)
	m.cfg.MockNextcloudDataPath = mockDataPath
}

// ReminderCLITest extends CLITest with reminder-specific helpers.
type ReminderCLITest struct {
	*CLITest
	reminderConfigPath   string
	notificationLogPath  string
	notificationCallback func(n interface{})
}

// NewCLITestWithReminder creates a new CLI test helper with reminder support.
func NewCLITestWithReminder(t *testing.T) *ReminderCLITest {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	notificationLogPath := filepath.Join(tmpDir, "notifications.log")
	reminderConfigPath := filepath.Join(tmpDir, "reminder-config.json")
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write a minimal default config to ensure isolation
	if err := os.WriteFile(configPath, []byte(defaultTestConfig), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg := &cmd.Config{
		NoPrompt:            true,
		DBPath:              dbPath,
		NotificationLogPath: notificationLogPath,
		NotificationMock:    true,
		ReminderConfigPath:  reminderConfigPath,
		CachePath:           cachePath,
		ConfigPath:          configPath,
	}

	return &ReminderCLITest{
		CLITest: &CLITest{
			t:      t,
			cfg:    cfg,
			tmpDir: tmpDir,
		},
		reminderConfigPath:  reminderConfigPath,
		notificationLogPath: notificationLogPath,
	}
}

// SetReminderConfig configures reminder settings for the test.
func (r *ReminderCLITest) SetReminderConfig(cfg interface{}) {
	jsonData, _ := json.Marshal(cfg)
	_ = os.WriteFile(r.reminderConfigPath, jsonData, 0644)
}

// SetNotificationCallback sets a callback to be called when notifications are sent.
func (r *ReminderCLITest) SetNotificationCallback(callback func(n interface{})) {
	r.notificationCallback = callback
	r.cfg.NotificationCallback = callback
}

// GetNotificationLog returns the contents of the notification log.
func (r *ReminderCLITest) GetNotificationLog() string {
	data, err := os.ReadFile(r.notificationLogPath)
	if err != nil {
		return ""
	}
	return string(data)
}

// ClearNotificationLog clears the notification log file.
func (r *ReminderCLITest) ClearNotificationLog() {
	_ = os.WriteFile(r.notificationLogPath, []byte{}, 0644)
}

// CacheCLITest extends CLITest with cache-specific helpers for testing list caching.
type CacheCLITest struct {
	*CLITest
	cachePath  string
	cacheTTL   time.Duration
	configPath string
}

// NewCLITestWithCache creates a new CLI test helper with cache support.
func NewCLITestWithCache(t *testing.T) *CacheCLITest {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	cacheDir := filepath.Join(tmpDir, "cache")
	cachePath := filepath.Join(cacheDir, "lists.json")
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("failed to create cache directory: %v", err)
	}

	// Write a minimal default config to ensure isolation
	if err := os.WriteFile(configPath, []byte(defaultTestConfig), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg := &cmd.Config{
		NoPrompt:   true,
		DBPath:     dbPath,
		CachePath:  cachePath,
		CacheTTL:   5 * time.Minute, // Default 5 minute TTL
		ConfigPath: configPath,
	}

	return &CacheCLITest{
		CLITest: &CLITest{
			t:      t,
			cfg:    cfg,
			tmpDir: tmpDir,
		},
		cachePath:  cachePath,
		cacheTTL:   5 * time.Minute,
		configPath: configPath,
	}
}

// CachePath returns the path to the cache file.
func (c *CacheCLITest) CachePath() string {
	return c.cachePath
}

// SetCacheTTL sets the cache TTL for testing (runtime override).
func (c *CacheCLITest) SetCacheTTL(ttl time.Duration) {
	c.cacheTTL = ttl
	c.cfg.CacheTTL = ttl
}

// SetCacheTTLViaConfig sets the cache TTL in the config file (tests config loading).
// This clears the runtime CacheTTL to force loading from config file.
func (c *CacheCLITest) SetCacheTTLViaConfig(ttlStr string) {
	c.t.Helper()

	// Read existing config
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		c.t.Fatalf("failed to read config file: %v", err)
	}

	// Append the cache_ttl config
	newConfig := string(data) + "cache_ttl: \"" + ttlStr + "\"\n"

	if err := os.WriteFile(c.configPath, []byte(newConfig), 0644); err != nil {
		c.t.Fatalf("failed to write config file: %v", err)
	}

	// Clear runtime TTL to force loading from config
	c.cfg.CacheTTL = 0
	c.cacheTTL = 0
}

// TrashCLITest extends CLITest with trash-specific helpers for testing auto-purge.
type TrashCLITest struct {
	*CLITest
	configPath string
}

// NewCLITestWithTrash creates a new CLI test helper with config file support for trash settings.
func NewCLITestWithTrash(t *testing.T) *TrashCLITest {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath := filepath.Join(tmpDir, "config.yaml")
	cachePath := filepath.Join(tmpDir, "cache", "lists.json")

	// Write initial config file with default backend for isolation
	initialConfig := "# test config\ndefault_backend: sqlite\n"
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cfg := &cmd.Config{
		NoPrompt:   true,
		DBPath:     dbPath,
		ConfigPath: configPath,
		CachePath:  cachePath,
	}

	return &TrashCLITest{
		CLITest: &CLITest{
			t:          t,
			cfg:        cfg,
			tmpDir:     tmpDir,
			configPath: configPath,
		},
		configPath: configPath,
	}
}

// SetTrashRetentionDays sets the trash.retention_days config value.
func (c *TrashCLITest) SetTrashRetentionDays(days int) {
	c.t.Helper()

	// Read existing config
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		c.t.Fatalf("failed to read config file: %v", err)
	}

	// Append the trash config section
	newConfig := string(data) + "trash:\n  retention_days: " + strconv.Itoa(days) + "\n"

	if err := os.WriteFile(c.configPath, []byte(newConfig), 0644); err != nil {
		c.t.Fatalf("failed to write config file: %v", err)
	}
}

// SetListDeletedAt modifies the deleted_at timestamp for a list by name.
// This is used to simulate lists that were deleted in the past.
func (c *TrashCLITest) SetListDeletedAt(listName string, deletedAt time.Time) {
	c.t.Helper()

	// Open the database directly
	db, err := openTestDB(c.cfg.DBPath)
	if err != nil {
		c.t.Fatalf("failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Update the deleted_at timestamp
	deletedAtStr := deletedAt.UTC().Format(time.RFC3339Nano)
	_, err = db.Exec("UPDATE task_lists SET deleted_at = ? WHERE LOWER(name) = LOWER(?)", deletedAtStr, listName)
	if err != nil {
		c.t.Fatalf("failed to update deleted_at: %v", err)
	}
}

// openTestDB opens the SQLite database for testing purposes.
func openTestDB(dbPath string) (*sql.DB, error) {
	return sql.Open("sqlite", dbPath)
}

// WaitFor polls a condition function until it returns true or timeout is reached.
// This is used to replace flaky time.Sleep calls in tests with condition-based waiting.
// The poll interval is 50ms by default.
func WaitFor(t *testing.T, timeout time.Duration, condition func() bool, description string) {
	t.Helper()
	WaitForWithInterval(t, timeout, 50*time.Millisecond, condition, description)
}

// WaitForWithInterval polls a condition function with a custom interval.
func WaitForWithInterval(t *testing.T, timeout, interval time.Duration, condition func() bool, description string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(interval)
	}
	t.Fatalf("WaitFor timed out after %v: %s", timeout, description)
}

// WaitForOutput polls until CLI output contains the expected string.
func (c *CLITest) WaitForOutput(timeout time.Duration, args []string, expected string) string {
	c.t.Helper()
	var lastOutput string
	WaitFor(c.t, timeout, func() bool {
		lastOutput, _, _ = c.Execute(args...)
		return strings.Contains(lastOutput, expected)
	}, fmt.Sprintf("output to contain %q", expected))
	return lastOutput
}

// WaitForSyncCount polls daemon status until sync count reaches the target value.
func (d *DaemonCLITest) WaitForSyncCount(timeout time.Duration, minCount int) string {
	d.t.Helper()
	var lastOutput string
	WaitFor(d.t, timeout, func() bool {
		lastOutput = d.MustExecute("-y", "sync", "daemon", "status")
		// Parse sync count from output
		if strings.Contains(lastOutput, "Sync count:") {
			lines := strings.Split(lastOutput, "\n")
			for _, line := range lines {
				if strings.Contains(line, "Sync count:") {
					var count int
					if _, err := fmt.Sscanf(line, "  Sync count: %d", &count); err == nil {
						return count >= minCount
					}
				}
			}
		}
		return false
	}, fmt.Sprintf("sync count to reach at least %d", minCount))
	return lastOutput
}
