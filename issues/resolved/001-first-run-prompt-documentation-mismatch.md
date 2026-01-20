# [001] First run config prompt documentation mismatch

## Type
doc-mismatch

## Category
user-journey

## Severity
low

## Steps to Reproduce
```bash
# Remove existing config
rm -rf ~/.config/todoat

# Run todoat for the first time
./todoat
```

## Expected Behavior
According to `docs/getting-started.md` (lines 24-30), the first run should show:

```
$ todoat
No configuration file found.
Do you want to copy config sample to ~/.config/todoat/config.yaml? (y/n)
```

## Actual Behavior
The config file is silently created without any prompt:

```
$ ./todoat
Available lists (N):
...
```

The config file is created at `~/.config/todoat/config.yaml` automatically without asking the user.

## Error Output
N/A - no error, just different behavior than documented

## Environment
- OS: Linux
- Runtime version: Go 1.25

## Possible Cause
The documentation describes a prompt-based flow that was either:
1. Never implemented
2. Removed when issue 001-config-file-not-created.md was resolved

The resolution of that issue added automatic config creation in `getBackend()`, which silently creates the config without prompting.

## Documentation Reference
- File: `docs/getting-started.md`
- Section: First Run
- Documented behavior: Interactive prompt asking to create config
- Actual behavior: Silent automatic config creation

## Related Files
- docs/getting-started.md:24-30
- cmd/todoat/cmd/todoat.go (getBackend function creates config silently)

## Recommended Fix
FIX DOCS - Update getting-started.md to remove the interactive prompt example and instead describe the actual behavior: config is automatically created with defaults on first run.

Suggested replacement text:

```markdown
## First Run

When you first run todoat, a configuration file is automatically created at `~/.config/todoat/config.yaml` with sensible defaults.

\`\`\`bash
$ todoat
Available lists (0):
(no lists yet)
\`\`\`
```

## Resolution

**Fixed in**: Already resolved (prior to this session)
**Fix description**: The documentation at `docs/getting-started.md` lines 24-31 was updated to correctly describe the actual behavior - automatic config creation without prompts.

### Verification Log
```bash
$ rm -rf ~/.config/todoat && ./todoat
Available lists (8):

NAME                 TASKS
Tasks                1
...
```
The config file was silently created at `~/.config/todoat/config.yaml` without any prompt.

The documentation now correctly states (lines 24-25):
> When you first run todoat, if no configuration file exists, it automatically creates one at `~/.config/todoat/config.yaml` with a sample configuration including documentation and examples.

**Matches expected behavior**: YES - Documentation now matches actual behavior.
