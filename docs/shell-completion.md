# Shell Completion

todoat supports tab completion for Bash, Zsh, Fish, and PowerShell shells.

## Quick Setup

### Zsh

Add to `~/.zshrc`:

```bash
eval "$(todoat completion zsh)"
```

Or save to file:

```bash
todoat completion zsh > "${fpath[1]}/_todoat"
```

### Bash

Add to `~/.bashrc`:

```bash
eval "$(todoat completion bash)"
```

Or save to file:

```bash
todoat completion bash > /etc/bash_completion.d/todoat
```

### Fish

Add to `~/.config/fish/config.fish`:

```fish
todoat completion fish | source
```

Or save to file:

```bash
todoat completion fish > ~/.config/fish/completions/todoat.fish
```

### PowerShell

Add to your PowerShell profile:

```powershell
todoat completion powershell | Out-String | Invoke-Expression
```

## What Gets Completed

### Commands

```bash
todoat <TAB>
# Shows: add, update, complete, delete, list, sync, view, version, etc.
```

### List Names

```bash
todoat <TAB>
# Shows: Work, Personal, Shopping (your task lists)
```

### Subcommands

```bash
todoat list <TAB>
# Shows: create, delete, show, trash, etc.
```

### Flags

```bash
todoat add --<TAB>
# Shows: --priority, --due-date, --description, --tags, etc.
```

### Flag Values

Some flags have value completion:

```bash
todoat -s <TAB>
# Shows: TODO, IN-PROGRESS, DONE, CANCELLED
```

## Reload After Changes

After modifying completion setup:

### Zsh

```bash
source ~/.zshrc
# Or
exec zsh
```

### Bash

```bash
source ~/.bashrc
# Or
exec bash
```

### Fish

```fish
source ~/.config/fish/config.fish
# Or
exec fish
```

## Troubleshooting

### Completion Not Working

1. Verify todoat is installed and in PATH:
   ```bash
   which todoat
   ```

2. Check completion script generates:
   ```bash
   todoat completion bash  # Should output script
   ```

3. Ensure shell completion is enabled:
   - Bash: `bash-completion` package installed
   - Zsh: `compinit` called in `.zshrc`

### Slow Completion

Completion fetches task lists from backend. If slow:

1. Enable sync for local caching:
   ```yaml
   sync:
     enabled: true
   ```

2. Run initial sync:
   ```bash
   todoat sync
   ```

### Outdated List Names

List names are cached. To refresh:

```bash
todoat sync
# Or restart your shell
```

## See Also

- [Getting Started](getting-started.md) - Initial setup
- [Task Management](task-management.md) - Commands to complete
