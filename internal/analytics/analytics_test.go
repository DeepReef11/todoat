package analytics

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// =============================================================================
// Analytics System Tests (067-analytics-system)
// =============================================================================

// TestTracker_TrackCommand verifies command tracking records events correctly
func TestTracker_TrackCommand(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "analytics.db")

	tracker, err := NewTracker(dbPath, true)
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	defer func() { _ = tracker.Close() }()

	// Track a successful command
	err = tracker.TrackCommand("add", "", "sqlite", []string{"--priority", "--tag"}, func() error {
		time.Sleep(10 * time.Millisecond) // Simulate some work
		return nil
	})
	if err != nil {
		t.Fatalf("TrackCommand() error = %v", err)
	}

	// Wait for async logging to complete
	time.Sleep(100 * time.Millisecond)

	// Query the database to verify the event was recorded
	events, err := tracker.QueryEvents("SELECT * FROM events WHERE command = 'add'")
	if err != nil {
		t.Fatalf("QueryEvents() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Command != "add" {
		t.Errorf("expected command = 'add', got %q", event.Command)
	}
	if event.Backend != "sqlite" {
		t.Errorf("expected backend = 'sqlite', got %q", event.Backend)
	}
	if !event.Success {
		t.Errorf("expected success = true, got false")
	}
	if event.DurationMs < 10 {
		t.Errorf("expected duration >= 10ms, got %d", event.DurationMs)
	}
	if event.Flags != `["--priority","--tag"]` {
		t.Errorf("expected flags = '[\"--priority\",\"--tag\"]', got %q", event.Flags)
	}
}

// TestTracker_TrackCommandError verifies error tracking and categorization
func TestTracker_TrackCommandError(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "analytics.db")

	tracker, err := NewTracker(dbPath, true)
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	defer func() { _ = tracker.Close() }()

	// Track a failed command
	testErr := errors.New("network timeout")
	err = tracker.TrackCommand("sync", "", "todoist", nil, func() error {
		return testErr
	})
	if err != testErr {
		t.Errorf("expected error = %v, got %v", testErr, err)
	}

	// Wait for async logging to complete
	time.Sleep(100 * time.Millisecond)

	// Query the database to verify the event was recorded with error
	events, err := tracker.QueryEvents("SELECT * FROM events WHERE command = 'sync'")
	if err != nil {
		t.Fatalf("QueryEvents() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Success {
		t.Errorf("expected success = false, got true")
	}
	if event.ErrorType == "" {
		t.Errorf("expected error_type to be set")
	}
}

// TestTracker_Cleanup verifies automatic retention cleanup works
func TestTracker_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "analytics.db")

	tracker, err := NewTracker(dbPath, true)
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	defer func() { _ = tracker.Close() }()

	// Insert an old event directly (simulating an event from 10 days ago)
	oldTimestamp := time.Now().Unix() - (10 * 86400) // 10 days ago
	_, err = tracker.db.Exec(`
		INSERT INTO events (timestamp, command, backend, success, duration_ms)
		VALUES (?, 'old_command', 'sqlite', 1, 100)
	`, oldTimestamp)
	if err != nil {
		t.Fatalf("failed to insert old event: %v", err)
	}

	// Insert a recent event
	recentTimestamp := time.Now().Unix() - (2 * 86400) // 2 days ago
	_, err = tracker.db.Exec(`
		INSERT INTO events (timestamp, command, backend, success, duration_ms)
		VALUES (?, 'recent_command', 'sqlite', 1, 100)
	`, recentTimestamp)
	if err != nil {
		t.Fatalf("failed to insert recent event: %v", err)
	}

	// Run cleanup with 7 day retention
	deleted, err := tracker.Cleanup(7)
	if err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	if deleted != 1 {
		t.Errorf("expected 1 event deleted, got %d", deleted)
	}

	// Verify old event was deleted
	events, err := tracker.QueryEvents("SELECT * FROM events WHERE command = 'old_command'")
	if err != nil {
		t.Fatalf("QueryEvents() error = %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected old event to be deleted, but found %d", len(events))
	}

	// Verify recent event still exists
	events, err = tracker.QueryEvents("SELECT * FROM events WHERE command = 'recent_command'")
	if err != nil {
		t.Fatalf("QueryEvents() error = %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected recent event to remain, but found %d", len(events))
	}
}

