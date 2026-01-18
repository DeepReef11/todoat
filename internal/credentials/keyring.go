package credentials

import (
	"fmt"
	"sync"
)

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
	// TODO: Use zalando/go-keyring or similar for production
	// For now, return an error indicating keyring is not available
	return fmt.Errorf("system keyring not available in this build")
}

// Get retrieves a password from the system keyring
func (s *systemKeyring) Get(service, account string) (string, error) {
	// TODO: Use zalando/go-keyring or similar for production
	return "", fmt.Errorf("system keyring not available in this build")
}

// Delete removes a password from the system keyring
func (s *systemKeyring) Delete(service, account string) error {
	// TODO: Use zalando/go-keyring or similar for production
	return fmt.Errorf("system keyring not available in this build")
}
