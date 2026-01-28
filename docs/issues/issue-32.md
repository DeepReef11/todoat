# Issue #32: Auto-sync not working after operations

**GitHub Issue**: #32
**Status**: Open
**Created**: 2026-01-27
**Labels**: None

## Description

Auto-sync doesn't work at all. Either it's completely broken or takes way too long to actually sync in the background.

## Steps to Reproduce

1. Configure a backend with sync enabled
2. Perform an operation that should trigger auto-sync (e.g., adding a task)
3. Observe that the background sync does not occur

## Expected Behavior

After operations that modify tasks, the background auto-sync should:
- Trigger automatically
- Complete within a reasonable timeframe
- Sync changes to the configured backend

## Actual Behavior

Auto-sync either:
- Does not trigger at all, OR
- Takes so long that it appears broken

## Workaround

Manual sync (`todoat sync`) works correctly.

## Context

This may be related to how background goroutines are managed or timing issues with the sync trigger. The manual sync functionality works, indicating the core sync logic is correct but the automatic triggering mechanism is broken.

## Acceptance Criteria

- [ ] Auto-sync triggers automatically after task operations
- [ ] Auto-sync completes within a reasonable timeframe (seconds, not minutes)
- [ ] Background sync errors are properly logged/reported
- [ ] Add tests to verify auto-sync behavior
