# Views Customization

Views control how tasks are displayed. This guide covers using built-in views and creating custom views.

## Using Views

### Select a View

```bash
# Use the "all" view
todoat MyList -v all

# Use a custom view
todoat MyList -v myview

# Default view (no flag needed)
todoat MyList
```

## Built-in Views

### Default View

Shows: status, summary, priority

```bash
todoat MyList
```

Output:
```
TODO         Buy groceries                              1 (High)
IN-PROGRESS  Write documentation                        5 (Medium)
DONE         Review pull request                        9 (Low)
```

### All View

Shows all available fields:

```bash
todoat MyList -v all
```

Output includes:
- Status
- Summary
- Description
- Priority
- Due date
- Start date
- Created timestamp
- Modified timestamp
- Completion timestamp
- Tags
- UID
- Parent task

## Custom Views

### View Location

Custom views are stored in: `~/.config/todoat/views/`

Each view is a YAML file named `{viewname}.yaml`.

### Creating a Custom View

Create a YAML file in the views directory. For example, `~/.config/todoat/views/urgent.yaml`:

```yaml
name: urgent
description: "High priority tasks due soon"
fields:
  - name: status
    width: 10
  - name: summary
    width: 50
  - name: due_date
    width: 12
  - name: priority
    width: 8
filters:
  - field: priority
    operator: lte
    value: 3
  - field: status
    operator: ne
    value: DONE
sort:
  - field: due_date
    direction: asc
  - field: priority
    direction: asc
```

### Available Fields

| Field | Description |
|-------|-------------|
| `status` | Task status (TODO, DONE, etc.) |
| `summary` | Task title |
| `description` | Detailed description |
| `priority` | Priority (1-9) |
| `due_date` | Due date |
| `start_date` | Start date |
| `created` | Creation timestamp |
| `modified` | Last modified timestamp |
| `completed` | Completion timestamp |
| `tags` | Categories/tags |
| `uid` | Unique identifier |
| `parent` | Parent task UID |

### Field Configuration

```yaml
fields:
  - name: summary
    width: 40        # Display width in characters
    align: left      # left, center, or right
    truncate: true   # Truncate if exceeds width
```

## Filtering

### Filter Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equals | `status eq TODO` |
| `ne` | Not equals | `status ne DONE` |
| `lt` | Less than | `priority lt 5` |
| `lte` | Less than or equal | `priority lte 3` |
| `gt` | Greater than | `created gt 2026-01-01` |
| `gte` | Greater than or equal | `due_date gte today` |
| `contains` | String contains | `summary contains meeting` |
| `in` | Value in list | `tags in [work, urgent]` |
| `not_in` | Not in list | `status not_in [DONE, CANCELLED]` |

### Date Filters

Use relative dates:
- `today` - Current date
- `tomorrow` - Next day
- `+7d` - 7 days from now
- `-3d` - 3 days ago
- `+2w` - 2 weeks from now

Example - Tasks due this week:

```yaml
filters:
  - field: due_date
    operator: gte
    value: today
  - field: due_date
    operator: lte
    value: "+7d"
```

### Tag Filters

```yaml
# Tasks with "work" tag
filters:
  - field: tags
    operator: contains
    value: work

# Tasks with any of these tags
filters:
  - field: tags
    operator: in
    value: [work, personal, urgent]
```

## Sorting

### Sort Configuration

```yaml
sort:
  - field: priority
    direction: asc   # asc or desc
  - field: due_date
    direction: asc
```

### Common Sort Patterns

Priority first:
```yaml
sort:
  - field: priority
    direction: asc
```

Due date first:
```yaml
sort:
  - field: due_date
    direction: asc
  - field: priority
    direction: asc
```

Newest first:
```yaml
sort:
  - field: created
    direction: desc
```

## Plugin Formatters

Custom scripts can format field values.

### Configuration

```yaml
fields:
  - name: status
    plugin:
      command: "~/.config/todoat/plugins/status-emoji.sh"
      timeout: 1000  # milliseconds
```

### Example Plugin (Bash)

`~/.config/todoat/plugins/status-emoji.sh`:

```bash
#!/bin/bash
json=$(cat)
status=$(echo "$json" | jq -r '.status')
case $status in
  "TODO") echo "[ ]";;
  "IN-PROGRESS") echo "[~]";;
  "DONE") echo "[x]";;
  "CANCELLED") echo "[-]";;
  *) echo "[ ]";;
esac
```

Make executable:
```bash
chmod +x ~/.config/todoat/plugins/status-emoji.sh
```

### Plugin Input

Plugins receive task data as JSON on stdin:

```json
{
  "uid": "task-123",
  "summary": "Buy groceries",
  "status": "TODO",
  "priority": 1,
  "due_date": "2026-01-15T14:00:00Z",
  "categories": ["shopping", "urgent"]
}
```

### Plugin Output

Output a single line to stdout.

## Managing Views

### List Views

```bash
todoat view list
```

Output:
```
Available views:
  - default (built-in)
  - all (built-in)
  - urgent (custom)
  - work (custom)
```

### Delete a View

```bash
rm ~/.config/todoat/views/myview.yaml
```

### Edit a View

```bash
nano ~/.config/todoat/views/myview.yaml
```

Changes take effect immediately.

## Example Views

### Work Tasks Due This Week

`~/.config/todoat/views/work-week.yaml`:

```yaml
name: work-week
description: "Work tasks due this week"
fields:
  - name: status
    width: 12
  - name: summary
    width: 40
  - name: due_date
    width: 12
  - name: priority
    width: 8
filters:
  - field: tags
    operator: contains
    value: work
  - field: due_date
    operator: lte
    value: "+7d"
  - field: status
    operator: ne
    value: DONE
sort:
  - field: due_date
    direction: asc
  - field: priority
    direction: asc
```

### Overdue Tasks

`~/.config/todoat/views/overdue.yaml`:

```yaml
name: overdue
description: "Tasks past their due date"
fields:
  - name: status
    width: 10
  - name: summary
    width: 50
  - name: due_date
    width: 12
filters:
  - field: due_date
    operator: lt
    value: today
  - field: status
    operator: ne
    value: DONE
sort:
  - field: due_date
    direction: asc
```

### Minimal View

`~/.config/todoat/views/minimal.yaml`:

```yaml
name: minimal
description: "Just task names"
fields:
  - name: summary
    width: 60
```

## See Also

- [Task Management](task-management.md) - Task operations
- [Tags](tags.md) - Using tags for filtering
