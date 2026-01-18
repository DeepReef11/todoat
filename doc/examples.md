# Examples

Practical examples and common workflows for using todoat.

## Quick Start Example

```bash
# Add some tasks to a shopping list
todoat Shopping add "Milk"
todoat Shopping add "Bread"
todoat Shopping add "Eggs"

# View your list
todoat Shopping

# Mark items as complete
todoat Shopping complete "Milk"
todoat Shopping complete "Bread"

# View again - completed items shown
todoat Shopping
```

## Work Task Management

```bash
# Add work tasks with priorities
todoat Work add "Finish quarterly report" -p 1
todoat Work add "Review pull requests" -p 2
todoat Work add "Update documentation" -p 3

# View tasks
todoat Work

# Start working on a task
todoat Work update "report" -s IN-PROGRESS

# Complete when done
todoat Work complete "report"
```

## Using Action Aliases

All actions have single-letter aliases for faster typing:

```bash
# Add a task (a = add)
todoat Tasks a "New task"

# Get/view tasks (g = get)
todoat Tasks g

# Update a task (u = update)
todoat Tasks u "task" -s DONE

# Complete a task (c = complete)
todoat Tasks c "task"

# Delete a task (d = delete)
todoat Tasks d "task"
```

## Managing Task Status

### Status Workflow

```bash
# New tasks start as TODO (NEEDS-ACTION)
todoat Project add "Design database schema"

# Mark as in progress when you start
todoat Project update "schema" -s IN-PROGRESS

# Mark as done when complete
todoat Project update "schema" -s DONE

# Or use the complete shortcut
todoat Project complete "schema"

# Cancel a task that's no longer needed
todoat Project update "old feature" -s CANCELLED
```

### Status Values

```bash
# Available statuses
todoat Project update "task" -s TODO
todoat Project update "task" -s IN-PROGRESS
todoat Project update "task" -s DONE
todoat Project update "task" -s CANCELLED
```

## Priority Management

Priority ranges from 0-9, where:
- `0` = Undefined (default)
- `1` = Highest priority
- `9` = Lowest priority

```bash
# Add high-priority task
todoat Work add "Critical bug fix" -p 1

# Add medium-priority task
todoat Work add "Feature request" -p 5

# Add low-priority task
todoat Work add "Nice to have" -p 9

# Change priority
todoat Work update "bug fix" -p 2
```

### Priority Filtering

```bash
# Show only priority 1 tasks
todoat Work get -p 1

# Show priorities 1, 2, or 3
todoat Work get -p 1,2,3

# Use aliases for common filters
todoat Work get -p high     # priorities 1-4
todoat Work get -p medium   # priority 5
todoat Work get -p low      # priorities 6-9
```

## Tags

Tags (also called categories) help organize tasks across lists:

```bash
# Add task with tags
todoat Work add "Review PR #123" --tag code-review
todoat Work add "Fix login bug" --tag bug,urgent
todoat Work add "Design meeting" --tag meeting --tag design

# Filter by tag (OR logic - shows tasks with ANY matching tag)
todoat Work get --tag bug
todoat Work get --tag bug,urgent           # Tasks with "bug" OR "urgent"
todoat Work get --tag code-review --tag bug # Same as above

# Update tags (replaces existing)
todoat Work update "PR" --tag code-review,approved

# Clear all tags from a task
todoat Work update "PR" --tag ""
```

### Tag Display

In text output, tags appear in curly braces:
```
Tasks in 'Work':
  [TODO] Fix login bug [P1] {bug,urgent}
  [IN-PROGRESS] Review PR #123 {code-review}
  [TODO] Design meeting {meeting,design}
```

In JSON output, tags are an array:
```bash
todoat Work get --json | jq '.tasks[].tags'
```

## Subtasks and Hierarchies

Organize complex projects with subtasks:

### Creating Subtasks with Path Notation

```bash
# Create a project hierarchy
todoat Project add "Website Redesign/Design Phase/Wireframes"
todoat Project add "Website Redesign/Design Phase/Color Palette"
todoat Project add "Website Redesign/Development/Frontend"
todoat Project add "Website Redesign/Development/Backend"

# View the tree structure
todoat Project
```

Output:
```
Tasks in 'Project':
  [TODO] Website Redesign
  └─ [TODO] Design Phase
       └─ [TODO] Wireframes
       └─ [TODO] Color Palette
  └─ [TODO] Development
       └─ [TODO] Frontend
       └─ [TODO] Backend
```

### Creating Subtasks with --parent Flag

```bash
# Create a parent task
todoat Work add "Code Review"

# Add subtasks using --parent
todoat Work add "Review authentication module" --parent "Code Review"
todoat Work add "Review database queries" --parent "Code Review"
todoat Work add "Review API endpoints" --parent "Code Review"
```

### Reorganizing Tasks

```bash
# Move a task under a different parent
todoat Work update "API endpoints" --parent "Backend"

# Make a subtask into a root-level task
todoat Work update "API endpoints" --no-parent
```

### Working with Subtasks

```bash
# Complete a subtask
todoat Project complete "Wireframes"

# Update a subtask status
todoat Project update "Frontend" -s IN-PROGRESS

# Delete a parent (also deletes all subtasks)
todoat Project delete "Design Phase"
```

### Using Literal Mode

If your task name contains `/`, use `--literal` to prevent hierarchy parsing:

```bash
# Creates a task named "Fix A/B test bug" (not a hierarchy)
todoat Work add "Fix A/B test bug" --literal

# Without --literal, this would create: Fix > A > B test bug
```

## Due Dates

Set start and due dates for task scheduling:

