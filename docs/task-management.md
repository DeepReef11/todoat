# Task Management

This guide covers all task operations in todoat: creating, viewing, updating, completing, and deleting tasks.

## Viewing Tasks

### Basic Usage

```bash
# View all tasks in a list
todoat MyList

# Explicit get action
todoat MyList get
```

### Filtering by Status

```bash
# Show only TODO tasks
todoat MyList -s TODO

# Show TODO and PROCESSING tasks
todoat MyList -s TODO,PROCESSING

# Using abbreviations (T=TODO, D=DONE, P=PROCESSING, C=CANCELLED)
todoat MyList -s T,P
```

### Using Custom Views

```bash
# Use the "all" view (shows all fields)
todoat MyList -v all

# Use a custom view
todoat MyList -v myview
```

## Adding Tasks

### Basic Task Creation

```bash
todoat MyList add "Task summary"

# Using abbreviation
todoat MyList a "Task summary"
```

### Task with Priority

Priority ranges from 0-9 where 1 is highest and 9 is lowest. 0 means undefined.

```bash
# High priority task
todoat MyList add "Urgent bug fix" -p 1

# Low priority task
todoat MyList add "Nice to have" -p 8
```

### Task with Dates

```bash
# Set due date
todoat MyList add "Report deadline" --due-date 2026-01-31

# Set start date
todoat MyList add "Project kickoff" --start-date 2026-02-01

# Both dates
todoat MyList add "Sprint work" --start-date 2026-01-15 --due-date 2026-01-29
```

### Relative Dates

Use human-friendly relative dates instead of absolute dates:

```bash
# Common relative dates
todoat MyList add "Do today" --due-date today
todoat MyList add "Do tomorrow" --due-date tomorrow

# Days from now
todoat MyList add "Due in a week" --due-date +7d

# Weeks and months
todoat MyList add "Due in 2 weeks" --due-date +2w
todoat MyList add "Due next month" --due-date +1m

# Days ago (useful for filtering)
todoat MyList --due-after -3d
```

| Format | Meaning |
|--------|---------|
| `today` | Current date |
| `tomorrow` | Next day |
| `yesterday` | Previous day |
| `+Nd` | N days from now |
| `-Nd` | N days ago |
| `+Nw` | N weeks from now |
| `+Nm` | N months from now |

### Task with Time

Set specific times for due dates and start dates:

```bash
# ISO 8601 format with time
todoat MyList add "Team meeting" --due-date 2026-01-20T14:30

# With timezone
todoat MyList add "Client call" --due-date "2026-01-20T14:30-05:00"

# Relative date with time
todoat MyList add "Morning standup" --due-date "tomorrow 09:00"
todoat MyList add "Friday meeting" --due-date "+2d 14:00"
```

Date-only input defaults to midnight (00:00). Tasks with time show the time component in output:

```
TODO   Team meeting                       Jan 20 14:30
TODO   Client call                        Jan 20 14:30
```

### Task with Description

```bash
todoat MyList add "Meeting prep" -d "Prepare slides and agenda for Monday standup"
```

### Task with Tags

```bash
# Single tag
todoat MyList add "Bug fix" --tags "urgent"

# Multiple tags
todoat MyList add "Feature request" --tags "feature,frontend,v2"
```

### Creating Subtasks

Create hierarchical task structures:

```bash
# Using parent flag
todoat MyList add "Write tests" -P "Feature Development"

# Using path syntax (auto-creates parent if needed)
todoat MyList add "Project/Phase 1/Design mockups"
```

To use a literal "/" in task name:

```bash
todoat MyList add -l "UI/UX Review"
```

### Combined Options

```bash
todoat MyList add "Critical deadline" \
  -p 1 \
  -d "Must complete before release" \
  --due-date 2026-01-20 \
  --tags "urgent,release"
```

## Updating Tasks

### Change Status

```bash
todoat MyList update "task name" -s DONE
todoat MyList update "task" -s PROCESSING
```

### Change Priority

```bash
todoat MyList update "task" -p 3
```

### Rename Task

```bash
todoat MyList update "old name" --summary "new name"
```

### Update Description

```bash
todoat MyList update "task" -d "Updated description text"
```

### Update Dates

```bash
# Change due date
todoat MyList update "task" --due-date 2026-02-15

# Clear due date
todoat MyList update "task" --due-date ""
```

### Update Tags

