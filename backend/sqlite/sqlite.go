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
	db *sql.DB
}

// New creates a new SQLite backend and initializes the database schema
func New(path string) (*Backend, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	b := &Backend{db: db}
	if err := b.initSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return b, nil
}

// initSchema creates the database tables if they don't exist
func (b *Backend) initSchema() error {
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

	// Enable foreign keys
	if _, err := b.db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return err
	}

	_, err := b.db.Exec(schema)
	if err != nil {
		return err
	}

	// Migration: add deleted_at column if it doesn't exist
	_, _ = b.db.Exec("ALTER TABLE task_lists ADD COLUMN deleted_at TEXT")

	return nil
}

// GetLists returns all active (non-deleted) task lists
func (b *Backend) GetLists(ctx context.Context) ([]backend.List, error) {
	rows, err := b.db.QueryContext(ctx, "SELECT id, name, color, modified FROM task_lists WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var lists []backend.List
	for rows.Next() {
		var l backend.List
		var modifiedStr string
		if err := rows.Scan(&l.ID, &l.Name, &l.Color, &modifiedStr); err != nil {
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

// GetList returns a specific active list by ID
func (b *Backend) GetList(ctx context.Context, listID string) (*backend.List, error) {
	var l backend.List
	var modifiedStr string
	err := b.db.QueryRowContext(ctx,
		"SELECT id, name, color, modified FROM task_lists WHERE id = ? AND deleted_at IS NULL",
		listID,
	).Scan(&l.ID, &l.Name, &l.Color, &modifiedStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	l.Modified, _ = time.Parse(time.RFC3339Nano, modifiedStr)
	return &l, nil
}

// GetListByName returns a specific active list by name (case-insensitive)
func (b *Backend) GetListByName(ctx context.Context, name string) (*backend.List, error) {
	var l backend.List
	var modifiedStr string
	err := b.db.QueryRowContext(ctx,
		"SELECT id, name, color, modified FROM task_lists WHERE LOWER(name) = LOWER(?) AND deleted_at IS NULL",
		name,
	).Scan(&l.ID, &l.Name, &l.Color, &modifiedStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	l.Modified, _ = time.Parse(time.RFC3339Nano, modifiedStr)
	return &l, nil
}

// CreateList creates a new task list
func (b *Backend) CreateList(ctx context.Context, name string) (*backend.List, error) {
	id := uuid.New().String()
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339Nano)

	_, err := b.db.ExecContext(ctx,
		"INSERT INTO task_lists (id, name, color, modified) VALUES (?, ?, '', ?)",
		id, name, nowStr,
	)
	if err != nil {
		return nil, err
	}

	return &backend.List{
		ID:       id,
		Name:     name,
		Color:    "",
		Modified: now,
	}, nil
}

// DeleteList soft-deletes a task list (moves to trash)
func (b *Backend) DeleteList(ctx context.Context, listID string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := b.db.ExecContext(ctx, "UPDATE task_lists SET deleted_at = ? WHERE id = ?", now, listID)
	return err
}

// GetDeletedLists returns all deleted (trashed) task lists
func (b *Backend) GetDeletedLists(ctx context.Context) ([]backend.List, error) {
	rows, err := b.db.QueryContext(ctx, "SELECT id, name, color, modified, deleted_at FROM task_lists WHERE deleted_at IS NOT NULL")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var lists []backend.List
	for rows.Next() {
		var l backend.List
		var modifiedStr string
		var deletedAtStr sql.NullString
		if err := rows.Scan(&l.ID, &l.Name, &l.Color, &modifiedStr, &deletedAtStr); err != nil {
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

// GetDeletedListByName returns a specific deleted list by name (case-insensitive)
func (b *Backend) GetDeletedListByName(ctx context.Context, name string) (*backend.List, error) {
	var l backend.List
	var modifiedStr string
	var deletedAtStr sql.NullString
	err := b.db.QueryRowContext(ctx,
		"SELECT id, name, color, modified, deleted_at FROM task_lists WHERE LOWER(name) = LOWER(?) AND deleted_at IS NOT NULL",
		name,
	).Scan(&l.ID, &l.Name, &l.Color, &modifiedStr, &deletedAtStr)

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

// RestoreList restores a deleted list from trash
func (b *Backend) RestoreList(ctx context.Context, listID string) error {
	_, err := b.db.ExecContext(ctx, "UPDATE task_lists SET deleted_at = NULL WHERE id = ?", listID)
	return err
}

// PurgeList permanently deletes a list and all its tasks
func (b *Backend) PurgeList(ctx context.Context, listID string) error {
	// First delete all tasks in this list
	_, err := b.db.ExecContext(ctx, "DELETE FROM tasks WHERE list_id = ?", listID)
	if err != nil {
		return err
	}

	_, err = b.db.ExecContext(ctx, "DELETE FROM task_lists WHERE id = ?", listID)
	return err
}

// GetTasks returns all tasks in a list
func (b *Backend) GetTasks(ctx context.Context, listID string) ([]backend.Task, error) {
	rows, err := b.db.QueryContext(ctx,
		`SELECT id, list_id, summary, description, status, priority, due_date, start_date, completed, created, modified, parent_id, categories
		 FROM tasks WHERE list_id = ?`,
		listID,
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

// GetTask returns a specific task
func (b *Backend) GetTask(ctx context.Context, listID, taskID string) (*backend.Task, error) {
	row := b.db.QueryRowContext(ctx,
		`SELECT id, list_id, summary, description, status, priority, due_date, start_date, completed, created, modified, parent_id, categories
		 FROM tasks WHERE list_id = ? AND id = ?`,
		listID, taskID,
	)

	t, err := scanTaskRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

// GetTaskByLocalID returns a task by its SQLite rowid (local ID)
func (b *Backend) GetTaskByLocalID(ctx context.Context, listID string, localID int64) (*backend.Task, error) {
	row := b.db.QueryRowContext(ctx,
		`SELECT id, list_id, summary, description, status, priority, due_date, start_date, completed, created, modified, parent_id, categories
		 FROM tasks WHERE list_id = ? AND rowid = ?`,
		listID, localID,
	)

	t, err := scanTaskRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

// GetTaskLocalID returns the SQLite rowid for a task
func (b *Backend) GetTaskLocalID(ctx context.Context, taskID string) (int64, error) {
	var localID int64
	err := b.db.QueryRowContext(ctx,
		`SELECT rowid FROM tasks WHERE id = ?`,
		taskID,
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
	var categoriesStr sql.NullString

	err := s.Scan(
		&t.ID, &t.ListID, &t.Summary, &t.Description, &t.Status,
		&t.Priority, &dueDateStr, &startDateStr, &completedStr, &createdStr, &modifiedStr, &t.ParentID, &categoriesStr,
	)
	if err != nil {
		return nil, err
	}

	parseDateStrings(&t, dueDateStr, startDateStr, completedStr, createdStr, modifiedStr)
	if categoriesStr.Valid {
		t.Categories = categoriesStr.String
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

// CreateTask adds a new task to a list
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

	_, err := b.db.ExecContext(ctx,
		`INSERT INTO tasks (id, list_id, summary, description, status, priority, due_date, start_date, completed, created, modified, parent_id, categories)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, listID, task.Summary, task.Description, status, task.Priority,
		dueDateStr, startDateStr, completedStr, nowStr, nowStr, task.ParentID, task.Categories,
	)
	if err != nil {
		return nil, err
	}

	return &backend.Task{
		ID:          id,
		ListID:      listID,
		Summary:     task.Summary,
		Description: task.Description,
		Status:      status,
		Priority:    task.Priority,
		DueDate:     task.DueDate,
		StartDate:   task.StartDate,
		Completed:   task.Completed,
		Created:     now,
		Modified:    now,
		ParentID:    task.ParentID,
		Categories:  task.Categories,
	}, nil
}

// UpdateTask modifies an existing task
func (b *Backend) UpdateTask(ctx context.Context, listID string, task *backend.Task) (*backend.Task, error) {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339Nano)

	dueDateStr := timeToNullString(task.DueDate)
	startDateStr := timeToNullString(task.StartDate)
	completedStr := timeToNullString(task.Completed)

	_, err := b.db.ExecContext(ctx,
		`UPDATE tasks SET summary = ?, description = ?, status = ?, priority = ?, due_date = ?, start_date = ?, completed = ?, modified = ?, parent_id = ?, categories = ?
		 WHERE id = ? AND list_id = ?`,
		task.Summary, task.Description, task.Status, task.Priority, dueDateStr, startDateStr, completedStr, nowStr, task.ParentID, task.Categories,
		task.ID, listID,
	)
	if err != nil {
		return nil, err
	}

	// Fetch the updated task to get all fields including Created
	return b.GetTask(ctx, listID, task.ID)
}

// DeleteTask removes a task
func (b *Backend) DeleteTask(ctx context.Context, listID, taskID string) error {
	_, err := b.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ? AND list_id = ?", taskID, listID)
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
				statusKey = "PROCESSING"
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
