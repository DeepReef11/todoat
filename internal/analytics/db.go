package analytics

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Schema for the analytics database
const schema = `
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,
    command TEXT NOT NULL,
    subcommand TEXT,
    backend TEXT,
    success INTEGER NOT NULL,
    duration_ms INTEGER,
    error_type TEXT,
    flags TEXT,
    created_at INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_timestamp ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_command ON events(command);
CREATE INDEX IF NOT EXISTS idx_backend ON events(backend);
CREATE INDEX IF NOT EXISTS idx_success ON events(success);
CREATE INDEX IF NOT EXISTS idx_created_at ON events(created_at);
`

// openDB opens or creates the analytics database at the given path
func openDB(dbPath string) (*sql.DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create analytics directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open analytics database: %w", err)
	}

	// Initialize schema
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize analytics schema: %w", err)
	}

	return db, nil
}
