# [053] Task Categories/Tags Management

## Summary
Implement CLI commands for managing task categories (tags), including adding, removing, and listing tags on tasks.

## Documentation Reference
- Primary: `docs/explanation/task-management.md`
- Section: Task Metadata (Categories field)
- Related: `docs/explanation/cli-interface.md`

## Dependencies
- Requires: [004] Task Commands
- Requires: [012] Tag Filtering

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestTaskAddWithTags` - `todoat MyList add "Task" --tags "work,urgent"` creates task with tags
- [ ] `TestTaskUpdateAddTag` - `todoat MyList update "Task" --add-tag "important"` adds tag to existing
- [ ] `TestTaskUpdateRemoveTag` - `todoat MyList update "Task" --remove-tag "urgent"` removes specific tag
- [ ] `TestTaskUpdateClearTags` - `todoat MyList update "Task" --tags ""` clears all tags
- [ ] `TestTaskTagsDisplay` - Tags shown in task listing with view that includes tags field
- [ ] `TestListTags` - `todoat tags` lists all unique tags across all tasks
- [ ] `TestListTagsJSON` - `todoat --json tags` returns JSON array of tags

### Functional Requirements
- [ ] `add --tags "tag1,tag2"` creates task with comma-separated tags
- [ ] `update --add-tag "tag"` appends tag without removing existing
- [ ] `update --remove-tag "tag"` removes specific tag
- [ ] `update --tags "tag1,tag2"` replaces all tags
- [ ] `update --tags ""` clears all tags
- [ ] `todoat tags` lists all unique tags in use
- [ ] `todoat tags --list "ListName"` shows tags only from specific list
- [ ] Tags are case-preserved but comparison is case-insensitive

## Implementation Notes

### Tag Storage
- **SQLite**: JSON array in categories column
- **Nextcloud**: CATEGORIES property (comma-separated in iCalendar)
- **Todoist**: Mapped to Todoist labels
- **Git**: Inline hashtags in markdown (`#tag`)

### Tag Normalization
- Trim whitespace from each tag
- Remove duplicates (case-insensitive)
- Preserve original case for display
- Empty tags are ignored

### Tags Command Output
```
Tags in use:
  work (15 tasks)
  urgent (8 tasks)
  home (5 tasks)
  project-x (3 tasks)
```

## Out of Scope
- Tag colors or icons
- Hierarchical tags (tag:subtag)
- Tag aliases or synonyms
- Tag auto-complete in TUI
