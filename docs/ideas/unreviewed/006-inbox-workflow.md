# [006] Inbox Workflow (Quick Capture)

## Summary
Add a dedicated inbox/quick capture workflow for rapidly adding tasks without specifying a list, then processing them later into appropriate lists.

## Source
Code analysis: Current `add` command requires specifying a list. The Getting Things Done (GTD) methodology emphasizes quick capture to an inbox, then later processing. This is a common workflow pattern in task apps.

## Motivation
When ideas or tasks come up, users want to capture them instantly without the friction of choosing a list or adding metadata. An inbox provides a trusted capture location for later processing, ensuring nothing gets lost.

## Current Behavior
```bash
# Must specify list when adding
todoat Work add "Quick idea"

# If unsure which list, user must choose or tasks get put in wrong places
```

## Proposed Behavior
```bash
# Quick capture to inbox (no list required)
todoat inbox add "Call back John"
todoat inbox "Schedule dentist"  # Short form
todoat i "Research vacation spots"  # Even shorter

# Review inbox
todoat inbox
# Output:
# Inbox (3 tasks):
#   1. [ ] Call back John
#   2. [ ] Schedule dentist
#   3. [ ] Research vacation spots

# Process inbox - move to list
todoat inbox process "Call back John" --to Work
# or interactive processing
todoat inbox process
# Prompts for each task: move to list, add metadata, or delete

# Bulk move
todoat inbox move-all --to Personal --filter-tag "personal"
```

## Estimated Value
medium - Matches GTD workflow, reduces friction for capture, popular in competitors (Todoist, Things)

## Estimated Effort
S - Inbox is just a special list, processing commands are variations of existing move/update

## Open Questions
- Inbox as special list or virtual collection?
- Per-backend inbox or global across backends?
- Should inbox tasks sync to remote backends (e.g., special Nextcloud calendar)?
- Auto-suggest list based on keywords/tags?
- Scheduled inbox review reminders?

## Related
- List management: docs/explanation/list-management.md
- Quick add patterns in Todoist, Things app

## Status
unreviewed
