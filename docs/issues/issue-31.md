# Issue #31: Regression: Custom backend names broken

**GitHub Issue**: #31
**Status**: Open
**Created**: 2026-01-27
**Labels**: None

## Description

The CLI is no longer able to use custom-named backends. This is a regression - it used to work.

## Steps to Reproduce

1. Configure a custom-named backend in `config.yml`:
   ```yaml
   nextcloud-test:
     type: nextcloud
     enabled: true
     host: "localhost:8080"
     username: "admin"
     allow_http: true
     insecure_skip_verify: true
     suppress_ssl_warning: true
   ```

2. Try to use the backend with the CLI flag:
   ```bash
   todoat -b nextcloud-test
   ```

3. Or set it as default in config.yml:
   ```yaml
   default_backend: nextcloud-test
   ```

## Expected Behavior

Custom-named backends should be recognized and usable via:
- The `-b` / `--backend` CLI flag
- The `default_backend` configuration option

## Actual Behavior

Custom-named backends are not recognized by the CLI.

## Context

This is a regression - custom backend names worked in previous versions. The issue affects users who have configured backends with custom names (not using the default backend type names like "nextcloud", "caldav", etc.).

## Acceptance Criteria

- [ ] Custom-named backends can be selected via `-b` / `--backend` CLI flag
- [ ] Custom-named backends work when set as `default_backend` in config.yml
- [ ] Existing backend type names continue to work as expected
- [ ] Add regression tests to prevent this from happening again
