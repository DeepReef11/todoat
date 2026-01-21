# [042] Task Description Flag

## Summary
Implement `-d, --description` flag for add and update commands to set task descriptions, supporting multi-line text input.

## Documentation Reference
- Primary: `docs/explanation/cli-interface.md`
- Section: Action Flags - Description

## Dependencies
- Requires: [004] Task Commands

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestAddTaskWithDescription` - `todoat -y MyList add "Task" -d "Detailed notes"` creates task with description
- [ ] `TestUpdateTaskDescription` - `todoat -y MyList update "Task" -d "Updated notes"` updates description
- [ ] `TestClearTaskDescription` - `todoat -y MyList update "Task" -d ""` clears description
- [ ] `TestDescriptionInJSON` - `todoat -y --json MyList` includes description field in output
- [ ] `TestDescriptionLongFlag` - `todoat -y MyList add "Task" --description "Notes"` works with long flag

### Functional Requirements
- Description field stored in task database
- Empty string clears description
- Multi-line text supported via quoted strings
- Description displayed in task detail views

## Implementation Notes
- Add `-d, --description` flag to add command
- Add `-d, --description` flag to update command
- Task struct already has Description field from backend interface
- Ensure SQLite backend stores/retrieves description

## Out of Scope
- Rich text/markdown formatting in descriptions
- Description search/filtering
