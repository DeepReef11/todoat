package nextcloud_test

import (
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// TestIssue012HTTPConfigInconsistency verifies that the built-in backend name "nextcloud"
// and custom backends with "type: nextcloud" respect the same config file settings.
//
// Issue: When two backends are configured identically in the config file:
//   - "nextcloud" (built-in name) ignores config file settings like allow_http
//   - "nextcloud-test" (custom name with type: nextcloud) respects these settings
//
// Both should behave identically since they have the same configuration.
func TestIssue012HTTPConfigInconsistency(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Configure two backends identically - one using built-in name "nextcloud",
	// the other using custom name "nextcloud-test" with type: nextcloud
	//
	// Both have allow_http: true, which should mean HTTP is used for both.
	cli.SetFullConfig(`
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    host: "localhost:8080"
    username: "admin"
    password: "testpass"
    allow_http: true
    insecure_skip_verify: true
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "localhost:8080"
    username: "admin"
    password: "testpass"
    allow_http: true
    insecure_skip_verify: true
default_backend: sqlite
no_prompt: true
`)

	// Clear environment variables to ensure config file is the only source of settings
	t.Setenv("TODOAT_NEXTCLOUD_HOST", "")
	t.Setenv("TODOAT_NEXTCLOUD_USERNAME", "")
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "")

	// Test with built-in name "nextcloud"
	_, stderrBuiltin, exitCodeBuiltin := cli.Execute("-y", "-b", "nextcloud", "list")
	errBuiltin := strings.ToLower(stderrBuiltin)

	// Test with custom name "nextcloud-test"
	_, stderrCustom, exitCodeCustom := cli.Execute("-y", "-b", "nextcloud-test", "list")
	errCustom := strings.ToLower(stderrCustom)

	// BUG behavior: "-b nextcloud" requires env vars and ignores config file entirely:
	//   "nextcloud backend requires TODOAT_NEXTCLOUD_HOST environment variable"
	//
	// Expected behavior: "-b nextcloud" should read from config file (same as "-b nextcloud-test")
	//   and fail with connection refused (not env var error)

	envVarError := "environment variable"

	if strings.Contains(errBuiltin, envVarError) && !strings.Contains(errCustom, envVarError) {
		t.Errorf("BUG: '-b nextcloud' requires env vars while '-b nextcloud-test' reads from config file.\n"+
			"Both should read config file since both are configured identically.\n"+
			"'-b nextcloud' stderr: %s\n"+
			"'-b nextcloud-test' stderr: %s", stderrBuiltin, stderrCustom)
	}

	// Both should behave the same since they have identical config
	httpsError := "server gave http response to https client"
	if strings.Contains(errBuiltin, httpsError) {
		t.Errorf("BUG: '-b nextcloud' ignores allow_http:true from config file.\n"+
			"Expected HTTP connection (same as '-b nextcloud-test'), but got HTTPS error.\n"+
			"stderr: %s\nexit code: %d", stderrBuiltin, exitCodeBuiltin)
	}

	if strings.Contains(errCustom, httpsError) {
		t.Errorf("Unexpected: '-b nextcloud-test' also got HTTPS error despite allow_http:true.\n"+
			"stderr: %s\nexit code: %d", stderrCustom, exitCodeCustom)
	}

	// Additional verification: both exit codes should be the same
	if exitCodeBuiltin != exitCodeCustom {
		t.Logf("Note: Exit codes differ - builtin=%d, custom=%d. "+
			"This may indicate different behavior between the two backends.", exitCodeBuiltin, exitCodeCustom)
	}
}

// TestIssue012BuiltinNextcloudReadsConfigFile verifies that when using "-b nextcloud"
// with config file settings (not just env vars), the config file is respected.
func TestIssue012BuiltinNextcloudReadsConfigFile(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set up config with credentials and allow_http in config file only
	cli.SetFullConfig(`
backends:
  nextcloud:
    type: nextcloud
    enabled: true
    host: "localhost:9999"
    username: "configuser"
    password: "configpass"
    allow_http: true
default_backend: sqlite
no_prompt: true
`)

	// Clear environment variables to ensure config file is the only source
	t.Setenv("TODOAT_NEXTCLOUD_HOST", "")
	t.Setenv("TODOAT_NEXTCLOUD_USERNAME", "")
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "")

	// Use the built-in nextcloud backend
	_, stderr, _ := cli.Execute("-y", "-b", "nextcloud", "list")
	errLower := strings.ToLower(stderr)

	// If the config file is being read, we should NOT get:
	// - "requires TODOAT_NEXTCLOUD_HOST environment variable" (config has host)
	// - "server gave HTTP response to HTTPS client" (config has allow_http: true)

	if strings.Contains(errLower, "environment variable") {
		t.Errorf("BUG: '-b nextcloud' is not reading credentials from config file.\n"+
			"Expected to use host/username/password from config file, but got env var error.\n"+
			"stderr: %s", stderr)
	}

	if strings.Contains(errLower, "server gave http response to https client") {
		t.Errorf("BUG: '-b nextcloud' is not reading allow_http from config file.\n"+
			"Expected HTTP connection (allow_http: true), but got HTTPS error.\n"+
			"stderr: %s", stderr)
	}
}
