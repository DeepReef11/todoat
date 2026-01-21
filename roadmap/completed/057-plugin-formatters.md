# [057] Plugin Formatters with External Scripts

## Summary
Implement plugin-based field formatting allowing external scripts (Bash, Python, Ruby, etc.) to transform task field values for custom display in views.

## Documentation Reference
- Primary: `docs/explanation/views-customization.md` (Plugin Formatters section)
- Related: `docs/explanation/cli-interface.md`

## Dependencies
- Requires: [015] Views Customization (view YAML structure must exist)

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestPluginFormatterStatus` - View with status plugin displays custom formatted status (e.g., emoji)
- [ ] `TestPluginFormatterPriority` - View with priority plugin displays custom priority (e.g., colored)
- [ ] `TestPluginFormatterDate` - View with date plugin displays relative dates ("2 days ago")
- [ ] `TestPluginTimeout` - Plugin exceeding timeout (default 1s) falls back to raw value
- [ ] `TestPluginError` - Plugin returning non-zero exit shows fallback value gracefully
- [ ] `TestPluginInvalidOutput` - Plugin returning invalid JSON falls back to raw value
- [ ] `TestPluginNotFound` - Non-existent plugin path shows warning and uses raw value

### Functional Requirements
- [ ] Plugins receive task data as JSON on stdin
- [ ] Plugin JSON input format includes: uid, summary, status, priority, due_date, etc.
- [ ] Plugins output formatted string on stdout (single line)
- [ ] Plugin timeout configurable per-field (default: 1000ms)
- [ ] Plugin environment variables configurable in view YAML
- [ ] Example plugins work: bash (status emoji), python (priority color), ruby (relative date)
- [ ] View YAML supports `plugin` field in field configuration:
  ```yaml
  fields:
    - name: status
      plugin:
        command: ~/.config/todoat/plugins/status-emoji.sh
        timeout: 500
  ```

### Output Requirements
- [ ] Plugin output replaces field value in display
- [ ] Failed plugins show raw value (no errors in output unless verbose)
- [ ] Performance stays acceptable (<100ms overhead per task row with plugins)

## Implementation Notes

### Plugin Interface
```
stdin: {"uid": "...", "summary": "...", "status": "TODO", "priority": 5, ...}
stdout: ✅ (single line, formatted value)
exit 0: success
exit non-zero: error, use fallback
```

### Example Plugin (bash)
```bash
#!/bin/bash
read -r task
status=$(echo "$task" | jq -r '.status')
case "$status" in
  "TODO") echo "📋";;
  "DONE") echo "✅";;
  "PROCESSING") echo "🔄";;
  "CANCELLED") echo "❌";;
  *) echo "$status";;
esac
```

### View YAML with Plugin
```yaml
name: emoji-status
fields:
  - name: status
    width: 3
    plugin:
      command: ~/.config/todoat/plugins/status-emoji.sh
      timeout: 500
      env:
        THEME: dark
  - name: summary
  - name: priority
```

## Out of Scope
- Plugin discovery/registration system
- Plugin marketplace or sharing
- Async/parallel plugin execution
- Plugin caching between rows (each row calls plugin fresh)
