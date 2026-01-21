# [015] Views and Customization System

## Summary
Implement custom view configurations supporting field selection, ordering, filtering, sorting, and plugin formatters. Views are defined in YAML files and selected via CLI `-v` flag.

## Documentation Reference
- Primary: `docs/explanation/views-customization.md`
- Related: `docs/explanation/cli-interface.md`, `docs/explanation/task-management.md`

## Dependencies
- Requires: [004] Task Commands
- Requires: [010] Configuration (for YAML parsing and XDG paths)

## Complexity
L

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestDefaultView` - `todoat MyList` displays tasks with default view (status, summary, priority)
- [ ] `TestAllView` - `todoat MyList -v all` displays all task metadata fields
- [ ] `TestCustomViewSelection` - `todoat MyList -v myview` loads view from `~/.config/todoat/views/myview.yaml`
- [ ] `TestViewListCommand` - `todoat view list` shows all available views (built-in and custom)
- [ ] `TestViewFieldOrdering` - Custom view with reordered fields displays columns in specified order
- [ ] `TestViewFiltering` - View with filters only shows matching tasks (e.g., status != DONE)
- [ ] `TestViewSorting` - View with sort rules orders tasks correctly (multi-level sort)
- [ ] `TestViewDateFilter` - View filters with relative dates (`today`, `+7d`, `-3d`) work correctly
- [ ] `TestViewTagFilter` - View filters on tags/categories work with `contains` and `in` operators
- [ ] `TestViewHierarchyPreserved` - Custom views maintain parent-child tree structure display
- [ ] `TestInvalidViewError` - Invalid view name shows helpful error message

## Implementation Notes
- Built-in views (default, all) are hard-coded, cannot be deleted
- Custom views stored in `~/.config/todoat/views/*.yaml`
- View loading uses lazy initialization with caching
- Filter operators: eq, ne, lt, lte, gt, gte, contains, in, not_in, regex
- Hierarchical display takes precedence over sorting (parents always before children)

## Out of Scope
- Interactive TUI view builder (separate roadmap item)
- Plugin formatters with external scripts (separate roadmap item)
- Real-time view preview during creation
