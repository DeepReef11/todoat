# [064] Fix documentation to match implemented CLI commands

## Summary
Several documented commands and flags don't match the actual CLI implementation. The documentation references commands/flags that don't exist or work differently.

## Documentation Reference
- Primary: `docs/list-management.md`, `docs/tags.md`, `docs/backends.md`

## Gap Type
broken (documentation mismatch)

## Current Behavior
Documentation shows commands that don't exist or work differently than implemented.

## Gaps to Fix

### 1. `list rename` vs `list update --name`
**Documented (incorrect):**
```bash
todoat list rename "Old Name" "New Name"
```

**Actual command:**
```bash
todoat list update "Old Name" --name "New Name"
```

**Files:** `docs/list-management.md`

### 2. `list show` vs `list info`
**Documented (incorrect):**
```bash
todoat list show "Work Tasks"
```

**Actual command:**
```bash
todoat list info "Work Tasks"
```

**Files:** `docs/list-management.md`

### 3. `--filter "tags:X"` vs `--tag X`
**Documented (incorrect):**
```bash
todoat MyList --filter "tags:urgent"
todoat Work -v all --filter "tags:project-x"
```

**Actual command:**
```bash
todoat MyList --tag urgent
todoat Work -v all --tag project-x
```

**Files:** `docs/tags.md`

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] Run documented examples from fixed docs and verify they work

### Functional Requirements
- [ ] Replace `list rename` examples with `list update --name`
- [ ] Replace `list show` examples with `list info`
- [ ] Replace `--filter "tags:X"` examples with `--tag X`
- [ ] Update any cross-references affected by these changes

## Implementation Notes
These are pure documentation fixes. No code changes required. The CLI already works correctly; the documentation just shows the wrong syntax.

Note: Consider also adding documentation about the commands that DO exist but aren't in the current documentation structure (e.g., `list export`, `list import`, `list vacuum`, `list stats`).
