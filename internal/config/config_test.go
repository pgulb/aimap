package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pgulb/aimap/internal/ignore"
)

func TestLoadNoIgnoreFile(t *testing.T) {
	dir := t.TempDir()

	cfg, err := Load(dir, ".aimapignore")
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Root != dir {
		t.Errorf("Root = %q, want %q", cfg.Root, dir)
	}
	if cfg.Matcher == nil {
		t.Fatal("expected non-nil matcher")
	}
}

func TestLoadWithIgnoreFile(t *testing.T) {
	dir := t.TempDir()

	ignoreContent := "*.pb.go\nbuild/\n"
	ignorePath := filepath.Join(dir, ".aimapignore")
	if err := os.WriteFile(ignorePath, []byte(ignoreContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir, ".aimapignore")
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Matcher.Ignored("api/v1/messages.pb.go") {
		t.Error("expected *.pb.go to be ignored")
	}
	if !cfg.Matcher.Ignored("build/output.bin") {
		t.Error("expected build/ to be ignored")
	}
}

func TestLoadResolvesAbsPath(t *testing.T) {
	cfg, err := Load(".", ".aimapignore")
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(cfg.Root) {
		t.Errorf("Root should be absolute, got %q", cfg.Root)
	}
}

func TestMatcherIsIgnoreMatcher(t *testing.T) {
	dir := t.TempDir()

	cfg, err := Load(dir, ".aimapignore")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := interface{}(cfg.Matcher).(*ignore.Matcher); !ok {
		t.Error("Matcher should be *ignore.Matcher")
	}
}

func TestEnsureIgnoreFileCreatesIfMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".aimapignore")

	if err := EnsureIgnoreFile(dir, ".aimapignore"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "# .aimapignore") {
		t.Errorf("expected header in new file, got: %s", content)
	}
}

func TestEnsureIgnoreFileDev(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".aimapignore_dev")

	if err := EnsureIgnoreFile(dir, ".aimapignore_dev"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, ".aimapignore_dev") {
		t.Errorf("expected dev header, got: %s", content)
	}
}

func TestEnsureIgnoreFilePreservesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".aimapignore")

	existing := "# .aimapignore\n*.log\n"
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureIgnoreFile(dir, ".aimapignore"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "*.log") {
		t.Error("expected existing patterns to be preserved")
	}
}

func TestEnsureIgnoreFilePrependsHeaderIfMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".aimapignore")

	existing := "*.log\n"
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureIgnoreFile(dir, ".aimapignore"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "# .aimapignore") {
		t.Errorf("expected header to be prepended, got: %s", content)
	}
	if !strings.Contains(content, "*.log") {
		t.Error("expected existing content to be preserved")
	}
}

func TestEnsureIgnoreFileNoopOnEmptyHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".aimapignore")

	existing := "# .aimapignore — ignore rules for aimap\n# Add patterns below\n"
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureIgnoreFile(dir, ".aimapignore"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if content != existing {
		t.Errorf("expected file unchanged, got: %s", content)
	}
}
