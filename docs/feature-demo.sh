#!/bin/bash
#
# todoat Feature Demo Script
#
# This script demonstrates all available todoat features.
# Run each section to see the feature in action.
#
# Prerequisites:
#   - todoat binary built and in PATH
#   - SQLite backend configured (default)
#
# Usage:
#   ./feature-demo.sh          # Run all demos
#   ./feature-demo.sh section  # Run specific section (e.g., ./feature-demo.sh tasks)
#
# Available sections:
#   version, help, config, lists, tasks, subtasks, dates, recurring,
#   tags, priority, views, filters, json, sync, tui, cleanup
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Demo list name
DEMO_LIST="DemoList"

print_header() {
    echo ""
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}============================================${NC}"
    echo ""
}

print_cmd() {
    echo -e "${YELLOW}> $1${NC}"
    eval "$1"
    echo ""
}

print_info() {
    echo -e "${GREEN}$1${NC}"
}

wait_for_enter() {
    echo -e "${YELLOW}Press Enter to continue...${NC}"
    read -r
}

# ============================================
# VERSION & HELP
# ============================================
demo_version() {
    print_header "VERSION & HELP"

    print_info "Show version information:"
    print_cmd "todoat version"

    print_info "Show general help:"
    print_cmd "todoat --help"

    print_info "Show help for specific commands:"
    print_cmd "todoat list --help"
}

# ============================================
# CONFIGURATION
# ============================================
demo_config() {
    print_header "CONFIGURATION"

    print_info "Show config file path:"
    print_cmd "todoat config path"

    print_info "Show all configuration:"
    print_cmd "todoat config get"

    print_info "Show specific config value:"
    print_cmd "todoat config get default_backend"

    print_info "Show backend auto-detection:"
    print_cmd "todoat --detect-backend"
}

# ============================================
# LIST MANAGEMENT
# ============================================
demo_lists() {
    print_header "LIST MANAGEMENT"

    print_info "View all lists:"
    print_cmd "todoat list"

    print_info "Create a demo list:"
    print_cmd "todoat list create '$DEMO_LIST' --description 'Demo tasks' --color '#00CC66'"

    print_info "View list details:"
    print_cmd "todoat list info '$DEMO_LIST'"

    print_info "Update list properties:"
    print_cmd "todoat list update '$DEMO_LIST' --color '#FF5733'"

    print_info "View all lists again:"
    print_cmd "todoat list"

    print_info "View database statistics:"
    print_cmd "todoat list stats"
}

# ============================================
# BASIC TASK MANAGEMENT
# ============================================
demo_tasks() {
    print_header "BASIC TASK MANAGEMENT"

    print_info "Add a simple task:"
    print_cmd "todoat '$DEMO_LIST' add 'Buy groceries'"

    print_info "Add task with abbreviation (a = add):"
    print_cmd "todoat '$DEMO_LIST' a 'Review code'"

    print_info "View tasks in list:"
    print_cmd "todoat '$DEMO_LIST'"

    print_info "View tasks with explicit get command:"
    print_cmd "todoat '$DEMO_LIST' get"

    print_info "Update task status to IN-PROGRESS:"
    print_cmd "todoat '$DEMO_LIST' update 'Review code' -s IN-PROGRESS"

    print_info "Rename a task:"
    print_cmd "todoat '$DEMO_LIST' update 'Buy groceries' --summary 'Buy groceries and supplies'"

    print_info "View tasks after updates:"
    print_cmd "todoat '$DEMO_LIST'"

    print_info "Complete a task:"
    print_cmd "todoat '$DEMO_LIST' complete 'Review code'"

    print_info "Complete using abbreviation (c = complete):"
    print_cmd "todoat '$DEMO_LIST' c 'groceries'"

    print_info "View all tasks including completed (status filter):"
    print_cmd "todoat '$DEMO_LIST' -s TODO,IN-PROGRESS,DONE"

    print_info "Delete a task:"
    print_cmd "todoat -y '$DEMO_LIST' delete 'groceries'"
}

