# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Nextcloud/CalDAV backend now parses and generates RELATED-TO property for parent-child task relationships (subtask support) (#29)
- `--parent` flag now accepts task UID directly when multiple tasks have the same name (#28)

### Fixed
- `config set analytics.enabled` and `config set analytics.retention_days` now work correctly (#078)
- Documentation command syntax examples now use correct list-based format (#001)

### Changed
- Default value for `analytics.enabled` changed from `false` to `true`
- Documentation improvements for configuration and synchronization
- Clarified that default view excludes completed tasks (use `-v all` to see them)
- Simplified feature documentation tables by removing redundant status column (all features are stable)
- Updated getting-started tutorial with `completion install` quick setup

### Added
- `completion install` command to auto-detect shell and install completion scripts (#035)
- `completion uninstall` command to remove installed completion scripts
- Tests for analytics configuration settings
- Issue tracking for config key and value validation bugs
- Integration tests for credential manager keyring flow (#010)
- Documentation for background pull sync on read operations
- Configurable `sync.background_pull_cooldown` option to control cooldown between background pull syncs (default: 30s, minimum: 5s) (#082)
