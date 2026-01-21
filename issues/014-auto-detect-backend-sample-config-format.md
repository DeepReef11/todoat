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
