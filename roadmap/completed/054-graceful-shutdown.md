# [054] Graceful Shutdown Handling

## Summary
Implement proper signal handling for graceful shutdown, ensuring database transactions are completed and resources are properly cleaned up when the application receives termination signals.

## Documentation Reference
- Primary: `docs/explanation/cli-interface.md`
- Section: Graceful Shutdown

## Dependencies
- Requires: [002] Core CLI with Cobra
- Requires: [003] SQLite Backend

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestGracefulShutdownSIGINT` - Application handles SIGINT cleanly without data loss
- [ ] `TestGracefulShutdownSIGTERM` - Application handles SIGTERM cleanly
- [ ] `TestShutdownDuringSync` - Sync in progress completes or rolls back safely
- [ ] `TestShutdownDuringWrite` - Database write completes or rolls back safely
- [ ] `TestShutdownExitCode` - Clean shutdown returns exit code 0

### Functional Requirements
- [ ] Register signal handlers for SIGINT and SIGTERM
- [ ] On signal: set shutdown flag to prevent new operations
- [ ] Complete or rollback any in-progress database transactions
- [ ] Close backend connections gracefully
- [ ] Stop background sync processes if running
- [ ] Flush any buffered log output
- [ ] Exit with code 0 on clean shutdown
- [ ] Exit with code 130 (128+2) if interrupted during blocking operation

## Implementation Notes

### Signal Handling Pattern
```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-sigChan
    // Set shutdown flag
    // Wait for active operations (with timeout)
    // Cleanup resources
    os.Exit(0)
}()
```

### Shutdown Sequence
1. Set global shutdown flag
2. Wait up to 5 seconds for active operations
3. Close database connections (triggers transaction rollback)
4. Stop auto-sync daemon if running
5. Close log file handles
6. Exit

### Context Cancellation
- Pass `context.Context` to long-running operations
- Cancel context on shutdown signal
- Operations check context for cancellation

## Out of Scope
- Windows-specific signal handling (Ctrl+C works via SIGINT)
- Recovery from partial operations after forced kill
- Checkpoint/resume for interrupted syncs
