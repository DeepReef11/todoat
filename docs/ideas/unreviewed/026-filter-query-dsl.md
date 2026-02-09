# [026] Filter Query DSL

## Summary
Add a lightweight query language for expressing complex task filters as a single string, enabling saved filters and advanced task selection.

## Source
Code analysis: Current CLI uses individual flags (`--filter-status`, `--filter-priority`, `--filter-tag`, `--filter-due-before`) which cannot express compound conditions. Users needing "priority=1 AND status=TODO" or "(tag=work OR tag=urgent) AND due < tomorrow" must use multiple commands or shell scripting.

## Motivation
Complex filtering is a common need for task management power users:
- "Show me all high-priority tasks that are due this week and not blocked"
- "Find tasks tagged 'review' that have been in-progress for more than 3 days"
- "Get all tasks in project X that are either done or cancelled"

Current flag-based filtering:
- Supports only AND conditions (multiple flags are ANDed together)
- Cannot express OR, NOT, or parenthesized conditions
- Cannot be saved/reused as named filters
- Verbose for complex queries

## Current Behavior
```bash
# Multiple flags are ANDed together
todoat Work --filter-status TODO --filter-priority 1 --filter-tag urgent
# Gets: status=TODO AND priority=1 AND tag=urgent

# No way to express OR conditions
# Cannot do: "priority 1 OR priority 2"
# Cannot do: "tag=work OR tag=home"

# Must use shell to combine queries
todoat Work --filter-priority 1 --json > /tmp/p1.json
todoat Work --filter-priority 2 --json > /tmp/p2.json
jq -s '.[0] + .[1]' /tmp/p1.json /tmp/p2.json
```

## Proposed Behavior
```bash
# Query DSL as a single argument
todoat Work --query 'priority=1 AND (status=TODO OR status=IN-PROGRESS)'
todoat Work -q 'tag=urgent OR due < "tomorrow"'
todoat Work -q 'NOT status=DONE AND created > "7 days ago"'

# Operators: =, !=, <, >, <=, >=, AND, OR, NOT, ()
# Fields: status, priority, tag, due, created, updated, summary (contains)

# Date expressions
todoat Work -q 'due < "next monday"'
todoat Work -q 'created > "2026-01-01"'
todoat Work -q 'updated >= "3 days ago"'

# String matching
todoat Work -q 'summary ~ "meeting"'  # contains
todoat Work -q 'description ~ "urgent"'

# Save queries as named filters in config
# config.yaml:
#   filters:
#     urgent-open: 'priority <= 2 AND status != DONE'
#     this-week: 'due >= "today" AND due <= "next sunday"'

todoat Work --filter urgent-open
todoat Work -f this-week

# Combine named filters
todoat Work -f urgent-open -f this-week
# Equivalent to: (urgent-open) AND (this-week)

# Use in views
todoat view create urgent-view --query 'priority <= 2 AND status = TODO'
```

## Estimated Value
high - Enables powerful, composable filtering that can be saved and reused; a foundational feature for advanced task management

## Estimated Effort
M - Requires parser for query language, integration with existing filter infrastructure, config extension for saved filters

## Related
- Idea #016 (Global Search) - searches across lists, could use query DSL internally
- Idea #009 (Multi-Backend Views) - aggregated views, could use query DSL for filtering
- Current filter flags: `cmd/todoat/cmd/todoat.go` filter implementation
- View system: `internal/views/` already has filter concepts

## Status
unreviewed
