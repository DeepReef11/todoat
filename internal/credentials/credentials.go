// Package credentials provides secure credential storage and retrieval
// for backend services (Nextcloud, Todoist, etc.) using OS-native keyrings
// with fallback to environment variables.
package credentials

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Source indicates where credentials were retrieved from
type Source string

const (
	SourceKeyring     Source = "keyring"
	SourceEnvironment Source = "environment"
	SourceConfigURL   Source = "config_url"
	SourceNone        Source = "none"
)

// CredentialInfo contains credential information returned by Get()
type CredentialInfo struct {
	Source   Source // Where credentials came from
	Backend  string // Backend name (e.g., "nextcloud")
	Username string // Username/account identifier
	Password string // Password (masked in display)
	Found    bool   // Whether credentials were found
}

// JSON serializes the credential info to JSON (password excluded for security)
func (c *CredentialInfo) JSON() ([]byte, error) {
	output := struct {
		Backend  string `json:"backend"`
		Username string `json:"username"`
		Source   string `json:"source"`
		Found    bool   `json:"found"`
	}{
		Backend:  c.Backend,
		Username: c.Username,
		Source:   string(c.Source),
		Found:    c.Found,
	}
	return json.Marshal(output)
}

// BackendConfig represents a backend configuration for listing
type BackendConfig struct {
	Name     string
	Username string
}

// BackendStatus represents the credential status for a backend
type BackendStatus struct {
	Backend        string
	Username       string
	HasCredentials bool
	Source         Source
}

// Keyring is the interface for keyring operations
type Keyring interface {
	Set(service, account, password string) error
	Get(service, account string) (string, error)
	Delete(service, account string) error
}

// Manager handles credential operations
type Manager struct {
	keyring Keyring
}

// ManagerOption is a functional option for Manager
type ManagerOption func(*Manager)

// WithKeyring sets a custom keyring implementation
func WithKeyring(k Keyring) ManagerOption {
	return func(m *Manager) {
		m.keyring = k
	}
}

// NewManager creates a new credential manager
func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{
		keyring: &systemKeyring{},
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// normalizeBackend normalizes backend names to lowercase
func normalizeBackend(backend string) string {
	return strings.ToLower(strings.TrimSpace(backend))
}

// serviceName returns the keyring service name for a backend
func serviceName(backend string) string {
	return fmt.Sprintf("todoat-%s", normalizeBackend(backend))
}

// Set stores credentials in the keyring
func (m *Manager) Set(ctx context.Context, backend, username, password string) error {
	backend = normalizeBackend(backend)
	service := serviceName(backend)
	return m.keyring.Set(service, username, password)
}

// Get retrieves credentials from available sources (keyring first, then env vars)
func (m *Manager) Get(ctx context.Context, backend, username string) (*CredentialInfo, error) {
	backend = normalizeBackend(backend)

	// Priority 1: Try keyring
	service := serviceName(backend)
	password, err := m.keyring.Get(service, username)
	if err == nil && password != "" {
		return &CredentialInfo{
			Source:   SourceKeyring,
			Backend:  backend,
			Username: username,
			Password: password,
			Found:    true,
		}, nil
	}

	// Priority 2: Try environment variables
	envPassword := m.getEnvPassword(backend, username)
	if envPassword != "" {
		return &CredentialInfo{
			Source:   SourceEnvironment,
			Backend:  backend,
			Username: username,
			Password: envPassword,
			Found:    true,
		}, nil
	}

	// Not found
	return &CredentialInfo{
		Source:   SourceNone,
		Backend:  backend,
		Username: username,
		Found:    false,
	}, nil
}

// getEnvPassword gets password/token from environment variables
func (m *Manager) getEnvPassword(backend, username string) string {
	upperBackend := strings.ToUpper(backend)

	// Priority 1: Check for TOKEN env var (e.g., TODOAT_TODOIST_TOKEN)
	// This is the preferred pattern for API tokens (Todoist, Google, etc.)
	tokenKey := fmt.Sprintf("TODOAT_%s_TOKEN", upperBackend)
	token := os.Getenv(tokenKey)
	if token != "" {
		return token
	}

	// Priority 2: Check for PASSWORD env var (e.g., TODOAT_NEXTCLOUD_PASSWORD)
	// This is used for username/password authentication (Nextcloud, etc.)
	envKey := fmt.Sprintf("TODOAT_%s_PASSWORD", upperBackend)
	password := os.Getenv(envKey)

	// Also check username matches if set
	userEnvKey := fmt.Sprintf("TODOAT_%s_USERNAME", upperBackend)
	envUsername := os.Getenv(userEnvKey)

	// If username is set in env and doesn't match, return empty
	if envUsername != "" && envUsername != username {
		return ""
	}

	return password
}

// Delete removes credentials from the keyring
func (m *Manager) Delete(ctx context.Context, backend, username string) error {
	backend = normalizeBackend(backend)
	service := serviceName(backend)

	err := m.keyring.Delete(service, username)
	// Idempotent: return nil if not found
	if err != nil && strings.Contains(err.Error(), "not found") {
		return nil
	}
	return err
}

// ListBackends returns the credential status for each configured backend
func (m *Manager) ListBackends(ctx context.Context, backends []BackendConfig) ([]BackendStatus, error) {
	var statuses []BackendStatus

	for _, bc := range backends {
		info, err := m.Get(ctx, bc.Name, bc.Username)
		if err != nil {
			return nil, err
		}

		statuses = append(statuses, BackendStatus{
			Backend:        bc.Name,
			Username:       bc.Username,
			HasCredentials: info.Found,
			Source:         info.Source,
		})
	}

	return statuses, nil
}

// PromptPassword prompts the user for a password with hidden input
// In production, this uses terminal.ReadPassword for actual hidden input
// For testing, it reads from the provided reader
func PromptPassword(reader io.Reader, writer io.Writer, backend, username string) (string, error) {
	_, _ = fmt.Fprintf(writer, "Enter password for %s (user: %s): ", backend, username)

	// For non-TTY input (testing), just read a line
	scanner := bufio.NewScanner(reader)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("no input received")
}
