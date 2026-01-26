# [013] Collaborative Task Sharing (Local Export)

## Summary
Enable sharing individual tasks or filtered task sets with others via shareable export formats (markdown, JSON, plain text) suitable for email, chat, or collaboration tools.

## Source
Code analysis: List export exists (`todoat list export`) but is oriented toward backup/migration. No quick way to share a subset of tasks in formats suitable for human communication or collaboration.

## Motivation
Users often need to share task information:
- "Here are my action items from the meeting"
- "These are the blockers for the release"
- "Status update for the project"

Exporting to human-readable formats for pasting into Slack, email, or docs enables collaboration without requiring others to use todoat.

## Current Behavior
```bash
# Full list export (backup-oriented)
todoat list export Work --format json > work.json

# No way to export filtered subset in readable format
# Must manually copy task info
```

## Proposed Behavior
```bash
# Export filtered tasks as markdown
todoat Work share --filter-tag "release-blockers"
# Copies to clipboard (or outputs):
#
# ## Release Blockers
# - [ ] Fix login bug (P1, due: Jan 28)
# - [ ] Update API docs (P2, due: Jan 30)
# - [ ] Security review (P1, due: Jan 29)

# Export as plain text checklist
todoat Work share --filter-tag "meeting" --format text
# Output:
# Action Items:
# [ ] Schedule follow-up call
# [ ] Send proposal draft
# [ ] Review budget

# Export with more detail
todoat Work share --filter-status "IN-PROGRESS" --format markdown --verbose
# Includes descriptions, dates, notes

# Interactive share from TUI
# Select tasks with 'm' (mark), then 's' to share

# Quick share single task
todoat Work share "Fix login bug" --format slack
# Output: formatted for Slack (with emoji, formatting)

# Copy to clipboard directly
todoat Work share --filter-priority 1 --clipboard
```

## Estimated Value
low - Useful for collaboration scenarios, but limited use for solo users

## Estimated Effort
S - Primarily formatting templates, output to clipboard integration

## Open Questions
- Output formats needed (markdown, plain, Slack, Jira, etc.)?
- Clipboard integration cross-platform (pbcopy, xclip, etc.)?
- Include task links if using remote backend?
- Two-way sharing (import from shared format)?
- Should this be separate command or extension of export?

## Related
- List export: `todoat list export`
- JSON output flag: `--json`

## Status
unreviewed
