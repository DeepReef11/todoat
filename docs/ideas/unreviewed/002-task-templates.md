# [002] Task Templates

## Summary
Allow users to define reusable task templates with pre-configured properties, enabling quick creation of common task types.

## Source
Code analysis: No template system exists. Users creating similar tasks repeatedly (e.g., "weekly review", "standup notes") must specify all properties each time.

## Motivation
Many tasks follow predictable patterns - weekly reviews, sprint planning, recurring meeting notes. Templates would let users define these patterns once and instantiate them quickly, saving time and ensuring consistency.

## Current Behavior
```bash
# Creating a weekly review task requires full specification each time
todoat Work add "Weekly Review" --priority 2 --tag "review" --recur weekly --description "Review accomplishments and plan next week"
```

## Proposed Behavior
```bash
# Define a template (stored in config or separate templates file)
todoat template create weekly-review \
  --summary "Weekly Review" \
  --priority 2 \
  --tag review \
  --recur weekly \
  --description "Review accomplishments and plan next week"

# Use template to create task
todoat Work add --template weekly-review

# Override template values as needed
todoat Work add --template weekly-review --summary "Q1 Review" --priority 1

# List available templates
todoat template list
```

## Estimated Value
medium - Saves time for users with repeating task patterns, ensures consistency

## Estimated Effort
S - Straightforward YAML storage, minimal new commands, reuses existing task creation logic

## Open Questions
- Store templates in config.yaml or separate templates.yaml?
- Should templates support variables (e.g., {{date}}, {{week}})?
- Should templates be shareable across backends?
- Template inheritance (e.g., base-meeting -> standup-meeting)?

## Related
- Internal config system at internal/config/config.go
- Task creation in cmd/todoat/cmd/todoat.go

## Status
unreviewed
