# [020] Shell Prompt Integration

## Summary
Provide a command or output format that enables shell prompts (bash, zsh, fish) to display pending task count or next due task, encouraging regular task review.

## Source
Code analysis: The app has excellent CLI integration and shell completion support but no way to embed task information into the user's shell prompt. Productivity workflows benefit from ambient task awareness.

## Motivation
Terminal-centric users spend significant time in their shell. Having pending task counts or upcoming deadlines visible in the prompt creates gentle, persistent reminders without requiring explicit task list checks. This is especially valuable for users who might otherwise forget to review their task lists throughout the day.

## Current Behavior
```bash
# Users must explicitly check tasks
$ todoat Work
# No ambient awareness of pending tasks

# Shell prompt shows no task context
user@host:~/project$
```

## Proposed Behavior
```bash
# New command for shell-friendly output
todoat prompt
# Output: 5 (pending task count, suitable for prompt embedding)

todoat prompt --format "{{.PendingCount}}/{{.OverdueCount}}"
# Output: 5/2

todoat prompt --next-due
# Output: "Submit report" (2h)

# Example shell prompt integration (user's .bashrc/.zshrc)
PS1='[$(todoat prompt 2>/dev/null || echo 0)] \u@\h:\w\$ '
# Result: [5] user@host:~/project$

# Fish shell integration
function fish_prompt
    set tasks (todoat prompt 2>/dev/null; or echo 0)
    echo "[$tasks] "(prompt_pwd)"> "
end
```

The command should:
- Be extremely fast (<50ms) by reading from local cache only
- Fail silently (exit 0, empty output) if backend unavailable
- Support customizable format templates
- Cache aggressively to avoid slowing down prompt rendering

## Estimated Value
medium - Creates ambient task awareness for terminal users; low effort integration with high daily visibility

## Estimated Effort
S - Simple command that reads cached task counts; no complex logic required

## Related
- Shell completion: `cmd/todoat/cmd/todoat.go:12186` (existing shell integration)
- Cache system: `internal/cache/` (for fast prompt response)
- Daily review idea: `docs/ideas/unreviewed/011-daily-review-report.md` (complementary ambient awareness)

## Status
unreviewed
