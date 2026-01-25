# [007] Doc Mismatch: --config flag Not Implemented

## Type
doc-mismatch

## Severity
critical

## Source
Code review - 2026-01-25

## Description
Documentation in `docs/explanation/configuration.md` describes the `--config` flag feature extensively, but this flag was never implemented. The code does not support `--config` flag.

Documentation claims:
- `--config` flag allows specifying custom configuration file location
- Multiple usage patterns documented (file path, directory path, relative paths)
- Integrated into "Config Precedence" table as priority 1

## Documentation Claims
From `docs/explanation/configuration.md`:

```markdown
| **1. Custom Path** | `--config` flag | Explicit path specified via command-line |

### Custom Config Path

**Purpose**: Allows users to specify a custom configuration file location using the `--config` flag...

**Usage Patterns**:
todoat --config /path/to/config.yaml MyList
todoat --config /path/to/config-dir/ MyList
todoat --config . MyList
```

## Actual Code
```bash
$ todoat --config /path/to/config.yaml list
Error: unknown flag: --config
```

The flag is not defined in `cmd/todoat/cmd/todoat.go`. A previous fix (roadmap/completed/008-dev-doc-missing-flags.md) removed references from `cli-interface.md` but `configuration.md` was missed.

## Impact
- Users following documentation cannot use the documented `--config` flag
- No workaround exists - must use XDG_CONFIG_HOME or default location
- Documentation provides false confidence in feature availability

## Steps to Reproduce
1. Run `todoat --config /tmp/config.yaml list`
2. Observe error: "unknown flag: --config"

## Expected Behavior
Either:
- Flag should work as documented, OR
- Documentation should remove all `--config` references

## Actual Behavior
Flag does not exist. Error: "unknown flag: --config"

## Files Affected
- docs/explanation/configuration.md (lines 55, 619-711, 1758)

## Recommended Fix
FIX DOCS - Remove the `--config` flag references from configuration.md:
1. Remove "Custom Path" from the Config Precedence table (line 55)
2. Remove or mark as "Planned" the entire "Custom Config Path" section (lines 619-711)
3. Update features-overview.md if it references this feature

## Related
- roadmap/081-fix-config-flag-docs-incomplete-removal.md (untracked file documenting this issue)
- roadmap/completed/008-dev-doc-missing-flags.md (partial fix that missed configuration.md)
