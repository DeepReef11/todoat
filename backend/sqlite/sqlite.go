package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
	"todoat/backend"
)

// Backend implements backend.TaskManager using SQLite
type Backend struct {
	db        *sql.DB
	backendID string // Identifies this backend instance for data isolation
}

// Migration represents a database schema migration
type Migration struct {
	Version int
	Name    string
	Up      func(db *sql.DB) error
}

// migrations is the ordered list of all schema migrations
var migrations = []Migration{
	{
		Version: 1,
		Name:    "initial_schema",
		Up: func(db *sql.DB) error {
			schema := `
				CREATE TABLE IF NOT EXISTS task_lists (
					id TEXT PRIMARY KEY,
					name TEXT NOT NULL,
					color TEXT DEFAULT '',
					modified TEXT NOT NULL,
					deleted_at TEXT
				);

				CREATE TABLE IF NOT EXISTS tasks (
					id TEXT PRIMARY KEY,
					list_id TEXT NOT NULL,
					summary TEXT NOT NULL,
					description TEXT DEFAULT '',
					status TEXT NOT NULL DEFAULT 'NEEDS-ACTION',
					priority INTEGER DEFAULT 0,
					due_date TEXT,
					start_date TEXT,
					completed TEXT,
					created TEXT NOT NULL,
					modified TEXT NOT NULL,
					parent_id TEXT DEFAULT '',
					categories TEXT DEFAULT '',
					FOREIGN KEY (list_id) REFERENCES task_lists(id) ON DELETE CASCADE
				);

				CREATE INDEX IF NOT EXISTS idx_tasks_list_id ON tasks(list_id);
				CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
			`
			_, err := db.Exec(schema)
			return err
		},
	},
	{
		Version: 2,
		Name:    "add_list_description",
		Up: func(db *sql.DB) error {
			_, err := db.Exec("ALTER TABLE task_lists ADD COLUMN description TEXT DEFAULT ''")
			return err
		},
	},
	{
		Version: 3,
		Name:    "add_recurrence_fields",
		Up: func(db *sql.DB) error {
			if _, err := db.Exec("ALTER TABLE tasks ADD COLUMN recurrence TEXT DEFAULT ''"); err != nil {
				return err
			}
			_, err := db.Exec("ALTER TABLE tasks ADD COLUMN recur_from_due INTEGER DEFAULT 1")
			return err
		},
	},
	{
		Version: 4,
		Name:    "add_backend_id_for_isolation",
		Up: func(db *sql.DB) error {
			// Add backend_id column to tasks table with default 'sqlite' for existing data
			if _, err := db.Exec("ALTER TABLE tasks ADD COLUMN backend_id TEXT NOT NULL DEFAULT 'sqlite'"); err != nil {
				return err
			}
			// Add backend_id column to task_lists table with default 'sqlite' for existing data
			if _, err := db.Exec("ALTER TABLE task_lists ADD COLUMN backend_id TEXT NOT NULL DEFAULT 'sqlite'"); err != nil {
				return err
			}
			// Create indexes for efficient backend filtering
			if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_backend_id ON tasks(backend_id)"); err != nil {
				return err
			}
			if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_task_lists_backend_id ON task_lists(backend_id)"); err != nil {
				return err
			}
			return nil
		},
	},
}

// New creates a new SQLite backend and initializes the database schema.
// Uses "sqlite" as the default backend ID for backward compatibility.
func New(path string) (*Backend, error) {
	return NewWithBackendID(path, "sqlite")
}

