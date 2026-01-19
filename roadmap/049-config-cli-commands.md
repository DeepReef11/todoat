# [049] Config CLI Commands

## Summary
Implement CLI commands for viewing and modifying configuration without manually editing YAML files, providing a user-friendly interface for common config operations.

## Documentation Reference
- Primary: `dev-doc/CONFIGURATION.md`
- Section: Feature Documentation (general)
- Related: `dev-doc/CLI_INTERFACE.md`

## Dependencies
- Requires: [010] Configuration System (config file exists)

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestConfigGet` - `todoat config get default_backend` returns current value
- [ ] `TestConfigGetNested` - `todoat config get sync.enabled` returns nested value
- [ ] `TestConfigGetAll` - `todoat config get` returns all config as YAML
- [ ] `TestConfigSet` - `todoat config set no_prompt true` updates config file
- [ ] `TestConfigSetNested` - `todoat config set sync.offline_mode auto` updates nested value
- [ ] `TestConfigSetValidation` - `todoat config set no_prompt invalid` returns ERROR with valid values
- [ ] `TestConfigPath` - `todoat config path` shows config file location
- [ ] `TestConfigEdit` - `todoat config edit` opens config in $EDITOR
- [ ] `TestConfigReset` - `todoat config reset` restores default config (with confirmation)
- [ ] `TestConfigJSON` - `todoat --json config get` returns JSON format

### Functional Requirements
- [ ] `config get [key]` - Display config value(s), supports dot notation for nested keys
- [ ] `config set <key> <value>` - Update config value with validation
- [ ] `config path` - Show path to active config file
- [ ] `config edit` - Open config file in system editor ($EDITOR or vi)
- [ ] `config reset` - Reset to default config (requires confirmation)
- [ ] All commands support `--json` output format

## Implementation Notes

### CLI Commands Structure
```bash
todoat config                    # Alias for 'config get'
todoat config get                # Show all config
todoat config get <key>          # Show specific key
todoat config set <key> <value>  # Set config value
todoat config path               # Show config file path
todoat config edit               # Open in editor
todoat config reset              # Reset to defaults
```

### Key Path Resolution
- Support dot notation: `sync.enabled`, `backends.sqlite.path`
- Support array indexing: `backend_priority[0]`
- Case-insensitive key matching

### Validation Rules
- Boolean fields: accept `true`/`false`, `yes`/`no`, `1`/`0`
- Enum fields: validate against allowed values
- Path fields: expand `~` and validate parent directory exists
- Required fields: prevent deletion

### Config File Update
- Read existing config
- Parse and validate new value
- Preserve comments where possible (use yaml.v3 with ast)
- Atomic write (write to temp, rename)

## Out of Scope
- Config encryption/secrets management
- Remote config sync
- Config profiles/environments
- Watch mode for config changes
