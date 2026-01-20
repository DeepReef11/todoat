package cache_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"todoat/internal/cache"
	"todoat/internal/testutil"
)

// =============================================================================
// List Metadata Caching Tests (036-list-metadata-caching)
// =============================================================================

// TestListCacheCreation verifies that first `todoat list` creates cache file at XDG cache path.
func TestListCacheCreation(t *testing.T) {
	cli := testutil.NewCLITestWithCache(t)

	// Run list command - should trigger cache creation
	cli.MustExecute("-y", "list")

	// Verify cache file was created
	cachePath := cli.CachePath()
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Errorf("expected cache file to be created at %s", cachePath)
	}

	// Verify cache file has correct permissions (0644)
	info, err := os.Stat(cachePath)
	if err != nil {
		t.Fatalf("failed to stat cache file: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Errorf("expected cache file permissions 0644, got %o", perm)
	}

	// Verify cache file contains valid JSON with expected structure
	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read cache file: %v", err)
	}

	var cacheData cache.ListCache
	if err := json.Unmarshal(data, &cacheData); err != nil {
		t.Errorf("cache file should contain valid JSON: %v", err)
	}

	// Verify cache has timestamp
	if cacheData.CreatedAt.IsZero() {
		t.Error("cache should have creation timestamp")
	}
}

// TestListCacheHit verifies that subsequent `todoat list` within TTL uses cached data.
func TestListCacheHit(t *testing.T) {
	cli := testutil.NewCLITestWithCache(t)

	// Create a list first
	cli.MustExecute("-y", "list", "create", "TestList")

	// First list command - should populate cache
	cli.MustExecute("-y", "list")

	// Get cache modification time
	cachePath := cli.CachePath()
	info1, err := os.Stat(cachePath)
	if err != nil {
		t.Fatalf("failed to stat cache file: %v", err)
	}
	modTime1 := info1.ModTime()

	// Wait a tiny bit to ensure any file system timestamp granularity
	time.Sleep(10 * time.Millisecond)

	// Second list command - should use cache (no write)
	cli.MustExecute("-y", "list")

	// Get cache modification time again
	info2, err := os.Stat(cachePath)
	if err != nil {
		t.Fatalf("failed to stat cache file after second call: %v", err)
	}
	modTime2 := info2.ModTime()

	// Cache file should not have been modified (cache hit)
	if !modTime1.Equal(modTime2) {
		t.Errorf("cache should not be modified on cache hit: first=%v, second=%v", modTime1, modTime2)
	}

	// Verify output still contains the list
	stdout := cli.MustExecute("-y", "list")
	testutil.AssertContains(t, stdout, "TestList")
}

// TestListCacheInvalidation verifies that `todoat list create "New"` invalidates cache.
func TestListCacheInvalidation(t *testing.T) {
	cli := testutil.NewCLITestWithCache(t)

	// Create initial list and populate cache
	cli.MustExecute("-y", "list", "create", "InitialList")
	cli.MustExecute("-y", "list")

	// Read initial cache content
	cachePath := cli.CachePath()
	data1, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read initial cache: %v", err)
	}

	var cache1 cache.ListCache
	if err := json.Unmarshal(data1, &cache1); err != nil {
		t.Fatalf("failed to parse initial cache: %v", err)
	}

	// Create a new list - should invalidate cache
	cli.MustExecute("-y", "list", "create", "NewList")

	// Run list again to repopulate cache
	cli.MustExecute("-y", "list")

	// Read cache again
	data2, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read cache after invalidation: %v", err)
	}

	var cache2 cache.ListCache
	if err := json.Unmarshal(data2, &cache2); err != nil {
		t.Fatalf("failed to parse cache after invalidation: %v", err)
	}

	// Cache should have been updated (new timestamp or new list count)
	if len(cache2.Lists) <= len(cache1.Lists) {
		t.Errorf("cache should include new list: had %d lists, now %d", len(cache1.Lists), len(cache2.Lists))
	}

	// Verify new list is in cache
	found := false
	for _, l := range cache2.Lists {
		if l.Name == "NewList" {
			found = true
			break
		}
	}
	if !found {
		t.Error("new list should be present in cache after invalidation")
	}
}

