package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pgulb/aimap/internal/ignore"
)

func TestClassifyFile(t *testing.T) {
	tests := []struct {
		path string
		want Language
	}{
		{"main.go", LanguageGo},
		{"internal/parser.go", LanguageGo},
		{"app.py", LanguagePython},
		{"setup.py", LanguagePython},
		{"README.md", LanguageOther},
		{"Makefile", LanguageOther},
		{"file", LanguageOther},
	}

	for _, tt := range tests {
		got := classifyFile(tt.path)
		if got != tt.want {
			t.Errorf("classifyFile(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestScan(t *testing.T) {
	dir := t.TempDir()

	// Create a directory structure.
	files := []string{
		"main.go",
		"internal/parser/parser.go",
		"internal/parser/parser_test.go",
		"app.py",
		"node_modules/package/index.js",
		".venv/bin/python",
		"__pycache__/module.pyc",
		"README.md",
	}

	for _, f := range files {
		p := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	m := ignore.NewMatcher(nil)
	s := NewScanner(m)
	results, err := s.Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Should find all .go and .py files that aren't ignored.
	// parser_test.go is also a .go file and contains symbols, so it's included.
	if len(results) != 4 {
		t.Fatalf("got %d files, want 4: %+v", len(results), results)
	}

	// Map results by base name for easier checking.
	byName := make(map[string]bool)
	for _, r := range results {
		base := filepath.Base(r.Path)
		byName[base] = true
	}

	if !byName["main.go"] {
		t.Error("expected main.go to be found")
	}
	if !byName["parser.go"] {
		t.Error("expected parser.go to be found")
	}
	if !byName["parser_test.go"] {
		t.Error("expected parser_test.go to be found")
	}
	if !byName["app.py"] {
		t.Error("expected app.py to be found")
	}
}