# ============================================
# SUBTASKS (HIERARCHICAL TASKS)
# ============================================
demo_subtasks() {
    print_header "SUBTASKS (HIERARCHICAL TASKS)"

    print_info "Create parent task:"
    print_cmd "todoat '$DEMO_LIST' add 'Project Alpha'"

    print_info "Create subtask using parent flag:"
    print_cmd "todoat '$DEMO_LIST' add 'Design mockups' -P 'Project Alpha'"

    print_info "Create subtask using path syntax:"
    print_cmd "todoat '$DEMO_LIST' add 'Project Alpha/Write documentation'"

    print_info "Create nested subtask:"
    print_cmd "todoat '$DEMO_LIST' add 'Project Alpha/Design mockups/Create wireframes'"

    print_info "View tasks with hierarchy (all view shows parent info):"
    print_cmd "todoat '$DEMO_LIST' -v all"

    print_info "Create task with literal slash (not hierarchy):"
    print_cmd "todoat '$DEMO_LIST' add -l 'UI/UX Review'"

    print_info "Move a subtask to root level:"
    print_cmd "todoat '$DEMO_LIST' update 'Write documentation' --no-parent"

    print_info "View tasks after hierarchy changes:"
    print_cmd "todoat '$DEMO_LIST'"
}

# ============================================
# DATES AND TIME
# ============================================
demo_dates() {
    print_header "DATES AND TIME"

    print_info "Add task with due date:"
    print_cmd "todoat '$DEMO_LIST' add 'Report deadline' --due-date 2026-02-15"

    print_info "Add task with relative due date (tomorrow):"
    print_cmd "todoat '$DEMO_LIST' add 'Quick review' --due-date tomorrow"

    print_info "Add task due in 7 days:"
    print_cmd "todoat '$DEMO_LIST' add 'Weekly review' --due-date +7d"

    print_info "Add task with start date:"
    print_cmd "todoat '$DEMO_LIST' add 'Sprint work' --start-date 2026-01-20 --due-date 2026-02-03"

    print_info "Add task with specific time:"
    print_cmd "todoat '$DEMO_LIST' add 'Team meeting' --due-date '2026-01-25T14:30'"

    print_info "View tasks with dates:"
    print_cmd "todoat '$DEMO_LIST' -v all"

    print_info "Filter tasks due today or later:"
    print_cmd "todoat '$DEMO_LIST' --due-after today"

    print_info "Filter tasks due within a week:"
    print_cmd "todoat '$DEMO_LIST' --due-after today --due-before +7d"

    print_info "Update task due date:"
    print_cmd "todoat '$DEMO_LIST' update 'Report deadline' --due-date 2026-02-28"

    print_info "Clear due date:"
    print_cmd "todoat '$DEMO_LIST' update 'Quick review' --due-date ''"
}

# ============================================
# RECURRING TASKS
# ============================================
demo_recurring() {
    print_header "RECURRING TASKS"

    print_info "Create daily recurring task:"
    print_cmd "todoat '$DEMO_LIST' add 'Daily standup' --recur daily --due-date today"

    print_info "Create weekly recurring task:"
    print_cmd "todoat '$DEMO_LIST' add 'Weekly report' --recur weekly --due-date +7d"

    print_info "Create task recurring every 3 days:"
    print_cmd "todoat '$DEMO_LIST' add 'Check logs' --recur 'every 3 days' --due-date today"

    print_info "Create task with recurrence from completion date:"
    print_cmd "todoat '$DEMO_LIST' add 'Water plants' --recur 'every 3 days' --recur-from-completion --due-date today"

    print_info "View recurring tasks (marked with indicator):"
    print_cmd "todoat '$DEMO_LIST'"

    print_info "Complete recurring task (creates next occurrence):"
    print_cmd "todoat '$DEMO_LIST' complete 'Daily standup'"

    print_info "View tasks after completing recurring task:"
    print_cmd "todoat '$DEMO_LIST'"

    print_info "Remove recurrence from a task:"
    print_cmd "todoat '$DEMO_LIST' update 'Check logs' --recur none"
}

