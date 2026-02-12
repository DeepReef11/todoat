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

# Show TODO and IN-PROGRESS tasks
todoat MyList -s TODO,IN-PROGRESS

# Using abbreviations (T=TODO, D=DONE, I=IN-PROGRESS, C=CANCELLED)
todoat MyList -s T,I
```

### Using Custom Views

```bash
# Use the "all" view (shows all fields)
todoat MyList -v all

# Use a custom view
todoat MyList -v myview
```

### Filtering by Date

Filter tasks by due date or creation date:

```bash
# Tasks due today or later
todoat MyList --due-after today

# Tasks due before next week
todoat MyList --due-before +7d

# Tasks due within a date range
todoat MyList --due-after 2026-01-15 --due-before 2026-01-31

# Tasks created in the last week
todoat MyList --created-after -7d

# Tasks created before a specific date
todoat MyList --created-before 2026-01-01
```

| Flag | Description |
|------|-------------|
| `--due-after` | Tasks due on or after this date (inclusive) |
| `--due-before` | Tasks due before this date (inclusive) |
| `--created-after` | Tasks created on or after this date |
| `--created-before` | Tasks created before this date |

Combine with other filters:

```bash
# High priority tasks due this week
todoat MyList -s TODO --due-after today --due-before +7d -p 1,2,3

# Completed tasks from the past month
todoat MyList -s DONE --created-after -30d
```

### Filtering by Priority

Filter tasks by priority level:

```bash
# Using range syntax
todoat MyList -p 1-4     # priorities 1 through 4

# Using comma-separated values
todoat MyList -p 1,2,3,4

# Using named levels
todoat MyList -p high    # priorities 1-4
todoat MyList -p medium  # priority 5
todoat MyList -p low     # priorities 6-9
```

### Filtering by Tag

```bash
# Tasks with a specific tag
todoat MyList --tag urgent

# Tasks with multiple tags (comma-separated)
todoat MyList --tag "work,important"
```

### Pagination

For large task lists, use pagination to limit output:

```bash
# Show first 20 tasks
todoat MyList --limit 20

# Show page 2 (with default page size of 50)
todoat MyList --page 2

# Custom page size
todoat MyList --page 1 --page-size 25

# Manual offset and limit
todoat MyList --offset 50 --limit 25
```

| Flag | Description |
|------|-------------|
| `--limit` | Maximum number of tasks to show |
| `--offset` | Number of tasks to skip |
| `--page` | Page number (1-indexed) |
| `--page-size` | Tasks per page (default: 50) |

Combine pagination with filters:

```bash
# First 10 high-priority TODO tasks
todoat MyList -s TODO -p high --limit 10

# Page 2 of tasks due this week
todoat MyList --due-before +7d --page 2
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

### Relative Dates with Time

Combine relative dates with specific times by adding a space-separated time component:

```bash
# Tomorrow at a specific time
todoat MyList add "Morning standup" --due-date "tomorrow 09:00"
todoat MyList add "Afternoon review" --due-date "tomorrow 14:30"

# Days from now with time
todoat MyList add "Team sync" --due-date "+2d 10:00"
todoat MyList add "Weekly review" --due-date "+7d 15:00"

# Today with specific time
todoat MyList add "Lunch meeting" --due-date "today 12:30"

# With seconds (optional)
todoat MyList add "Precise event" --due-date "tomorrow 09:00:00"
```

**Syntax**: `<relative-date> HH:MM` or `<relative-date> HH:MM:SS`

| Component | Format | Range |
|-----------|--------|-------|
| Hours | 1-2 digits | 0-23 |
| Minutes | 2 digits | 00-59 |
| Seconds | 2 digits (optional) | 00-59 |

**Note**: Relative dates with time always use local timezone. Timezone offsets (like `-05:00`) are not supported with relative dates. For timezone-aware scheduling, use absolute ISO 8601 format instead.

### Task with Time (Absolute Dates)

Set specific times using ISO 8601 format:

```bash
# ISO 8601 format with time (local timezone)
todoat MyList add "Team meeting" --due-date 2026-01-20T14:30

# With seconds
todoat MyList add "Precise meeting" --due-date 2026-01-20T14:30:00

# With timezone offset
todoat MyList add "Client call" --due-date "2026-01-20T14:30-05:00"

# UTC timezone
todoat MyList add "Server maintenance" --due-date "2026-01-20T14:30Z"
```

Date-only input defaults to midnight (00:00). Tasks with time show the time component in output:

```
TODO   Team meeting                       Jan 20 14:30
TODO   Client call                        Jan 20 14:30
```

### Recurring Tasks

Create tasks that automatically regenerate when completed:

```bash
# Basic recurrence patterns
todoat MyList add "Daily standup" --recur daily
todoat MyList add "Weekly review" --recur weekly
todoat MyList add "Monthly report" --recur monthly
todoat MyList add "Annual review" --recur yearly

# Custom intervals
todoat MyList add "Check logs" --recur "every 3 days"

# Recurring task with due date
todoat MyList add "Team meeting" --recur weekly --due-date "2026-01-20"
```

By default, the next occurrence is calculated from the original due date. To base it on when you complete the task:

```bash
# Next due date = completion date + interval
todoat MyList add "Water plants" --recur "every 3 days" --recur-from-completion
```

Recurring tasks display with an `[R]` indicator in the task list:

```
[TODO] Daily standup [R]                   Jan 20
[TODO] One-time task                       Jan 21
```

When you complete a recurring task, a new task is automatically created with the next due date:

```bash
todoat MyList complete "Daily standup"
# Original task marked DONE
# New task created with tomorrow's date
```

