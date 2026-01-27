package views

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSetupViewsFolder tests the views folder setup functionality
func TestSetupViewsFolder(t *testing.T) {
	t.Run("creates folder and default files when not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		viewsDir := filepath.Join(tmpDir, "views")

		// Folder should not exist
		_, err := os.Stat(viewsDir)
		if !os.IsNotExist(err) {
			t.Fatal("views folder should not exist yet")
		}

		// Setup views folder
		created, err := SetupViewsFolder(viewsDir)
		if err != nil {
			t.Fatalf("SetupViewsFolder failed: %v", err)
		}
		if !created {
			t.Error("expected created to be true")
		}

		// Check folder was created
		info, err := os.Stat(viewsDir)
		if err != nil {
			t.Fatalf("views folder should exist: %v", err)
		}
		if !info.IsDir() {
			t.Error("views should be a directory")
		}

		// Check default.yaml was created
		defaultPath := filepath.Join(viewsDir, "default.yaml")
		if _, err := os.Stat(defaultPath); err != nil {
			t.Errorf("default.yaml should exist: %v", err)
		}

		// Check all.yaml was created
		allPath := filepath.Join(viewsDir, "all.yaml")
		if _, err := os.Stat(allPath); err != nil {
			t.Errorf("all.yaml should exist: %v", err)
		}

		// Verify default.yaml content is loadable and has DONE filter
		loader := NewLoader(viewsDir)
		view, err := loader.LoadView("default")
		if err != nil {
			t.Fatalf("failed to load default view: %v", err)
		}
		if view.Name != "default" {
			t.Errorf("expected view name 'default', got %q", view.Name)
		}

		// Check that default view has the DONE filter
		hasDoneFilter := false
		for _, f := range view.Filters {
			if f.Field == "status" && f.Operator == "ne" && f.Value == "DONE" {
				hasDoneFilter = true
				break
			}
		}
		if !hasDoneFilter {
			t.Error("default.yaml should have filter to exclude DONE tasks")
		}
	})

	t.Run("returns false when folder already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		viewsDir := filepath.Join(tmpDir, "views")

		// Create folder manually
		if err := os.MkdirAll(viewsDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Setup should not create anything
		created, err := SetupViewsFolder(viewsDir)
		if err != nil {
			t.Fatalf("SetupViewsFolder failed: %v", err)
		}
		if created {
			t.Error("expected created to be false when folder already exists")
		}

		// Files should NOT be created (empty folder)
		defaultPath := filepath.Join(viewsDir, "default.yaml")
		if _, err := os.Stat(defaultPath); !os.IsNotExist(err) {
			t.Error("default.yaml should not be created when folder already exists")
		}
	})
}

// TestUserViewOverridesBuiltin tests that user views override built-in views
func TestUserViewOverridesBuiltin(t *testing.T) {
	tmpDir := t.TempDir()
	viewsDir := filepath.Join(tmpDir, "views")

	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a custom default.yaml that shows DONE tasks (no filter)
	customDefault := `name: default
description: My custom default view (shows all tasks)
fields:
  - name: status
    width: 12
  - name: summary
    width: 50
# No filter - shows DONE tasks too
`
	if err := os.WriteFile(filepath.Join(viewsDir, "default.yaml"), []byte(customDefault), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewLoader(viewsDir)
	view, err := loader.LoadView("default")
	if err != nil {
		t.Fatalf("failed to load default view: %v", err)
	}

	// Should use user's custom description
	if view.Description != "My custom default view (shows all tasks)" {
		t.Errorf("expected custom description, got %q", view.Description)
	}

	// Should NOT have the DONE filter (user's version has no filters)
	if len(view.Filters) != 0 {
		t.Errorf("expected no filters in user's default view, got %d filters", len(view.Filters))
	}
}

// TestFallbackToBuiltinWhenUserViewMissing tests fallback behavior
func TestFallbackToBuiltinWhenUserViewMissing(t *testing.T) {
	tmpDir := t.TempDir()
	viewsDir := filepath.Join(tmpDir, "views")

	// Create empty views folder (no files)
	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatal(err)
	}

	loader := NewLoader(viewsDir)

	// Should fall back to built-in default
	view, err := loader.LoadView("default")
	if err != nil {
		t.Fatalf("failed to load default view: %v", err)
	}

	// Should be the built-in version (has DONE filter)
	hasDoneFilter := false
	for _, f := range view.Filters {
		if f.Field == "status" && f.Operator == "ne" && f.Value == "DONE" {
			hasDoneFilter = true
			break
		}
	}
	if !hasDoneFilter {
		t.Error("built-in default view should have DONE filter")
	}

	// Should fall back to built-in all
	allView, err := loader.LoadView("all")
	if err != nil {
		t.Fatalf("failed to load all view: %v", err)
	}

	// Built-in "all" view has many fields
	if len(allView.Fields) < 10 {
		t.Errorf("built-in all view should have many fields, got %d", len(allView.Fields))
	}
}

