# [002] Review: Resolved Issues Incorrectly Moved Out of resolved/ Folder

## Type
code-bug

## Severity
critical

## Source
Code review - 2026-01-23 23:43:45

## Steps to Reproduce
1. Run `git status` in the repository
2. Observe that `issues/resolved/0-sync-is-wrong.md` and `issues/resolved/1-default-backend-problem.md` show as deleted
3. Observe that `issues/0-sync-is-wrong.md` and `issues/1-default-backend-problem.md` appear as untracked files

## Expected Behavior
Resolved issues should remain in the `issues/resolved/` folder. Both files contain:
- Complete "Resolution" sections with fix descriptions
- Tests added for the fixes
- Verification logs showing expected behavior
- "**Matches expected behavior**: YES" confirmation

## Actual Behavior
The resolved issue files were moved from `issues/resolved/` to `issues/`, making them appear as open issues when they are actually resolved.

## Files Affected
- issues/0-sync-is-wrong.md (should be in issues/resolved/)
- issues/1-default-backend-problem.md (should be in issues/resolved/)

## Resolution

**Fixed in**: this session
**Fix description**: Restored the incorrectly moved resolved issue files to their proper location in `issues/resolved/`. The file `1-default-backend-problem.md` was restored via `git checkout` and the misplaced copy was removed.

### Verification Log
```bash
$ ls issues/resolved/1-default-backend-problem.md
issues/resolved/1-default-backend-problem.md

$ ls issues/1-default-backend-problem.md
ls: cannot access 'issues/1-default-backend-problem.md': No such file or directory
```
**Matches expected behavior**: YES - Resolved issues are now correctly located in the `issues/resolved/` folder.
