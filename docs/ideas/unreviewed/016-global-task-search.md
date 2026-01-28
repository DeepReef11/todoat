# [016] Global Task Search

## Summary
Add ability to search for tasks across all lists within a backend, without needing to specify which list to search in.

## Source
Documentation analysis: Current CLI requires specifying a list for all task operations (`todoat ListName <action>`). Users who don't remember which list contains a task must manually search each list or use shell scripting to iterate.

## Motivation
When you remember a task exists but not which list it's in, finding it requires:
1. Guessing which list (trial and error)
2. Manually checking multiple lists
3. Writing shell scripts to iterate through lists

A global search command would let users quickly find tasks by summary, tag, or other criteria across all lists.

## Current Behavior
```bash
# Must specify a list - no global search
todoat Work "meeting"          # Only searches Work list
todoat Personal "meeting"      # Must check each list manually

# Workaround requires shell scripting
for list in $(todoat list --json | jq -r '.[].name'); do
  todoat "$list" | grep -i "meeting"
done
```

## Proposed Behavior
```bash
# Search all lists for a task
todoat search "meeting"
# Output:
# Work       Team standup meeting       TODO    1 (High)
# Personal   Dentist meeting            TODO    -
# Projects   Q1 planning meeting        DONE    2 (Medium)

# Search with filters
todoat search "meeting" --status TODO
todoat search --tag urgent
todoat search --priority 1
todoat search --due-before "next week"

# JSON output for scripting
todoat search "meeting" --json
# [{"list": "Work", "summary": "Team standup meeting", ...}]

# Search returns list name with each result
# Enables: todoat Work update "Team standup" ...
```

## Estimated Value
medium - Common use case when managing multiple lists, reduces friction in finding tasks

## Estimated Effort
S - Iterates through existing GetTasks per list, filters results, displays with list column. No new backend methods needed.

## Related
- Idea #009 (Multi-Backend Views) - cross-backend aggregation (different scope: backends vs lists)
- Task search: already exists per-list in `cmd/todoat/cmd/todoat.go`
- List iteration: `GetLists()` already available

## Status
unreviewed
