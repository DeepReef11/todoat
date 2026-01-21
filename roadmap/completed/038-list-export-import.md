# [038] List Export/Import

## Summary
Implement list export and import functionality for SQLite backend, enabling backup, restore, and data portability across todoat instances.

## Documentation Reference
- Primary: `docs/explanation/list-management.md` (Export/Import section)
- Related: `docs/explanation/backend-system.md`

## Dependencies
- Requires: [003] SQLite Backend
- Requires: [007] List Commands

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestListExportSQLite` - `todoat list export "MyList" --format sqlite` creates standalone db file
- [ ] `TestListExportJSON` - `todoat list export "MyList" --format json` creates JSON file
- [ ] `TestListExportCSV` - `todoat list export "MyList" --format csv` creates CSV file
- [ ] `TestListExportICalendar` - `todoat list export "MyList" --format ical` creates .ics file
- [ ] `TestListImport` - `todoat list import backup.db` restores list from exported file

### Functional Requirements
- [ ] `todoat list export "<name>" --format <format>` command exports a list
- [ ] Supported export formats:
  - `sqlite` - Standalone SQLite database file
  - `json` - JSON array of tasks with metadata
  - `csv` - Comma-separated values for spreadsheet import
  - `ical` - iCalendar VTODO format for calendar apps
- [ ] Default export path: `./<list-name>.<ext>` (overridable with `--output`)
- [ ] Export includes all task metadata (UID, summary, description, status, priority, dates, categories, parent relationships)
- [ ] `todoat list import <file>` imports from any supported format
- [ ] Import detects format from file extension or `--format` flag
- [ ] Conflict handling: `--skip`, `--replace`, `--rename` (default: prompt)
- [ ] Progress indicator for large exports/imports

### Output Requirements
- [ ] Export success message with file path and task count
- [ ] Import success message with imported task count
- [ ] JSON mode returns `{"action": "export/import", "file": "...", "task_count": N}`

## Implementation Notes
- Use SQLite ATTACH DATABASE for sqlite format export
- Use encoding/json for JSON format
- Use encoding/csv for CSV format
- Use existing iCalendar serialization from Nextcloud backend for ical format
- Preserve task UIDs during import to maintain references
- Handle parent-child relationships correctly (import parents before children)

## Out of Scope
- Incremental/differential exports
- Scheduled automatic backups
- Cloud storage integration (S3, etc.)
- Export to proprietary formats (Todoist JSON, etc.)
