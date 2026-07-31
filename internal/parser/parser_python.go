package parser

import (
	"strings"
	"unicode"

	"github.com/pgulb/aimap/internal/symbol"
)

// ParsePythonFile parses a Python source string and extracts symbols.
// Uses a line-based state machine to detect function and class definitions.
func ParsePythonFile(path, source string) ([]symbol.Symbol, error) {
	lines := strings.Split(source, "\n")
	var syms []symbol.Symbol

	// Track current class context for methods.
	var currentClass string
	var currentClassIndent int

	// Collect pending doc comments (comments preceding a definition).
	var pendingComment string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and pure comments (we collect them).
		if trimmed == "" {
			continue
		}

		// Track comments as potential doc comments.
		if strings.HasPrefix(trimmed, "#") {
			pendingComment = collectComment(trimmed, pendingComment)
			continue
		}

		indent := leadingSpaces(line)
		stripped := strings.TrimLeftFunc(line, unicode.IsSpace)

		// Class definition.
		if strings.HasPrefix(stripped, "class ") {
			name := extractPythonName(stripped, "class ")
			if name != "" {
				s := symbol.Symbol{
					Name:      name,
					Kind:      symbol.Class,
					FilePath:  path,
					LineStart: i + 1,
					LineEnd:   i + 1,
					Comment:   pendingComment,
				}
				syms = append(syms, s)

				currentClass = name
				currentClassIndent = indent
			}
			pendingComment = ""
			continue
		}

		// Function definition.
		if strings.HasPrefix(stripped, "def ") {
			name := extractPythonName(stripped, "def ")
			if name != "" {
				kind := symbol.Function
				parent := ""
				endLine := findFunctionEnd(lines, i+1, indent)

				// If we're inside a class, this is a method.
				if currentClass != "" && indent > currentClassIndent {
					kind = symbol.Method
					parent = currentClass
				}

				s := symbol.Symbol{
					Name:      name,
					Kind:      kind,
					FilePath:  path,
					LineStart: i + 1,
					LineEnd:   endLine,
					Comment:   pendingComment,
					Parent:    parent,
				}
				syms = append(syms, s)
			}
			pendingComment = ""
			continue
		}

		// Top-level assignment (simple variables/constants).
		if currentClass == "" && strings.Contains(stripped, "=") && !strings.HasPrefix(stripped, "if ") &&
			!strings.HasPrefix(stripped, "for ") && !strings.HasPrefix(stripped, "while ") &&
			!strings.HasPrefix(stripped, "elif ") && !strings.HasPrefix(stripped, "else") &&
			!strings.HasPrefix(stripped, "try") && !strings.HasPrefix(stripped, "except") &&
			!strings.HasPrefix(stripped, "with ") && !strings.HasPrefix(stripped, "import ") &&
			!strings.HasPrefix(stripped, "from ") && !strings.HasPrefix(stripped, "return ") &&
			!strings.HasPrefix(stripped, "raise ") && !strings.HasPrefix(stripped, "pass") &&
			!strings.HasPrefix(stripped, "def ") && !strings.HasPrefix(stripped, "class ") &&
			!strings.HasPrefix(stripped, "@") {

			name := extractAssignName(stripped)
			if name != "" && isTopLevel(indent) {
				s := symbol.Symbol{
					Name:      name,
					Kind:      symbol.Variable,
					FilePath:  path,
					LineStart: i + 1,
					LineEnd:   i + 1,
					Comment:   pendingComment,
				}
				syms = append(syms, s)
			}
		}

		pendingComment = ""
	}

	return syms, nil
}

// collectComment accumulates lines into a pending doc comment.
func collectComment(line, existing string) string {
	comment := strings.TrimSpace(strings.TrimPrefix(line, "#"))
	if existing == "" {
		return comment
	}
	return existing + "\n" + comment
}

// extractPythonName extracts the name after a keyword (def or class).
func extractPythonName(line, keyword string) string {
	after := strings.TrimSpace(strings.TrimPrefix(line, keyword))
	parenIdx := strings.Index(after, "(")
	colonIdx := strings.Index(after, ":")
	var name string
	if parenIdx >= 0 && (colonIdx < 0 || parenIdx < colonIdx) {
		name = strings.TrimSpace(after[:parenIdx])
	} else if colonIdx >= 0 {
		name = strings.TrimSpace(after[:colonIdx])
	} else {
		name = strings.TrimSpace(after)
	}
	// Remove generic type annotations like [T].
	if bracketIdx := strings.Index(name, "["); bracketIdx >= 0 {
		name = name[:bracketIdx]
	}
	return name
}

// extractAssignName gets the variable name from a top-level assignment.
func extractAssignName(line string) string {
	eqIdx := strings.Index(line, "=")
	if eqIdx <= 0 {
		return ""
	}
	name := strings.TrimSpace(line[:eqIdx])
	// Must be a simple identifier, not a complex expression.
	for _, r := range name {
		if !unicode.IsLetter(r) && r != '_' && !unicode.IsDigit(r) {
			return ""
		}
	}
	if name == "" {
		return ""
	}
	// Skip dunder variables like __all__, __version__ etc as they're typically
	// metadata, not meaningful symbols for navigation.
	if strings.HasPrefix(name, "__") && strings.HasSuffix(name, "__") {
		return ""
	}
	return name
}

// findFunctionEnd finds the last line of a function body given the starting line (1-based) and indentation.
func findFunctionEnd(lines []string, startLine int, baseIndent int) int {
	end := startLine
	for i := startLine; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			continue
		}
		indent := leadingSpaces(lines[i])
		if indent <= baseIndent && trimmed != "" {
			break
		}
		end = i + 1
	}
	if end < startLine {
		return startLine
	}
	return end
}

// leadingSpaces counts the number of leading spaces in a line.
func leadingSpaces(line string) int {
	count := 0
	for _, r := range line {
		if r == ' ' {
			count++
		} else if r == '\t' {
			count += 8
		} else {
			break
		}
	}
	return count
}

// isTopLevel checks if indentation indicates top-level scope.
func isTopLevel(indent int) bool {
	return indent == 0
}
