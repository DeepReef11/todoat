# auto_detect_backend Config Format Inconsistency

## Summary
Documentation shows conflicting formats for `auto_detect_backend` configuration.

## Config Formats Found

### Format 1: Simple boolean (config.sample.yaml and docs/backends.md)
```yaml
auto_detect_backend: false
```

### Format 2: Nested object (mentioned in user issue comment)
```yaml
auto_detect_backend:
  enabled: true
  backend_priority:
    - git
    - nextcloud-prod
    - sqlite
```

### Format 3: Boolean with separate backend_priority (docs/backends.md:361-365)
```yaml
auto_detect_backend: true

backend_priority:
  - git
  - nextcloud
```

## Questions
1. Which format is correct/supported?
2. If multiple formats are supported, are they all documented?
3. Is `backend_priority` inside or outside `auto_detect_backend`?

## Reference
- `internal/config/config.sample.yaml` - uses Format 1
- `docs/backends.md` - uses Format 1 and Format 3
- User reported seeing Format 2 somewhere in documentation
- Example config: `issues/examples/user-config-multi-backend.yaml`

## Impact
Users may use incorrect config format, leading to unexpected behavior or silent config errors.

## Resolution

**Fixed in**: this session
**Fix description**: Clarified documentation to match actual implementation

### Answers to Questions

1. **Which format is correct?** Format 1 (simple boolean). The code only supports `auto_detect_backend: bool`.
2. **Are multiple formats supported?** No, only simple boolean is supported.
3. **Is `backend_priority` inside or outside?** It's a separate top-level field, but it's **not implemented** - just documented as a planned feature.

### Changes Made

1. **docs/backends.md**:
   - Removed `backend_priority` from Selection Priority list (was item #4)
   - Removed dedicated "Backend Priority" section
   - Removed `backend_priority` from example configurations

2. **internal/config/config.sample.yaml**:
   - Updated comment to clarify `backend_priority` is a "planned feature - not yet implemented"

### Verification Log
```bash
$ grep -n "auto_detect_backend\|backend_priority" docs/backends.md internal/config/config.sample.yaml
docs/backends.md:207:auto_detect_backend: true
docs/backends.md:346:auto_detect_backend: true
internal/config/config.sample.yaml:49:auto_detect_backend: false
internal/config/config.sample.yaml:52:# backend_priority:
```

Documentation now shows only the simple boolean format for `auto_detect_backend`, and `backend_priority` is only mentioned in config.sample.yaml as a commented-out planned feature.

**Matches expected behavior**: YES
