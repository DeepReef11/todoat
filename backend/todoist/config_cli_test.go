package todoist_test

import (
	"strings"
	"testing"

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
// the CLI attempts to use Todoist backend (not SQLite).
// This test reproduces issue #001 - default_backend config is ignored.
func TestDefaultBackendTodoistUsedCLI(t *testing.T) {
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

	// Clear the token env var to ensure we get a credentials error
	t.Setenv("TODOAT_TODOIST_TOKEN", "")

	// When default_backend is todoist but no credentials, we should get an error
	// about Todoist credentials, NOT silently fall back to SQLite
	stdout, stderr, exitCode := cli.Execute("-y", "list")

	// With the bug (issue #001), this would succeed using SQLite silently
	// After the fix, this should fail with a Todoist credentials error
	combined := stdout + stderr

	if exitCode == 0 {
		// If it succeeds, check if it's using SQLite (bug) or Todoist
		// The test should fail if we see SQLite being used
		if !strings.Contains(combined, "todoist") && !strings.Contains(combined, "Todoist") {
			t.Errorf("default_backend is set to todoist but CLI silently used SQLite. Output: %s", combined)
		}
	} else {
		// If it fails, verify it's because of Todoist credentials (correct behavior)
		lowerCombined := strings.ToLower(combined)
		if !strings.Contains(lowerCombined, "todoist") && !strings.Contains(lowerCombined, "token") && !strings.Contains(lowerCombined, "credential") {
			t.Errorf("expected error about Todoist credentials, got: %s", combined)
		}
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
