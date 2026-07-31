package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pgulb/aimap/internal/ignore"
)

const ignoreHeader = "# .aimapignore — ignore rules for aimap\n# Add patterns below (one per line, gitignore syntax):\n"

const ignoreHeaderDev = "# .aimapignore_dev — ignore rules for aimap (dev mode)\n# Add patterns below (one per line, gitignore syntax):\n"

// Config holds the runtime configuration for aimap.
type Config struct {
	Root    string
	Matcher *ignore.Matcher
}

// Load reads configuration from the project root directory.
// ignoreFileName is the name of the ignore file (e.g. ".aimapignore" or ".aimapignore_dev").
func Load(root, ignoreFileName string) (*Config, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	ignorePath := filepath.Join(absRoot, ignoreFileName)
	matcher, err := ignore.LoadMatcherFromFile(ignorePath)
	if err != nil {
		return nil, err
	}

	return &Config{
		Root:    absRoot,
		Matcher: matcher,
	}, nil
}

// EnsureIgnoreFile creates the given ignore file in root if it does not exist.
// If it exists but is missing the aimap header, prepends it.
func EnsureIgnoreFile(root, ignoreFileName string) error {
	path := filepath.Join(root, ignoreFileName)

	header := ignoreHeader
	if ignoreFileName == ".aimapignore_dev" {
		header = ignoreHeaderDev
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return os.WriteFile(path, []byte(header), 0644)
	}

	if hasHeader(data) {
		return nil
	}

	newContent := header + string(data)
	return os.WriteFile(path, []byte(newContent), 0644)
}

// hasHeader checks if the first non-empty, non-whitespace line starts with the
// expected header prefix.
func hasHeader(data []byte) bool {
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		return strings.HasPrefix(trimmed, "# .aimapignore")
	}
	return false
}
