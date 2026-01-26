// Package credentials provides secure credential storage and retrieval
package credentials

import (
	"context"
	"os"
	"testing"
)

// TestIssue010KeyringCredentialE2EFlow tests the complete end-to-end flow:
// 1. Store credentials via Manager.Set (simulating `credentials set`)
// 2. Retrieve credentials with username from config
// 3. Verify the password is found from keyring
//
// This is a regression test for issue #010 - integration tests don't verify
// real keyring credential flow. The test ensures that:
// - Credentials stored with a specific backend name and username can be retrieved
// - The retrieval works correctly when username matches the stored account
// - Backend name normalization is consistent between set and get
func TestIssue010KeyringCredentialE2EFlow(t *testing.T) {
	// Use mock keyring for testing (real keyring not available in CI)
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Step 1: Store credentials (simulates: todoat credentials set nextcloud-test admin --password secret)
	backendName := "nextcloud-test"
	username := "admin"
	password := "secretPassword123"

	err := manager.Set(context.Background(), backendName, username, password)
	if err != nil {
		t.Fatalf("Failed to store credentials: %v", err)
	}

	// Verify credentials were stored with correct service name
	// Service name format: todoat-{backend}
	storedPassword, err := mockKeyring.Get("todoat-nextcloud-test", username)
	if err != nil {
		t.Fatalf("Credentials not stored correctly in keyring: %v", err)
	}
	if storedPassword != password {
		t.Errorf("Stored password mismatch: got %q, want %q", storedPassword, password)
	}

	// Step 2: Retrieve credentials using the same backend name and username
	// This simulates what buildNextcloudConfigWithKeyring does
	credInfo, err := manager.Get(context.Background(), backendName, username)
	if err != nil {
		t.Fatalf("Failed to retrieve credentials: %v", err)
	}

	// Step 3: Verify the credentials were retrieved correctly
	if !credInfo.Found {
		t.Error("Expected credentials to be found, but Found=false")
	}
	if credInfo.Source != SourceKeyring {
		t.Errorf("Expected source %q, got %q", SourceKeyring, credInfo.Source)
	}
	if credInfo.Username != username {
		t.Errorf("Expected username %q, got %q", username, credInfo.Username)
	}
	if credInfo.Password != password {
		t.Errorf("Expected password %q, got %q", password, credInfo.Password)
	}
	if credInfo.Backend != backendName {
		t.Errorf("Expected backend %q, got %q", backendName, credInfo.Backend)
	}
}

// TestIssue010ConfigKeyringInteraction tests the interaction between config
// username and keyring lookup. This verifies that:
// - Username from config is used correctly for keyring lookup
// - Missing username in config results in failed keyring lookup
// - Username mismatch between config and keyring fails lookup
func TestIssue010ConfigKeyringInteraction(t *testing.T) {
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Store credentials with specific username
	backendName := "nextcloud"
	storedUsername := "configuser"
	password := "configPassword123"

	err := manager.Set(context.Background(), backendName, storedUsername, password)
	if err != nil {
		t.Fatalf("Failed to store credentials: %v", err)
	}

	// Test case 1: Config username matches keyring - should succeed
	t.Run("matching_username", func(t *testing.T) {
		credInfo, err := manager.Get(context.Background(), backendName, storedUsername)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if !credInfo.Found {
			t.Error("Expected credentials to be found with matching username")
		}
		if credInfo.Password != password {
			t.Errorf("Expected password %q, got %q", password, credInfo.Password)
		}
	})

	// Test case 2: Config username doesn't match keyring - should not find (no env fallback)
	t.Run("mismatched_username", func(t *testing.T) {
		// Clear any environment variables that could interfere
		origPass := os.Getenv("TODOAT_NEXTCLOUD_PASSWORD")
		defer func() { _ = os.Setenv("TODOAT_NEXTCLOUD_PASSWORD", origPass) }()
		_ = os.Unsetenv("TODOAT_NEXTCLOUD_PASSWORD")

		credInfo, err := manager.Get(context.Background(), backendName, "wronguser")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if credInfo.Found {
			t.Error("Expected credentials NOT to be found with mismatched username")
		}
	})

	// Test case 3: Empty config username - should not find keyring credentials
	t.Run("empty_username", func(t *testing.T) {
		// Clear environment to ensure no fallback
		origPass := os.Getenv("TODOAT_NEXTCLOUD_PASSWORD")
		defer func() { _ = os.Setenv("TODOAT_NEXTCLOUD_PASSWORD", origPass) }()
		_ = os.Unsetenv("TODOAT_NEXTCLOUD_PASSWORD")

		credInfo, err := manager.Get(context.Background(), backendName, "")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		// With empty username, keyring lookup will fail (different account)
		// and env fallback won't find anything either
		if credInfo.Found && credInfo.Source == SourceKeyring {
			t.Error("Expected keyring lookup to fail with empty username")
		}
	})
}

