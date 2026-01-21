# [031] CI/CD Integration

## Summary
Configure GitHub Actions workflows for automated testing, including unit tests on all PRs and integration tests with Docker-based Nextcloud on merges to main.

## Documentation Reference
- Primary: `docs/explanation/integration-testing.md` (CI/CD Integration section)
- Secondary: `docs/explanation/test-driven-dev.md`

## Dependencies
- Requires: [030] Integration Test Infrastructure

## Complexity
M

## Acceptance Criteria

### CLI Tests Required
- [ ] `TestGitHubWorkflowExists` - `.github/workflows/ci.yml` file exists
- [ ] `TestWorkflowRunsUnitTests` - Workflow includes `go test ./...` step
- [ ] `TestWorkflowRunsLint` - Workflow includes golangci-lint step
- [ ] `TestWorkflowBuild` - Workflow includes `go build` step

### CI Workflow Requirements
- [ ] `WorkflowTriggersOnPR` - Workflow triggers on pull requests to main
- [ ] `WorkflowTriggersOnPush` - Workflow triggers on pushes to main
- [ ] `WorkflowUsesGoSetup` - Uses actions/setup-go with Go 1.22+
- [ ] `WorkflowCachesModules` - Caches Go modules for faster builds

### Integration Test Workflow Requirements
- [ ] `IntegrationWorkflowExists` - Separate workflow for integration tests
- [ ] `IntegrationUsesNextcloudService` - Docker service container for Nextcloud
- [ ] `IntegrationWaitsForHealth` - Waits for Nextcloud health check
- [ ] `IntegrationRunsOnMainOnly` - Integration tests only run on main branch merges

## Implementation Notes

### Files to Create

1. **`.github/workflows/ci.yml`** - Main CI workflow
   ```yaml
   name: CI
   on:
     push:
       branches: [main]
     pull_request:
       branches: [main]
   jobs:
     test:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with:
             go-version: '1.22'
         - run: go test ./...
         - run: go build ./cmd/todoat
   ```

2. **`.github/workflows/integration.yml`** - Integration test workflow
   ```yaml
   name: Integration Tests
   on:
     push:
       branches: [main]
   jobs:
     nextcloud:
       runs-on: ubuntu-latest
       services:
         nextcloud:
           image: nextcloud:latest
           ports:
             - 8080:80
           env:
             NEXTCLOUD_ADMIN_USER: admin
             NEXTCLOUD_ADMIN_PASSWORD: adminpass
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
         - run: go test -tags=integration ./backend/nextcloud
   ```

### Workflow Features
- Go module caching for faster CI runs
- golangci-lint for code quality
- Build verification
- Unit tests on all PRs
- Integration tests on main merges only (to avoid API limits)
- Secrets for Todoist token (optional, for API tests)

### Secret Configuration
| Secret | Purpose |
|--------|---------|
| `TODOAT_TODOIST_TOKEN` | Todoist API token for integration tests (optional) |

## Out of Scope
- Release automation (versioning, tagging)
- Container image publishing
- Code coverage badges
- Dependabot configuration
- Google Tasks / Microsoft To Do integration tests in CI (require OAuth complexity)