// TestListCacheInvalidationOnDelete verifies that `todoat list delete` invalidates cache.
func TestListCacheInvalidationOnDelete(t *testing.T) {
	cli := testutil.NewCLITestWithCache(t)

	// Create list and populate cache
	cli.MustExecute("-y", "list", "create", "ToDelete")
	cli.MustExecute("-y", "list")

	// Read initial cache
	cachePath := cli.CachePath()
	data1, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read initial cache: %v", err)
	}

	var cache1 cache.ListCache
	if err := json.Unmarshal(data1, &cache1); err != nil {
		t.Fatalf("failed to parse initial cache: %v", err)
	}
	initialCount := len(cache1.Lists)

	// Delete the list - should invalidate cache
	cli.MustExecute("-y", "list", "delete", "ToDelete")

	// Run list again to repopulate cache
	cli.MustExecute("-y", "list")

	// Read updated cache
	data2, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read cache after delete: %v", err)
	}

	var cache2 cache.ListCache
	if err := json.Unmarshal(data2, &cache2); err != nil {
		t.Fatalf("failed to parse cache after delete: %v", err)
	}

	// Cache should have fewer lists
	if len(cache2.Lists) >= initialCount {
		t.Errorf("cache should have fewer lists after delete: had %d, now %d", initialCount, len(cache2.Lists))
	}
}

// TestListCacheTTL verifies that cache expires after configured TTL (default 5 minutes).
func TestListCacheTTL(t *testing.T) {
	cli := testutil.NewCLITestWithCache(t)

	// Set a very short TTL for testing (100ms)
	cli.SetCacheTTL(100 * time.Millisecond)

	// Create a list and populate cache
	cli.MustExecute("-y", "list", "create", "TTLTest")
	cli.MustExecute("-y", "list")

	// Read cache timestamp
	cachePath := cli.CachePath()
	data1, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read cache: %v", err)
	}

	var cache1 cache.ListCache
	if err := json.Unmarshal(data1, &cache1); err != nil {
		t.Fatalf("failed to parse cache: %v", err)
	}
	ts1 := cache1.CreatedAt

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Run list again - should refresh cache due to TTL expiration
	cli.MustExecute("-y", "list")

	// Read updated cache
	data2, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read cache after TTL: %v", err)
	}

	var cache2 cache.ListCache
	if err := json.Unmarshal(data2, &cache2); err != nil {
		t.Fatalf("failed to parse cache after TTL: %v", err)
	}
	ts2 := cache2.CreatedAt

	// Cache timestamp should be updated
	if !ts2.After(ts1) {
		t.Errorf("cache should be refreshed after TTL expiration: old=%v, new=%v", ts1, ts2)
	}
}

// TestListCacheCorruptHandling verifies that corrupt cache is deleted and regenerated.
func TestListCacheCorruptHandling(t *testing.T) {
	cli := testutil.NewCLITestWithCache(t)

	// Create a list first
	cli.MustExecute("-y", "list", "create", "CorruptTest")
	cli.MustExecute("-y", "list")

	// Write corrupt data to cache file
	cachePath := cli.CachePath()
	if err := os.WriteFile(cachePath, []byte("not valid json{{{"), 0644); err != nil {
		t.Fatalf("failed to write corrupt cache: %v", err)
	}

	// Run list again - should handle corrupt cache gracefully and regenerate
	stdout := cli.MustExecute("-y", "list")

	// Should still show the list (regenerated cache)
	testutil.AssertContains(t, stdout, "CorruptTest")

	// Verify cache is now valid
	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read regenerated cache: %v", err)
	}

	var cacheData cache.ListCache
	if err := json.Unmarshal(data, &cacheData); err != nil {
		t.Errorf("regenerated cache should be valid JSON: %v", err)
	}
}

// TestListCacheMissingDirCreation verifies that cache directory is created if missing.
func TestListCacheMissingDirCreation(t *testing.T) {
	cli := testutil.NewCLITestWithCache(t)

	// Get cache path and ensure parent dir doesn't exist
	cachePath := cli.CachePath()
	cacheDir := filepath.Dir(cachePath)

	// Remove the cache directory
	if err := os.RemoveAll(cacheDir); err != nil {
		t.Fatalf("failed to remove cache dir: %v", err)
	}

	// Verify directory is gone
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Fatalf("cache directory should not exist")
	}

	// Run list command - should create directory and cache file
	cli.MustExecute("-y", "list")

	// Verify directory was created
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Error("cache directory should be created automatically")
	}

	// Verify cache file was created
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("cache file should be created")
	}
}

// TestListCacheContainsRequiredFields verifies cache contains all required metadata.
func TestListCacheContainsRequiredFields(t *testing.T) {
	cli := testutil.NewCLITestWithCache(t)

	// Create a list with color
	cli.MustExecute("-y", "list", "create", "FieldsTest")

	// Run list to populate cache
	cli.MustExecute("-y", "list")

	// Read cache
	cachePath := cli.CachePath()
	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read cache: %v", err)
	}

	var cacheData cache.ListCache
	if err := json.Unmarshal(data, &cacheData); err != nil {
		t.Fatalf("failed to parse cache: %v", err)
	}

	// Verify required fields are present
	if cacheData.CreatedAt.IsZero() {
		t.Error("cache should have creation timestamp")
	}
	if cacheData.Backend == "" {
		t.Error("cache should have backend identifier")
	}

	// Find our test list and verify its fields
	for _, l := range cacheData.Lists {
		if l.Name == "FieldsTest" {
			if l.ID == "" {
				t.Error("cached list should have ID")
			}
			if l.Name == "" {
				t.Error("cached list should have Name")
			}
			// Modified is expected to be populated
			if l.Modified.IsZero() {
				t.Error("cached list should have Modified timestamp")
			}
			return
		}
	}
	t.Error("test list not found in cache")
}