// NewWithBackendID creates a new SQLite backend with a specific backend ID.
// The backend ID is used to isolate data between different backends sharing
// the same database file (e.g., sync cache scenario in Issue #007).
func NewWithBackendID(path string, backendID string) (*Backend, error) {
	if backendID == "" {
		backendID = "sqlite"
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	b := &Backend{db: db, backendID: backendID}
	if err := b.initSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return b, nil
}

// initSchema runs database migrations to ensure the schema is up to date
func (b *Backend) initSchema() error {
	// Configure SQLite for concurrent access
	// Set busy timeout to 5 seconds (5000ms) to wait when database is locked
	if _, err := b.db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		return err
	}

	// Enable WAL mode for better concurrent access (allows concurrent reads and writes)
	if _, err := b.db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return err
	}

	// Enable foreign keys
	if _, err := b.db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return err
	}

	// Create schema_version table if it doesn't exist
	_, err := b.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_version table: %w", err)
	}

	// Get current schema version
	currentVersion, err := b.getSchemaVersionInternal()
	if err != nil {
		return fmt.Errorf("failed to get schema version: %w", err)
	}

	// Apply pending migrations in order
	for _, m := range migrations {
		if m.Version <= currentVersion {
			continue
		}

		// Run the migration
		if err := m.Up(b.db); err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w", m.Version, m.Name, err)
		}

		// Record the migration
		_, err = b.db.Exec("INSERT OR REPLACE INTO schema_version (version, applied_at) VALUES (?, CURRENT_TIMESTAMP)", m.Version)
		if err != nil {
			return fmt.Errorf("failed to record migration %d: %w", m.Version, err)
		}
	}

	return nil
}

