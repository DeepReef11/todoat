# [002] Duplicate Error Messages

## Type
code-bug

## Category
error-handling

## Severity
low

## Steps to Reproduce
```bash
todoat list info "NonExistentList"
```

## Expected Behavior
A single error message should be displayed.

## Actual Behavior
The error message is displayed twice:

```
$ todoat list info "NonExistentList"
Error: list 'NonExistentList' not found
Error: list 'NonExistentList' not found
```

## Error Output
```
Error: list 'NonExistentList' not found
Error: list 'NonExistentList' not found
```

## Environment
- OS: Linux
- Runtime version: Go 1.21+

## Possible Cause
The error is likely being printed by the command handler and then again by the Cobra error handling mechanism. The error may be returned from a function and then also printed directly.

## Related Files
- cmd/todoat/list_commands.go (likely location)

## Recommended Fix
FIX CODE - Ensure that errors are either printed directly OR returned to Cobra for handling, not both. Typically, returning the error to Cobra is preferred as it provides consistent error formatting across all commands.
