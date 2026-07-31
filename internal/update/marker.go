package update

import (
	"os"
	"path/filepath"
	"runtime"
)

const markerFilename = "will_autoupdate"

// MarkerDir returns the platform-specific directory for the self-update marker file.
func MarkerDir() string {
	switch runtime.GOOS {
	case "linux":
		if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
			return filepath.Join(d, "aimap")
		}
		return filepath.Join(os.Getenv("HOME"), ".config", "aimap")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "aimap")
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "aimap")
	default:
		return filepath.Join(os.Getenv("HOME"), ".config", "aimap")
	}
}

// markerPath returns the full path to the marker file.
// This is a variable so tests can override it.
var markerPath = func() string {
	return filepath.Join(MarkerDir(), markerFilename)
}

// CanSelfUpdate checks whether the self-update marker file exists.
func CanSelfUpdate() bool {
	_, err := os.Stat(markerPath())
	return err == nil
}

// CreateMarker creates the self-update marker file and its parent directory.
func CreateMarker() error {
	dir := filepath.Dir(markerPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.Create(markerPath())
	if err != nil {
		return err
	}
	return f.Close()
}
