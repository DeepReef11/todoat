# [068] View Create Filter Flags

## Summary
Add `--filter-status` and `--filter-priority` flags to the `view create` command to allow creating views with filters from the command line.

## Documentation Reference
- Primary: `docs/reference/cli.md`
- Section: Line 258 (view create example)
- Issue: `issues/071-cli-reference-view-create-flags-mismatch.md`

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] `TestViewCreate_FilterStatus` - Verify --filter-status flag creates view with status filter
- [ ] `TestViewCreate_FilterPriority` - Verify --filter-priority flag creates view with priority filter
- [ ] `TestViewCreate_CombinedFilters` - Verify multiple filter flags can be combined

### Functional Requirements
- [ ] `--filter-status` flag accepts comma-separated status values (e.g., "TODO,IN-PROGRESS")
- [ ] `--filter-priority` flag accepts priority values (e.g., "high", "1-3", "low")
- [ ] Flags work with `-y` (non-interactive mode)
- [ ] Filters saved to view YAML file correctly
- [ ] Created views work correctly when used with task listing

## Implementation Notes
- Add flags to view create command in todoat.go
- Parse status values as comma-separated list
- Parse priority as "high", "medium", "low" or numeric ranges
- Update view YAML writer to include filter configuration
- Follow existing flag pattern used for `--fields` and `--sort`

## Out of Scope
- Interactive filter configuration (already exists in TUI builder)
- Filter by tags (can be added later)
- Filter by dates (can be added later)
