# [033] Default View Configuration

## Summary
Add configuration option to set a default custom view, so users don't need to specify `-v` flag repeatedly for their preferred view.

## Documentation Reference
- Primary: `dev-doc/CONFIGURATION.md` (View Defaults section)
- Secondary: `dev-doc/VIEWS_CUSTOMIZATION.md`

## Dependencies
- Requires: [015] Views Customization
- Requires: [010] Configuration

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestDefaultViewConfig` - `default_view: myview` in config is respected
- [ ] `TestDefaultViewFallback` - Falls back to "default" if configured view not found
- [ ] `TestDefaultViewOverride` - `-v` flag overrides config default
- [ ] `TestDefaultViewBuiltin` - Can set built-in views as default ("all", "default")
- [ ] `TestDefaultViewCustom` - Can set custom view from `~/.config/todoat/views/` as default
- [ ] `TestDefaultViewMissing` - Warning shown if configured default view doesn't exist

## Implementation Notes

### Configuration
```yaml
# Set default view (optional)
default_view: myview  # Custom view name

# Or use built-in view
default_view: all     # Show all fields by default
```

### View Resolution Order
1. Command-line flag: `-v custom-view` (highest priority)
2. Config default: `default_view: myview`
3. Built-in default: "default" view (fallback)

### Files to Modify
1. `internal/config/config.go` - Add `DefaultView` field
2. `cmd/todoat/root.go` - Apply default view if no `-v` flag
3. `internal/views/loader.go` - Validate default view exists

### Validation
- Check if view file exists in `~/.config/todoat/views/`
- Check if view is a built-in ("default", "all")
- Log warning if view not found but continue with fallback

## Out of Scope
- Per-list default views
- Per-backend default views
- Default view creation wizard
