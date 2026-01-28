package config_test

import (
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// TestIssue31_CustomBackendNamesBroken reproduces issue #31 - custom backend names broken
// The issue reports that custom-named backends are not recognized by the CLI.
func TestIssue31_CustomBackendNamesBroken(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set up config EXACTLY as described in the issue - with custom backend under backends:
	cli.SetFullConfig(`
backends:
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "localhost:8080"
    username: "admin"
    allow_http: true
    insecure_skip_verify: true
    suppress_ssl_warning: true
default_backend: nextcloud-test
no_prompt: true
`)

	// Set password via environment variable
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "testpass")

	// Test case 1: Using -b flag with custom backend name
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "nextcloud-test", "list")
	combined := strings.ToLower(stdout + stderr)

	// Should NOT fail with "unknown backend" error
	if strings.Contains(combined, "unknown backend") {
		t.Errorf("FAIL: Custom backend 'nextcloud-test' via -b flag was not recognized. Output: %s", stdout+stderr)
	} else {
		t.Logf("PASS: -b nextcloud-test was recognized (exit code %d)", exitCode)
	}

	// Test case 2: Using default_backend setting
	stdout2, stderr2, exitCode2 := cli.Execute("-y", "list")
	combined2 := strings.ToLower(stdout2 + stderr2)

	if strings.Contains(combined2, "unknown backend") {
		t.Errorf("FAIL: Custom backend 'nextcloud-test' via default_backend was not recognized. Output: %s", stdout2+stderr2)
	} else {
		t.Logf("PASS: default_backend=nextcloud-test was recognized (exit code %d)", exitCode2)
	}
}

// TestIssue31_TopLevelBackendConfig tests if custom backends at the top level of config work
// This follows the exact format from the issue report.
func TestIssue31_TopLevelBackendConfig(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set up config EXACTLY as described in the issue - at TOP LEVEL (not under backends:)
	// This is how the issue report shows it
	cli.SetFullConfig(`
nextcloud-test:
  type: nextcloud
  enabled: true
  host: "localhost:8080"
  username: "admin"
  allow_http: true
  insecure_skip_verify: true
  suppress_ssl_warning: true
default_backend: nextcloud-test
no_prompt: true
`)

	// Set password via environment variable
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "testpass")

	// Test case: Using -b flag with custom backend name
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "nextcloud-test", "list")
	combined := strings.ToLower(stdout + stderr)

	t.Logf("Output: %s", stdout+stderr)
	t.Logf("Exit code: %d", exitCode)

	// Should NOT fail with "unknown backend" error
	if strings.Contains(combined, "unknown backend") {
		t.Errorf("FAIL: Custom backend 'nextcloud-test' at top level was not recognized. Output: %s", stdout+stderr)
	}
}
