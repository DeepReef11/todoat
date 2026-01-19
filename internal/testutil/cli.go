// Package testutil provides shared test utilities for CLI testing across packages.
// This enables co-located CLI tests while maintaining consistent test infrastructure.
package testutil

import (
	"bytes"
	"database/sql"
	"encoding/json"
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
	t          *testing.T
	cfg        *cmd.Config
	tmpDir     string
	configPath string // Optional path to config file for SetConfigValue
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
	logFile        string
	configPath     string
	daemonInterval time.Duration
	offlineMode    bool
}

// NewCLITestWithDaemon creates a new CLI test helper with daemon support.
func NewCLITestWithDaemon(t *testing.T) *DaemonCLITest {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	notificationLogPath := filepath.Join(tmpDir, "notifications.log")
	pidFile := filepath.Join(tmpDir, "daemon.pid")
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
		logFile:        logFile,
		configPath:     configPath,
		daemonInterval: 5 * time.Minute,
		offlineMode:    false,
	}
}

// PIDFilePath returns the path to the daemon PID file.
func (d *DaemonCLITest) PIDFilePath() string {
	return d.pidFile
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
	cachePath string
	cacheTTL  time.Duration
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
		cachePath: cachePath,
		cacheTTL:  5 * time.Minute,
	}
}

// CachePath returns the path to the cache file.
func (c *CacheCLITest) CachePath() string {
	return c.cachePath
}

// SetCacheTTL sets the cache TTL for testing.
func (c *CacheCLITest) SetCacheTTL(ttl time.Duration) {
	c.cacheTTL = ttl
	c.cfg.CacheTTL = ttl
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
