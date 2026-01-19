# [051] Version Command

## Summary
Implement a version command that displays the application version, build information, and optionally checks for updates.

## Documentation Reference
- Primary: `dev-doc/CLI_INTERFACE.md`
- Section: CLI Features (Version Command mentioned)

## Dependencies
- Requires: [002] Core CLI with Cobra

## Complexity
S

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestVersionCommand` - `todoat version` displays version string
- [ ] `TestVersionVerbose` - `todoat version -v` shows extended build info
- [ ] `TestVersionJSON` - `todoat --json version` returns JSON with version fields
- [ ] `TestVersionShort` - `todoat --version` works as alias

### Functional Requirements
- [ ] `todoat version` displays: version number, build date, git commit
- [ ] `todoat --version` works as shorthand
- [ ] `todoat version -v` includes: Go version, OS/Arch, build tags
- [ ] Version info embedded at build time via ldflags
- [ ] JSON output includes all version fields

## Implementation Notes

### Version Information Structure
```go
type VersionInfo struct {
    Version   string `json:"version"`    // Semantic version (e.g., "1.2.3")
    Commit    string `json:"commit"`     // Git commit hash (short)
    BuildDate string `json:"build_date"` // ISO8601 build timestamp
    GoVersion string `json:"go_version"` // Go compiler version
    Platform  string `json:"platform"`   // OS/Arch (e.g., "linux/amd64")
}
```

### Build-Time Injection
Update Makefile to inject version info:
```makefile
VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)
```

## Out of Scope
- Auto-update functionality
- Version comparison with remote releases
- Checking for newer versions