# ============================================
# TAGS (CATEGORIES)
# ============================================
demo_tags() {
    print_header "TAGS (CATEGORIES)"

    print_info "Add task with single tag:"
    print_cmd "todoat '$DEMO_LIST' add 'Bug fix' --tags 'urgent'"

    print_info "Add task with multiple tags:"
    print_cmd "todoat '$DEMO_LIST' add 'Feature request' --tags 'feature,frontend,v2'"

    print_info "View all tags in use:"
    print_cmd "todoat tags"

    print_info "View tags for specific list:"
    print_cmd "todoat tags --list '$DEMO_LIST'"

    print_info "Add tag to existing task:"
    print_cmd "todoat '$DEMO_LIST' update 'Bug fix' --add-tag 'backend'"

    print_info "Remove tag from task:"
    print_cmd "todoat '$DEMO_LIST' update 'Bug fix' --remove-tag 'urgent'"

    print_info "Filter tasks by tag:"
    print_cmd "todoat '$DEMO_LIST' --tag feature"

    print_info "Replace all tags on a task:"
    print_cmd "todoat '$DEMO_LIST' update 'Feature request' --tags 'redesigned,priority'"

    print_info "Clear all tags from a task:"
    print_cmd "todoat '$DEMO_LIST' update 'Bug fix' --tags ''"
}

# ============================================
# PRIORITY
# ============================================
demo_priority() {
    print_header "PRIORITY"

    print_info "Add high priority task (1 = highest):"
    print_cmd "todoat '$DEMO_LIST' add 'Critical bug' -p 1"

    print_info "Add medium priority task:"
    print_cmd "todoat '$DEMO_LIST' add 'Important update' -p 5"

    print_info "Add low priority task:"
    print_cmd "todoat '$DEMO_LIST' add 'Nice to have' -p 8"

    print_info "View tasks (shows priority):"
    print_cmd "todoat '$DEMO_LIST'"

    print_info "Filter by high priority (1-4):"
    print_cmd "todoat '$DEMO_LIST' -p 1,2,3,4"

    print_info "Filter by priority using named level:"
    print_cmd "todoat '$DEMO_LIST' -p high"

    print_info "Update task priority:"
    print_cmd "todoat '$DEMO_LIST' update 'Nice to have' -p 2"
}

# ============================================
# VIEWS
# ============================================
demo_views() {
    print_header "VIEWS"

    print_info "Use default view:"
    print_cmd "todoat '$DEMO_LIST'"

    print_info "Use 'all' view (shows all fields):"
    print_cmd "todoat '$DEMO_LIST' -v all"

    print_info "List available views:"
    print_cmd "todoat view list"

    print_info "Create a custom view:"
    print_cmd "todoat view create urgent-tasks -y --fields 'status,summary,priority,due_date' --sort 'priority:asc'"

    print_info "Use custom view:"
    print_cmd "todoat '$DEMO_LIST' -v urgent-tasks"
}

# ============================================
# FILTERS
# ============================================
demo_filters() {
    print_header "FILTERING"

    print_info "Filter by status - TODO only:"
    print_cmd "todoat '$DEMO_LIST' -s TODO"

    print_info "Filter by status - multiple statuses:"
    print_cmd "todoat '$DEMO_LIST' -s TODO,IN-PROGRESS"

    print_info "Filter by status using abbreviations:"
    print_cmd "todoat '$DEMO_LIST' -s T,I"

    print_info "Filter by due date range:"
    print_cmd "todoat '$DEMO_LIST' --due-after today --due-before +30d"

    print_info "Filter by creation date:"
    print_cmd "todoat '$DEMO_LIST' --created-after -7d"

    print_info "Combine multiple filters:"
    print_cmd "todoat '$DEMO_LIST' -s TODO -p high --due-after today"
}

# ============================================
# JSON OUTPUT
# ============================================
demo_json() {
    print_header "JSON OUTPUT"

    print_info "List tasks as JSON:"
    print_cmd "todoat --json '$DEMO_LIST'"

    print_info "List all lists as JSON:"
    print_cmd "todoat --json list"

    print_info "Add task with JSON response:"
    print_cmd "todoat --json '$DEMO_LIST' add 'JSON test task'"

    print_info "Get tags as JSON:"
    print_cmd "todoat --json tags"

    print_info "Get config as JSON:"
    print_cmd "todoat --json config get"
}

