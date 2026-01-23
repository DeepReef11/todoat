package todoist_test

import (
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// TestIssue002KeyringCredentialsNotDetected reproduces issue #002:
// When token is stored in the keyring via 'credentials set todoist token',
// the backend should detect and use it instead of requiring the
// TODOAT_TODOIST_TOKEN environment variable.
//
// User journey from issue:
//
//	todoat credentials set todoist token --prompt
//	Enter password for todoist (user: token):
//	Credentials stored in system keyring
//
//	todoat -b todoist
//	Error: todoist backend requires TODOAT_TODOIST_TOKEN environment variable
//
// Expected: Should try to use keyring credentials, not require env var.
func TestIssue002KeyringCredentialsNotDetected(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set up config with todoist backend ONLY (no sqlite fallback)
	// This forces the CLI to use todoist or fail, making the error visible
	cli.SetFullConfig(`
backends:
  todoist:
    enabled: true
default_backend: todoist
no_prompt: true
`)

	// Clear the environment variable to ensure we're testing keyring-only path
	t.Setenv("TODOAT_TODOIST_TOKEN", "")

	// When no env var is set but keyring might have credentials,
	// the CLI should NOT say "requires TODOAT_TODOIST_TOKEN environment variable"
	// Instead, it should either:
	// 1. Try to use keyring credentials (and fail with "no credentials found" if empty)
	// 2. Or fail with a message that acknowledges keyring as an option
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "todoist", "list")

	combined := stdout + stderr

	// We expect the command to fail since no credentials exist
	// But the error should NOT exclusively mention the environment variable
	if exitCode == 0 {
		t.Logf("Command succeeded unexpectedly. Output: %s", combined)
		return
	}

	// BUG CHECK: The error message should NOT exclusively mention the environment variable
	// as if that's the ONLY way to provide credentials.
	// The message "requires TODOAT_TODOIST_TOKEN environment variable" ignores keyring.
	if strings.Contains(combined, "requires TODOAT_TODOIST_TOKEN environment variable") {
		t.Errorf("BUG: Error message only mentions environment variable, ignoring keyring credentials. "+
			"Should try keyring first or mention both options. Got: %s", combined)
	}
}

// TestIssue002TodoistWithKeyringCredentials tests that when token is available
// via keyring (simulated by env var as proxy), the backend should be used
// without requiring the specific TODOAT_TODOIST_TOKEN env var pattern.
//
// This verifies the fix works correctly: the CLI should use the credentials
// manager to retrieve credentials from keyring/env, similar to how nextcloud does it.
func TestIssue002TodoistWithKeyringCredentials(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set up config with todoist as default
	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  todoist:
    enabled: true
default_backend: todoist
no_prompt: true
`)

	// Set token via env var (this simulates keyring providing the token)
	// The fix should use credentials manager which checks both keyring and env
	t.Setenv("TODOAT_TODOIST_TOKEN", "fake-test-token-for-unit-test")

	// With credentials available, it should try to use Todoist backend
	// The API call will fail (fake token), but the error should be about the API,
	// not about missing credentials
	stdout, stderr, exitCode := cli.Execute("-y", "list")

	combined := stdout + stderr
	lowerCombined := strings.ToLower(combined)

	// Should NOT show warning about missing credentials
	if strings.Contains(lowerCombined, "requires todoat_todoist_token") {
		t.Errorf("Should not require env var when token is provided. Got: %s", combined)
	}

	// If exit code is non-zero, it should be due to API error (invalid token),
	// not missing credentials
	if exitCode != 0 {
		// Any error about auth/unauthorized/invalid token is expected with fake token
		// But error about "missing" or "requires" credentials is a bug
		if strings.Contains(lowerCombined, "missing") && strings.Contains(lowerCombined, "credential") {
			t.Errorf("Should not report missing credentials when token is set. Got: %s", combined)
		}
	}
}

// TestIssue002DetectBackendWithKeyring tests that --detect-backend includes
// todoist when credentials are available via the credentials manager
// (which checks keyring and env vars).
func TestIssue002DetectBackendWithKeyring(t *testing.T) {
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

	// Set the token (simulating keyring-provided credentials)
	t.Setenv("TODOAT_TODOIST_TOKEN", "test-token-value")

	stdout := cli.MustExecute("-y", "--detect-backend")

	// Should show todoist in the list of detected backends
	if !strings.Contains(stdout, "todoist") {
		t.Errorf("expected todoist to appear in --detect-backend output when credentials are available, got: %s", stdout)
	}
}