// TestViewListShowsOverrideStatus tests that ListViews shows which views are overridden
func TestViewListShowsOverrideStatus(t *testing.T) {
	tmpDir := t.TempDir()
	viewsDir := filepath.Join(tmpDir, "views")

	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create user's default.yaml (overrides built-in)
	customDefault := `name: default
description: Custom default
fields:
  - name: status
`
	if err := os.WriteFile(filepath.Join(viewsDir, "default.yaml"), []byte(customDefault), 0644); err != nil {
		t.Fatal(err)
	}

	// Create user's custom view
	customView := `name: myview
description: My custom view
fields:
  - name: summary
`
	if err := os.WriteFile(filepath.Join(viewsDir, "myview.yaml"), []byte(customView), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewLoader(viewsDir)
	views, err := loader.ListViews()
	if err != nil {
		t.Fatalf("ListViews failed: %v", err)
	}

	// Find each view and check its properties
	var defaultView, allView, myView *ViewInfo
	for i := range views {
		switch views[i].Name {
		case "default":
			defaultView = &views[i]
		case "all":
			allView = &views[i]
		case "myview":
			myView = &views[i]
		}
	}

	// default: should show as user-defined, overrides built-in
	if defaultView == nil {
		t.Fatal("default view not found in list")
	}
	if defaultView.BuiltIn {
		t.Error("default view should NOT be marked as built-in when user file exists")
	}
	if !defaultView.Overrides {
		t.Error("default view should be marked as overriding built-in")
	}

	// all: should show as built-in (no user file)
	if allView == nil {
		t.Fatal("all view not found in list")
	}
	if !allView.BuiltIn {
		t.Error("all view should be marked as built-in")
	}
	if allView.Overrides {
		t.Error("all view should NOT be marked as overriding (no user file)")
	}

	// myview: should show as user-defined, does not override built-in
	if myView == nil {
		t.Fatal("myview not found in list")
	}
	if myView.BuiltIn {
		t.Error("myview should NOT be marked as built-in")
	}
	if myView.Overrides {
		t.Error("myview should NOT be marked as overriding (no built-in with that name)")
	}
}

// TestUserAllViewOverridesBuiltin tests that user can override the "all" view too
func TestUserAllViewOverridesBuiltin(t *testing.T) {
	tmpDir := t.TempDir()
	viewsDir := filepath.Join(tmpDir, "views")

	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a custom all.yaml
	customAll := `name: all
description: My minimal all view
fields:
  - name: status
  - name: summary
`
	if err := os.WriteFile(filepath.Join(viewsDir, "all.yaml"), []byte(customAll), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewLoader(viewsDir)
	view, err := loader.LoadView("all")
	if err != nil {
		t.Fatalf("failed to load all view: %v", err)
	}

	// Should use user's version with only 2 fields
	if len(view.Fields) != 2 {
		t.Errorf("expected 2 fields from user's all.yaml, got %d", len(view.Fields))
	}

	if view.Description != "My minimal all view" {
		t.Errorf("expected custom description, got %q", view.Description)
	}
}
