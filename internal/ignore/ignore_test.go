package ignore

import (
	"strings"
	"testing"
)

func TestBuiltinPatterns(t *testing.T) {
	m := NewMatcher(nil)

	tests := []struct {
		path    string
		ignored bool
	}{
		{"node_modules/package/index.js", true},
		{".venv/bin/python", true},
		{"venv/bin/activate", true},
		{".git/HEAD", true},
		{"src/main.go", false},
		{"__pycache__/module.pyc", true},
		{"internal/parser/parser.go", false},
		{".DS_Store", true},
		{".idea/workspace.xml", true},
	}

	for _, tt := range tests {
		got := m.Ignored(tt.path)
		if got != tt.ignored {
			t.Errorf("Ignored(%q) = %v, want %v", tt.path, got, tt.ignored)
		}
	}
}

func TestCustomPatterns(t *testing.T) {
	m := NewMatcher([]string{"*.pb.go", "testdata/"})

	tests := []struct {
		path    string
		ignored bool
	}{
		{"api/v1/messages.pb.go", true},
		{"testdata/sample.go", true},
		{"testdata/", true},
		{"internal/scanner/scanner.go", false},
	}

	for _, tt := range tests {
		got := m.Ignored(tt.path)
		if got != tt.ignored {
			t.Errorf("Ignored(%q) = %v, want %v", tt.path, got, tt.ignored)
		}
	}
}

func TestParsePatterns(t *testing.T) {
	input := "node_modules\n.git\n# this is a comment\n*.pyc\n\nvendor\n"
	patterns, err := parsePatterns(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"node_modules", ".git", "*.pyc", "vendor"}
	if len(patterns) != len(expected) {
		t.Fatalf("got %d patterns, want %d: %v", len(patterns), len(expected), patterns)
	}
	for i, p := range patterns {
		if p != expected[i] {
			t.Errorf("pattern[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestLoadMatcherFromFileMissing(t *testing.T) {
	m, err := LoadMatcherFromFile("/tmp/nonexistent-aimap-ignore-file")
	if err != nil {
		t.Fatal(err)
	}
	if m == nil {
		t.Fatal("expected non-nil matcher")
	}
	if !m.Ignored("node_modules") {
		t.Error("expected built-in node_modules to be ignored")
	}
}
