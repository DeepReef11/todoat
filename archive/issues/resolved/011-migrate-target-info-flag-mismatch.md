# [011] Example Mismatch: migrate --target-info flag requires argument

## Type
doc-mismatch

## Category
user-journey

## Severity
high

## Location
- File: `docs/backends.md`
- Line: 405
- Context: README documentation

## Documented Command
```bash
todoat migrate --to nextcloud --target-info
```

## Actual Result
```bash
$ todoat migrate --to nextcloud --target-info
Error: flag needs an argument: --target-info
```

## Working Alternative (if known)
The `--target-info` flag requires a string argument. The intended usage is unclear from the current implementation:
```bash
todoat migrate --target-info <something>
```

## Recommended Fix
FIX EXAMPLE or FIX CODE - Either:
1. Update the documentation to show the correct flag usage with its required argument
2. Or change the code to make `--target-info` a boolean flag as the documentation implies

## Impact
Users following this example will see: "Error: flag needs an argument: --target-info"

## Resolution

**Fixed in**: this session
**Fix description**: Updated docs/backends.md to show correct `--target-info` usage with backend argument
**Test added**: N/A (doc fix only - existing tests already use correct syntax)

### Verification Log
```bash
$ todoat migrate --help | grep target-info
      --target-info string   Show tasks in target backend
```
Documentation now correctly shows:
```bash
todoat migrate --target-info nextcloud --list Work
```
**Matches expected behavior**: YES