```bash
# Add a tag
todoat MyList update "task" --add-tag "important"

# Remove a tag
todoat MyList update "task" --remove-tag "old-tag"

# Replace all tags
todoat MyList update "task" --tags "new,tags"

# Clear all tags
todoat MyList update "task" --tags ""
```

## Completing Tasks

Mark tasks as done with a single command:

```bash
todoat MyList complete "task name"

# Using abbreviation
todoat MyList c "task"
```

This is equivalent to `update "task" -s DONE` but also sets the completion timestamp.

## Deleting Tasks

```bash
todoat MyList delete "task name"

# Using abbreviation
todoat MyList d "task"
```

Note: Task deletion is permanent. Unlike lists, tasks cannot be restored from trash.

## Task Search and Matching

todoat uses intelligent matching to find tasks:

### Exact Match

If your search string exactly matches a task name (case-insensitive), it's selected immediately.

### Partial Match

If no exact match, todoat searches for tasks containing your search string:

```bash
# Matches "Buy groceries", "grocery list", etc.
todoat MyList complete "groceries"
```

### Single Match

When one task matches, todoat asks for confirmation:

```
Found: Buy groceries [TODO, Priority: 3]. Is this correct? (y/n)
```

### Multiple Matches

When multiple tasks match, you choose from a menu:

```
Multiple tasks found matching "review":
1. Review PR #456 (TODO, Priority: 2)
2. Code review guidelines (DONE)
3. Review meeting notes (PROCESSING)
Select task (1-3, or 'c' to cancel):
```

### Direct Selection by UID

For scripts or unambiguous operations:

```bash
# Select by backend UID (for synced tasks)
todoat MyList complete --uid "550e8400-e29b-41d4-a716-446655440000"

# Select by local ID (requires sync enabled)
todoat MyList complete --local-id 42
```

## Task Status Values

| Status | Meaning |
|--------|---------|
| TODO | Task not started |
| PROCESSING | Task in progress |
| DONE | Task completed |
| CANCELLED | Task abandoned |

### Status Abbreviations

Use single letters in commands:

| Abbreviation | Status |
|--------------|--------|
| T | TODO |
| P | PROCESSING |
| D | DONE |
| C | CANCELLED |

## Task Priority Values

| Priority | Meaning |
|----------|---------|
| 0 | Undefined (no priority) |
| 1 | Highest priority |
| 2-4 | High priority |
| 5 | Medium priority |
| 6-8 | Low priority |
| 9 | Lowest priority |

## Scripting and Automation

### Non-Interactive Mode

Use `-y` or `--no-prompt` for scripts:

```bash
# Delete without confirmation
todoat -y MyList delete "task"

# Complete first match without prompt
todoat -y MyList complete "task"
```

### JSON Output

Use `--json` for machine-readable output:

```bash
# List tasks as JSON
todoat --json MyList

# Add task with JSON response
todoat --json MyList add "New task"
```

### Result Codes

Commands return result codes for scripting:

| Code | Meaning |
|------|---------|
| `ACTION_COMPLETED` | Operation succeeded |
| `ACTION_INCOMPLETE` | Multiple matches found |
| `INFO_ONLY` | Display-only (no changes) |
| `ERROR` | Operation failed |

Example script usage:

```bash
result=$(todoat -y MyList complete "task" 2>&1 | tail -1)
if [ "$result" = "ACTION_COMPLETED" ]; then
    echo "Task completed"
fi
```

## Examples

### Daily Workflow

```bash
# Morning: Check today's tasks
todoat Work -s TODO,PROCESSING

# Add new tasks
todoat Work add "Morning standup" -p 2
todoat Work add "Review PR #789" --due-date $(date +%Y-%m-%d)

# Mark tasks complete
todoat Work complete "standup"

# End of day: Review completed
todoat Work -s DONE
```

### Project Management

```bash
# Create project structure
todoat Work add "Project Alpha"
todoat Work add "Project Alpha/Design"
todoat Work add "Project Alpha/Development"
todoat Work add "Project Alpha/Testing"

# Add subtasks
todoat Work add "Create wireframes" -P "Project Alpha/Design" -p 2
todoat Work add "Write unit tests" -P "Project Alpha/Testing"

# View project
todoat Work -v all
```

## See Also

- [List Management](list-management.md) - Creating and managing lists
- [Tags](tags.md) - Organizing tasks with categories
- [Views](views.md) - Customizing task display