# ============================================
# SYNC (if enabled)
# ============================================
demo_sync() {
    print_header "SYNCHRONIZATION"

    print_info "Check sync status:"
    print_cmd "todoat sync status" || echo "Sync not configured or not available"

    print_info "View sync queue (pending operations):"
    print_cmd "todoat sync queue" || echo "Sync not configured or not available"

    print_info "View sync conflicts:"
    print_cmd "todoat sync conflicts" || echo "Sync not configured or no conflicts"

    print_info "Note: Full sync commands require a remote backend configured."
    echo "Commands available:"
    echo "  todoat sync           - Full sync with remote"
    echo "  todoat sync daemon start - Start background sync"
    echo "  todoat sync daemon stop  - Stop background sync"
}

# ============================================
# REMINDERS
# ============================================
demo_reminders() {
    print_header "REMINDERS"

    print_info "List upcoming reminders:"
    print_cmd "todoat reminder list" || echo "No reminders configured"

    print_info "Check reminder status:"
    print_cmd "todoat reminder status" || echo "Reminder system status"

    print_info "Note: Reminder commands work with tasks that have due dates."
    echo "Commands available:"
    echo "  todoat reminder check            - Check for due reminders"
    echo "  todoat reminder disable 'Task'   - Disable reminders for task"
    echo "  todoat reminder dismiss 'Task'   - Dismiss current reminder"
}

# ============================================
# NOTIFICATIONS
# ============================================
demo_notifications() {
    print_header "NOTIFICATIONS"

    print_info "Show notification commands help:"
    print_cmd "todoat notification --help"
}

# ============================================
# CREDENTIALS
# ============================================
demo_credentials() {
    print_header "CREDENTIALS MANAGEMENT"

    print_info "List configured credentials:"
    print_cmd "todoat credentials list"

    print_info "Note: Credential management for backends."
    echo "Commands available:"
    echo "  todoat credentials set <backend> <user> --prompt    - Store credential"
    echo "  todoat credentials update <backend> <user> --prompt - Update credential"
    echo "  todoat credentials get <backend> <user>             - Check credential"
    echo "  todoat credentials delete <backend> <user>          - Remove credential"
}

# ============================================
# MIGRATION
# ============================================
demo_migration() {
    print_header "MIGRATION"

    print_info "Show migration commands help:"
    print_cmd "todoat migrate --help"

    print_info "Migration examples:"
    echo "  todoat migrate --from sqlite --to nextcloud           - Migrate all lists"
    echo "  todoat migrate --from sqlite --to nextcloud --list Work - Migrate single list"
    echo "  todoat migrate --from sqlite --to nextcloud --dry-run - Preview migration"
    echo "  todoat migrate --target-info nextcloud --list Work    - Check target contents"
    echo ""
    echo "Supported backends: sqlite, nextcloud, todoist, file"
    echo ""
    echo "Migration preserves:"
    echo "  - Task summary and description"
    echo "  - Priority and status"
    echo "  - Due dates and start dates"
    echo "  - Tags/categories"
    echo "  - Parent-child relationships"
}

# ============================================
# TUI (Interactive Mode)
# ============================================
demo_tui() {
    print_header "TERMINAL USER INTERFACE (TUI)"

    print_info "Note: The TUI is an interactive interface. Launch manually to explore."
    echo ""
    echo "To launch: todoat tui"
    echo ""
    echo "TUI Keyboard shortcuts:"
    echo "  Tab      - Switch between lists and tasks panes"
    echo "  j/k      - Move down/up"
    echo "  a        - Add task"
    echo "  e        - Edit task"
    echo "  c        - Complete/uncomplete task"
    echo "  d        - Delete task"
    echo "  /        - Filter tasks"
    echo "  ?        - Show help"
    echo "  q        - Quit"
    echo ""
    echo "Launch TUI with specific backend:"
    echo "  todoat -b sqlite tui"
}

