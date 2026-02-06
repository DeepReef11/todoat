# [022] Quick-Add Aliases (Saved Commands)

## Summary
Allow users to define named aliases for frequently-used add commands with pre-configured flags, reducing repetitive typing for common task patterns.

## Source
Code analysis: Users adding similar tasks repeatedly (daily standups, weekly reviews, sprint tasks) must type the same flags each time. Unlike idea 002 (templates which pre-fill task fields), this is about saving command patterns.

## Motivation
Power users develop muscle memory for specific task patterns:
- `todoat Work add "standup" --recur daily --tag meeting`
- `todoat Personal add "..." --due-date saturday --tag weekend`
- `todoat Inbox add "..." --priority 2`

Typing these flags repeatedly is tedious and error-prone. Aliases would let users save command patterns and invoke them with short names.

## Current Behavior
```bash
# User must type full command every time
todoat Work add "Monday standup" --recur weekly --tag meeting --due-date monday
todoat Work add "Tuesday standup" --recur weekly --tag meeting --due-date tuesday
# Repetitive flag typing
```

## Proposed Behavior
```bash
# Define an alias (saved to config)
todoat alias add standup "Work add --recur daily --tag meeting"

# Use the alias
todoat standup "Monday standup"
# Expands to: todoat Work add "Monday standup" --recur daily --tag meeting

# List aliases
todoat alias list
# Output:
# standup -> Work add --recur daily --tag meeting
# weekend -> Personal add --due-date saturday --tag weekend
# inbox   -> Inbox add --priority 2

# Delete an alias
todoat alias delete standup

# Aliases can include list name or be list-agnostic
todoat alias add quick "add --priority 2"  # No list - uses current default
todoat Work quick "Review PR"
# Expands to: todoat Work add "Review PR" --priority 2
```

Aliases stored in `~/.config/todoat/aliases.yaml`:
```yaml
aliases:
  standup: "Work add --recur daily --tag meeting"
  weekend: "Personal add --due-date saturday --tag weekend"
```

## Estimated Value
medium - Significant time savings for power users with repetitive task patterns

## Estimated Effort
S - Config file storage, simple string expansion, new command group

## Related
- Idea 002: Task templates (pre-fills task fields, this saves command patterns)
- Config system: `internal/config/` (alias storage)
- CLI structure: `cmd/todoat/cmd/todoat.go` (command expansion)

## Status
unreviewed
