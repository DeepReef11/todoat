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
