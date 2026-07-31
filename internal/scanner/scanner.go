package scanner

import (
	"io/fs"
	"path/filepath"

	"github.com/pgulb/aimap/internal/ignore"
)

// Language identifies the programming language of a source file.
type Language int

const (
	LanguageGo     Language = iota
	LanguagePython
	LanguageOther
)

// FileInfo describes a source file to be parsed.
type FileInfo struct {
	Path     string
	Language Language
}

// Scanner walks a directory tree and returns files to parse.
type Scanner struct {
	matcher *ignore.Matcher
}

// NewScanner creates a Scanner that uses the given matcher for ignore rules.
func NewScanner(matcher *ignore.Matcher) *Scanner {
	return &Scanner{matcher: matcher}
}

// Scan walks root and returns all non-ignored source files.
func (s *Scanner) Scan(root string) ([]FileInfo, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	err = filepath.WalkDir(rootAbs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		// Check ignore rules for both files and directories.
		if s.matcher.Ignored(rel) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		lang := classifyFile(path)
		if lang == LanguageOther {
			return nil
		}

		files = append(files, FileInfo{
			Path:     path,
			Language: lang,
		})
		return nil
	})
	return files, err
}

func classifyFile(path string) Language {
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return LanguageGo
	case ".py":
		return LanguagePython
	default:
		return LanguageOther
	}
}
