# auto_detect_backend Sample Config Missing Full Format

## Summary
The config sample shows `auto_detect_backend: false` but doesn't document the full nested format with `enabled` and other options.

## Current Sample Format
```yaml
auto_detect_backend: false
```

## Expected Sample Format
Should show the full structure:
```yaml
auto_detect_backend:
  enabled: false
  # backend_priority:
  #   - git
  #   - nextcloud
  #   - sqlite
```

## Impact
Users don't know about the `enabled` sub-option and other available settings without digging into documentation.

## Related
- Issue #005: auto_detect_backend config format inconsistency (if still open)
- `internal/config/config.sample.yaml`

## Suggested Fix
Update `config.sample.yaml` to show the full nested structure, even if commented out.

## Resolution

**Status**: Closed - Won't Fix
**Reason**: The requested nested format does not match the actual implementation

### Analysis

The issue requests showing a nested format:
```yaml
auto_detect_backend:
  enabled: false
  # backend_priority: ...
```

However, examining the code in `internal/config/config.go:31`:
```go
AutoDetectBackend bool `yaml:"auto_detect_backend"`
```

The `auto_detect_backend` field is implemented as a simple boolean, not a nested struct. The current sample config (`auto_detect_backend: false`) is **correct** because it matches the actual code implementation.

This was explicitly clarified in the resolution of issue #005 (auto_detect_backend config format inconsistency):
- "Which format is correct? Format 1 (simple boolean). The code only supports `auto_detect_backend: bool`."
- "backend_priority" is a separate planned feature that's not yet implemented

The sample config already correctly shows:
1. `auto_detect_backend: false` (boolean) - line 49
2. `backend_priority` as a commented "planned feature - not yet implemented" - lines 51-54

**No changes needed** - the sample config accurately reflects the implementation.

**Matches expected behavior**: YES (sample config already correct)