// TestIssue010BackendNameConsistency tests that backend names are normalized
// consistently between credential set and get operations.
func TestIssue010BackendNameConsistency(t *testing.T) {
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	username := "testuser"
	password := "testPassword"

	testCases := []struct {
		name        string
		setBackend  string
		getBackend  string
		shouldMatch bool
	}{
		{
			name:        "exact_match",
			setBackend:  "nextcloud-test",
			getBackend:  "nextcloud-test",
			shouldMatch: true,
		},
		{
			name:        "case_insensitive_lowercase",
			setBackend:  "NEXTCLOUD-TEST",
			getBackend:  "nextcloud-test",
			shouldMatch: true,
		},
		{
			name:        "case_insensitive_uppercase",
			setBackend:  "nextcloud-test",
			getBackend:  "NEXTCLOUD-TEST",
			shouldMatch: true,
		},
		{
			name:        "mixed_case",
			setBackend:  "NextCloud-Test",
			getBackend:  "nextcloud-test",
			shouldMatch: true,
		},
		{
			name:        "different_backend",
			setBackend:  "nextcloud-test",
			getBackend:  "nextcloud-prod",
			shouldMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear any previous state
			_ = manager.Delete(context.Background(), tc.setBackend, username)
			_ = manager.Delete(context.Background(), tc.getBackend, username)

			// Store credentials
			err := manager.Set(context.Background(), tc.setBackend, username, password)
			if err != nil {
				t.Fatalf("Failed to set credentials: %v", err)
			}

			// Retrieve credentials
			credInfo, err := manager.Get(context.Background(), tc.getBackend, username)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}

			if tc.shouldMatch {
				if !credInfo.Found {
					t.Errorf("Expected credentials to be found (set with %q, get with %q)",
						tc.setBackend, tc.getBackend)
				}
				if credInfo.Password != password {
					t.Errorf("Password mismatch: got %q, want %q", credInfo.Password, password)
				}
			} else {
				if credInfo.Found && credInfo.Source == SourceKeyring {
					t.Errorf("Expected credentials NOT to be found (set with %q, get with %q)",
						tc.setBackend, tc.getBackend)
				}
			}
		})
	}
}

// TestIssue010CustomBackendCredentialFlow tests the credential flow for custom
// backend names (e.g., "work-nextcloud" instead of just "nextcloud").
func TestIssue010CustomBackendCredentialFlow(t *testing.T) {
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	// Test with custom backend names that might be used in config files
	customBackends := []struct {
		backendName string
		username    string
	}{
		{"work-nextcloud", "work-user"},
		{"personal-nextcloud", "personal-user"},
		{"my-todoist", "token"},
		{"company-google", "token"},
	}

	password := "customPassword123"

	for _, cb := range customBackends {
		t.Run(cb.backendName, func(t *testing.T) {
			// Store credentials
			err := manager.Set(context.Background(), cb.backendName, cb.username, password)
			if err != nil {
				t.Fatalf("Failed to set credentials for %s: %v", cb.backendName, err)
			}

			// Verify storage format
			expectedService := "todoat-" + cb.backendName
			storedPass, err := mockKeyring.Get(expectedService, cb.username)
			if err != nil {
				t.Fatalf("Credentials not stored with expected service name %q: %v",
					expectedService, err)
			}
			if storedPass != password {
				t.Errorf("Stored password mismatch: got %q, want %q", storedPass, password)
			}

			// Retrieve and verify
			credInfo, err := manager.Get(context.Background(), cb.backendName, cb.username)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}
			if !credInfo.Found {
				t.Error("Expected credentials to be found")
			}
			if credInfo.Backend != cb.backendName {
				t.Errorf("Backend name mismatch: got %q, want %q",
					credInfo.Backend, cb.backendName)
			}
		})
	}
}