// getSchemaVersionInternal returns the current schema version (0 if no migrations applied)
func (b *Backend) getSchemaVersionInternal() (int, error) {
	var version int
	err := b.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// GetSchemaVersion returns the current schema version
func (b *Backend) GetSchemaVersion() (int, error) {
	return b.getSchemaVersionInternal()
}

// GetLists returns all active (non-deleted) task lists for this backend
func (b *Backend) GetLists(ctx context.Context) ([]backend.List, error) {
	rows, err := b.db.QueryContext(ctx,
		"SELECT id, name, color, description, modified FROM task_lists WHERE deleted_at IS NULL AND backend_id = ?",
		b.backendID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var lists []backend.List
	for rows.Next() {
		var l backend.List
		var modifiedStr string
		if err := rows.Scan(&l.ID, &l.Name, &l.Color, &l.Description, &modifiedStr); err != nil {
			return nil, err
		}
		l.Modified, _ = time.Parse(time.RFC3339Nano, modifiedStr)
		lists = append(lists, l)
	}

	if lists == nil {
		lists = []backend.List{}
	}
	return lists, rows.Err()
}

// GetList returns a specific active list by ID for this backend
func (b *Backend) GetList(ctx context.Context, listID string) (*backend.List, error) {
	var l backend.List
	var modifiedStr string
	err := b.db.QueryRowContext(ctx,
		"SELECT id, name, color, description, modified FROM task_lists WHERE id = ? AND deleted_at IS NULL AND backend_id = ?",
		listID, b.backendID,
	).Scan(&l.ID, &l.Name, &l.Color, &l.Description, &modifiedStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	l.Modified, _ = time.Parse(time.RFC3339Nano, modifiedStr)
	return &l, nil
}

// GetListByName returns a specific active list by name (case-insensitive) for this backend
func (b *Backend) GetListByName(ctx context.Context, name string) (*backend.List, error) {
	var l backend.List
	var modifiedStr string
	err := b.db.QueryRowContext(ctx,
		"SELECT id, name, color, description, modified FROM task_lists WHERE LOWER(name) = LOWER(?) AND deleted_at IS NULL AND backend_id = ?",
		name, b.backendID,
	).Scan(&l.ID, &l.Name, &l.Color, &l.Description, &modifiedStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	l.Modified, _ = time.Parse(time.RFC3339Nano, modifiedStr)
	return &l, nil
}

// CreateList creates a new task list for this backend
func (b *Backend) CreateList(ctx context.Context, name string) (*backend.List, error) {
	id := uuid.New().String()
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339Nano)

	_, err := b.db.ExecContext(ctx,
		"INSERT INTO task_lists (id, name, color, description, modified, backend_id) VALUES (?, ?, '', '', ?, ?)",
		id, name, nowStr, b.backendID,
	)
	if err != nil {
		return nil, err
	}

	return &backend.List{
		ID:          id,
		Name:        name,
		Color:       "",
		Description: "",
		Modified:    now,
	}, nil
}

// UpdateList updates an existing task list for this backend
func (b *Backend) UpdateList(ctx context.Context, list *backend.List) (*backend.List, error) {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339Nano)

	_, err := b.db.ExecContext(ctx,
		"UPDATE task_lists SET name = ?, color = ?, description = ?, modified = ? WHERE id = ? AND deleted_at IS NULL AND backend_id = ?",
		list.Name, list.Color, list.Description, nowStr, list.ID, b.backendID,
	)
	if err != nil {
		return nil, err
	}

	list.Modified = now
	return list, nil
}

// DeleteList soft-deletes a task list (moves to trash) for this backend
func (b *Backend) DeleteList(ctx context.Context, listID string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := b.db.ExecContext(ctx, "UPDATE task_lists SET deleted_at = ? WHERE id = ? AND backend_id = ?", now, listID, b.backendID)
	return err
}

// GetDeletedLists returns all deleted (trashed) task lists for this backend
func (b *Backend) GetDeletedLists(ctx context.Context) ([]backend.List, error) {
	rows, err := b.db.QueryContext(ctx,
		"SELECT id, name, color, description, modified, deleted_at FROM task_lists WHERE deleted_at IS NOT NULL AND backend_id = ?",
		b.backendID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var lists []backend.List
	for rows.Next() {
		var l backend.List
		var modifiedStr string
		var deletedAtStr sql.NullString
		if err := rows.Scan(&l.ID, &l.Name, &l.Color, &l.Description, &modifiedStr, &deletedAtStr); err != nil {
			return nil, err
		}
		l.Modified, _ = time.Parse(time.RFC3339Nano, modifiedStr)
		if deletedAtStr.Valid {
			t, _ := time.Parse(time.RFC3339Nano, deletedAtStr.String)
			l.DeletedAt = &t
		}
		lists = append(lists, l)
	}

	if lists == nil {
		lists = []backend.List{}
	}
	return lists, rows.Err()
}

// GetDeletedListByName returns a specific deleted list by name (case-insensitive) for this backend
func (b *Backend) GetDeletedListByName(ctx context.Context, name string) (*backend.List, error) {
	var l backend.List
	var modifiedStr string
	var deletedAtStr sql.NullString
	err := b.db.QueryRowContext(ctx,
		"SELECT id, name, color, description, modified, deleted_at FROM task_lists WHERE LOWER(name) = LOWER(?) AND deleted_at IS NOT NULL AND backend_id = ?",
		name, b.backendID,
	).Scan(&l.ID, &l.Name, &l.Color, &l.Description, &modifiedStr, &deletedAtStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	l.Modified, _ = time.Parse(time.RFC3339Nano, modifiedStr)
	if deletedAtStr.Valid {
		t, _ := time.Parse(time.RFC3339Nano, deletedAtStr.String)
		l.DeletedAt = &t
	}
	return &l, nil
}

// RestoreList restores a deleted list from trash for this backend
func (b *Backend) RestoreList(ctx context.Context, listID string) error {
	_, err := b.db.ExecContext(ctx, "UPDATE task_lists SET deleted_at = NULL WHERE id = ? AND backend_id = ?", listID, b.backendID)
	return err
}

// PurgeList permanently deletes a list and all its tasks for this backend
func (b *Backend) PurgeList(ctx context.Context, listID string) error {
	// First delete all tasks in this list for this backend
	_, err := b.db.ExecContext(ctx, "DELETE FROM tasks WHERE list_id = ? AND backend_id = ?", listID, b.backendID)
	if err != nil {
		return err
	}

	_, err = b.db.ExecContext(ctx, "DELETE FROM task_lists WHERE id = ? AND backend_id = ?", listID, b.backendID)
	return err
}

// GetTasks returns all tasks in a list for this backend
func (b *Backend) GetTasks(ctx context.Context, listID string) ([]backend.Task, error) {
	rows, err := b.db.QueryContext(ctx,
		`SELECT id, list_id, summary, description, status, priority, due_date, start_date, completed, created, modified, parent_id, categories, recurrence, recur_from_due
		 FROM tasks WHERE list_id = ? AND backend_id = ?`,
		listID, b.backendID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tasks []backend.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *t)
	}

	if tasks == nil {
		tasks = []backend.Task{}
	}
	return tasks, rows.Err()
}

// GetTask returns a specific task for this backend
func (b *Backend) GetTask(ctx context.Context, listID, taskID string) (*backend.Task, error) {
	row := b.db.QueryRowContext(ctx,
		`SELECT id, list_id, summary, description, status, priority, due_date, start_date, completed, created, modified, parent_id, categories, recurrence, recur_from_due
		 FROM tasks WHERE list_id = ? AND id = ? AND backend_id = ?`,
		listID, taskID, b.backendID,
	)

	t, err := scanTaskRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

// GetTaskByLocalID returns a task by its SQLite rowid (local ID) for this backend
func (b *Backend) GetTaskByLocalID(ctx context.Context, listID string, localID int64) (*backend.Task, error) {
	row := b.db.QueryRowContext(ctx,
		`SELECT id, list_id, summary, description, status, priority, due_date, start_date, completed, created, modified, parent_id, categories, recurrence, recur_from_due
		 FROM tasks WHERE list_id = ? AND rowid = ? AND backend_id = ?`,
		listID, localID, b.backendID,
	)

	t, err := scanTaskRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

// GetTaskLocalID returns the SQLite rowid for a task for this backend
func (b *Backend) GetTaskLocalID(ctx context.Context, taskID string) (int64, error) {
	var localID int64
	err := b.db.QueryRowContext(ctx,
		`SELECT rowid FROM tasks WHERE id = ? AND backend_id = ?`,
		taskID, b.backendID,
	).Scan(&localID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return localID, err
}

// timeToNullString converts a *time.Time to sql.NullString for database storage.
func timeToNullString(t *time.Time) sql.NullString {
	if t == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: t.Format(time.RFC3339Nano), Valid: true}
}

// parseOptionalDate parses a nullable date string and returns a pointer to time.Time.
func parseOptionalDate(str sql.NullString) *time.Time {
	if str.Valid && str.String != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, str.String); err == nil {
			return &parsed
		}
	}
	return nil
}

// parseDateStrings parses the nullable date strings and populates the task's date fields.
func parseDateStrings(t *backend.Task, dueDateStr, startDateStr, completedStr, createdStr, modifiedStr sql.NullString) {
	if createdStr.Valid {
		t.Created, _ = time.Parse(time.RFC3339Nano, createdStr.String)
	}
	if modifiedStr.Valid {
		t.Modified, _ = time.Parse(time.RFC3339Nano, modifiedStr.String)
	}
	t.DueDate = parseOptionalDate(dueDateStr)
	t.StartDate = parseOptionalDate(startDateStr)
	t.Completed = parseOptionalDate(completedStr)
}

// scanner is an interface satisfied by both *sql.Rows and *sql.Row
type scanner interface {
	Scan(dest ...any) error
}

// scanTaskFrom scans a task from any scanner (Rows or Row)
func scanTaskFrom(s scanner) (*backend.Task, error) {
	var t backend.Task
	var dueDateStr, startDateStr, completedStr, createdStr, modifiedStr sql.NullString
	var categoriesStr, recurrenceStr sql.NullString
	var recurFromDue sql.NullInt64

	err := s.Scan(
		&t.ID, &t.ListID, &t.Summary, &t.Description, &t.Status,
		&t.Priority, &dueDateStr, &startDateStr, &completedStr, &createdStr, &modifiedStr, &t.ParentID, &categoriesStr,
		&recurrenceStr, &recurFromDue,
	)
	if err != nil {
		return nil, err
	}

	parseDateStrings(&t, dueDateStr, startDateStr, completedStr, createdStr, modifiedStr)
	if categoriesStr.Valid {
		t.Categories = categoriesStr.String
	}
	if recurrenceStr.Valid {
		t.Recurrence = recurrenceStr.String
	}
	if recurFromDue.Valid {
		t.RecurFromDue = recurFromDue.Int64 == 1
	} else {
		t.RecurFromDue = true // default to from due date
	}
	return &t, nil
}

// scanTask scans a task from a Rows result
func scanTask(rows *sql.Rows) (*backend.Task, error) {
	return scanTaskFrom(rows)
}

// scanTaskRow scans a task from a Row result
func scanTaskRow(row *sql.Row) (*backend.Task, error) {
	return scanTaskFrom(row)
}

// CreateTask adds a new task to a list for this backend
func (b *Backend) CreateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	id := uuid.New().String()
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339Nano)

	dueDateStr := timeToNullString(task.DueDate)
	startDateStr := timeToNullString(task.StartDate)
	completedStr := timeToNullString(task.Completed)

	status := task.Status
	if status == "" {
		status = backend.StatusNeedsAction
	}

	// Convert bool to int for SQLite storage
	recurFromDueInt := 1
	if !task.RecurFromDue {
		recurFromDueInt = 0
	}

	_, err := b.db.ExecContext(ctx,
		`INSERT INTO tasks (id, list_id, summary, description, status, priority, due_date, start_date, completed, created, modified, parent_id, categories, recurrence, recur_from_due, backend_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, listID, task.Summary, task.Description, status, task.Priority,
		dueDateStr, startDateStr, completedStr, nowStr, nowStr, task.ParentID, task.Categories, task.Recurrence, recurFromDueInt, b.backendID,
	)
	if err != nil {
		return nil, err
	}

	return &backend.Task{
		ID:           id,
		ListID:       listID,
		Summary:      task.Summary,
		Description:  task.Description,
		Status:       status,
		Priority:     task.Priority,
		DueDate:      task.DueDate,
		StartDate:    task.StartDate,
		Completed:    task.Completed,
		Created:      now,
		Modified:     now,
		ParentID:     task.ParentID,
		Categories:   task.Categories,
		Recurrence:   task.Recurrence,
		RecurFromDue: task.RecurFromDue,
	}, nil
}

// UpdateTask modifies an existing task for this backend
func (b *Backend) UpdateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339Nano)

	dueDateStr := timeToNullString(task.DueDate)
	startDateStr := timeToNullString(task.StartDate)
	completedStr := timeToNullString(task.Completed)

	// Convert bool to int for SQLite storage
	recurFromDueInt := 1
	if !task.RecurFromDue {
		recurFromDueInt = 0
	}

	_, err := b.db.ExecContext(ctx,
		`UPDATE tasks SET summary = ?, description = ?, status = ?, priority = ?, due_date = ?, start_date = ?, completed = ?, modified = ?, parent_id = ?, categories = ?, recurrence = ?, recur_from_due = ?
		 WHERE id = ? AND list_id = ? AND backend_id = ?`,
		task.Summary, task.Description, task.Status, task.Priority, dueDateStr, startDateStr, completedStr, nowStr, task.ParentID, task.Categories, task.Recurrence, recurFromDueInt,
		task.ID, listID, b.backendID,
	)
	if err != nil {
		return nil, err
	}

	// Fetch the updated task to get all fields including Created
	return b.GetTask(ctx, listID, task.ID)
}

// DeleteTask removes a task for this backend
func (b *Backend) DeleteTask(ctx context.Context, listID, taskID string) error {
	_, err := b.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ? AND list_id = ? AND backend_id = ?", taskID, listID, b.backendID)
	return err
}

// Close closes the database connection
func (b *Backend) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

// DatabaseStats contains statistics about the SQLite database
type DatabaseStats struct {
	TotalTasks        int            `json:"total_tasks"`
	Lists             []ListStats    `json:"lists"`
	ByStatus          map[string]int `json:"by_status"`
	DatabaseSizeBytes int64          `json:"database_size_bytes"`
	LastVacuum        *time.Time     `json:"last_vacuum,omitempty"`
}

// ListStats contains task statistics for a single list
type ListStats struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// VacuumResult contains the result of a vacuum operation
type VacuumResult struct {
	SizeBefore int64 `json:"size_before"`
	SizeAfter  int64 `json:"size_after"`
	Reclaimed  int64 `json:"reclaimed"`
}

// Stats returns database statistics
func (b *Backend) Stats(ctx context.Context, listName string) (*DatabaseStats, error) {
	stats := &DatabaseStats{
		ByStatus: make(map[string]int),
	}

	// Get all lists
	lists, err := b.GetLists(ctx)
	if err != nil {
		return nil, err
	}

	// If specific list requested, filter
	if listName != "" {
		var filtered []backend.List
		for _, l := range lists {
			if l.Name == listName {
				filtered = append(filtered, l)
				break
			}
		}
		lists = filtered
	}

	// Get stats per list
	for _, l := range lists {
		tasks, err := b.GetTasks(ctx, l.ID)
		if err != nil {
			return nil, err
		}
		stats.Lists = append(stats.Lists, ListStats{Name: l.Name, Count: len(tasks)})
		stats.TotalTasks += len(tasks)

		// Count by status
		for _, t := range tasks {
			statusKey := string(t.Status)
			// Map internal status to display status
			switch t.Status {
			case backend.StatusNeedsAction:
				statusKey = "TODO"
			case backend.StatusCompleted:
				statusKey = "DONE"
			case backend.StatusInProgress:
				statusKey = "IN-PROGRESS"
			case backend.StatusCancelled:
				statusKey = "CANCELLED"
			}
			stats.ByStatus[statusKey]++
		}
	}

	// Get database size using PRAGMA
	var pageCount, pageSize int64
	if err := b.db.QueryRowContext(ctx, "PRAGMA page_count").Scan(&pageCount); err != nil {
		return nil, err
	}
	if err := b.db.QueryRowContext(ctx, "PRAGMA page_size").Scan(&pageSize); err != nil {
		return nil, err
	}
	stats.DatabaseSizeBytes = pageCount * pageSize

	// Try to get last vacuum time from metadata (if table exists)
	var lastVacuumStr string
	err = b.db.QueryRowContext(ctx, "SELECT value FROM metadata WHERE key = 'last_vacuum'").Scan(&lastVacuumStr)
	if err == nil {
		if t, err := time.Parse(time.RFC3339, lastVacuumStr); err == nil {
			stats.LastVacuum = &t
		}
	}

	return stats, nil
}

// Vacuum runs the SQLite VACUUM command to reclaim space
func (b *Backend) Vacuum(ctx context.Context) (*VacuumResult, error) {
	result := &VacuumResult{}

	// Get size before vacuum
	var pageCount, pageSize int64
	if err := b.db.QueryRowContext(ctx, "PRAGMA page_count").Scan(&pageCount); err != nil {
		return nil, err
	}
	if err := b.db.QueryRowContext(ctx, "PRAGMA page_size").Scan(&pageSize); err != nil {
		return nil, err
	}
	result.SizeBefore = pageCount * pageSize

	// Run VACUUM
	if _, err := b.db.ExecContext(ctx, "VACUUM"); err != nil {
		return nil, fmt.Errorf("vacuum failed: %w", err)
	}

	// Get size after vacuum
	if err := b.db.QueryRowContext(ctx, "PRAGMA page_count").Scan(&pageCount); err != nil {
		return nil, err
	}
	result.SizeAfter = pageCount * pageSize
	result.Reclaimed = result.SizeBefore - result.SizeAfter

	// Store last vacuum time in metadata table
	b.ensureMetadataTable(ctx)
	now := time.Now().UTC().Format(time.RFC3339)
	_, _ = b.db.ExecContext(ctx, "INSERT OR REPLACE INTO metadata (key, value) VALUES ('last_vacuum', ?)", now)

	return result, nil
}

// ensureMetadataTable creates the metadata table if it doesn't exist
func (b *Backend) ensureMetadataTable(ctx context.Context) {
	_, _ = b.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS metadata (key TEXT PRIMARY KEY, value TEXT)`)
}

// DetectableBackend wraps Backend with auto-detection capabilities
type DetectableBackend struct {
	*Backend
	dbPath string
}

// NewDetectable creates a DetectableBackend for the given database path.
// It ensures the parent directory exists before opening the database,
// since SQLite is designed to be "always available" as a fallback.
func NewDetectable(dbPath string) (*DetectableBackend, error) {
	// Ensure parent directory exists before opening database
	// This prevents confusing "out of memory" errors on fresh installs
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	be, err := New(dbPath)
	if err != nil {
		return nil, err
	}
	return &DetectableBackend{
		Backend: be,
		dbPath:  dbPath,
	}, nil
}

// CanDetect returns true - SQLite is always available as a fallback
func (d *DetectableBackend) CanDetect() (bool, error) {
	return true, nil
}

// DetectionInfo returns information about the SQLite database
func (d *DetectableBackend) DetectionInfo() string {
	return d.dbPath + " (always available)"
}

// Verify interface compliance at compile time
var _ backend.TaskManager = (*Backend)(nil)
var _ backend.DetectableBackend = (*DetectableBackend)(nil)
