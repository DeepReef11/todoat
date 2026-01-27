# 035 - Auto-Install Shell Completion

## Summary

Add a `todoat completion install` command that automatically detects the user's shell and installs the completion script to the appropriate location, instead of requiring users to manually pipe the output.

## Dependencies

None

## Acceptance Criteria

### CLI Tests Required

1. **Auto-detect shell and install**
   ```bash
   todoat completion install
   # Detects current shell from $SHELL
   # Installs to appropriate location
   # Output: "Completion installed for zsh at ~/.config/todoat/completions/_todoat"
   #         "Run 'source ~/.zshrc' or restart your shell to enable"
   ```

2. **Explicit shell specification**
   ```bash
   todoat completion install --shell bash
   # Installs bash completion regardless of current shell
   ```

3. **Prompt for confirmation (without -y)**
   ```bash
   todoat completion install
   # Output: "Install zsh completion to /usr/local/share/zsh/site-functions/_todoat? [Y/n]"
   ```

4. **Silent install with -y**
   ```bash
   todoat -y completion install
   # No prompt, just installs
   ```

5. **Show install location without installing**
   ```bash
   todoat completion install --dry-run
   # Output: "Would install zsh completion to: /usr/local/share/zsh/site-functions/_todoat"
   ```

6. **Handle permission issues gracefully**
   ```bash
   todoat completion install
   # If /etc/bash_completion.d not writable:
   # "Cannot write to /etc/bash_completion.d (permission denied)"
   # "Alternative: Install to ~/.local/share/bash-completion/completions/todoat"
   # "Proceed with alternative location? [Y/n]"
   ```

7. **Uninstall option**
   ```bash
   todoat completion uninstall
   # Removes installed completion script
   # "Removed zsh completion from /usr/local/share/zsh/site-functions/_todoat"
   ```

## Implementation Notes

### Shell Detection

```go
func detectShell() string {
    // Check $SHELL environment variable
    shell := os.Getenv("SHELL")
    if shell != "" {
        // Extract shell name from path (e.g., /bin/zsh -> zsh)
        return filepath.Base(shell)
    }

    // Fallback: check parent process on Unix
    // Fallback: check $ComSpec on Windows for PowerShell detection

    return "unknown"
}
```

### Install Locations by Shell and OS

**Bash:**
| OS | Primary Location | Fallback (user-writable) |
|----|------------------|--------------------------|
| Linux | `/etc/bash_completion.d/todoat` | `~/.local/share/bash-completion/completions/todoat` |
| macOS | `$(brew --prefix)/etc/bash_completion.d/todoat` | `~/.local/share/bash-completion/completions/todoat` |

**Zsh:**
| OS | Primary Location | Fallback (user-writable) |
|----|------------------|--------------------------|
| Linux | `/usr/local/share/zsh/site-functions/_todoat` | `~/.config/todoat/completions/_todoat` (add to fpath) |
| macOS | `$(brew --prefix)/share/zsh/site-functions/_todoat` | `~/.config/todoat/completions/_todoat` |

**Fish:**
| OS | Location |
|----|----------|
| All | `~/.config/fish/completions/todoat.fish` |

**PowerShell:**
| OS | Location |
|----|----------|
| All | Write to profile, or `~/.config/todoat/completions/todoat.ps1` with profile instruction |

### Command Structure

```
todoat completion
├── bash         - Output bash completion script (existing)
├── zsh          - Output zsh completion script (existing)
├── fish         - Output fish completion script (existing)
├── powershell   - Output powershell completion script (existing)
├── install      - Auto-install completion for detected/specified shell
│   ├── --shell  - Specify shell (bash, zsh, fish, powershell)
│   └── --dry-run - Show where it would install without installing
└── uninstall    - Remove installed completion script
```

### Installation Flow

```go
func installCompletion(shell string, autoYes bool) error {
    // 1. Determine install location
    loc := getInstallLocation(shell, runtime.GOOS)

    // 2. Check if location is writable
    if !isWritable(loc.Primary) {
        if !autoYes {
            // Prompt for fallback location
        }
        loc = loc.Fallback
    }

    // 3. Confirm with user (unless -y)
    if !autoYes {
        fmt.Printf("Install %s completion to %s? [Y/n] ", shell, loc)
        // Read confirmation
    }

    // 4. Generate completion script
    script := generateCompletionScript(shell)

    // 5. Create parent directories if needed
    os.MkdirAll(filepath.Dir(loc), 0755)

    // 6. Write script
    os.WriteFile(loc, []byte(script), 0644)

    // 7. Print post-install instructions
    printPostInstallInstructions(shell, loc)

    return nil
}
```

### Post-Install Instructions

After installation, print shell-specific instructions:

**Zsh (fallback location):**
```
Completion installed to ~/.config/todoat/completions/_todoat

Add to your ~/.zshrc:
  fpath=(~/.config/todoat/completions $fpath)
  autoload -U compinit; compinit

Then run: source ~/.zshrc
```

**Bash (fallback location):**
```
Completion installed to ~/.local/share/bash-completion/completions/todoat

This location should be auto-loaded if bash-completion is installed.
If not working, add to ~/.bashrc:
  source ~/.local/share/bash-completion/completions/todoat

Then run: source ~/.bashrc
```

## Documentation Updates

Update `docs/how-to/shell-completion.md`:

```markdown
## Quick Setup

The easiest way to set up completion:

```bash
todoat completion install
```

This automatically detects your shell and installs completion to the appropriate location.

### Manual Setup

If you prefer manual control, or the auto-install doesn't work for your setup:
[existing manual instructions...]
```

## Files to Modify

- `cmd/todoat/cmd/completion.go` (or create new file) - Add install/uninstall commands
- `docs/how-to/shell-completion.md` - Update with auto-install instructions
- `cmd/todoat/cmd/todoat_test.go` - Add tests for install command
