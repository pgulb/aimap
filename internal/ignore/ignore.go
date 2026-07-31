package ignore

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var builtinPatterns = []string{
	"node_modules",
	".venv",
	"venv",
	".git",
	".gitignore",
	".dockerignore",
	"__pycache__",
	"*.pyc",
	"*.pyo",
	"*.so",
	"*.dll",
	"*.dylib",
	".DS_Store",
	"Thumbs.db",
	".idea",
	".vscode",
	"*.exe",
	"*.bin",
	"vendor",
}

// Matcher checks whether a given path should be ignored.
type Matcher struct {
	patterns []string
}

// NewMatcher creates a Matcher with the built-in ignore patterns combined with
// any user-provided patterns.
func NewMatcher(patterns []string) *Matcher {
	merged := make([]string, 0, len(builtinPatterns)+len(patterns))
	merged = append(merged, builtinPatterns...)
	merged = append(merged, patterns...)
	return &Matcher{patterns: merged}
}

// LoadMatcherFromFile reads ignore patterns from path (one per line, gitignore-style)
// and returns a Matcher combining them with built-in patterns.
// If path does not exist, a Matcher with only built-in patterns is returned.
func LoadMatcherFromFile(path string) (*Matcher, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewMatcher(nil), nil
		}
		return nil, err
	}
	defer f.Close()

	patterns, err := parsePatterns(f)
	if err != nil {
		return nil, err
	}
	return NewMatcher(patterns), nil
}

func parsePatterns(r io.Reader) ([]string, error) {
	var patterns []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, scanner.Err()
}

// Ignored checks whether path should be skipped. path is relative to the project root.
// It matches the path's base name and any component against the pattern list.
func (m *Matcher) Ignored(path string) bool {
	// Normalize to forward slashes.
	path = filepath.ToSlash(path)
	base := filepath.Base(path)

	for _, pattern := range m.patterns {
		pat := filepath.ToSlash(pattern)

		// Match the full path or any component.
		if matchPath(pat, path) {
			return true
		}
		// Match base name.
		if matchPath(pat, base) {
			return true
		}
	}
	return false
}

// matchPath checks if a single pattern matches the given path.
// Supports * and ? wildcards and trailing / for directory-only matches.
func matchPath(pattern, path string) bool {
	// Check for directory-only pattern (trailing /).
	dirOnly := strings.HasSuffix(pattern, "/")
	if dirOnly {
		pattern = strings.TrimSuffix(pattern, "/")
	}

	matched, err := filepath.Match(pattern, path)
	if err != nil {
		return false
	}
	if matched {
		return true
	}

	// If pattern has no slash, it matches against any path component.
	if !strings.Contains(pattern, "/") {
		parts := strings.Split(path, "/")
		for _, part := range parts {
			matched, err := filepath.Match(pattern, part)
			if err == nil && matched {
				return true
			}
		}
	}

	return false
}
