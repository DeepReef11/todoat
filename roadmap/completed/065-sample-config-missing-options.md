# [065] Sample Config Missing Options

## Summary
Add missing configuration options (`allow_http`, `auto_detect_backend`) to the sample config file with proper documentation.

## Documentation Reference
- Primary: `issues/0-allow-http-not-in-sample.md`
- Related: `dev-doc/CONFIGURATION.md`

## Dependencies
- Requires: none

## Complexity
S

## Acceptance Criteria

### Tests Required
- [ ] `TestSampleConfigContainsAllOptions` - Verify sample config includes all documented options

### Functional Requirements
- [ ] Sample config includes `allow_http` option with inline comment explaining its purpose
- [ ] Sample config includes `auto_detect_backend` option with `enabled: false` default
- [ ] Sample config includes backend priority configuration
- [ ] All new options have brief inline comments explaining their purpose

## Implementation Notes
- Check `dev-doc/CONFIGURATION.md` and `dev-doc/BACKEND_SYSTEM.md` for complete list of config options
- Sample config location: typically embedded or in config/ directory
- Use YAML comments to explain security implications of `allow_http`

## Out of Scope
- Adding new configuration features
- Modifying existing config behavior