Remove recurrence from an existing task:

```bash
todoat MyList update "Daily standup" --recur none
```

| Pattern | Meaning |
|---------|---------|
| `daily` | Every day |
| `weekly` | Every week |
| `monthly` | Every month |
| `yearly` | Every year |
| `every N days` | Every N days |
| `every N weeks` | Every N weeks |
| `every N months` | Every N months |
| `none` | Remove recurrence |

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
todoat MyList update "task" -s IN-PROGRESS
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

### Update Parent Relationship

```bash
# Move task under a parent
todoat MyList update "subtask" -P "New Parent"

# Move task under a nested parent using path syntax
todoat MyList update "subtask" -P "Project Alpha/Design"

# Make task a root-level task (remove parent)
todoat MyList update "subtask" --no-parent
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

## Bulk Operations

Operate on multiple tasks at once using glob patterns. Bulk operations work with hierarchical task structures.

### Pattern Syntax

| Pattern | Meaning |
|---------|---------|
| `Parent/*` | Direct children of Parent only |
| `Parent/**` | All descendants of Parent (children, grandchildren, etc.) |

### Bulk Complete

Complete multiple tasks with a single command:

```bash
# Complete direct children only
todoat MyList complete "Parent/*"

# Complete all descendants (children, grandchildren, etc.)
todoat MyList complete "Parent/**"
```

Example with a project hierarchy:

```bash
# Create project structure
todoat MyList add "Release v2.0"
todoat MyList add "Feature A" -P "Release v2.0"
todoat MyList add "Feature B" -P "Release v2.0"
todoat MyList add "Task A1" -P "Feature A"

# Complete all release tasks at once
todoat MyList complete "Release v2.0/**"
# Output: Completed 3 tasks under "Release v2.0"
```

### Bulk Update

Update properties on multiple tasks:

```bash
# Set high priority on all subtasks
todoat MyList update "Project/**" -p 1

# Update direct children only
todoat MyList update "Project/*" -p 2
```

### Bulk Delete

Delete multiple tasks with a single command:

```bash
# Delete direct children only
todoat MyList delete "Parent/*"

# Delete all descendants
todoat MyList delete "Parent/**"
```

Note: When deleting direct children with `/*`, grandchildren are also deleted (cascade delete).

### Output and Feedback

Bulk operations display the count of affected tasks:

```
Completed 5 tasks under "Release v2.0"
```

### JSON Output

Use `--json` for machine-readable bulk operation results:

```bash
todoat --json MyList complete "Parent/**"
```

JSON response structure:

```json
{
  "result": "ACTION_COMPLETED",
  "affected_count": 3,
  "pattern": "**",
  "parent": "Parent"
}
```

### Error Handling

| Scenario | Result Code | Description |
|----------|-------------|-------------|
| Parent not found | `ERROR` | The specified parent task doesn't exist |
| No children match | `INFO_ONLY` | Parent exists but has no children (count: 0) |
| Success | `ACTION_COMPLETED` | Tasks were modified |

Examples:

```bash
# Error: Parent doesn't exist
todoat MyList complete "NonExistent/*"
# Result: ERROR - task "NonExistent" not found

# Info: Parent exists but has no children
todoat MyList complete "LeafTask/*"
# Result: INFO_ONLY - 0 tasks affected
```

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

When exactly one task matches, todoat uses it directly:

```bash
todoat MyList complete "groceries"
# Completes "Buy groceries" immediately
```

### Multiple Matches

When multiple tasks match your search term, todoat enters interactive selection mode:

```
Multiple tasks match 'review':
Filter (or press Enter to show all):

  1) Review PR #456 [TODO, P2, due: 2026-02-01]
  2) Code review guidelines [TODO, tags: docs]
  3) Review meeting notes [TODO, P5]
Select (0 to cancel):
```

Type a filter term to narrow the list, or press Enter to show all matches. Then select by number.

By default, interactive selection only shows tasks relevant to the action (e.g., only TODO/IN-PROGRESS tasks for complete). To include completed and cancelled tasks in the selection:

```yaml
# config.yaml
ui:
  interactive_prompt_for_all_tasks: true
```

If you're running in non-interactive mode (`-y` / `--no-prompt`), todoat shows an error with UIDs instead:

```
Error: multiple tasks match 'review'. Use --uid to specify:
  - Review PR #456 [P2, due: 2026-02-01] (UID: 550e8400-e29b-41d4-a716-446655440000)
  - Code review guidelines [desc: "Guidelines for code reviews..."] (UID: 660f9500-f39c-52e5-b827-557766551111)
  - Review meeting notes [P5] (UID: 770a0600-a40d-63f6-c938-668877662222)
```

Re-run the command with `--uid` to specify which task:

```bash
todoat MyList complete --uid "550e8400-e29b-41d4-a716-446655440000"
```

### Direct Selection by UID

For scripts or when you know the task ID:

```bash
# Select by backend UID (bypasses summary search)
todoat MyList complete --uid "550e8400-e29b-41d4-a716-446655440000"

# Select by local ID (requires sync enabled)
todoat MyList complete --local-id 42
```

## Task Status Values

| Status | Meaning |
|--------|---------|
| TODO | Task not started |
| IN-PROGRESS | Task in progress |
| DONE | Task completed |
| CANCELLED | Task abandoned |

### Status Abbreviations

Use single letters in commands:

| Abbreviation | Status |
|--------------|--------|
| T | TODO |
| I | IN-PROGRESS |
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
todoat Work -s TODO,IN-PROGRESS

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
- [CLI Reference](../reference/cli.md) - Complete command reference
