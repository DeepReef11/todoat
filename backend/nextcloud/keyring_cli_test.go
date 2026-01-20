package nextcloud_test

import (
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// TestIssue007KeyringCredentialsNotUsedForValidation reproduces issue #007:
// When host and username are configured in the config file and password is stored
// in the keyring, the backend validation should use these sources instead of
// requiring all values as environment variables.
func TestIssue007KeyringCredentialsNotUsedForValidation(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set up config with nextcloud host and username, but no password in config
	// This mimics: user has config file + password stored in keyring
	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  nextcloud:
    enabled: true
    host: "fake-test-server.local"
    username: "admin"
default_backend: nextcloud
no_prompt: true
`)

	// Clear all environment variables to ensure we're testing config + keyring only
	t.Setenv("TODOAT_NEXTCLOUD_HOST", "")
	t.Setenv("TODOAT_NEXTCLOUD_USERNAME", "")
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "")

	// When host and username are in config but password is missing everywhere,
	// the warning should mention only password is missing (not host/username).
	// Current buggy behavior: warns about all three env vars being missing
	stdout, stderr, exitCode := cli.Execute("-y", "list")

	combined := stdout + stderr
	lowerCombined := strings.ToLower(combined)

	// Should succeed using fallback (sqlite)
	if exitCode != 0 {
		t.Errorf("expected exit code 0 (graceful fallback), got %d. Output: %s", exitCode, combined)
	}

	// Should show warning about nextcloud being unavailable
	if !strings.Contains(lowerCombined, "warning") || !strings.Contains(lowerCombined, "nextcloud") {
		t.Errorf("expected warning about nextcloud being unavailable, got: %s", combined)
	}

	// BUG CHECK: The warning should NOT mention TODOAT_NEXTCLOUD_HOST or TODOAT_NEXTCLOUD_USERNAME
	// because those values ARE configured in the config file.
	// Only the password should be mentioned as missing.
	if strings.Contains(combined, "TODOAT_NEXTCLOUD_HOST") {
		t.Errorf("BUG: Warning mentions TODOAT_NEXTCLOUD_HOST even though host is in config file. "+
			"Config values should be read. Got: %s", combined)
	}
	if strings.Contains(combined, "TODOAT_NEXTCLOUD_USERNAME") {
		t.Errorf("BUG: Warning mentions TODOAT_NEXTCLOUD_USERNAME even though username is in config file. "+
			"Config values should be read. Got: %s", combined)
	}

	// The warning should guide user to set password (via keyring or env)
	// This is the EXPECTED behavior after fix
	if !strings.Contains(combined, "password") && !strings.Contains(combined, "PASSWORD") {
		t.Logf("Note: Warning should mention password as the missing credential. Got: %s", combined)
	}
}

// TestIssue007NextcloudWithConfigAndKeyring tests that when host/username are
// in config and password is available (from env as proxy for keyring test),
// the backend should be used.
func TestIssue007NextcloudWithConfigAndKeyring(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set up config with nextcloud host and username
	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  nextcloud:
    enabled: true
    host: "fake-test-server.local"
    username: "admin"
default_backend: nextcloud
no_prompt: true
`)

	// Clear host and username env vars to ensure config values are used
	t.Setenv("TODOAT_NEXTCLOUD_HOST", "")
	t.Setenv("TODOAT_NEXTCLOUD_USERNAME", "")
	// Set password via env (this simulates keyring providing the password)
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "testpassword")

	// The backend should now be initialized using:
	// - host from config: fake-test-server.local
	// - username from config: admin
	// - password from env (simulating keyring): testpassword
	stdout, stderr, exitCode := cli.Execute("-y", "list")

	combined := stdout + stderr
	lowerCombined := strings.ToLower(combined)

	// BUG CHECK: If the code uses the config values for host/username,
	// it should try to connect to "fake-test-server.local" which will fail
	// with a connection error (not a "missing credentials" warning).
	// If it falls back to sqlite with a warning about env vars, that's the bug.

	if strings.Contains(lowerCombined, "todoat_nextcloud_host") {
		t.Errorf("BUG: Still asking for TODOAT_NEXTCLOUD_HOST even though host is in config. Got: %s", combined)
	}

	if strings.Contains(lowerCombined, "todoat_nextcloud_username") {
		t.Errorf("BUG: Still asking for TODOAT_NEXTCLOUD_USERNAME even though username is in config. Got: %s", combined)
	}

	// If we get a sqlite fallback warning, that's the bug
	if strings.Contains(lowerCombined, "sqlite") && strings.Contains(lowerCombined, "warning") {
		t.Errorf("BUG: Fell back to sqlite instead of using nextcloud with config values. Got: %s", combined)
	}

	// Expected: either connection error (correct behavior - tried to use nextcloud)
	// or success with empty list (if somehow connection worked)
	_ = exitCode // exit code varies based on connection behavior
}

// TestIssue009AllowHTTPNotReadFromConfig reproduces issue #009:
// When allow_http: true is set in the config file, the backend should use
// HTTP protocol instead of HTTPS. Currently, allow_http is not being read
// from the config, so HTTPS is always used regardless of config setting.
func TestIssue009AllowHTTPNotReadFromConfig(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set up config with allow_http: true for HTTP server
	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  nextcloud:
    enabled: true
    host: "localhost:8080"
    username: "testuser"
    password: "testpass"
    allow_http: true
    insecure_skip_verify: true
default_backend: nextcloud
no_prompt: true
`)

	// Clear all env vars to ensure config values are used
	t.Setenv("TODOAT_NEXTCLOUD_HOST", "")
	t.Setenv("TODOAT_NEXTCLOUD_USERNAME", "")
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "")

	// Run a command that will try to connect to the backend
	stdout, stderr, _ := cli.Execute("-y", "list")

	combined := stdout + stderr

	// BUG CHECK: If allow_http is NOT being read from config, the backend
	// will try to use HTTPS (https://localhost:8080) instead of HTTP.
	// This will cause an error like:
	// "http: server gave HTTP response to HTTPS client"
	//
	// If allow_http IS properly read, we would see a different error
	// (connection refused, no route to host, etc. since no server is running)

	if strings.Contains(combined, "server gave HTTP response to HTTPS client") {
		t.Errorf("BUG: allow_http: true in config is being ignored. "+
			"Backend tried HTTPS instead of HTTP. Got: %s", combined)
	}

	// We expect either:
	// 1. Connection refused (HTTP tried, no server running) - CORRECT
	// 2. No route to host / connection timeout - CORRECT
	// 3. HTTPS error message - BUG (allow_http not read)

	// Additional check: if the error mentions https:// URL, that's wrong
	if strings.Contains(combined, "https://localhost:8080") {
		t.Errorf("BUG: allow_http: true in config is being ignored. "+
			"URL shows https:// instead of http://. Got: %s", combined)
	}
}
