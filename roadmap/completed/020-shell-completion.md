# [020] Shell Completion

## Summary
Implement shell completion scripts for Bash, Zsh, Fish, and PowerShell, enabling tab completion for commands, subcommands, list names, and task summaries.

## Documentation Reference
- Primary: `docs/explanation/cli-interface.md#shell-completion`
- Related: `docs/explanation/cli-interface.md`

## Dependencies
- Requires: [002] Core CLI (Cobra framework provides completion infrastructure)
- Requires: [007] List Commands (for list name completion)

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestCompletionBash` - `todoat completion bash` outputs valid Bash completion script
- [ ] `TestCompletionZsh` - `todoat completion zsh` outputs valid Zsh completion script
- [ ] `TestCompletionFish` - `todoat completion fish` outputs valid Fish completion script
- [ ] `TestCompletionPowerShell` - `todoat completion powershell` outputs valid PowerShell completion script
- [ ] `TestCompletionHelp` - `todoat completion --help` shows usage instructions for each shell
- [ ] `TestCompletionInstallInstructions` - Each completion subcommand outputs installation instructions

### Manual Verification
- [ ] Bash: Commands complete with TAB (todoat li<TAB> → todoat list)
- [ ] Zsh: List names complete dynamically (todoat MyL<TAB> → todoat MyList)
- [ ] Fish: Subcommands shown in completion menu
- [ ] All shells: Flags complete (todoat --<TAB> shows available flags)

## Implementation Notes
- Cobra provides built-in completion infrastructure via `cmd.GenBashCompletion()`
- Add custom completion functions for dynamic values (list names, task summaries)
- Include installation instructions in command output for each shell
- Test completion scripts don't error when sourced

## Out of Scope
- Auto-installation of completion scripts
- IDE/editor integration
- Custom completion for third-party plugins
