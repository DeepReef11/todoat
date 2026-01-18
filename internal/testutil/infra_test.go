// Package testutil provides shared test utilities.
// infra_test.go contains tests for project infrastructure (docker-compose, Makefile).
package testutil

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// getProjectRoot returns the project root directory.
func getProjectRoot(t *testing.T) string {
	t.Helper()
	// Get the path of this test file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get test file path")
	}
	// Navigate from internal/testutil/ to project root
	testDir := filepath.Dir(filename)
	return filepath.Join(testDir, "..", "..")
}

// =============================================================================
// Infrastructure Tests (030-integration-test-infrastructure)
// =============================================================================

// TestDockerComposeExists verifies docker-compose.yml file exists in project root
func TestDockerComposeExists(t *testing.T) {
	projectRoot := getProjectRoot(t)
	dockerComposePath := filepath.Join(projectRoot, "docker-compose.yml")

	if _, err := os.Stat(dockerComposePath); os.IsNotExist(err) {
		t.Errorf("docker-compose.yml not found at %s", dockerComposePath)
	}

	// Verify it contains essential services
	content, err := os.ReadFile(dockerComposePath)
	if err != nil {
		t.Fatalf("failed to read docker-compose.yml: %v", err)
	}

	contentStr := string(content)

	// Check for nextcloud service
	if !strings.Contains(contentStr, "nextcloud") {
		t.Error("docker-compose.yml should contain nextcloud service")
	}

	// Check for healthcheck
	if !strings.Contains(contentStr, "healthcheck") {
		t.Error("docker-compose.yml should contain healthcheck for nextcloud")
	}

	// Check for environment variables
	if !strings.Contains(contentStr, "NEXTCLOUD_ADMIN_USER") {
		t.Error("docker-compose.yml should define NEXTCLOUD_ADMIN_USER")
	}
	if !strings.Contains(contentStr, "NEXTCLOUD_ADMIN_PASSWORD") {
		t.Error("docker-compose.yml should define NEXTCLOUD_ADMIN_PASSWORD")
	}
}

// TestMakefileDockerUp verifies `make docker-up` target exists and is documented
func TestMakefileDockerUp(t *testing.T) {
	projectRoot := getProjectRoot(t)
	makefilePath := filepath.Join(projectRoot, "Makefile")

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("failed to read Makefile: %v", err)
	}

	contentStr := string(content)

	// Check for docker-up target
	if !regexp.MustCompile(`^docker-up:`).MatchString(contentStr) &&
		!regexp.MustCompile(`\ndocker-up:`).MatchString(contentStr) {
		t.Error("Makefile should contain docker-up target")
	}

	// Check that docker-up uses docker-compose
	if !strings.Contains(contentStr, "docker-compose up") && !strings.Contains(contentStr, "docker compose up") {
		t.Error("docker-up target should use docker-compose up")
	}
}

// TestMakefileDockerDown verifies `make docker-down` target exists
func TestMakefileDockerDown(t *testing.T) {
	projectRoot := getProjectRoot(t)
	makefilePath := filepath.Join(projectRoot, "Makefile")

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("failed to read Makefile: %v", err)
	}

	contentStr := string(content)

	// Check for docker-down target
	if !regexp.MustCompile(`^docker-down:`).MatchString(contentStr) &&
		!regexp.MustCompile(`\ndocker-down:`).MatchString(contentStr) {
		t.Error("Makefile should contain docker-down target")
	}

	// Check that docker-down uses docker-compose
	if !strings.Contains(contentStr, "docker-compose down") && !strings.Contains(contentStr, "docker compose down") {
		t.Error("docker-down target should use docker-compose down")
	}
}

// TestMakefileTestIntegration verifies `make test-integration` target exists
func TestMakefileTestIntegration(t *testing.T) {
	projectRoot := getProjectRoot(t)
	makefilePath := filepath.Join(projectRoot, "Makefile")

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("failed to read Makefile: %v", err)
	}

	contentStr := string(content)

	// Check for test-integration target
	if !regexp.MustCompile(`^test-integration:`).MatchString(contentStr) &&
		!regexp.MustCompile(`\ntest-integration:`).MatchString(contentStr) {
		t.Error("Makefile should contain test-integration target")
	}

	// Check that test-integration uses integration tag
	if !strings.Contains(contentStr, "-tags=integration") {
		t.Error("test-integration target should use -tags=integration")
	}
}

// TestMakefileTestNextcloud verifies `make test-nextcloud` target exists
func TestMakefileTestNextcloud(t *testing.T) {
	projectRoot := getProjectRoot(t)
	makefilePath := filepath.Join(projectRoot, "Makefile")

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("failed to read Makefile: %v", err)
	}

	contentStr := string(content)

	// Check for test-nextcloud target
	if !regexp.MustCompile(`^test-nextcloud:`).MatchString(contentStr) &&
		!regexp.MustCompile(`\ntest-nextcloud:`).MatchString(contentStr) {
		t.Error("Makefile should contain test-nextcloud target")
	}

	// Check that test-nextcloud targets nextcloud package
	if !strings.Contains(contentStr, "backend/nextcloud") {
		t.Error("test-nextcloud target should target backend/nextcloud package")
	}
}

// TestMakefileTestTodoist verifies `make test-todoist` target exists
func TestMakefileTestTodoist(t *testing.T) {
	projectRoot := getProjectRoot(t)
	makefilePath := filepath.Join(projectRoot, "Makefile")

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("failed to read Makefile: %v", err)
	}

	contentStr := string(content)

	// Check for test-todoist target
	if !regexp.MustCompile(`^test-todoist:`).MatchString(contentStr) &&
		!regexp.MustCompile(`\ntest-todoist:`).MatchString(contentStr) {
		t.Error("Makefile should contain test-todoist target")
	}

	// Check that test-todoist targets todoist package
	if !strings.Contains(contentStr, "backend/todoist") {
		t.Error("test-todoist target should target backend/todoist package")
	}
}
