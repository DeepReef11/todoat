package nextcloud_test

import (
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// TestNextcloudConfigSetDefaultBackendCLI verifies 'config set default_backend nextcloud' works
// This test reproduces issue #2 - nextcloud is not supported as default_backend
func TestNextcloudConfigSetDefaultBackendCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set nextcloud backend as enabled first
	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  nextcloud:
    enabled: true
default_backend: sqlite
no_prompt: true
`)

	// Now try to set default_backend to nextcloud
	stdout, stderr, exitCode := cli.Execute("-y", "config", "set", "default_backend", "nextcloud")

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", exitCode, stderr)
	}

	// Verify the output shows the change
	if !strings.Contains(stdout, "nextcloud") {
		t.Errorf("expected output to contain 'nextcloud', got: %s", stdout)
	}
}

// TestDefaultBackendNextcloudUsedCLI verifies that when default_backend is set to nextcloud,
// the CLI attempts to use Nextcloud backend (not SQLite).
// This test reproduces issue #2 - config doesn't seem to be read by app.
func TestDefaultBackendNextcloudUsedCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  nextcloud:
    enabled: true
default_backend: nextcloud
no_prompt: true
`)

	// Clear the nextcloud env vars to ensure we get a credentials error
	t.Setenv("TODOAT_NEXTCLOUD_HOST", "")
	t.Setenv("TODOAT_NEXTCLOUD_USERNAME", "")
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "")

	// When default_backend is nextcloud but no credentials, we should get an error
	// about Nextcloud credentials, NOT silently fall back to SQLite
	stdout, stderr, exitCode := cli.Execute("-y", "list")

	// With the bug (issue #2), this would succeed using SQLite silently
	// After the fix, this should fail with a Nextcloud credentials error
	combined := stdout + stderr

	if exitCode == 0 {
		// If it succeeds, check if it's using SQLite (bug) or Nextcloud
		// The test should fail if we see SQLite being used
		if !strings.Contains(combined, "nextcloud") && !strings.Contains(combined, "Nextcloud") {
			t.Errorf("default_backend is set to nextcloud but CLI silently used SQLite. Output: %s", combined)
		}
	} else {
		// If it fails, verify it's because of Nextcloud credentials (correct behavior)
		lowerCombined := strings.ToLower(combined)
		if !strings.Contains(lowerCombined, "nextcloud") && !strings.Contains(lowerCombined, "host") && !strings.Contains(lowerCombined, "credential") {
			t.Errorf("expected error about Nextcloud credentials, got: %s", combined)
		}
	}
}

// TestDefaultBackendNextcloudWithCredentialsCLI verifies that when default_backend is nextcloud
// and credentials are available, the CLI attempts to use Nextcloud backend.
func TestDefaultBackendNextcloudWithCredentialsCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    enabled: true
  nextcloud:
    enabled: true
default_backend: nextcloud
no_prompt: true
`)

	// Set fake credentials to trigger Nextcloud backend usage
	t.Setenv("TODOAT_NEXTCLOUD_HOST", "fake-test-host.local")
	t.Setenv("TODOAT_NEXTCLOUD_USERNAME", "testuser")
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "testpass")

	// When default_backend is nextcloud with credentials, it should try to use Nextcloud
	// The API call will fail (fake host), but the error should be about connectivity,
	// not about missing credentials or falling back to SQLite
	stdout, stderr, exitCode := cli.Execute("-y", "list")

	combined := stdout + stderr

	// The command should fail because the host is invalid (connection error)
	if exitCode == 0 {
		// If it succeeds, it might be using SQLite (bug behavior)
		// Check if output mentions Nextcloud-specific things
		if strings.Contains(combined, "No lists found") {
			t.Errorf("expected Nextcloud backend to be used (would fail with connection error), but SQLite empty response was returned")
		}
	} else {
		// If it fails, verify it's a connection error (correct behavior)
		// We should NOT see "No lists found" or SQLite-related messages
		if strings.Contains(combined, "No lists found") {
			t.Errorf("expected Nextcloud connection error, but got SQLite empty list response")
		}
	}
}

// TestBackendFlagNextcloudCLI verifies that --backend nextcloud flag works
func TestBackendFlagNextcloudCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Clear credentials
	t.Setenv("TODOAT_NEXTCLOUD_HOST", "")
	t.Setenv("TODOAT_NEXTCLOUD_USERNAME", "")
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "")

	// Try to use nextcloud backend via flag
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "nextcloud", "list")

	if exitCode == 0 {
		t.Errorf("expected non-zero exit code when using nextcloud without credentials, got 0")
	}

	// Should get an error about missing credentials or host
	combined := strings.ToLower(stdout + stderr)
	if !strings.Contains(combined, "nextcloud") {
		t.Errorf("expected error mentioning nextcloud, got: %s", stdout+stderr)
	}
}
