package update

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMarkerDir(t *testing.T) {
	dir := MarkerDir()
	if dir == "" {
		t.Fatal("MarkerDir() returned empty")
	}
}

func TestCanSelfUpdateFalseByDefault(t *testing.T) {
	origDir := markerPath
	defer func() { markerPath = origDir }()

	tmpDir := t.TempDir()
	markerPath = func() string {
		return filepath.Join(tmpDir, "will_autoupdate")
	}

	got := CanSelfUpdate()
	if got {
		t.Error("CanSelfUpdate() should be false when no marker exists")
	}
}

func TestCreateAndCheckMarker(t *testing.T) {
	origDir := markerPath
	defer func() { markerPath = origDir }()

	tmpDir := t.TempDir()
	markerPath = func() string {
		return filepath.Join(tmpDir, "will_autoupdate")
	}

	if CanSelfUpdate() {
		t.Fatal("should not be able to self-update before creating marker")
	}

	if err := CreateMarker(); err != nil {
		t.Fatal(err)
	}

	if !CanSelfUpdate() {
		t.Fatal("should be able to self-update after creating marker")
	}

	// Second call should be idempotent.
	if err := CreateMarker(); err != nil {
		t.Fatal(err)
	}
}

// Test that the marker file is created at the expected path.
func TestCreateMarkerCreatesFile(t *testing.T) {
	origDir := markerPath
	defer func() { markerPath = origDir }()

	tmpDir := t.TempDir()
	expectedPath := filepath.Join(tmpDir, "will_autoupdate")
	markerPath = func() string {
		return expectedPath
	}

	if err := CreateMarker(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(expectedPath); err != nil {
		t.Errorf("expected marker file at %s: %v", expectedPath, err)
	}
}
