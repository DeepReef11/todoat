# Installation

How to install todoat on your system.

## Requirements

- Go 1.21 or later

## From Source

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/yourusername/todoat.git
cd todoat

# Build the binary
make build

# The binary will be in ./bin/todoat
```

### Install to PATH

```bash
# Install to $GOPATH/bin
go install ./cmd/todoat
```

Or manually copy the binary:

```bash
# Copy to a directory in your PATH
sudo cp ./bin/todoat /usr/local/bin/
```

## Verify Installation

```bash
# Check version
todoat --version

# Show help
todoat --help
```

## Shell Completion

todoat supports shell completion for bash, zsh, fish, and PowerShell.

### Bash

```bash
# Add to ~/.bashrc
source <(todoat completion bash)

# Or generate and save the script
todoat completion bash > /etc/bash_completion.d/todoat
```

### Zsh

```bash
# Add to ~/.zshrc
source <(todoat completion zsh)

# Or generate and add to fpath
todoat completion zsh > "${fpath[1]}/_todoat"
```

### Fish

```bash
todoat completion fish > ~/.config/fish/completions/todoat.fish
```

### PowerShell

```powershell
todoat completion powershell | Out-String | Invoke-Expression
```

## Data Location

todoat stores data following XDG conventions:

- **Configuration**: `~/.config/todoat/config.yaml`
- **Database**: `~/.local/share/todoat/tasks.db` (SQLite database)
- **Custom Views**: `~/.config/todoat/views/`

If XDG environment variables are set, those paths are used instead:
- `$XDG_CONFIG_HOME/todoat/` for configuration
- `$XDG_DATA_HOME/todoat/` for data

The directories are created automatically on first use.

## Uninstall

To uninstall todoat:

```bash
# Remove the binary
rm $(which todoat)

# Optionally remove configuration and data
rm -rf ~/.config/todoat
rm -rf ~/.local/share/todoat
```

---
*Last updated: 2026-01-19*
