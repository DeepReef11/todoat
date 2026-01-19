# Views

Views allow you to customize how tasks are displayed. You can use built-in views or create custom views with specific fields, filters, and sorting.

## Using Views

Use the `--view` or `-v` flag to apply a view when listing tasks:

```bash
# Use default view (can be omitted)
todoat MyList get --view default

# Use the 'all' view to see all task fields
todoat MyList get -v all

# Use a custom view
todoat MyList get --view my-custom-view
```

## Built-in Views

todoat includes two built-in views:

| View | Description |
|------|-------------|
| `default` | Standard task display showing status, summary, and priority |
| `all` | Comprehensive display showing all task metadata |

## Listing Views

To see all available views (both built-in and custom):

```bash
todoat view list
```

Output:
```
Available views:
  - default (built-in)
  - all (built-in)
  - my-custom-view (custom)
```

## Creating Custom Views

Custom views are YAML files stored in:

```
~/.config/todoat/views/
```

Each view file should have a `.yaml` extension. The filename (without extension) becomes the view name.

### View Structure

```yaml
# Example: ~/.config/todoat/views/work.yaml
name: work
description: Work tasks view with priority sorting

# Fields to display
fields:
  - name: status
    width: 12
  - name: summary
    width: 50
  - name: priority
    width: 10
  - name: due_date
    width: 12
    format: "2006-01-02"

# Optional: Filter conditions
filters:
  - field: status
    operator: ne
    value: DONE

# Optional: Sort order
sort:
  - field: priority
    direction: asc
  - field: due_date
    direction: asc
```

### Available Fields

| Field | Description |
|-------|-------------|
| `status` | Task status (TODO, IN-PROGRESS, DONE, CANCELLED) |
| `summary` | Task title/name |
| `description` | Task description |
| `priority` | Priority level (0-9) |
| `due_date` | Due date |
| `start_date` | Start date |
| `created` | Creation timestamp |
| `modified` | Last modified timestamp |
| `completed` | Completion timestamp |
| `tags` | Task tags/categories |
| `uid` | Unique task identifier |
| `parent` | Parent task ID (for subtasks) |

### Field Options

Each field can have the following options:

| Option | Type | Description |
|--------|------|-------------|
| `name` | string | Field name (required) |
| `width` | int | Column width in characters |
| `align` | string | Text alignment: `left`, `center`, `right` |
| `format` | string | Date format string (Go time format) |
| `truncate` | bool | Truncate text if exceeds width |

### Filter Operators

Filters allow you to show only tasks matching specific criteria:

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equals | `{field: status, operator: eq, value: TODO}` |
| `ne` | Not equals | `{field: status, operator: ne, value: DONE}` |
| `lt` | Less than | `{field: priority, operator: lt, value: 5}` |
| `lte` | Less than or equal | `{field: priority, operator: lte, value: 3}` |
| `gt` | Greater than | `{field: priority, operator: gt, value: 5}` |
| `gte` | Greater than or equal | `{field: priority, operator: gte, value: 7}` |
| `contains` | Contains substring | `{field: summary, operator: contains, value: review}` |
| `in` | In list | `{field: status, operator: in, value: [TODO, IN-PROGRESS]}` |
| `not_in` | Not in list | `{field: status, operator: not_in, value: [DONE, CANCELLED]}` |
| `regex` | Regex match | `{field: summary, operator: regex, value: "^Fix.*"}` |

### Sort Direction

Sort rules control the order tasks are displayed:

| Direction | Description |
|-----------|-------------|
| `asc` | Ascending (A-Z, 0-9, oldest first) |
| `desc` | Descending (Z-A, 9-0, newest first) |

## Example Views

### High Priority Tasks

```yaml
# ~/.config/todoat/views/urgent.yaml
name: urgent
description: High priority incomplete tasks

fields:
  - name: status
    width: 12
  - name: summary
    width: 50
  - name: due_date
    width: 12

filters:
  - field: priority
    operator: lte
    value: 3
  - field: status
    operator: ne
    value: DONE

sort:
  - field: priority
    direction: asc
  - field: due_date
    direction: asc
```

### Due Soon

```yaml
# ~/.config/todoat/views/due-soon.yaml
name: due-soon
description: Tasks with due dates, sorted by date

fields:
  - name: status
  - name: summary
  - name: due_date
  - name: priority

filters:
  - field: due_date
    operator: ne
    value: null

sort:
  - field: due_date
    direction: asc
```

### Minimal View

```yaml
# ~/.config/todoat/views/minimal.yaml
name: minimal
description: Just task names

fields:
  - name: summary
```

### All Details

```yaml
# ~/.config/todoat/views/detailed.yaml
name: detailed
description: All task information

fields:
  - name: uid
    width: 8
  - name: status
    width: 12
  - name: summary
    width: 40
  - name: priority
    width: 8
  - name: tags
    width: 20
  - name: due_date
    width: 12
  - name: start_date
    width: 12
  - name: created
    width: 20
  - name: modified
    width: 20
```

## Views Directory

The views directory is located at:

```
~/.config/todoat/views/
```

You can create this directory manually if it doesn't exist:

```bash
mkdir -p ~/.config/todoat/views
```

## Combining Views with Filters

Views can be combined with command-line filters:

```bash
# Use urgent view, but only show tasks tagged 'work'
todoat MyList get --view urgent --tag work

# Use detailed view, filtered to IN-PROGRESS status
todoat MyList get -v detailed -s IN-PROGRESS
```

Command-line filters are applied **after** view filters.

---
*Last updated: 2026-01-19*
