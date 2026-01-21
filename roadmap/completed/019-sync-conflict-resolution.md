# [019] Sync Conflict Resolution

## Summary
Implement conflict detection and resolution strategies for synchronization, including server-wins, local-wins, merge, and keep-both strategies with configurable defaults.

## Documentation Reference
- Primary: `docs/explanation/synchronization.md#conflict-resolution`
- Related: `docs/explanation/configuration.md`

## Dependencies
- Requires: [018] Synchronization Core System

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestConflictDetection` - Sync detects when local and remote have both changed same task
- [ ] `TestConflictServerWins` - With `conflict_strategy: server-wins`, remote changes override local
- [ ] `TestConflictLocalWins` - With `conflict_strategy: local-wins`, local changes override remote
- [ ] `TestConflictMerge` - With `conflict_strategy: merge`, non-conflicting fields are combined
- [ ] `TestConflictKeepBoth` - With `conflict_strategy: keep-both`, duplicate task created
- [ ] `TestConflictStatusDisplay` - `todoat sync status` shows count of conflicts needing attention
- [ ] `TestConflictList` - `todoat sync conflicts` lists all unresolved conflicts with details
- [ ] `TestConflictResolve` - `todoat sync conflicts resolve [task-uid] --strategy server-wins` resolves specific conflict
- [ ] `TestConflictDefaultStrategy` - Default strategy is configurable in config.yaml
- [ ] `TestConflictJSONOutput` - `todoat --json sync conflicts` returns conflicts in JSON format

## Implementation Notes
- Compare modified timestamps to detect conflicts
- Merge strategy should handle field-level merging (e.g., different fields changed on each side)
- Keep-both creates new task with suffix (e.g., "Task Name (conflict copy)")
- Store conflict metadata in database for later resolution

## Out of Scope
- Interactive conflict resolution wizard (TUI)
- Three-way merge with common ancestor
- Automatic conflict resolution based on ML/heuristics
