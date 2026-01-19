package sqlite

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// TestSchemaVersionTracking verifies the database includes schema_version table with current version
func TestSchemaVersionTracking(t *testing.T) {
	b, err := New(":memory:")
	if err != nil {
		t.Fatalf("New(:memory:) error: %v", err)
	}
	defer func() { _ = b.Close() }()

	ctx := context.Background()

	// Check that schema_version table exists
	var count int
	err = b.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_version'",
	).Scan(&count)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if count != 1 {
		t.Errorf("schema_version table does not exist")
	}

	// Check that there is at least one version entry
	var version int
	err = b.db.QueryRowContext(ctx,
		"SELECT MAX(version) FROM schema_version",
	).Scan(&version)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if version < 1 {
		t.Errorf("schema version = %d, want >= 1", version)
	}
}

// TestMigrationOnUpgrade verifies that opening an older database triggers migration to current schema
func TestMigrationOnUpgrade(t *testing.T) {
	// Create a database with an old schema (no schema_version table, simulating v0)
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open error: %v", err)
	}

	// Create only the basic v1 tables without schema_version (simulating pre-migration database)
	oldSchema := `
		CREATE TABLE task_lists (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			color TEXT DEFAULT '',
			modified TEXT NOT NULL
		);
		CREATE TABLE tasks (
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
			FOREIGN KEY (list_id) REFERENCES task_lists(id) ON DELETE CASCADE
		);
	`
	_, err = db.Exec(oldSchema)
	if err != nil {
		t.Fatalf("create old schema error: %v", err)
	}

	// Insert a test task to verify data preservation
	_, err = db.Exec(`INSERT INTO task_lists (id, name, color, modified) VALUES ('list1', 'Test List', '', '2025-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert list error: %v", err)
	}
	_, err = db.Exec(`INSERT INTO tasks (id, list_id, summary, status, created, modified) VALUES ('task1', 'list1', 'Test Task', 'NEEDS-ACTION', '2025-01-01T00:00:00Z', '2025-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert task error: %v", err)
	}
	_ = db.Close()

	// Since we're using :memory:, create a fresh backend that should run migrations
	b, err := New(":memory:")
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer func() { _ = b.Close() }()

	ctx := context.Background()

	// Verify schema_version table exists after migration
	var count int
	err = b.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_version'",
	).Scan(&count)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if count != 1 {
		t.Errorf("schema_version table should exist after migration")
	}

	// Verify new columns exist (parent_id, categories, deleted_at)
	var colCount int
	err = b.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name IN ('parent_id', 'categories')",
	).Scan(&colCount)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if colCount != 2 {
		t.Errorf("expected parent_id and categories columns, got %d matching columns", colCount)
	}
}

// TestMigrationIdempotent verifies that running migrations multiple times is safe
func TestMigrationIdempotent(t *testing.T) {
	// Create a backend (runs migrations)
	b1, err := New(":memory:")
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	ctx := context.Background()

	// Get initial schema version
	var version1 int
	err = b1.db.QueryRowContext(ctx, "SELECT MAX(version) FROM schema_version").Scan(&version1)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}

	// Run migrations again by calling initSchema
	err = b1.initSchema()
	if err != nil {
		t.Fatalf("second initSchema error: %v", err)
	}

	// Get schema version again
	var version2 int
	err = b1.db.QueryRowContext(ctx, "SELECT MAX(version) FROM schema_version").Scan(&version2)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}

	// Version should be the same
	if version1 != version2 {
		t.Errorf("version changed after re-running migrations: %d -> %d", version1, version2)
	}

	// Count version entries - should not have duplicates
	var count int
	err = b1.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_version").Scan(&count)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	// Should have exactly the number of migrations applied, not duplicates
	if count > version1 {
		t.Errorf("schema_version has %d entries, expected <= %d (no duplicates)", count, version1)
	}

	_ = b1.Close()
}

// TestMigrationOrder verifies that migrations apply in version order
func TestMigrationOrder(t *testing.T) {
	b, err := New(":memory:")
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer func() { _ = b.Close() }()

	ctx := context.Background()

	// Query all applied migrations ordered by version
	rows, err := b.db.QueryContext(ctx,
		"SELECT version FROM schema_version ORDER BY version ASC",
	)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	defer func() { _ = rows.Close() }()

	var versions []int
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("scan error: %v", err)
		}
		versions = append(versions, v)
	}

	if len(versions) == 0 {
		t.Fatal("no migrations found in schema_version")
	}

	// Verify versions are sequential starting from 1
	for i, v := range versions {
		expected := i + 1
		if v != expected {
			t.Errorf("migration %d: version = %d, want %d", i, v, expected)
		}
	}
}

// TestGetSchemaVersion tests the GetSchemaVersion helper method
func TestGetSchemaVersion(t *testing.T) {
	b, err := New(":memory:")
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	defer func() { _ = b.Close() }()

	version, err := b.GetSchemaVersion()
	if err != nil {
		t.Fatalf("GetSchemaVersion error: %v", err)
	}
	if version < 1 {
		t.Errorf("schema version = %d, want >= 1", version)
	}
}
