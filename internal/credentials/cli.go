package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// CLIHandler handles CLI commands for credential management
type CLIHandler struct {
	manager *Manager
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
}

// NewCLIHandler creates a new CLI handler for credential commands
func NewCLIHandler(manager *Manager, stdin io.Reader, stdout, stderr io.Writer) *CLIHandler {
	return &CLIHandler{
		manager: manager,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
	}
}

// Set stores credentials in the keyring
// When prompt is true, it prompts for password input
func (h *CLIHandler) Set(backend, username string, prompt bool) error {
	var password string
	var err error

	if prompt {
		password, err = PromptPassword(h.stdin, h.stdout, backend, username)
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
	} else {
		return fmt.Errorf("--prompt flag is required for secure password input")
	}

	ctx := context.Background()
	err = h.manager.Set(ctx, backend, username, password)
	if err != nil {
		// Check if keyring is not available and provide helpful guidance
		if errors.Is(err, ErrKeyringNotAvailable) {
			return h.keyringNotAvailableError(backend)
		}
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	_, _ = fmt.Fprintf(h.stdout, "Credentials stored in system keyring\n")
	return nil
}

// keyringNotAvailableError returns a helpful error message when keyring is not available
func (h *CLIHandler) keyringNotAvailableError(backend string) error {
	upperBackend := strings.ToUpper(backend)

	msg := fmt.Sprintf(`System keyring not available in this build.

Alternative: Use environment variables instead:

For %s, set one of these environment variables:
  export TODOAT_%s_TOKEN="your-api-token"
  export TODOAT_%s_PASSWORD="your-password"

Environment variables are automatically detected by todoat.
Run 'todoat credentials list' to verify credentials are detected.

For more information, see: https://github.com/yourusername/todoat#credentials
`, backend, upperBackend, upperBackend)

	return errors.New(msg)
}

// Get retrieves and displays credential information
func (h *CLIHandler) Get(backend, username string, jsonOutput bool) error {
	ctx := context.Background()
	info, err := h.manager.Get(ctx, backend, username)
	if err != nil {
		return fmt.Errorf("failed to get credentials: %w", err)
	}

	if jsonOutput {
		return h.outputGetJSON(info)
	}

	return h.outputGetText(info)
}

// outputGetJSON outputs credential info as JSON
func (h *CLIHandler) outputGetJSON(info *CredentialInfo) error {
	jsonBytes, err := info.JSON()
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(h.stdout, string(jsonBytes))
	return nil
}

// outputGetText outputs credential info as text
func (h *CLIHandler) outputGetText(info *CredentialInfo) error {
	if !info.Found {
		_, _ = fmt.Fprintf(h.stdout, "No credentials found for %s/%s\n", info.Backend, info.Username)
		_, _ = fmt.Fprintf(h.stdout, "Searched:\n")
		_, _ = fmt.Fprintf(h.stdout, "  - System keyring: Not found\n")
		_, _ = fmt.Fprintf(h.stdout, "  - Environment variables: Not found\n")
		_, _ = fmt.Fprintf(h.stdout, "\nSuggestion: Run 'todoat credentials set %s %s --prompt'\n", info.Backend, info.Username)
		return nil
	}

	_, _ = fmt.Fprintf(h.stdout, "Source: %s\n", info.Source)
	_, _ = fmt.Fprintf(h.stdout, "Username: %s\n", info.Username)
	_, _ = fmt.Fprintf(h.stdout, "Password: ******** (hidden)\n")
	_, _ = fmt.Fprintf(h.stdout, "Backend: %s\n", info.Backend)
	_, _ = fmt.Fprintf(h.stdout, "Status: Available\n")
	return nil
}

// Delete removes credentials from the keyring
func (h *CLIHandler) Delete(backend, username string) error {
	ctx := context.Background()
	err := h.manager.Delete(ctx, backend, username)
	if err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	_, _ = fmt.Fprintf(h.stdout, "Credentials removed from system keyring\n")
	return nil
}

// List displays credential status for all configured backends
func (h *CLIHandler) List(backends []BackendConfig, jsonOutput bool) error {
	ctx := context.Background()
	statuses, err := h.manager.ListBackends(ctx, backends)
	if err != nil {
		return fmt.Errorf("failed to list credentials: %w", err)
	}

	if jsonOutput {
		return h.outputListJSON(statuses)
	}

	return h.outputListText(statuses)
}

// outputListJSON outputs backend statuses as JSON
func (h *CLIHandler) outputListJSON(statuses []BackendStatus) error {
	type statusJSON struct {
		Backend        string `json:"backend"`
		Username       string `json:"username"`
		HasCredentials bool   `json:"has_credentials"`
		Source         string `json:"source,omitempty"`
	}

	var output []statusJSON
	for _, s := range statuses {
		entry := statusJSON{
			Backend:        s.Backend,
			Username:       s.Username,
			HasCredentials: s.HasCredentials,
		}
		if s.HasCredentials {
			entry.Source = string(s.Source)
		}
		output = append(output, entry)
	}

	jsonBytes, err := json.Marshal(output)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(h.stdout, string(jsonBytes))
	return nil
}

// outputListText outputs backend statuses as text
func (h *CLIHandler) outputListText(statuses []BackendStatus) error {
	_, _ = fmt.Fprintf(h.stdout, "Backend Credentials:\n\n")
	_, _ = fmt.Fprintf(h.stdout, "%-20s %-20s %-15s %s\n", "BACKEND", "USERNAME", "STATUS", "SOURCE")

	for _, s := range statuses {
		status := "Not configured"
		source := "-"
		if s.HasCredentials {
			status = "Available"
			source = string(s.Source)
		}
		_, _ = fmt.Fprintf(h.stdout, "%-20s %-20s %-15s %s\n", s.Backend, s.Username, status, source)
	}

	return nil
}
