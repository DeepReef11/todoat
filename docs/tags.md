# Tags (Categories)

Tags help organize tasks across lists. This guide covers adding, managing, and filtering by tags.

## Adding Tags to Tasks

### When Creating a Task

```bash
# Single tag
todoat MyList add "Bug fix" --tags "urgent"

# Multiple tags (comma-separated)
todoat MyList add "Feature request" --tags "feature,frontend,v2"
```

### Adding Tags to Existing Tasks

```bash
# Add a single tag
todoat MyList update "Bug fix" --add-tag "important"

# Add multiple tags
todoat MyList update "Bug fix" --add-tag "backend" --add-tag "critical"
```

### Removing Tags

```bash
# Remove specific tag
todoat MyList update "Bug fix" --remove-tag "urgent"
```

### Replacing All Tags

```bash
# Replace with new tags
todoat MyList update "Bug fix" --tags "new,tags"

# Clear all tags
todoat MyList update "Bug fix" --tags ""
```

## Viewing Tags

### List All Tags

```bash
todoat tags
```

Output:
```
Tags in use:
  work (15 tasks)
  urgent (8 tasks)
  home (5 tasks)
  project-x (3 tasks)
```

### Tags in Specific List

```bash
todoat tags --list "Work"
```

### JSON Output

```bash
todoat --json tags
```

```json
{
  "tags": [
    {"name": "work", "count": 15},
    {"name": "urgent", "count": 8},
    {"name": "home", "count": 5}
  ],
  "result": "INFO_ONLY"
}
```

## Filtering by Tags

### View Tasks with Tag

```bash
# Tasks with "urgent" tag
todoat MyList --filter "tags:urgent"
```

### Using Views

Create a view that filters by tags:

```yaml
# ~/.config/todoat/views/work.yaml
name: work
fields:
  - name: status
  - name: summary
  - name: tags
filters:
  - field: tags
    operator: contains
    value: work
```

Use it:

```bash
todoat MyList -v work
```

### Multiple Tag Filters

In view YAML:

```yaml
# Tasks with ANY of these tags
filters:
  - field: tags
    operator: in
    value: [work, urgent]

# Tasks with ALL of these tags (multiple filters)
filters:
  - field: tags
    operator: contains
    value: work
  - field: tags
    operator: contains
    value: urgent
```

## Tag Best Practices

### Consistent Naming

- Use lowercase: `work` instead of `Work`
- Use hyphens: `project-alpha` instead of `project alpha`
- Keep short: `urgent` instead of `very-urgent-please-fix`

### Common Tag Patterns

**By Context:**
- `work`, `personal`, `home`

**By Priority:**
- `urgent`, `important`, `someday`

**By Project:**
- `project-alpha`, `project-beta`

**By Action:**
- `waiting`, `blocked`, `review`

**By Time:**
- `quick`, `long-term`, `daily`

## Tag Storage by Backend

Tags are stored differently depending on your backend:

| Backend | Storage |
|---------|---------|
| SQLite | JSON array in categories column |
| Nextcloud | CATEGORIES property (iCalendar) |
| Todoist | Mapped to Todoist labels |
| Git | Inline hashtags in markdown (`#tag`) |

Tags sync between backends when using synchronization.

## Examples

### Project Management

```bash
# Tag tasks by project phase
todoat Work add "Create mockups" --tags "project-x,design"
todoat Work add "Build API" --tags "project-x,backend"
todoat Work add "Write tests" --tags "project-x,testing"

# View all project-x tasks
todoat Work -v all --filter "tags:project-x"
```

### Priority System

```bash
# Mark urgent tasks
todoat Work update "Server down" --add-tag "urgent"

# Create urgent view
# In ~/.config/todoat/views/urgent.yaml
name: urgent
filters:
  - field: tags
    operator: contains
    value: urgent
  - field: status
    operator: ne
    value: DONE
sort:
  - field: priority
    direction: asc

# Use it
todoat Work -v urgent
```

### Waiting/Blocked Tasks

```bash
# Mark blocked tasks
todoat Work update "Review PR" --add-tag "waiting"

# Later, clear the tag
todoat Work update "Review PR" --remove-tag "waiting"
```

## See Also

- [Task Management](task-management.md) - Task operations
- [Views](views.md) - Filtering views by tags
- [Synchronization](sync.md) - Tag sync across backends
