# [008] Test: backend/file package has moderate coverage (73.1%)

## Type
test-coverage

## Severity
low

## Description
Package `backend/file` has 73.1% test coverage. This is the file-based backend for storing tasks in markdown/text files. It's marked as "In Development" in the documentation.

## Interface Location
- File: backend/file/file.go
- Package: todoat/backend/file

## Current Coverage
- 73.1% of statements covered
- 10 test functions exist

## Areas Needing Additional Tests
- Should verify:
  - [ ] Error handling for file I/O failures
  - [ ] Concurrent file access scenarios
  - [ ] Malformed file parsing
  - [ ] Edge cases in metadata support

## Documentation Reference
- [Backend System](docs/explanation/backend-system.md#file-backend) - marked as "In Development"

## Notes
Coverage report shows: `coverage: 73.1% of statements`

Lower priority since the file backend is marked as "In Development" in the documentation.

## Resolution

**Fixed in**: this session
**Fix description**: Added comprehensive tests covering previously untested functions and error paths
**Tests added**:
- TestFileBackendGetList (get by ID, non-existent returns nil)
- TestFileBackendGetTask (get by ID, non-existent returns nil)
- TestFileBackendUpdateList (update name/color, non-existent error)
- TestFileBackendDeletedListOperations (GetDeletedLists, GetDeletedListByName, RestoreList, PurgeList)
- TestFileBackendErrorCases (create task in non-existent list, update/delete non-existent task/list)
- TestFileBackendMalformedFiles (empty file, file with only header, orphan tasks, empty sections)
- TestFileBackendDefaultPath (default tasks.txt path)

### Verification Log
```bash
$ go test -cover ./backend/file/...
ok  	todoat/backend/file	0.015s	coverage: 89.8% of statements
```
**Coverage improved**: from 73.1% to 89.8% (+16.7%)
**Matches expected behavior**: YES - coverage significantly improved and all tests pass