// TestAnalytics_Disabled verifies analytics can be disabled via config
func TestAnalytics_Disabled(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "analytics.db")

	// Create tracker with enabled=false
	tracker, err := NewTracker(dbPath, false)
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	defer func() { _ = tracker.Close() }()

	// Track a command - should not record anything
	callCount := 0
	err = tracker.TrackCommand("add", "", "sqlite", []string{"--priority"}, func() error {
		callCount++
		return nil
	})
	if err != nil {
		t.Fatalf("TrackCommand() error = %v", err)
	}

	// The function should still be called
	if callCount != 1 {
		t.Errorf("expected function to be called once, got %d", callCount)
	}

	// Wait for any potential async logging
	time.Sleep(100 * time.Millisecond)

	// Verify no events were recorded
	events, err := tracker.QueryEvents("SELECT * FROM events")
	if err != nil {
		t.Fatalf("QueryEvents() error = %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events when disabled, got %d", len(events))
	}
}

// TestAnalytics_EnvironmentOverride verifies TODOAT_ANALYTICS_ENABLED override
func TestAnalytics_EnvironmentOverride(t *testing.T) {
	// Test that environment variable can disable analytics
	t.Run("env disables analytics", func(t *testing.T) {
		t.Setenv("TODOAT_ANALYTICS_ENABLED", "false")

		enabled := IsEnabledFromEnv(true) // config says enabled
		if enabled {
			t.Errorf("expected analytics disabled by env, got enabled")
		}
	})

	t.Run("env enables analytics", func(t *testing.T) {
		t.Setenv("TODOAT_ANALYTICS_ENABLED", "true")

		enabled := IsEnabledFromEnv(false) // config says disabled
		if !enabled {
			t.Errorf("expected analytics enabled by env, got disabled")
		}
	})

	t.Run("no env uses config value", func(t *testing.T) {
		// Ensure env var is not set
		_ = os.Unsetenv("TODOAT_ANALYTICS_ENABLED")

		// Config disabled
		enabled := IsEnabledFromEnv(false)
		if enabled {
			t.Errorf("expected analytics disabled (from config), got enabled")
		}

		// Config enabled
		enabled = IsEnabledFromEnv(true)
		if !enabled {
			t.Errorf("expected analytics enabled (from config), got disabled")
		}
	})
}

// TestTracker_DatabaseCreation verifies analytics database is created at correct path
func TestTracker_DatabaseCreation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "analytics.db")

	tracker, err := NewTracker(dbPath, true)
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	defer func() { _ = tracker.Close() }()

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("database file not created at %s", dbPath)
	}

	// Verify schema was created
	var tableName string
	err = tracker.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='events'").Scan(&tableName)
	if err != nil {
		t.Fatalf("failed to query schema: %v", err)
	}
	if tableName != "events" {
		t.Errorf("expected table 'events', got %q", tableName)
	}
}

// TestTracker_Subcommand verifies subcommand is tracked correctly
func TestTracker_Subcommand(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "analytics.db")

	tracker, err := NewTracker(dbPath, true)
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	defer func() { _ = tracker.Close() }()

	// Track command with subcommand
	err = tracker.TrackCommand("list", "view", "sqlite", nil, func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("TrackCommand() error = %v", err)
	}

	// Wait for async logging
	time.Sleep(100 * time.Millisecond)

	events, err := tracker.QueryEvents("SELECT * FROM events WHERE command = 'list'")
	if err != nil {
		t.Fatalf("QueryEvents() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Subcommand != "view" {
		t.Errorf("expected subcommand = 'view', got %q", events[0].Subcommand)
	}
}

// Helper function to access db for tests
func (t *Tracker) QueryEvents(query string) ([]Event, error) {
	rows, err := t.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []Event
	for rows.Next() {
		var e Event
		var flagsNull sql.NullString
		var errorTypeNull sql.NullString
		var subcommandNull sql.NullString
		var backendNull sql.NullString
		var durationNull sql.NullInt64
		var createdAt int64

		err := rows.Scan(
			&e.ID,
			&e.Timestamp,
			&e.Command,
			&subcommandNull,
			&backendNull,
			&e.Success,
			&durationNull,
			&errorTypeNull,
			&flagsNull,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}

		if subcommandNull.Valid {
			e.Subcommand = subcommandNull.String
		}
		if backendNull.Valid {
			e.Backend = backendNull.String
		}
		if durationNull.Valid {
			e.DurationMs = durationNull.Int64
		}
		if errorTypeNull.Valid {
			e.ErrorType = errorTypeNull.String
		}
		if flagsNull.Valid {
			e.Flags = flagsNull.String
		}

		events = append(events, e)
	}

	return events, rows.Err()
}
