# [001] Natural Language Task Input

## Summary
Add support for parsing natural language task input that automatically extracts due dates, priorities, and tags from the task summary.

## Source
Code analysis: Current `add` command requires explicit flags for each property (--due-date, --priority, --tag). Competitors like Todoist support natural language parsing.

## Motivation
Users frequently want to quickly add tasks without remembering flag syntax. Natural language input reduces friction and matches how people naturally think about tasks. This is especially valuable for terminal users who want speed over explicit configuration.

## Current Behavior
```bash
# Current: explicit flags required
todoat Work add "Submit report" --due-date 2026-02-01 --priority 1 --tag work

# Must know all flags and their formats
```

## Proposed Behavior
```bash
# Natural language parsing
todoat Work add "Submit report tomorrow high priority #work"
# Automatically parses: due=tomorrow, priority=high, tag=work

# Additional examples
todoat Work add "Call mom next friday p2"     # due=next friday, priority=2
todoat Work add "Review PR by end of day"     # due=today EOD
todoat Work add "Weekly standup every monday" # recurring=weekly on monday
```

## Estimated Value
medium - Reduces friction for task entry, matches user expectations from modern task apps

## Estimated Effort
M - Requires date parsing library, pattern matching for priorities/tags, integration with existing add flow

## Open Questions
- Should natural language be the default or opt-in via flag (`--smart` or `--nl`)?
- How to handle ambiguous input (e.g., "review 2 documents" - is 2 a priority?)
- Should it support localized date parsing?
- Which date parsing library to use (e.g., dateparser, natural)?

## Related
- Todoist quick add: https://todoist.com/help/articles/use-natural-language
- Taskwarrior natural language: https://taskwarrior.org/docs/

## Status
unreviewed