```bash
# Add task with due date
todoat Work add "Submit report" --due-date 2026-01-31

# Add task with start and due date
todoat Work add "Project phase 1" --start-date 2026-01-20 --due-date 2026-02-15

# Update due date
todoat Work update "report" --due-date 2026-02-01

# Clear due date
todoat Work update "report" --due-date ""

# Combine with priority
todoat Work add "Urgent deadline" -p 1 --due-date 2026-01-25
```

## Multiple Lists

```bash
# Create different lists for different contexts
todoat Work add "Finish report"
todoat Personal add "Buy groceries"
todoat "Home Improvement" add "Fix leaky faucet"

# View specific list
todoat Work
todoat Personal
todoat "Home Improvement"
```

## Task Matching

todoat uses intelligent matching - you don't need the exact task name:

```bash
# Add a task
todoat Work add "Review quarterly financial report"

# Complete using partial match
todoat Work complete "quarterly"

# Or even shorter
todoat Work complete "report"

# Just be specific enough to match one task
```

## Scripting with No-Prompt Mode

Use `-y` flag for automation:

```bash
#!/bin/bash
# Example script to add tasks from a file

while read task; do
    todoat -y Imported add "$task"
done < tasks.txt
```

```bash
# Delete without confirmation
todoat -y Cleanup delete "old task"
```

### Result Codes

In no-prompt mode, todoat outputs result codes for scripting:

```bash
# Check if action completed
output=$(todoat -y Work add "Task" 2>&1)
if echo "$output" | grep -q "ACTION_COMPLETED"; then
    echo "Task added successfully"
fi
```

Result codes:
- `ACTION_COMPLETED` - Modification succeeded
- `INFO_ONLY` - Read-only operation
- `ERROR` - Operation failed

## JSON Output

Use `--json` for machine-readable output:

```bash
# Get tasks as JSON
todoat Work get --json

# Pretty print with jq
todoat Work get --json | jq '.'

# Get task count
todoat Work get --json | jq '.count'

# Get task summaries
todoat Work get --json | jq '.tasks[].summary'

# View all lists as JSON
todoat list --json
```

### JSON in Scripts

```bash
#!/bin/bash
# Example: Count tasks by status

TODO_COUNT=$(todoat Work get -s TODO --json | jq '.count')
DONE_COUNT=$(todoat Work get -s DONE --json | jq '.count')

echo "Todo: $TODO_COUNT, Done: $DONE_COUNT"
```

```bash
#!/bin/bash
# Example: Add task and capture ID

RESULT=$(todoat Work add "New task" --json)
TASK_ID=$(echo "$RESULT" | jq -r '.task.uid')
echo "Created task with ID: $TASK_ID"
```

## List Management

View and manage task lists:

```bash
# View all lists
todoat list

# Create a new list
todoat list create "Projects"

# Get list info
todoat list info "Projects"

# Lists as JSON
todoat list --json | jq '.[] | {name, tasks}'
```

### Deleting and Restoring Lists

Lists can be soft-deleted (moved to trash) and restored:

```bash
# Delete a list (moves to trash)
todoat list delete "OldProjects"

# View deleted lists
todoat list trash

# Restore a deleted list
todoat list trash restore "OldProjects"

# Permanently delete a list
todoat list trash purge "OldProjects"
```

### List Cleanup Example

```bash
# Create some lists
todoat list create "Temp"
todoat list create "Archive"

# Delete old lists
todoat list delete "Temp"
todoat list delete "Archive"

# Check trash
todoat list trash

# Restore one
todoat list trash restore "Archive"

# Purge the other
todoat list trash purge "Temp"
```

## Renaming Tasks

```bash
# Rename a task using --summary
todoat Work add "Fix bug"
todoat Work update "bug" --summary "Fix critical authentication bug"

# The task is now named "Fix critical authentication bug"
```

## Daily Workflow Example

```bash
# Morning: Review tasks
todoat Today

# Add new tasks as they come up
todoat Today add "Morning standup" -p 1
todoat Today add "Code review" -p 2
todoat Today add "Email responses" -p 3

# Work through the day
todoat Today update "standup" -s IN-PROGRESS
todoat Today complete "standup"

todoat Today update "Code review" -s IN-PROGRESS
todoat Today complete "Code review"

# End of day: Check remaining
todoat Today
```

## Using Views

Views let you customize how tasks are displayed:

```bash
# Use built-in 'all' view to see all task fields
todoat Work get --view all

# Use short form
todoat Work get -v all

# List available views
todoat view list
```

### Custom Views

Create custom views in `~/.config/todoat/views/`:

```bash
# Create views directory
mkdir -p ~/.config/todoat/views

# Create a custom view file
cat > ~/.config/todoat/views/urgent.yaml << 'EOF'
name: urgent
description: High priority tasks only

fields:
  - name: status
  - name: summary
  - name: priority
  - name: due_date

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
EOF

# Use your custom view
todoat Work get --view urgent
```

### Combining Views with Filters

```bash
# Use urgent view, but only show tasks tagged 'work'
todoat Project get --view urgent --tag work

# Use all view filtered to in-progress status
todoat Project get -v all -s IN-PROGRESS
```

## Tips

### Use Short List Names

```bash
# Instead of
todoat "Work Projects" add "Task"

# Use
todoat Work add "Task"
```

### Use Aliases for Speed

```bash
# Fast task addition
todoat W a "Quick task"

# Fast completion
todoat W c "Quick task"
```

### Quote Task Names with Spaces

```bash
# Always quote task names with spaces
todoat Work add "This is a task with spaces"
todoat Work complete "task with spaces"
```

---
*Last updated: 2026-01-18*
