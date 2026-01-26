# Views Customization

**Category:** Display & Output Formatting
**Related Features:** [Task Management](task-management.md), [CLI Interface](cli-interface.md)

## Overview

The Views Customization system allows users to create personalized task display configurations that control which fields are shown, how they're formatted, and which tasks are included. This flexible system supports custom field selection, ordering, filtering, sorting, and even external plugin-based formatters for complete display control.

---

## Table of Contents

1. [View Concepts](#view-concepts)
2. [Built-in Views](#built-in-views)
3. [Custom View Creation](#custom-view-creation)
4. [Field Selection and Ordering](#field-selection-and-ordering)
5. [Filtering Tasks](#filtering-tasks)
6. [Sorting Tasks](#sorting-tasks)
7. [Plugin Formatters](#plugin-formatters)
8. [Interactive View Builder](#interactive-view-builder)
9. [View Storage and Management](#view-storage-and-management)
10. [Technical Architecture](#technical-architecture)

---

## View Concepts

### What is a View?

A **view** is a saved configuration that defines:
- **Which fields** to display (e.g., status, summary, due date, priority)
- **How to format** each field (e.g., emoji status, colored priorities)
- **Which tasks** to show (filters by status, priority, tags, dates)
- **How to order** tasks (sorting rules)
- **Display layout** (field widths, alignment, hierarchy display)

### Purpose

Views solve several user needs:
1. **Focused workflows**: Show only relevant fields for specific contexts (e.g., "due today" view shows only due date and summary)
2. **Information density**: Control how much detail to display (minimal vs. comprehensive)
3. **Personalization**: Customize output to match personal preferences or team standards
4. **Reusability**: Save and reuse display configurations across sessions
5. **Context switching**: Quickly switch between different task perspectives

### Key Benefits

- **No command-line clutter**: Complex filters and formatting saved in named views
- **Consistency**: Same view produces same output format every time
- **Shareability**: View YAML files can be shared with team members
- **Extensibility**: Plugin system allows unlimited customization
- **Performance**: Views with filters reduce output processing time

---

## Built-in Views

### Default View

**Purpose:** Standard task display for everyday use, showing only active (non-completed) tasks

**Configuration:**
```yaml
name: default
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
```

**Key Behavior:**
- **Filters out DONE/completed tasks** by default
- Use `-v all` to see completed tasks

**When Used:**
- When no view is specified
- General task browsing
- Quick task overview

**Output Example:**
```
TODO         Buy groceries                              1 (High)
IN-PROCESS   Write documentation                        5 (Medium)
```

Note: Completed tasks are hidden. Use `todoat MyList -v all` to see all tasks including completed ones.

### All View

**Purpose:** Comprehensive display showing all task metadata

**Configuration:**
```yaml
name: all
fields:
  - status
  - summary
  - description
  - priority
  - due_date
  - start_date
  - created
  - modified
  - completed
  - tags
  - uid
  - parent
```

**When Used:**
- Debugging task data
- Full task inspection
- Export preparation
- Detailed analysis

**How to Use:**
```bash
todoat MyList -v all
```

**Output Includes:**
- All timestamps (creation, modification, completion)
- Full description text
- Parent-child relationships (UIDs)
- All tags/categories
- Complete priority scale

---

## Custom View Creation

### Purpose

Custom views enable users to create specialized display configurations tailored to specific workflows or use cases.

### How It Works

#### Method 1: Interactive Builder (TUI)

**Step-by-Step:**
1. User runs view creation command
2. Terminal UI (TUI) launches with builder interface
3. User selects fields to include
4. User configures field options (width, alignment, format)
5. User adds filters (optional)
6. User sets sorting rules (optional)
7. Builder generates YAML file
8. View is saved and ready to use

**Command:**
```bash
todoat view create myview
```

**Builder Features:**
- Arrow key navigation
- Checkbox field selection
- Inline field configuration
- Filter builder with autocomplete
- Sort rule builder
- Real-time preview (if enabled)
- Validation and error checking

**User Journey:**
```
1. Run: todoat view create work_today
2. TUI displays available fields
3. User selects: status, summary, due_date, priority
4. User adds filter: due_date = today, status != DONE
5. User sets sort: due_date ASC, priority ASC
6. User saves configuration
7. View saved to ~/.config/todoat/views/work_today.yaml
8. User runs: todoat MyList -v work_today
```

#### Method 2: Manual YAML Creation

**Purpose:** Direct view configuration for advanced users or automation

**Steps:**
1. Create YAML file in `~/.config/todoat/views/`
2. Define view configuration (see [Field Selection](#field-selection-and-ordering))
3. Save file with `.yaml` extension
4. View becomes immediately available

**Example File:** `~/.config/todoat/views/urgent.yaml`
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
    format: "2006-01-02"
  - name: priority
    width: 8
filters:
  - field: priority
    operator: lte
    value: 3
  - field: due_date
    operator: lte
    value: "+7d"
sort:
  - field: due_date
    direction: asc
  - field: priority
    direction: asc
```

### Prerequisites

- Todoat configuration directory must exist (`~/.config/todoat/`)
- Views directory is auto-created if missing
- Valid YAML syntax required for manual creation

### Outputs/Results

- YAML view file saved to `~/.config/todoat/views/`
- View appears in `todoat view list` output
- View can be used with `-v` or `--view` flag
- Validation errors shown if configuration is invalid

---

## Field Selection and Ordering

### Available Fields

| Field Name | Description | Data Type | Example |
|------------|-------------|-----------|---------|
| `status` | Task completion state | String | TODO, DONE, IN-PROCESS |
| `summary` | Task title/name | String | "Buy groceries" |
| `description` | Detailed task information | Text | "Get milk, eggs, bread from store" |
| `priority` | Task importance (0-9) | Integer | 1 (highest) to 9 (lowest) |
| `due_date` | When task is due | DateTime | 2026-01-15T14:00:00Z |
| `start_date` | When to start task | DateTime | 2026-01-10T09:00:00Z |
| `created` | Task creation timestamp | DateTime | 2026-01-01T12:00:00Z |
| `modified` | Last modification timestamp | DateTime | 2026-01-14T16:30:00Z |
| `completed` | Completion timestamp | DateTime | 2026-01-13T11:00:00Z |
| `tags` | Task categories/labels | Array | ["work", "urgent"] |
| `uid` | Unique task identifier | String | "task-abc123" |
| `parent` | Parent task UID (for subtasks) | String | "task-xyz789" |

### Field Configuration Options

Each field supports these configuration properties:

```yaml
fields:
  - name: summary           # Required: field identifier
    width: 40              # Optional: display width in characters
    align: left            # Optional: left|center|right
    format: "custom"       # Optional: format string (date fields)
    truncate: true         # Optional: truncate if exceeds width
    plugin:                # Optional: custom formatter (see Plugin Formatters)
      command: "/path/to/script.sh"
```

### Field Ordering

**Purpose:** Control the left-to-right display sequence of fields

**How It Works:**
- Fields are displayed in the order they appear in the YAML `fields` array
- First field in array appears leftmost in output
- Last field appears rightmost

**Example:**
```yaml
fields:
  - name: status      # Column 1
  - name: priority    # Column 2
  - name: summary     # Column 3
  - name: due_date    # Column 4
```

**Output:**
```
TODO     1 (High)    Buy groceries                    2026-01-15
DONE     5 (Medium)  Review documentation              2026-01-14
```

### Hierarchical Display Support

**Purpose:** Maintain parent-child task relationships in custom views

**How It Works:**
- Views automatically detect subtask relationships via `parent_uid`
- Tree structure rendered with box-drawing characters
- Indentation shows nesting level
- Parent tasks always appear before children (regardless of sort)

**Configuration:**
```yaml
fields:
  - name: summary
    width: 50
hierarchy:
  enabled: true        # Default: true if parent field exists
  indent_size: 2       # Spaces per nesting level
  show_connectors: true  # Box-drawing characters (├─, └─)
```

**Output Example:**
```
Project Alpha
├─ Design phase
│  ├─ Create mockups
│  └─ User testing
└─ Development phase
   ├─ Backend API
   └─ Frontend UI
```

---

## Filtering Tasks

### Purpose

Filters reduce the task set to only those matching specific criteria, enabling focused views for particular workflows or contexts.

### Filter Operators

| Operator | Description | Example | Use Case |
|----------|-------------|---------|----------|
| `eq` | Equals | `status eq TODO` | Exact matches |
| `ne` | Not equals | `status ne DONE` | Exclusion |
| `lt` | Less than | `priority lt 5` | Range filtering |
| `lte` | Less than or equal | `priority lte 3` | Inclusive ranges |
| `gt` | Greater than | `created gt 2026-01-01` | Threshold filtering |
| `gte` | Greater than or equal | `due_date gte today` | Date ranges |
| `contains` | String contains | `summary contains meeting` | Text search |
| `in` | Value in list | `tags in [work, urgent]` | Multi-value matching |
| `not_in` | Value not in list | `status not_in [DONE, CANCELLED]` | Exclusion lists |
| `regex` | Regular expression | `summary regex ^Project` | Pattern matching |

### Filter Configuration

**Basic Filter:**
```yaml
filters:
  - field: status
    operator: eq
    value: TODO
```

**Multiple Filters (AND logic):**
```yaml
filters:
  - field: status
    operator: ne
    value: DONE
  - field: priority
    operator: lte
    value: 3
  - field: tags
    operator: contains
    value: work
```

**All filters must match** for a task to be included (implicit AND).

### Date Filter Special Values

**Relative Dates:**
- `today` - Current date (00:00:00)
- `tomorrow` - Next day
- `yesterday` - Previous day
- `+Nd` - N days from now (e.g., `+7d` = one week ahead)
- `-Nd` - N days ago (e.g., `-3d` = three days ago)
- `+Nw` - N weeks from now (e.g., `+2w` = two weeks ahead)
- `+Nm` - N months from now (e.g., `+1m` = one month ahead)

**Example: Tasks Due This Week**
```yaml
filters:
  - field: due_date
    operator: gte
    value: today
  - field: due_date
    operator: lte
    value: "+7d"
```

### Tag/Category Filtering

**Single Tag:**
```yaml
filters:
  - field: tags
    operator: contains
    value: urgent
```

**Multiple Tags (OR logic):**
```yaml
filters:
  - field: tags
    operator: in
    value: [work, personal, urgent]
```

**Exclude Tags:**
```yaml
filters:
  - field: tags
    operator: not_in
    value: [archive, someday]
```

### Complex Filter Examples

**Active High-Priority Tasks:**
```yaml
filters:
  - field: status
    operator: not_in
    value: [DONE, CANCELLED]
  - field: priority
    operator: lte
    value: 2
```

**Overdue Tasks:**
```yaml
filters:
  - field: due_date
    operator: lt
    value: today
  - field: status
    operator: ne
    value: DONE
```

**Recently Created Work Tasks:**
```yaml
filters:
  - field: created
    operator: gte
    value: "-7d"
  - field: tags
    operator: contains
    value: work
```

### User Journey: Creating Filtered View

1. User wants to see only urgent work tasks due this week
2. User runs: `todoat view create urgent_work`
3. TUI builder opens
4. User selects fields: status, summary, due_date, priority
5. User clicks "Add Filter" button
6. User selects: field=tags, operator=contains, value=work
7. User clicks "Add Filter" again
8. User selects: field=tags, operator=contains, value=urgent
9. User clicks "Add Filter" again
10. User selects: field=due_date, operator=lte, value=+7d
11. User clicks "Add Filter" again
12. User selects: field=status, operator=ne, value=DONE
13. User saves view
14. User runs: `todoat MyList -v urgent_work`
15. Output shows only matching tasks

---

## Sorting Tasks

### Purpose

Sorting controls the order in which filtered tasks appear, enabling prioritized or chronological views.

### Sort Configuration

**Basic Sort:**
```yaml
sort:
  - field: priority
    direction: asc    # asc (ascending) or desc (descending)
```

**Multi-Level Sort:**
```yaml
sort:
  - field: priority
    direction: asc
  - field: due_date
    direction: asc
  - field: summary
    direction: asc
```

**Sort Logic:**
- First sort rule is primary
- Subsequent rules break ties
- Hierarchical tasks: parents always before children (overrides sort)

### Common Sort Patterns

**Priority-First:**
```yaml
sort:
  - field: priority
    direction: asc    # 1 (high) to 9 (low)
  - field: summary
    direction: asc
```

**Deadline-Driven:**
```yaml
sort:
  - field: due_date
    direction: asc    # Soonest first
  - field: priority
    direction: asc
```

**Chronological Creation:**
```yaml
sort:
  - field: created
    direction: desc   # Newest first
```

**Alphabetical:**
```yaml
sort:
  - field: summary
    direction: asc
```

**Status Then Priority:**
```yaml
sort:
  - field: status
    direction: asc    # CANCELLED, DONE, IN-PROCESS, TODO (alphabetical)
  - field: priority
    direction: asc
```

### Sortable Fields

All fields in [Field Selection](#available-fields) are sortable:
- **String fields**: Alphabetical (case-insensitive)
- **Integer fields**: Numerical
- **DateTime fields**: Chronological
- **Array fields** (tags): Lexicographic comparison of joined strings
- **Null values**: Always sorted last (regardless of direction)

### Hierarchical Sort Interaction

**Important:** Parent-child relationships always override sort rules.

**Example:**
```yaml
sort:
  - field: priority
    direction: asc
```

**With Tasks:**
- Task A (priority 1, no parent)
- Task B (priority 3, parent of C)
- Task C (priority 2, child of B)

**Output Order:**
1. Task A (priority 1)
2. Task B (priority 3) - parent displayed first
3. Task C (priority 2) - child displayed after parent

Despite C having higher priority than B, it appears after B because hierarchy takes precedence.

---

## Plugin Formatters

### Purpose

Plugin formatters allow external scripts (in any language) to customize how individual fields are displayed, enabling unlimited formatting possibilities beyond built-in formatters.

### How It Works

**Data Flow:**
1. View renderer reaches a field with plugin configuration
2. Renderer serializes task data to JSON
3. JSON sent to plugin script via stdin
4. Script processes data and outputs formatted string to stdout
5. Renderer captures output and displays in field column
6. Timeout enforced (default 5 seconds, configurable per plugin)

**Architecture:**
```
Task Data → JSON Serialization → stdin → Plugin Script → stdout → Field Display
                                           ↓
                                    (timeout enforced)
```

### Plugin Configuration

**In View YAML:**
```yaml
fields:
  - name: status
    plugin:
      command: "/path/to/status-emoji.sh"
      args: ["--style", "emoji"]
      timeout: 1000        # milliseconds (default: 5000)
      env:
        CUSTOM_VAR: "value"
```

**Field Options:**
- `command` (required): Absolute path to executable script
- `args` (optional): Command-line arguments passed to script
- `timeout` (optional): Max execution time in milliseconds
- `env` (optional): Environment variables set for script

### JSON Input Format

**Structure:**
```json
{
  "uid": "task-123",
  "summary": "Buy groceries",
  "description": "Get milk, eggs, bread",
  "status": "TODO",
  "priority": 1,
  "due_date": "2026-01-15T14:00:00Z",
  "start_date": null,
  "created": "2026-01-10T09:00:00Z",
  "modified": "2026-01-14T11:30:00Z",
  "completed": null,
  "categories": ["shopping", "urgent"],
  "parent_uid": ""
}
```

**All fields included** in JSON (even if null) for complete context.

### Plugin Script Requirements

**Interface Contract:**
1. Read JSON from stdin
2. Parse JSON (any parser/library)
3. Format field value (custom logic)
4. Write single-line output to stdout
5. Exit with code 0 for success
6. Exit with non-zero for error (falls back to default formatter)

**Language Agnostic:** Works with bash, python, ruby, go, node.js, etc.

### Example Plugins

#### Bash: Status Emoji Formatter

**File:** `status-emoji.sh`
```bash
#!/bin/bash

# Read JSON from stdin
json=$(cat)

# Extract status field
status=$(echo "$json" | jq -r '.status')

# Map status to emoji
case $status in
  "TODO")
    echo "⏳ To Do"
    ;;
  "IN-PROCESS")
    echo "🔄 In Progress"
    ;;
  "DONE")
    echo "✅ Done"
    ;;
  "CANCELLED")
    echo "❌ Cancelled"
    ;;
  *)
    echo "❓ Unknown"
    ;;
esac
```

**Usage:**
```yaml
fields:
  - name: status
    plugin:
      command: "~/.config/todoat/plugins/status-emoji.sh"
```

**Output:**
```
⏳ To Do       Buy groceries
✅ Done        Review documentation
🔄 In Progress Write tests
```

#### Python: Priority Color Formatter

**File:** `priority-color.py`
```python
#!/usr/bin/env python3
import json
import sys

# Read JSON from stdin
task = json.load(sys.stdin)

priority = task.get('priority', 9)

# Color-coded priority with terminal escape codes
if priority <= 2:
    print(f"\033[91m{priority} (Urgent)\033[0m")  # Red
elif priority <= 5:
    print(f"\033[93m{priority} (High)\033[0m")    # Yellow
elif priority <= 7:
    print(f"\033[92m{priority} (Medium)\033[0m")  # Green
else:
    print(f"\033[90m{priority} (Low)\033[0m")     # Gray
```

**Usage:**
```yaml
fields:
  - name: priority
    plugin:
      command: "~/.config/todoat/plugins/priority-color.py"
```

#### Ruby: Relative Date Formatter

**File:** `relative-date.rb`
```ruby
#!/usr/bin/env ruby
require 'json'
require 'time'

# Read JSON from stdin
task = JSON.parse(STDIN.read)

due_date_str = task['due_date']
exit 0 if due_date_str.nil? || due_date_str.empty?

due_date = Time.parse(due_date_str)
now = Time.now

days_diff = ((due_date - now) / 86400).round

# Format as relative date
if days_diff < 0
  puts "#{days_diff.abs} days overdue"
elsif days_diff == 0
  puts "Due today"
elsif days_diff == 1
  puts "Due tomorrow"
elsif days_diff <= 7
  puts "Due in #{days_diff} days"
else
  puts due_date.strftime("%Y-%m-%d")
end
```

**Usage:**
```yaml
fields:
  - name: due_date
    plugin:
      command: "~/.config/todoat/plugins/relative-date.rb"
```

**Output:**
```
Buy groceries          Due today
Review documentation   Due in 3 days
File taxes            Due in 14 days
```

### Error Handling

**Timeout:**
- Plugin exceeds configured timeout
- Error message displayed: `[timeout]`
- Falls back to default formatter for field

**Execution Error:**
- Non-zero exit code
- stderr captured and logged (debug mode)
- Falls back to default formatter

**Invalid Output:**
- Empty stdout
- Multi-line output (only first line used)
- Falls back to default formatter

**Security:**
- Plugins run with user's permissions
- No sandboxing (user responsible for plugin security)
- Command injection prevented (args passed as array, not shell string)

### Plugin Discovery

**Location:** Any filesystem path (user specifies absolute path in view YAML)

**Recommended Structure:**
```
~/.config/todoat/plugins/
├── status-emoji.sh
├── priority-color.py
├── relative-date.rb
└── tag-badges.js
```

**Permissions:** Plugins must be executable (`chmod +x plugin.sh`)

### Performance Considerations

- **Timeout**: Keep timeouts short (1-2 seconds) for responsive output
- **Caching**: Plugins called once per task per field (no automatic caching)
- **Parallelization**: Not currently parallelized (sequential execution)
- **Large Lists**: Many tasks × many plugin fields = slower rendering

**Best Practice:** Use plugins sparingly for large task lists, or add filters to reduce task count.

---

## Interactive View Builder

### Purpose

The Terminal UI (TUI) builder provides a user-friendly interface for creating custom views without manually writing YAML.

### How It Works

**Launch:**
```bash
todoat view create myview
```

**Interface Components:**

1. **Field Selection Panel**
   - Checkbox list of all available fields
   - Arrow keys navigate
   - Space bar toggles selection
   - Selected fields highlighted

2. **Field Configuration Panel**
   - Shows selected fields in order
   - Enter key opens field options
   - Configure width, alignment, format, plugin
   - Up/Down arrows reorder fields

3. **Filter Builder Panel**
   - "Add Filter" button
   - Dropdown for field selection
   - Dropdown for operator selection
   - Input for value (with validation)
   - "Remove" button for each filter
   - Multiple filters supported

4. **Sort Builder Panel**
   - "Add Sort Rule" button
   - Dropdown for field selection
   - Toggle for direction (asc/desc)
   - Up/Down arrows change priority
   - "Remove" button for each rule

5. **Preview Panel (optional)**
   - Shows sample output using current configuration
   - Updates in real-time as changes made
   - Requires active task list (prompts for selection)

6. **Action Buttons**
   - "Save" - Writes YAML file
   - "Cancel" - Discards changes
   - "Reset" - Clears all configuration

### User Journey

**Creating "Due This Week" View:**

1. Run: `todoat view create due_this_week`
2. TUI displays field selection panel
3. User navigates with arrow keys to "status" field
4. User presses Space to select
5. User navigates to "summary" field → Space
6. User navigates to "due_date" field → Space
7. User navigates to "priority" field → Space
8. User presses Tab to move to field configuration panel
9. User presses Enter on "due_date" field
10. User sets format to "2006-01-02" (date format)
11. User sets width to 12
12. User presses Tab to move to filter builder
13. User clicks "Add Filter" (or presses 'a')
14. User selects field: due_date
15. User selects operator: gte
16. User enters value: today
17. User clicks "Add Filter" again
18. User selects field: due_date, operator: lte, value: +7d
19. User clicks "Add Filter" again
20. User selects field: status, operator: ne, value: DONE
21. User presses Tab to move to sort builder
22. User clicks "Add Sort Rule"
23. User selects field: due_date, direction: asc
24. User clicks "Add Sort Rule"
25. User selects field: priority, direction: asc
26. User presses Tab to preview panel (optional)
27. User selects task list to preview
28. Preview shows sample output
29. User presses Tab to action buttons
30. User clicks "Save"
31. Success message: "View saved to ~/.config/todoat/views/due_this_week.yaml"
32. TUI exits
33. User runs: `todoat MyList -v due_this_week`

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| Arrow Keys | Navigate elements |
| Tab | Next panel |
| Shift+Tab | Previous panel |
| Space | Toggle checkbox |
| Enter | Open/confirm |
| Esc | Cancel/close |
| a | Add filter/sort (in builder panels) |
| d | Delete filter/sort (in builder panels) |
| Ctrl+S | Quick save |
| Ctrl+C | Cancel and exit |

### Validation

**Real-time Checks:**
- Field names must be valid (from available fields list)
- Operators must match field type (e.g., `lte` only for numbers/dates)
- Date values validated (relative dates parsed)
- Width values must be positive integers
- At least one field must be selected

**Error Display:**
- Invalid values highlighted in red
- Error message shown at bottom of screen
- Cannot save until all errors resolved

### Prerequisites

- Todoat installed and configured
- Terminal with minimum 80x24 character size
- Arrow key and color support (most modern terminals)

### Outputs/Results

- YAML view file created in `~/.config/todoat/views/`
- View immediately available for use
- Success message with file path shown

---

## View Storage and Management

### Storage Location

**Path:** `~/.config/todoat/views/`

**File Format:** YAML files with `.yaml` extension

**Naming:** View name matches filename (e.g., `urgent.yaml` → view name is "urgent")

### Views Folder Setup

On first use, if the `views/` folder does not exist, todoat will prompt you to create it:

```
Views folder not found. Create with default views? [Y/n]
```

**What happens when you confirm:**
1. Creates `~/.config/todoat/views/` directory
2. Copies `default.yaml` with the built-in default view configuration
3. Copies `all.yaml` with the built-in all view configuration

**With `-y` flag:** The folder is created silently without prompting.

**If folder exists (even if empty):** No prompt is shown. The application uses built-in views as fallback.

### Overriding Built-in Views

You can customize the built-in `default` and `all` views by creating your own versions:

**To customize the default view:**
```bash
# Create or edit ~/.config/todoat/views/default.yaml
nano ~/.config/todoat/views/default.yaml
```

**Example: Default view that shows completed tasks:**
```yaml
name: default
description: My custom default view (shows all tasks including completed)
fields:
  - name: status
    width: 12
  - name: summary
    width: 50
  - name: priority
    width: 10
# No filters - shows all tasks including DONE
```

**View Loading Priority:**
1. Check `~/.config/todoat/views/{name}.yaml` first
2. If not found, fall back to built-in view

This means your custom `default.yaml` or `all.yaml` will override the built-in versions.

**To restore built-in behavior:**
```bash
# Delete your custom override
rm ~/.config/todoat/views/default.yaml
# Built-in default view will be used again
```

### View Listing

**Command:**
```bash
todoat view list
```

**Output:**
```
Available views:
  - default (built-in)
  - all (built-in)
  - urgent (custom)
  - work_today (custom)
  - due_this_week (custom)
```

**How It Works:**
1. Scans `~/.config/todoat/views/` directory
2. Parses each `.yaml` file
3. Validates view configuration
4. Lists valid views with type (built-in vs. custom)
5. Invalid views shown with error message

### View Deletion

**Manual Deletion:**
```bash
rm ~/.config/todoat/views/urgent.yaml
```

**Effect:** View no longer appears in `view list`, cannot be used with `-v` flag

**Note:** Built-in views (default, all) are hard-coded in the application and always available as fallback. However, you can override them by creating your own `default.yaml` or `all.yaml` files (see [Overriding Built-in Views](#overriding-built-in-views)).

### View Editing

**Manual Editing:**
```bash
nano ~/.config/todoat/views/urgent.yaml
```

**Changes take effect immediately** (no restart required).

**Validation on Use:**
- Invalid YAML causes error when view is used
- Error message shows which view and what's invalid
- Falls back to default view if current view is invalid

### View Sharing

**Export:**
1. Copy view file from `~/.config/todoat/views/`
2. Share file with team (email, git repo, shared drive)
3. Document any required plugins (if view uses custom formatters)

**Import:**
1. Receive view YAML file
2. Copy to `~/.config/todoat/views/` directory
3. Install any required plugins (if applicable)
4. Run `todoat view list` to verify
5. Use with `todoat MyList -v imported_view`

**Team Workflow:**
```
team-views/
├── README.md           (describes each view)
├── sprint-board.yaml   (sprint planning view)
├── daily-standup.yaml  (daily standup view)
└── plugins/            (shared custom formatters)
    └── sprint-status.sh
```

---

## Technical Architecture

### Key Modules

**View Types** (`internal/views/types.go`):
- `View` struct: Main configuration container
- `Field` struct: Field configuration (name, width, alignment, plugin)
- `Filter` struct: Filter rules (field, operator, value)
- `SortRule` struct: Sorting configuration (field, direction)
- YAML unmarshaling tags for configuration loading

**View Renderer** (`internal/views/renderer.go`):
- `Renderer` interface: Pluggable rendering system
- `DefaultRenderer`: Standard table-based output
- `RenderView(tasks, view)`: Main rendering function
- Field formatters: Built-in formatters for each field type
- Plugin executor: Calls external formatter scripts
- Width calculation: Dynamic column sizing
- Hierarchy renderer: Tree structure with box-drawing

**Filter Engine** (`internal/views/filter.go`):
- `FilterTask(task, filters)`: Applies all filters to single task
- Operator implementations: eq, ne, lt, lte, gt, gte, contains, in, not_in, regex
- Type-aware filtering: String, int, date, array handling
- Relative date parser: Converts "today", "+7d", etc. to absolute dates

**Sort Engine** (`internal/views/filter.go`):
- `SortTasks(tasks, rules)`: Multi-level sorting
- Hierarchy preservation: Parents always before children
- Type-aware comparison: String, int, date, array
- Null handling: Always sorted last
- Stable sort: Maintains original order for equal elements

**Plugin Formatter** (`internal/views/formatters/plugin.go`):
- `PluginFormatter` struct: Manages plugin execution
- `Format(task, config)`: Main plugin execution
- JSON serialization: Task to JSON
- Process management: Spawns plugin process, pipes stdin/stdout
- Timeout enforcement: Context-based cancellation
- Error handling: Captures stderr, handles non-zero exits
- Fallback: Returns default format on error

**View Builder** (`internal/views/builder/`):
- TUI components: Field selector, filter builder, sort builder
- Event handlers: Keyboard and mouse input
- Validation: Real-time configuration checking
- YAML generator: Converts TUI state to view YAML

### Data Flow

**Using a View:**
```
1. CLI parses -v flag
2. Load view YAML from ~/.config/todoat/views/
3. Parse YAML into View struct
4. Fetch tasks from backend (see Task Management)
5. Apply filters: FilterTask() on each task
6. Apply sort: SortTasks() on filtered set
7. Render view: RenderView() with sorted tasks
   ├─ For each task:
   │  ├─ For each field:
   │  │  ├─ Check if plugin configured
   │  │  ├─ If plugin: execute plugin, use output
   │  │  └─ If no plugin: use built-in formatter
   │  └─ Assemble row
   └─ Output table
8. Display to user
```

**Creating a View (Interactive):**
```
1. CLI invokes view create command
2. Launch TUI builder
3. User interacts with builder (select fields, add filters, etc.)
4. On save:
   ├─ Validate configuration
   ├─ Convert TUI state to View struct
   ├─ Marshal to YAML
   ├─ Write to ~/.config/todoat/views/[name].yaml
   └─ Show success message
5. Exit TUI
```

### Integration Points

- **[Task Management](task-management.md)**: Views display tasks fetched via backend TaskManager interface
- **[Backend System](backend-system.md)**: Views work with tasks from any backend (Nextcloud, SQLite, File)
- **[Subtasks Hierarchy](subtasks-hierarchy.md)**: Views render hierarchical relationships with tree display
- **[CLI Interface](cli-interface.md)**: Views invoked via `-v` or `--view` flag on task display commands

### Performance Characteristics

**View Loading:**
- YAML parsing: < 1ms per view (cached in memory)
- View validation: < 1ms

**Filtering:**
- Linear scan: O(n) where n = total task count
- Typical: 1000 tasks filtered in < 10ms
- Regex filters: Slower (compile pattern per filter)

**Sorting:**
- Multi-level sort: O(n log n)
- Typical: 1000 tasks sorted in < 20ms
- Hierarchical sort: Additional O(n) pass for parent-child grouping

**Rendering:**
- Table rendering: O(n × f) where n = tasks, f = fields
- Plugin execution: Adds latency per plugin per task per field
  - Example: 100 tasks × 1 plugin field @ 100ms = 10 seconds
  - **Recommendation**: Use plugins sparingly or reduce task count with filters

**Overall Performance:**
- Views without plugins: 1000 tasks render in < 100ms
- Views with plugins: Depends on plugin execution time
- Large task sets (>5000): Use pagination with `--limit`, `--offset`, `--page`, and `--page-size` flags

**Pagination:**
For large task sets, pagination prevents slow rendering and memory issues:
```bash
# Show first 50 tasks (default page size)
todoat MyList --limit 50

# Show page 2 with 50 tasks per page
todoat MyList --page 2

# Custom page size
todoat MyList --page 1 --page-size 100
```

---

## Related Features

- **[Task Management](task-management.md)**: Views display task data fetched via CRUD operations
- **[Subtasks Hierarchy](subtasks-hierarchy.md)**: Views maintain parent-child relationships in display
- **[CLI Interface](cli-interface.md)**: Views invoked via command-line flags and arguments
- **[Backend System](backend-system.md)**: Views work seamlessly with any task backend
- **[List Management](list-management.md)**: Views can be scoped to specific task lists

---

## Summary

The Views Customization system provides a powerful, flexible way to control task display:
- **Built-in views** for common use cases (default, all)
- **Custom views** created via interactive TUI builder or manual YAML
- **Field selection and ordering** for precise output control
- **Filtering** to reduce task set to relevant items
- **Sorting** for prioritized or chronological display
- **Plugin formatters** for unlimited customization with external scripts
- **YAML storage** for portability and version control
- **Hierarchical display** preserving parent-child relationships

This system enables users to create workflows tailored to their specific needs, whether for personal task management, team collaboration, or specialized reporting.
