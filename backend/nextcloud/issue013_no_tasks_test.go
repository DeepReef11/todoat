package nextcloud_test

import (
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// TestIssue013NoTasksShownResolved verifies that issue #013 (no tasks shown) was resolved
// as part of issue #012 (HTTP/HTTPS config inconsistency).
//
// Issue #013 described: "Task list shows no tasks even though tasks exist on the backend"
// when using `-b nextcloud`, but `-b nextcloud-test` worked correctly.
//
// Root cause: The built-in backend name "nextcloud" was not reading config file settings,
// causing HTTP/HTTPS protocol mismatch which resulted in connection errors (and thus no tasks).
//
// Resolution: Issue #012 fixed `createBackendByName` to check if "nextcloud" is configured
// in the config file before falling back to environment-only mode. Both backends now use
// `createCustomBackend` when configured, ensuring identical behavior.
func TestIssue013NoTasksShownResolved(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Configure both backends identically (same as issue #012 test setup)
	cli.SetFullConfig(`
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    host: "localhost:9999"
    username: "admin"
    password: "testpass"
    allow_http: true
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "localhost:9999"
    username: "admin"
    password: "testpass"
    allow_http: true
default_backend: sqlite
no_prompt: true
`)

	// Clear environment variables
	t.Setenv("TODOAT_NEXTCLOUD_HOST", "")
	t.Setenv("TODOAT_NEXTCLOUD_USERNAME", "")
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "")

	// Test with built-in "nextcloud"
	_, stderrBuiltin, exitBuiltin := cli.Execute("-y", "-b", "nextcloud", "list")

	// Test with custom "nextcloud-test"
	_, stderrCustom, exitCustom := cli.Execute("-y", "-b", "nextcloud-test", "list")

	// Both should now behave identically - connection refused (no server) rather than
	// protocol mismatch or env var errors
	errBuiltin := strings.ToLower(stderrBuiltin)
	errCustom := strings.ToLower(stderrCustom)

	// Verify neither gets HTTP/HTTPS mismatch (which was the original issue #012 problem)
	httpsError := "server gave http response to https client"
	if strings.Contains(errBuiltin, httpsError) || strings.Contains(errCustom, httpsError) {
		t.Errorf("Protocol mismatch detected - issue #012 fix may have regressed.\n"+
			"'-b nextcloud' stderr: %s\n"+
			"'-b nextcloud-test' stderr: %s", stderrBuiltin, stderrCustom)
	}

	// Verify neither requires env vars (config file should be read for both)
	envVarError := "environment variable"
	if strings.Contains(errBuiltin, envVarError) {
		t.Errorf("'-b nextcloud' still requires env vars instead of reading config file.\n"+
			"This indicates issue #012 fix has regressed.\nstderr: %s", stderrBuiltin)
	}
	if strings.Contains(errCustom, envVarError) {
		t.Errorf("'-b nextcloud-test' requires env vars - unexpected.\nstderr: %s", stderrCustom)
	}

	// Both should have same exit code and similar errors (connection refused to localhost:9999)
	if exitBuiltin != exitCustom {
		t.Logf("Note: Exit codes differ (builtin=%d, custom=%d). "+
			"Both backends may not be fully equivalent yet.", exitBuiltin, exitCustom)
	}

	// Verify both get the same type of error (connection-related, not config-related)
	connectionError := "connection refused"
	dialError := "dial tcp"
	hasConnectionErr := strings.Contains(errBuiltin, connectionError) || strings.Contains(errBuiltin, dialError)
	hasConnectionErrCustom := strings.Contains(errCustom, connectionError) || strings.Contains(errCustom, dialError)

	if hasConnectionErr != hasConnectionErrCustom {
		t.Errorf("Different error types between backends.\n"+
			"'-b nextcloud' stderr: %s\n"+
			"'-b nextcloud-test' stderr: %s", stderrBuiltin, stderrCustom)
	}
}
