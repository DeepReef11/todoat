# [067] Documentation uses db_path but config uses path for SQLite

## Type
doc-mismatch

## Category
other

## Severity
medium

## Steps to Reproduce
```bash
# Documentation shows this format:
cat > ~/.config/todoat/config.yaml << EOF
backends:
  sqlite:
    type: sqlite
    enabled: true
    db_path: "~/my-tasks/tasks.db"  # From docs
EOF

# But the actual config struct uses 'path':
cat > ~/.config/todoat/config.yaml << EOF
backends:
  sqlite:
    type: sqlite
    enabled: true
    path: "~/my-tasks/tasks.db"  # Actual field name
EOF
```

## Expected Behavior
Documentation should match the actual configuration field names.

## Actual Behavior
Documentation uses `db_path` but the actual YAML field name is `path`.

The Go struct definition in `internal/config/config.go:67-70`:
```go
type SQLiteConfig struct {
    Enabled bool   `yaml:"enabled"`
    Path    string `yaml:"path"`
}
```

## Error Output
No error - the `db_path` field is silently ignored, and the default path is used.

## Environment
- OS: Linux
- Runtime version: Go (any version)

## Possible Cause
Documentation was written with a different field name than what was implemented.

## Documentation Reference
- File: `docs/getting-started.md`
- Section: Minimal SQLite Configuration
- Documented key: `db_path`
- Actual key: `path`

Files with incorrect `db_path` references:
- `docs/getting-started.md:47` - `db_path: ""`
- `docs/backends.md:120` - `db_path: ""`
- `docs/backends.md:132` - `db_path: "~/my-tasks/tasks.db"`
- `docs/configuration.md:87` - `todoat config set backends.sqlite.db_path`

The sample config (`internal/config/config.sample.yaml:6`) correctly uses `path`.

## Related Files
- `docs/getting-started.md`
- `docs/backends.md`
- `docs/configuration.md`
- `internal/config/config.go`
- `internal/config/config.sample.yaml`

## Recommended Fix
FIX DOCS - Update documentation files to use `path` instead of `db_path` to match the actual configuration field name.

## Resolution

**Fixed in**: commit 196d9be
**Fix description**: Updated all documentation files to use `path` instead of `db_path` for SQLite configuration

### Verification Log
```bash
$ grep -n "db_path" docs/getting-started.md docs/backends.md docs/configuration.md
No db_path references found

$ grep -n "path:" docs/getting-started.md docs/backends.md docs/configuration.md
docs/getting-started.md:47:    path: ""  # Uses default location
docs/backends.md:120:    path: ""  # Empty = default location
docs/backends.md:132:    path: "~/my-tasks/tasks.db"

$ grep -n "backends.sqlite" docs/configuration.md
87:todoat config set backends.sqlite.path "~/my-tasks/tasks.db"
```
**Matches expected behavior**: YES - All documentation now uses `path` instead of `db_path`