# ============================================
# EXPORT/IMPORT
# ============================================
demo_export_import() {
    print_header "EXPORT & IMPORT"

    print_info "Export list to JSON:"
    print_cmd "todoat list export '$DEMO_LIST' --format json --output /tmp/demo-export.json"

    print_info "View exported file:"
    print_cmd "cat /tmp/demo-export.json | head -50"

    print_info "Note: Import and other export formats available."
    echo "Commands available:"
    echo "  todoat list export 'List' --format csv --output file.csv"
    echo "  todoat list export 'List' --format ical --output file.ics"
    echo "  todoat list import file.json"
}

# ============================================
# NON-INTERACTIVE / SCRIPTING
# ============================================
demo_scripting() {
    print_header "NON-INTERACTIVE MODE (SCRIPTING)"

    print_info "Delete without confirmation using -y flag:"
    print_cmd "todoat -y '$DEMO_LIST' add 'Temp task for scripting demo'"
    print_cmd "todoat -y '$DEMO_LIST' delete 'Temp task for scripting demo'"

    print_info "Combine JSON output with jq for scripting:"
    print_cmd "todoat --json '$DEMO_LIST' | jq '.tasks | length'" || echo "jq not installed"

    print_info "Direct task selection by UID (example):"
    echo "  todoat MyList complete --uid 'task-uuid-here'"
    echo "  todoat MyList complete --local-id 42"
}

# ============================================
# CLEANUP
# ============================================
demo_cleanup() {
    print_header "CLEANUP"

    print_info "Delete demo list (move to trash):"
    print_cmd "todoat -y list delete '$DEMO_LIST'"

    print_info "View trashed lists:"
    print_cmd "todoat list trash"

    print_info "Restore from trash:"
    print_cmd "todoat list trash restore '$DEMO_LIST'"

    print_info "Delete again and purge permanently:"
    print_cmd "todoat -y list delete '$DEMO_LIST'"
    print_cmd "todoat list trash purge '$DEMO_LIST'"

    print_info "Compact database:"
    print_cmd "todoat list vacuum"

    print_info "Demo cleanup complete!"
}

# ============================================
# MAIN
# ============================================
run_all() {
    demo_version
    demo_config
    demo_lists
    demo_tasks
    demo_subtasks
    demo_dates
    demo_recurring
    demo_tags
    demo_priority
    demo_views
    demo_filters
    demo_json
    demo_sync
    demo_reminders
    demo_notifications
    demo_credentials
    demo_migration
    demo_export_import
    demo_scripting
    demo_tui
    demo_cleanup
}

# Parse arguments
if [ $# -eq 0 ]; then
    run_all
else
    case "$1" in
        version)      demo_version ;;
        help)         demo_version ;;
        config)       demo_config ;;
        lists)        demo_lists ;;
        tasks)        demo_tasks ;;
        subtasks)     demo_subtasks ;;
        dates)        demo_dates ;;
        recurring)    demo_recurring ;;
        tags)         demo_tags ;;
        priority)     demo_priority ;;
        views)        demo_views ;;
        filters)      demo_filters ;;
        json)         demo_json ;;
        sync)         demo_sync ;;
        reminders)    demo_reminders ;;
        notifications) demo_notifications ;;
        credentials)  demo_credentials ;;
        migration)    demo_migration ;;
        export)       demo_export_import ;;
        scripting)    demo_scripting ;;
        tui)          demo_tui ;;
        cleanup)      demo_cleanup ;;
        all)          run_all ;;
        *)
            echo "Unknown section: $1"
            echo "Available sections: version, help, config, lists, tasks, subtasks, dates, recurring,"
            echo "                    tags, priority, views, filters, json, sync, reminders,"
            echo "                    notifications, credentials, migration, export, scripting, tui, cleanup, all"
            exit 1
            ;;
    esac
fi

echo ""
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}  Feature Demo Complete!${NC}"
echo -e "${GREEN}============================================${NC}"
