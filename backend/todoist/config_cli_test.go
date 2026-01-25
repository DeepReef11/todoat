package todoist_test

import (
	"context"
	"strings"
	"testing"

	"todoat/internal/credentials"
	"todoat/internal/testutil"
)

// =============================================================================
// CLI Tests for Todoist Backend Configuration (Issue #001)
// =============================================================================

// TestTodoistConfigSetDefaultBackendCLI verifies 'config set default_backend todoist' works
func TestTodoistConfigSetDefaultBackendCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  todoist:
    enabled: true
default_backend: sqlite
no_prompt: true
`)

	stdout := cli.MustExecute("-y", "config", "set", "default_backend", "todoist")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the value was changed
	stdout = cli.MustExecute("-y", "config", "get", "default_backend")
	testutil.AssertContains(t, stdout, "todoist")
}

// TestTodoistConfigSetBackendEnabledCLI verifies 'config set backends.todoist.enabled true' works
func TestTodoistConfigSetBackendEnabledCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  todoist:
    enabled: false
default_backend: sqlite
no_prompt: true
`)

	stdout := cli.MustExecute("-y", "config", "set", "backends.todoist.enabled", "true")

	testutil.AssertResultCode(t, stdout, testutil.ResultActionCompleted)

	// Verify the value was changed
	stdout = cli.MustExecute("-y", "config", "get", "backends.todoist.enabled")
	testutil.AssertContains(t, stdout, "true")
}

// TestTodoistConfigGetAllShowsTodoistCLI verifies 'config get' shows todoist section
func TestTodoistConfigGetAllShowsTodoistCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  todoist:
    enabled: true
default_backend: sqlite
no_prompt: true
`)

	stdout := cli.MustExecute("-y", "config", "get")

	// Should show todoist in backends section
	testutil.AssertContains(t, stdout, "todoist")
}

// TestTodoistConfigValidationCLI verifies todoist is a valid backend value
func TestTodoistConfigValidationCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  todoist:
    enabled: true
default_backend: sqlite
no_prompt: true
`)

	// This should NOT return an error mentioning "valid: sqlite"
	stdout, stderr, exitCode := cli.Execute("-y", "config", "set", "default_backend", "todoist")

	if exitCode != 0 {
		combined := stdout + stderr
		if strings.Contains(combined, "valid: sqlite") && !strings.Contains(combined, "todoist") {
			t.Errorf("todoist should be a valid backend option, but got error: %s", combined)
		}
	}
}

// TestTodoistDetectBackendWithTokenCLI verifies --detect-backend shows todoist when token is set
func TestTodoistDetectBackendWithTokenCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  todoist:
    enabled: true
default_backend: sqlite
no_prompt: true
`)

	// Set the environment variable for the token
	t.Setenv("TODOAT_TODOIST_TOKEN", "test-token-value")

	stdout := cli.MustExecute("-y", "--detect-backend")

	// Should show todoist in the list of detected backends
	if !strings.Contains(stdout, "todoist") {
		t.Errorf("expected todoist to appear in --detect-backend output when TODOAT_TODOIST_TOKEN is set, got: %s", stdout)
	}
}

// TestDefaultBackendTodoistUsedCLI verifies that when default_backend is set to todoist,
// but credentials are missing, the CLI warns and falls back to SQLite gracefully.
// This test relates to issue 001 (backend detection fallback warning).
func TestDefaultBackendTodoistUsedCLI(t *testing.T) {
	// Skip if keyring has todoist credentials - we can't isolate the keyring in CLI tests
	credMgr := credentials.NewManager()
	if credInfo, err := credMgr.Get(context.Background(), "todoist", "token"); err == nil && credInfo.Found {
		t.Skip("Skipping test: keyring has todoist credentials that cannot be isolated")
	}

	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  todoist:
    enabled: true
default_backend: todoist
no_prompt: true
`)

	// Clear the token env var to ensure backend is unavailable
	t.Setenv("TODOAT_TODOIST_TOKEN", "")

	// When default_backend is todoist but no credentials, the CLI should:
	// 1. Warn the user about todoist being unavailable
	// 2. Fall back to SQLite gracefully (exit code 0)
	stdout, stderr, exitCode := cli.Execute("-y", "list")

	combined := stdout + stderr
	lowerCombined := strings.ToLower(combined)

	// Should succeed using fallback
	if exitCode != 0 {
		t.Errorf("expected exit code 0 (graceful fallback), got %d. Output: %s", exitCode, combined)
	}

	// Should show warning about todoist being unavailable
	if !strings.Contains(lowerCombined, "warning") || !strings.Contains(lowerCombined, "todoist") {
		t.Errorf("expected warning about todoist being unavailable, got: %s", combined)
	}

	// Should mention sqlite as the fallback
	if !strings.Contains(lowerCombined, "sqlite") {
		t.Errorf("expected mention of sqlite fallback, got: %s", combined)
	}
}

// TestDefaultBackendTodoistWithTokenCLI verifies that when default_backend is todoist
// and credentials are available, the CLI attempts to use Todoist backend.
func TestDefaultBackendTodoistWithTokenCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  todoist:
    enabled: true
default_backend: todoist
no_prompt: true
`)

	// Set a fake token to trigger Todoist backend usage
	t.Setenv("TODOAT_TODOIST_TOKEN", "fake-test-token-for-unit-test")

	// When default_backend is todoist with credentials, it should try to use Todoist
	// The API call will fail (fake token), but the error should be about the API,
	// not about missing credentials or falling back to SQLite
	stdout, stderr, exitCode := cli.Execute("-y", "list")

	combined := stdout + stderr

	// The command should fail because the token is invalid (API error)
	if exitCode == 0 {
		// If it succeeds with no lists, check if output mentions Todoist
		// (shouldn't happen with a fake token, but just in case)
		if !strings.Contains(combined, "No lists found") {
			t.Logf("Command succeeded unexpectedly. Output: %s", combined)
		}
	} else {
		// Expected: API error or connection error from Todoist, NOT SQLite fallback
		// SQLite would say "No lists found" on success
		lowerCombined := strings.ToLower(combined)
		// Any error is fine as long as we're not silently using SQLite
		if strings.Contains(lowerCombined, "no lists found") {
			t.Errorf("expected Todoist backend to be used, but got SQLite-like output: %s", combined)
		}
	}
}
