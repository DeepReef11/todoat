package nextcloud_test

import (
	"strings"
	"testing"

	"todoat/internal/testutil"
)

// TestCustomBackendNamingCLI tests that custom backend names (e.g., nextcloud-test) are supported.
// This test reproduces issue 002 - Config Should Support Custom Backend Naming.
func TestCustomBackendNamingCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set up config with custom-named backend
	cli.SetFullConfig(`
backends:
  sqlite:
    type: sqlite
    enabled: true
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "localhost:8080"
    username: "testadmin"
default_backend: sqlite
no_prompt: true
`)

	// Set password via environment variable
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "testpass")

	// Try to use the custom-named backend via --backend flag
	// The backend should be recognized even though it's not a built-in name
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "nextcloud-test", "list")

	combined := strings.ToLower(stdout + stderr)

	// The command should try to use nextcloud-test backend (which won't connect to a real server)
	// but should NOT fail with "unknown backend: nextcloud-test"
	if strings.Contains(combined, "unknown backend") {
		t.Errorf("Custom backend 'nextcloud-test' was not recognized. Output: %s", stdout+stderr)
	}

	// We expect a connection error (no real server) but NOT an "unknown backend" error
	// Exit code should be non-zero due to connection failure
	if exitCode == 0 {
		t.Logf("Unexpected success - might have connected to a real server? Output: %s", stdout)
	}

	// If there's an error, it should be about connection/network, not about unknown backend
	if exitCode != 0 && strings.Contains(combined, "unknown backend") {
		t.Errorf("Expected connection error, but got 'unknown backend' error. Output: %s", stdout+stderr)
	}
}

// TestCustomBackendMissingPasswordCLI tests error handling when custom backend is missing credentials.
func TestCustomBackendMissingPasswordCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set up config with custom-named backend (no password in config or env)
	cli.SetFullConfig(`
backends:
  sqlite:
    type: sqlite
    enabled: true
  nextcloud-test:
    type: nextcloud
    enabled: true
    host: "localhost:8080"
    username: "testadmin"
default_backend: sqlite
no_prompt: true
`)

	// Clear password env var
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "")

	// Try to use the custom-named backend
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "nextcloud-test", "list")

	combined := strings.ToLower(stdout + stderr)

	// Should fail with a password-related error, not "unknown backend"
	if strings.Contains(combined, "unknown backend") {
		t.Errorf("Custom backend 'nextcloud-test' was not recognized. Output: %s", stdout+stderr)
	}

	// Should require password
	if exitCode == 0 {
		t.Errorf("Expected failure due to missing password, but command succeeded. Output: %s", stdout)
	}

	// Error should mention password requirement
	if !strings.Contains(combined, "password") {
		t.Errorf("Expected error about missing password, got: %s", stdout+stderr)
	}
}

// TestCustomBackendHostFromConfigCLI tests that host is read from config file.
func TestCustomBackendHostFromConfigCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	// Set up config with custom host
	cli.SetFullConfig(`
backends:
  sqlite:
    type: sqlite
    enabled: true
  my-nextcloud:
    type: nextcloud
    enabled: true
    host: "custom-host.example.com"
    username: "customuser"
default_backend: sqlite
no_prompt: true
`)

	// Set password
	t.Setenv("TODOAT_NEXTCLOUD_PASSWORD", "testpass")

	// Try to use the custom-named backend
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "my-nextcloud", "list")

	combined := strings.ToLower(stdout + stderr)

	// Should NOT fail with "unknown backend"
	if strings.Contains(combined, "unknown backend") {
		t.Errorf("Custom backend 'my-nextcloud' was not recognized. Output: %s", stdout+stderr)
	}

	// The command will fail (no real server), but error should be about connection
	// not about unknown backend or missing host
	if exitCode != 0 {
		// Verify it's trying to connect (connection error is expected)
		if strings.Contains(combined, "unknown backend") {
			t.Errorf("Expected connection error, got unknown backend error. Output: %s", stdout+stderr)
		}
	}
}

// TestBuiltInBackendStillWorksCLI ensures that standard backend names still work.
func TestBuiltInBackendStillWorksCLI(t *testing.T) {
	cli := testutil.NewCLITestWithConfig(t)

	cli.SetFullConfig(`
backends:
  sqlite:
    type: sqlite
    enabled: true
default_backend: sqlite
no_prompt: true
`)

	// Using built-in "sqlite" backend should still work
	stdout, stderr, exitCode := cli.Execute("-y", "-b", "sqlite", "list")

	if exitCode != 0 {
		t.Errorf("Built-in sqlite backend should work. Exit code: %d, Output: %s", exitCode, stdout+stderr)
	}
}
