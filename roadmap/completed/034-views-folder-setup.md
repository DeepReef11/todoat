# 034 - Views Folder Setup with User-Overridable Built-ins

## Summary

When the `~/.config/todoat/views/` folder does not exist, prompt the user to create it with default and all view YAML files. This allows users to customize built-in views by editing these files. Change view loading order to check user views first, then fall back to built-in.

## Dependencies

None

## Acceptance Criteria

### CLI Tests Required

1. **First run without views folder prompts user**
   ```bash
   # When views/ folder doesn't exist and -y not used
   rm -rf ~/.config/todoat/views
   todoat MyList
   # Should prompt: "Views folder not found. Create with default views? [Y/n]"
   # On 'y' or Enter: creates views/ with default.yaml and all.yaml
   ```

2. **With -y flag, auto-creates views folder**
   ```bash
   rm -rf ~/.config/todoat/views
   todoat -y MyList
   # Should silently create views/ folder with default.yaml and all.yaml
   # No prompt shown
   ```

3. **Empty views folder does NOT trigger prompt**
   ```bash
   rm -rf ~/.config/todoat/views
   mkdir ~/.config/todoat/views
   todoat MyList
   # Should NOT prompt - folder exists (even if empty)
   # Uses built-in views as fallback
   ```

4. **User view overrides built-in**
   ```bash
   # Create custom default.yaml that shows DONE tasks
   cat > ~/.config/todoat/views/default.yaml << 'EOF'
   name: default
   description: My custom default view (shows all tasks)
   fields:
     - name: status
       width: 12
     - name: summary
       width: 50
   # No filter - shows DONE tasks too
   EOF

   todoat MyList
   # Should use user's default.yaml, not built-in
   # DONE tasks should be visible
   ```

5. **Fallback to built-in when user view missing**
   ```bash
   mkdir -p ~/.config/todoat/views
   rm -f ~/.config/todoat/views/default.yaml
   todoat MyList
   # Should fall back to built-in default view
   ```

6. **View list shows override status**
   ```bash
   todoat view list
   # Output should indicate which views are overridden:
   # - default (user-defined, overrides built-in)
   # - all (built-in)
   # - myview (user-defined)
   ```

## Implementation Notes

### View Loading Order Change

Current behavior in `internal/views/loader.go`:
```go
// LoadView checks built-in first, then disk
switch strings.ToLower(name) {
case "default", "":
    return DefaultView(), nil  // Always returns built-in
case "all":
    return AllView(), nil      // Always returns built-in
}
// Then tries disk...
```

New behavior:
```go
// LoadView checks disk first for overrides, then built-in
if l.viewsDir != "" {
    viewPath := filepath.Join(l.viewsDir, name+".yaml")
    if _, err := os.Stat(viewPath); err == nil {
        // User file exists - load it (even for "default" and "all")
        return l.loadFromDisk(name)
    }
}
// Fall back to built-in
switch strings.ToLower(name) {
case "default", "":
    return DefaultView(), nil
case "all":
    return AllView(), nil
}
```

### Views Folder Setup

Create function to initialize views folder:

```go
// SetupViewsFolder creates views directory with default view files
// Returns true if folder was created, false if already existed
func SetupViewsFolder(viewsDir string) (bool, error) {
    if _, err := os.Stat(viewsDir); err == nil {
        return false, nil // Already exists
    }

    if err := os.MkdirAll(viewsDir, 0755); err != nil {
        return false, err
    }

    // Write default.yaml
    defaultYAML := `name: default
description: Standard task display for everyday use (excludes completed tasks)
fields:
  - name: status
    width: 12
  - name: summary
    width: 40
  - name: priority
    width: 10
filters:
  - field: status
    operator: ne
    value: DONE
`
    if err := os.WriteFile(filepath.Join(viewsDir, "default.yaml"), []byte(defaultYAML), 0644); err != nil {
        return false, err
    }

    // Write all.yaml
    allYAML := `name: all
description: Comprehensive display showing all task metadata
fields:
  - name: status
  - name: summary
  - name: description
  - name: priority
  - name: due_date
  - name: start_date
  - name: created
  - name: modified
  - name: completed
  - name: tags
  - name: uid
  - name: parent
`
    if err := os.WriteFile(filepath.Join(viewsDir, "all.yaml"), []byte(allYAML), 0644); err != nil {
        return false, err
    }

    return true, nil
}
```

### Prompt Integration

In main command flow (before rendering tasks):

```go
// Check if views folder needs setup
viewsDir := filepath.Join(configDir, "views")
if _, err := os.Stat(viewsDir); os.IsNotExist(err) {
    if autoYes {
        // -y flag: silently create
        views.SetupViewsFolder(viewsDir)
    } else {
        // Prompt user
        fmt.Print("Views folder not found. Create with default views? [Y/n] ")
        // Read input, create if confirmed
    }
}
```

### Documentation Updates

Update `docs/explanation/views-customization.md`:

1. Add section on views folder setup
2. Document that default view filters DONE tasks
3. Document how to override built-in views
4. Fix example output (remove DONE task from default view example)

## Files to Modify

- `internal/views/loader.go` - Change loading order, add setup function
- `internal/views/types.go` - Add YAML generation for built-in views
- `cmd/todoat/root.go` or equivalent - Add views folder check/prompt
- `docs/explanation/views-customization.md` - Update documentation
