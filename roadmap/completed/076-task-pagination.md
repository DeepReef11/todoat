# [076] Task Pagination

## Summary
Implement pagination for task listing to improve performance and usability with large task sets (>5000 tasks).

## Documentation Reference
- Primary: `docs/explanation/views-customization.md`
- Section: "Performance Characteristics" - notes "Large task sets (>5000): Consider pagination (not yet implemented)"

## Dependencies
- Requires: [015] Views Customization (completed)

## Complexity
M

## Acceptance Criteria

### Tests Required
- [ ] `TestPaginationDefault` - `todoat MyList` shows first page with default page size
- [ ] `TestPaginationWithLimit` - `todoat MyList --limit 20` limits output to 20 tasks
- [ ] `TestPaginationWithOffset` - `todoat MyList --offset 20 --limit 20` shows second page
- [ ] `TestPaginationPageFlag` - `todoat MyList --page 2` shows second page
- [ ] `TestPaginationWithFilters` - Pagination works correctly with view filters
- [ ] `TestPaginationWithSort` - Pagination preserves sort order
- [ ] `TestPaginationTotalCount` - Output includes total task count for UI navigation
- [ ] `TestPaginationJSONMetadata` - JSON output includes pagination metadata

### Functional Requirements
- Default behavior unchanged for small task sets (<100 tasks)
- `--limit N` flag to limit number of tasks shown
- `--offset N` flag to skip first N tasks
- `--page N` convenience flag (calculates offset from page number)
- `--page-size N` configurable page size (default: 50)
- Output footer shows pagination info: "Showing 1-50 of 5123 tasks"
- JSON output includes `total`, `page`, `pageSize`, `hasMore` metadata

## Implementation Notes
- Modify `internal/views/renderer.go` to support pagination
- Add pagination flags to task list commands
- Implement efficient database queries with LIMIT/OFFSET
- Consider cursor-based pagination for very large sets
- Ensure hierarchical view maintains parent-child grouping across pages

## Out of Scope
- Infinite scroll (TUI feature)
- Keyset/cursor pagination (can be future optimization)
- Per-view pagination settings