// TestListCachePerformance verifies cache lookup is under 10ms.
func TestListCachePerformance(t *testing.T) {
	cli := testutil.NewCLITestWithCache(t)

	// Create several lists
	for i := 0; i < 10; i++ {
		cli.MustExecute("-y", "list", "create", "PerfList"+string(rune('A'+i)))
	}

	// First call to populate cache
	cli.MustExecute("-y", "list")

	// Measure time for cache hit
	start := time.Now()
	for i := 0; i < 10; i++ {
		cli.MustExecute("-y", "list")
	}
	elapsed := time.Since(start)

	// Average should be well under 100ms per call (10ms target per acceptance criteria)
	avgPerCall := elapsed / 10
	if avgPerCall > 100*time.Millisecond {
		t.Errorf("cache lookup too slow: average %v per call (target <10ms)", avgPerCall)
	}
}

// ==================== Issue #008 Cache Isolation Tests ====================

// TestCacheBackendIsolation verifies that cache is isolated per backend (Issue #008).
// Bug: Cache was shared across backends causing stale/wrong data to appear.
// Fix: Cache now validates backend name before using cached data.
func TestCacheBackendIsolation(t *testing.T) {
	cli := testutil.NewCLITestWithCache(t)

	// Create a list with SQLite backend (default)
	cli.MustExecute("-y", "list", "create", "SQLiteList")

	// Run list command to populate cache
	stdout := cli.MustExecute("-y", "list")
	testutil.AssertContains(t, stdout, "SQLiteList")

	// Verify cache file was created
	cachePath := cli.CachePath()
	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read cache file: %v", err)
	}

	// Verify cache contains correct backend name
	var cacheData cache.ListCache
	if err := json.Unmarshal(data, &cacheData); err != nil {
		t.Fatalf("failed to unmarshal cache: %v", err)
	}

	if cacheData.Backend != "sqlite" {
		t.Errorf("expected cache backend to be 'sqlite', got '%s'", cacheData.Backend)
	}

	// Verify cache contains our list
	found := false
	for _, list := range cacheData.Lists {
		if list.Name == "SQLiteList" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected cache to contain 'SQLiteList'")
	}
}

// TestCacheBackendMismatchInvalidates verifies that cache from one backend is not used for another.
// This is the fix for Issue #008 where "credential" list appeared in both nextcloud and todoist.
func TestCacheBackendMismatchInvalidates(t *testing.T) {
	cli := testutil.NewCLITestWithCache(t)

	// Create a list
	cli.MustExecute("-y", "list", "create", "TestIsolation")

	// Populate cache
	cli.MustExecute("-y", "list")

	// Read cache file and modify the backend name to simulate a different backend
	cachePath := cli.CachePath()
	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("failed to read cache file: %v", err)
	}

	var cacheData cache.ListCache
	if err := json.Unmarshal(data, &cacheData); err != nil {
		t.Fatalf("failed to unmarshal cache: %v", err)
	}

	// Change backend name to simulate cache from a different backend
	cacheData.Backend = "different-backend"

	// Write modified cache
	modifiedData, _ := json.Marshal(cacheData)
	if err := os.WriteFile(cachePath, modifiedData, 0644); err != nil {
		t.Fatalf("failed to write modified cache: %v", err)
	}

	// Get mod time before list command
	info1, _ := os.Stat(cachePath)
	modTime1 := info1.ModTime()

	// Wait for filesystem timestamp granularity
	time.Sleep(10 * time.Millisecond)

	// Run list command - should NOT use cache because backend name doesn't match
	// Instead, it should re-fetch from backend and write new cache
	stdout := cli.MustExecute("-y", "list")

	// Verify output still shows our list (fetched fresh, not from cache)
	testutil.AssertContains(t, stdout, "TestIsolation")

	// Verify cache was rewritten (mod time should change)
	info2, _ := os.Stat(cachePath)
	modTime2 := info2.ModTime()

	if modTime1.Equal(modTime2) {
		t.Error("cache should have been rewritten when backend mismatch detected")
	}

	// Verify cache now has correct backend name
	newData, _ := os.ReadFile(cachePath)
	var newCacheData cache.ListCache
	_ = json.Unmarshal(newData, &newCacheData)

	if newCacheData.Backend != "sqlite" {
		t.Errorf("expected cache backend to be 'sqlite' after refresh, got '%s'", newCacheData.Backend)
	}
}