// TestIssue010KeyringEnvironmentPriority tests that keyring takes priority
// over environment variables when both are available.
func TestIssue010KeyringEnvironmentPriority(t *testing.T) {
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	backendName := "nextcloud"
	username := "priorityuser"
	keyringPassword := "keyringPass"
	envPassword := "envPass"

	// Store in keyring
	err := manager.Set(context.Background(), backendName, username, keyringPassword)
	if err != nil {
		t.Fatalf("Failed to set credentials: %v", err)
	}

	// Set environment variable
	origPass := os.Getenv("TODOAT_NEXTCLOUD_PASSWORD")
	origUser := os.Getenv("TODOAT_NEXTCLOUD_USERNAME")
	defer func() {
		_ = os.Setenv("TODOAT_NEXTCLOUD_PASSWORD", origPass)
		_ = os.Setenv("TODOAT_NEXTCLOUD_USERNAME", origUser)
	}()
	_ = os.Setenv("TODOAT_NEXTCLOUD_PASSWORD", envPassword)
	_ = os.Setenv("TODOAT_NEXTCLOUD_USERNAME", username)

	// Retrieve - keyring should win
	credInfo, err := manager.Get(context.Background(), backendName, username)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !credInfo.Found {
		t.Fatal("Expected credentials to be found")
	}
	if credInfo.Source != SourceKeyring {
		t.Errorf("Expected source %q (keyring should take priority), got %q",
			SourceKeyring, credInfo.Source)
	}
	if credInfo.Password != keyringPassword {
		t.Errorf("Expected keyring password %q, got %q (env password: %q)",
			keyringPassword, credInfo.Password, envPassword)
	}
}

// TestIssue010DeleteCredentialFlow tests that deleted credentials are properly
// removed and subsequent lookups fail appropriately.
func TestIssue010DeleteCredentialFlow(t *testing.T) {
	mockKeyring := NewMockKeyring()
	manager := NewManager(WithKeyring(mockKeyring))

	backendName := "nextcloud-delete-test"
	username := "deleteuser"
	password := "deletePassword"

	// Clear environment to isolate keyring behavior
	origPass := os.Getenv("TODOAT_NEXTCLOUD-DELETE-TEST_PASSWORD")
	defer func() { _ = os.Setenv("TODOAT_NEXTCLOUD-DELETE-TEST_PASSWORD", origPass) }()
	_ = os.Unsetenv("TODOAT_NEXTCLOUD-DELETE-TEST_PASSWORD")

	// Store credentials
	err := manager.Set(context.Background(), backendName, username, password)
	if err != nil {
		t.Fatalf("Failed to set credentials: %v", err)
	}

	// Verify stored
	credInfo, err := manager.Get(context.Background(), backendName, username)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !credInfo.Found {
		t.Fatal("Expected credentials to be found after set")
	}

	// Delete credentials
	err = manager.Delete(context.Background(), backendName, username)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted - should not find from keyring
	credInfo, err = manager.Get(context.Background(), backendName, username)
	if err != nil {
		t.Fatalf("Get after delete failed: %v", err)
	}
	if credInfo.Found && credInfo.Source == SourceKeyring {
		t.Error("Expected credentials NOT to be found from keyring after delete")
	}
}
