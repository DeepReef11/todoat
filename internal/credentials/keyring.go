package credentials

import (
	"errors"
	"fmt"
	"sync"

	"github.com/zalando/go-keyring"
)

// ErrKeyringNotAvailable is returned when the system keyring is not available
var ErrKeyringNotAvailable = errors.New("system keyring not available in this build")

// MockKeyring is a test implementation of the Keyring interface
type MockKeyring struct {
	mu    sync.RWMutex
	store map[string]map[string]string // service -> account -> password
}

// NewMockKeyring creates a new mock keyring for testing
func NewMockKeyring() *MockKeyring {
	return &MockKeyring{
		store: make(map[string]map[string]string),
	}
}

// Set stores a password in the mock keyring
func (m *MockKeyring) Set(service, account, password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.store[service] == nil {
		m.store[service] = make(map[string]string)
	}
	m.store[service][account] = password
	return nil
}

// Get retrieves a password from the mock keyring
func (m *MockKeyring) Get(service, account string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if accounts, ok := m.store[service]; ok {
		if password, ok := accounts[account]; ok {
			return password, nil
		}
	}
	return "", fmt.Errorf("password not found for %s/%s", service, account)
}

// Delete removes a password from the mock keyring
func (m *MockKeyring) Delete(service, account string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if accounts, ok := m.store[service]; ok {
		if _, ok := accounts[account]; ok {
			delete(accounts, account)
			return nil
		}
	}
	return fmt.Errorf("password not found for %s/%s", service, account)
}

// systemKeyring is the real keyring implementation using the OS keyring
type systemKeyring struct{}

// Set stores a password in the system keyring
func (s *systemKeyring) Set(service, account, password string) error {
	err := keyring.Set(service, account, password)
	if err != nil {
		// Check if it's a "not available" error
		if isKeyringNotAvailable(err) {
			return ErrKeyringNotAvailable
		}
		return err
	}
	return nil
}

// Get retrieves a password from the system keyring
func (s *systemKeyring) Get(service, account string) (string, error) {
	password, err := keyring.Get(service, account)
	if err != nil {
		if isKeyringNotAvailable(err) {
			return "", ErrKeyringNotAvailable
		}
		return "", err
	}
	return password, nil
}

// Delete removes a password from the system keyring
func (s *systemKeyring) Delete(service, account string) error {
	err := keyring.Delete(service, account)
	if err != nil {
		if isKeyringNotAvailable(err) {
			return ErrKeyringNotAvailable
		}
		return err
	}
	return nil
}

// isKeyringNotAvailable checks if the error indicates keyring is not available
func isKeyringNotAvailable(err error) bool {
	// go-keyring returns specific error types when keyring is not available
	// On Linux without D-Bus/Secret Service, it returns an error
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Common patterns for "keyring not available":
	// - "exec: \"dbus-launch\": executable file not found" (no D-Bus)
	// - "The name org.freedesktop.secrets was not provided" (no Secret Service)
	// - "Cannot autolaunch D-Bus without X11" (headless)
	return !errors.Is(err, keyring.ErrNotFound) &&
		(contains(errStr, "dbus") ||
			contains(errStr, "secrets") ||
			contains(errStr, "X11") ||
			contains(errStr, "not found") && contains(errStr, "executable"))
}

// contains is a simple helper for string containment check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
