# Shell Completion

todoat supports tab completion for Bash, Zsh, Fish, and PowerShell shells.

## Setup

### Zsh

First, ensure shell completion is enabled in your environment:

```bash
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

Load completions in your current session:

```bash
source <(todoat completion zsh)
```

Or save to file for permanent setup:

**Linux:**
```bash
todoat completion zsh > "${fpath[1]}/_todoat"
```

**macOS:**
```bash
todoat completion zsh > $(brew --prefix)/share/zsh/site-functions/_todoat
```

### Bash

This requires the `bash-completion` package. Install via your package manager if needed.

Load completions in your current session:

```bash
source <(todoat completion bash)
```

Or save to file for permanent setup:

**Linux:**
```bash
todoat completion bash > /etc/bash_completion.d/todoat
```

**macOS:**
```bash
todoat completion bash > $(brew --prefix)/etc/bash_completion.d/todoat
```

### Fish

Load completions in your current session:

```fish
todoat completion fish | source
```

Or save to file for permanent setup:

```bash
todoat completion fish > ~/.config/fish/completions/todoat.fish
```

### PowerShell

Load completions in your current session:

```powershell
todoat completion powershell | Out-String | Invoke-Expression
```

For permanent setup, add the above command to your PowerShell profile.

## What Gets Completed

### Commands

```bash
todoat <TAB>
# Shows: analytics, completion, config, credentials, help, list, migrate,
#        notification, reminder, sync, tags, tui, version, view
```

### List Names

```bash
todoat <TAB>
# Shows: Work, Personal, Shopping (your task lists)
```

### Subcommands

```bash
todoat list <TAB>
# Shows: create, delete, export, import, info, stats, trash, update, vacuum
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

- [Getting Started](../tutorials/getting-started.md) - Initial setup
- [Task Management](task-management.md) - Commands to complete
