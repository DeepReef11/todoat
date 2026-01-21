package analytics

import (
	"database/sql"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

// Tracker handles analytics event recording
type Tracker struct {
	db      *sql.DB
	enabled bool
	mu      sync.Mutex
}

// NewTracker creates a new analytics tracker.
// If enabled is false, tracking is disabled but the database is still created.
func NewTracker(dbPath string, enabled bool) (*Tracker, error) {
	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
	}

	return &Tracker{
		db:      db,
		enabled: enabled,
	}, nil
}

// Close closes the database connection
func (t *Tracker) Close() error {
	if t.db != nil {
		return t.db.Close()
	}
	return nil
}

// TrackCommand wraps command execution with analytics tracking.
// The provided function is always executed, but events are only recorded
// when analytics is enabled.
func (t *Tracker) TrackCommand(cmd, subcmd, backend string, flags []string, fn func() error) error {
	if !t.enabled {
		return fn()
	}

	start := time.Now()
	err := fn()
	duration := time.Since(start).Milliseconds()

	event := Event{
		Timestamp:  time.Now().Unix(),
		Command:    cmd,
		Subcommand: subcmd,
		Backend:    backend,
		Success:    err == nil,
		DurationMs: duration,
	}

	if flags != nil {
		flagsJSON, _ := json.Marshal(flags)
		event.Flags = string(flagsJSON)
	}

	if err != nil {
		event.ErrorType = categorizeError(err)
	}

	// Log asynchronously to avoid slowing down commands
	go t.logEvent(event)

	return err
}

// logEvent records an event to the database
func (t *Tracker) logEvent(event Event) {
	t.mu.Lock()
	defer t.mu.Unlock()

	_, _ = t.db.Exec(`
		INSERT INTO events (timestamp, command, subcommand, backend, success, duration_ms, error_type, flags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, event.Timestamp, event.Command, nullString(event.Subcommand), nullString(event.Backend),
		boolToInt(event.Success), event.DurationMs, nullString(event.ErrorType), nullString(event.Flags))
}

// Cleanup removes events older than the specified retention period.
// Returns the number of deleted events.
func (t *Tracker) Cleanup(retentionDays int) (int64, error) {
	cutoff := time.Now().Unix() - int64(retentionDays*86400)

	result, err := t.db.Exec("DELETE FROM events WHERE timestamp < ?", cutoff)
	if err != nil {
		return 0, err
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	// Vacuum to reclaim space
	_, _ = t.db.Exec("VACUUM")

	return deleted, nil
}

// categorizeError categorizes an error into a general type
func categorizeError(err error) string {
	if err == nil {
		return ""
	}

	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "network") || strings.Contains(errStr, "connection"):
		return "network"
	case strings.Contains(errStr, "auth") || strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "forbidden"):
		return "auth"
	case strings.Contains(errStr, "not found"):
		return "not_found"
	case strings.Contains(errStr, "invalid") || strings.Contains(errStr, "validation"):
		return "validation"
	default:
		return "unknown"
	}
}

// nullString returns nil for empty strings, otherwise the string pointer
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// boolToInt converts a bool to 1 (true) or 0 (false)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
