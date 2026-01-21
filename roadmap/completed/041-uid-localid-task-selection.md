# [041] UID/Local-ID Task Selection

## Summary
Implement `--uid` and `--local-id` flags for direct task selection, bypassing summary-based search for unambiguous task operations in scripts and automation workflows.

## Documentation Reference
- Primary: `docs/explanation/cli-interface.md` (Task Selection Flags, No-Prompt Mode sections)
- Related: `docs/explanation/task-management.md` (Task Identifiers)

## Dependencies
- Requires: [004] Task Commands (update, complete, delete operations)
- Requires: [018] Synchronization Core (for UID assignment and local_id)

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestUpdateByUID` - `todoat MyList update --uid "550e8400..." -s DONE` updates task by UID
- [ ] `TestCompleteByUID` - `todoat MyList complete --uid "550e8400..."` completes task by UID
- [ ] `TestDeleteByUID` - `todoat MyList delete --uid "550e8400..."` deletes task by UID
- [ ] `TestUpdateByLocalID` - `todoat MyList update --local-id 42 -s DONE` updates task by SQLite ID
- [ ] `TestCompleteByLocalID` - `todoat MyList complete --local-id 42` completes task by ID
- [ ] `TestDeleteByLocalID` - `todoat MyList delete --local-id 42` deletes task by ID
- [ ] `TestUIDNotFound` - `todoat MyList update --uid "nonexistent" -s DONE` returns ERROR
- [ ] `TestLocalIDNotFound` - `todoat MyList update --local-id 99999 -s DONE` returns ERROR
- [ ] `TestLocalIDRequiresSync` - `--local-id` returns error when sync not enabled
- [ ] `TestUIDRequiresSyncedTask` - `--uid` only works for tasks with backend-assigned UID

### Functional Requirements
- [ ] `--uid` flag accepts backend-assigned UUID string
- [ ] `--local-id` flag accepts integer (SQLite internal ID)
- [ ] UID lookup bypasses summary search entirely
- [ ] Local-ID lookup requires sync to be enabled
- [ ] Flags are mutually exclusive with positional task summary argument
- [ ] Clear error messages when task not found by ID

### JSON Output Enhancement
- [ ] Multiple matches response includes `local_id` and `uid` fields:
  ```json
  {
    "matches": [
      {"local_id": 42, "uid": "550e8400-...", "summary": "Task A", "synced": true},
      {"local_id": 43, "uid": null, "summary": "Task B", "synced": false}
    ],
    "result": "ACTION_INCOMPLETE",
    "message": "Multiple tasks match. Use --uid or --local-id to specify exact task."
  }
  ```
- [ ] Unsynced tasks show `uid: null` and `synced: false`
- [ ] Synced tasks show actual UID and `synced: true`

### Workflow Support
- [ ] Script workflow: search → receive IDs → operate with specific ID
- [ ] `--uid` preferred for cross-session operations (stable identifier)
- [ ] `--local-id` works for all tasks including unsynced

## Implementation Notes
- Add `--uid` and `--local-id` flags to update, complete, delete commands
- Lookup by UID: query `tasks` table where `uid = ?`
- Lookup by local-id: query `tasks` table where `id = ?`
- Return descriptive error if no match found
- Validate sync enabled before accepting `--local-id`

## Out of Scope
- UID generation for unsynced tasks (handled by sync system)
- Batch operations with multiple UIDs
- UID-based task references in subtask creation (`-P --uid`)
