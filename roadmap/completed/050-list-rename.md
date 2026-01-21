# [050] List Rename Command

## Summary
Implement the ability to rename task lists via CLI command, updating both local storage and remote backends as needed.

## Documentation Reference
- Primary: `docs/explanation/list-management.md`
- Section: List Properties

## Dependencies
- Requires: [007] List Management Commands
- Requires: [013] List Management

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestListRename` - `todoat list update "OldName" --name "NewName"` renames list
- [ ] `TestListRenameNotFound` - Renaming non-existent list returns ERROR
- [ ] `TestListRenameDuplicate` - Renaming to existing name returns ERROR with suggestion
- [ ] `TestListRenameJSON` - `todoat --json list update "OldName" --name "NewName"` returns JSON with result
- [ ] `TestListRenameNoPrompt` - `todoat -y list update "Partial" --name "NewName"` handles partial match

### Functional Requirements
- [ ] `list update <name> --name <new-name>` renames list
- [ ] Validates new name is non-empty and unique
- [ ] Updates local SQLite and remote backend (if sync enabled)
- [ ] Preserves all other list properties (color, description)
- [ ] Works with partial name matching (with confirmation)

## Implementation Notes

### Backend-Specific Behavior
- **SQLite**: UPDATE query on `task_lists` table
- **Nextcloud**: PROPPATCH request to update `calendar:displayname` property
- **Todoist**: API call to update project name
- **Git**: Rename markdown heading for list section

### Validation Rules
- New name must be non-empty
- New name must be unique within backend
- New name length should not exceed backend limits

## Out of Scope
- Renaming multiple lists at once
- List name history/audit trail
