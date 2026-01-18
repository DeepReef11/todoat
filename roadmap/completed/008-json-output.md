# 008: JSON Output Mode

## Summary
Implement `--json` flag for machine-parseable JSON output, enabling scripting and automation integration.

## Documentation Reference
- Primary: `dev-doc/CLI_INTERFACE.md`
- Sections: JSON Output Mode

## Dependencies
- Requires: 002-core-cli.md (CLI framework with global flags)
- Requires: 004-task-commands.md (commands to output)
- Requires: 006-cli-tests.md (result codes to include)
- Blocked by: none

## Complexity
**S (Small)** - Flag handling and JSON marshalling for existing data structures

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestJSONFlagParsing` - `--json` flag is recognized and sets output mode
- [ ] `TestListTasksJSON` - `todoat -y --json MyList` returns valid JSON with tasks array
- [ ] `TestAddTaskJSON` - `todoat -y --json MyList add "Task"` returns JSON with task info and result
- [ ] `TestUpdateTaskJSON` - `todoat -y --json MyList update "Task" -s DONE` returns JSON with updated task
- [ ] `TestDeleteTaskJSON` - `todoat -y --json MyList delete "Task"` returns JSON with result
- [ ] `TestErrorJSON` - `todoat -y --json NonExistent` returns JSON error with result: "ERROR"
- [ ] `TestJSONResultCodes` - All JSON responses include "result" field (ACTION_COMPLETED, INFO_ONLY, ERROR)

### Unit Tests (if needed)
- [ ] JSON marshalling of Task struct produces expected fields
- [ ] JSON marshalling of TaskList struct produces expected fields

### Manual Verification
- [ ] JSON output can be piped to `jq` for processing
- [ ] All commands produce valid JSON when `--json` flag is used

## Implementation Notes

### JSON Response Structures

**List Tasks Response:**
```json
{
  "tasks": [
    {
      "uid": "550e8400-...",
      "summary": "Task name",
      "status": "TODO",
      "priority": 1
    }
  ],
  "list": "MyList",
  "count": 1,
  "result": "INFO_ONLY"
}
```

**Action Response:**
```json
{
  "action": "add",
  "task": {
    "uid": "550e8400-...",
    "summary": "New task"
  },
  "result": "ACTION_COMPLETED"
}
```

**Error Response:**
```json
{
  "error": "List 'NonExistent' not found",
  "code": 1,
  "result": "ERROR"
}
```

### Required Changes
1. Add `--json` persistent flag to root command
2. Create JSON response structs in `internal/cli/output/`
3. Modify output functions to check JSON mode and format accordingly
4. Ensure all paths through code respect JSON output mode

### Key Patterns
- Check `cmd.Flags().GetBool("json")` before output
- Use `json.MarshalIndent()` for readable output
- Include `result` field in all responses for consistency

## Out of Scope
- JSON input (reading tasks from JSON) - not planned
- YAML output - not planned
- CSV output - not planned
