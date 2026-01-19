# [052] List Color and Description Update

## Summary
Implement CLI commands to update list properties including color and description, supporting visual customization of task lists.

## Documentation Reference
- Primary: `dev-doc/LIST_MANAGEMENT.md`
- Section: List Properties

## Dependencies
- Requires: [007] List Management Commands
- Requires: [013] List Management

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestListUpdateColor` - `todoat list update "Work" --color "#FF5733"` sets list color
- [ ] `TestListUpdateDescription` - `todoat list update "Work" --description "Work tasks"` sets description
- [ ] `TestListUpdateMultiple` - `todoat list update "Work" --color "#FF5733" --description "Text"` updates both
- [ ] `TestListUpdateColorValidation` - Invalid hex color returns ERROR with format hint
- [ ] `TestListShowProperties` - `todoat list show "Work"` displays all properties
- [ ] `TestListUpdateJSON` - `todoat --json list update "Work" --color "#FF5733"` returns JSON

### Functional Requirements
- [ ] `list update <name> --color <hex>` updates list color
- [ ] `list update <name> --description <text>` updates list description
- [ ] `list show <name>` displays all list properties (id, name, color, description, task count)
- [ ] Color validation accepts standard hex formats (#RGB, #RRGGBB)
- [ ] Empty string for description clears the field
- [ ] Changes sync to remote backend when sync enabled

## Implementation Notes

### Color Validation
- Accept formats: `#RGB`, `#RRGGBB`, `RGB`, `RRGGBB`
- Normalize to `#RRGGBB` for storage
- Case-insensitive hex digits

### Backend-Specific Behavior
- **SQLite**: UPDATE on task_lists table
- **Nextcloud**: PROPPATCH for `calendar:calendar-color` and `calendar:calendar-description`
- **Todoist**: API call (color mapped to Todoist color IDs)

### List Show Output Format
```
List: Work Tasks
  ID: abc-123-def
  Color: #FF5733
  Description: Work-related tasks
  Tasks: 42
  Created: 2026-01-15
```

## Out of Scope
- Color picker UI
- Named colors (e.g., "red", "blue")
- List icons or emoji
