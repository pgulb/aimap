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
	
	// Track decorators for current definition (e.g., @property, @staticmethod, @classmethod).
	var pendingDecorators []string

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

		// Decorator lines — collect them for method classification, then skip.
		if strings.HasPrefix(stripped, "@") {
			// Extract decorator name (e.g., "property" from "@property")
			decoratorName := extractDecoratorName(stripped)
			if decoratorName != "" {
				pendingDecorators = append(pendingDecorators, decoratorName)
			}
			continue
		}

		// Class definition.
		if strings.HasPrefix(stripped, "class ") {
			name := extractPythonName(stripped, "class ")
			if name != "" {
				// Extract docstring if present
				docstring := extractDocstringAfter(lines, i, indent)
				comment := pendingComment
				if docstring != "" && comment == "" {
					comment = docstring
				} else if docstring != "" && comment != "" {
					comment = comment + "\n" + docstring
				}
				
				s := symbol.Symbol{
					Name:      name,
					Kind:      symbol.Class,
					FilePath:  path,
					LineStart: i + 1,
					LineEnd:   i + 1,
					Comment:   comment,
				}
				syms = append(syms, s)

				currentClass = name
				currentClassIndent = indent
			}
			pendingComment = ""
			pendingDecorators = nil
			continue
		}

		// Function definition (both def and async def).
		if strings.HasPrefix(stripped, "def ") || strings.HasPrefix(stripped, "async def ") {
			keyword := "def "
			if strings.HasPrefix(stripped, "async ") {
				keyword = "async def "
			}
			name := extractPythonName(stripped, keyword)
			if name != "" {
				kind := symbol.Function
				parent := ""
				endLine := findFunctionEnd(lines, i+1, indent)

				// If we're inside a class, this is a method.
				if currentClass != "" && indent > currentClassIndent {
					kind = symbol.Method
					parent = currentClass
					
					// Check for special decorators to override method kind if needed
					// (Currently we keep as Method, but could distinguish @property, etc. if needed)
				}

				// Extract docstring if present
				docstring := extractDocstringAfter(lines, i, indent)
				comment := pendingComment
				if docstring != "" && comment == "" {
					comment = docstring
				} else if docstring != "" && comment != "" {
					comment = comment + "\n" + docstring
				}

				s := symbol.Symbol{
					Name:      name,
					Kind:      kind,
					FilePath:  path,
					LineStart: i + 1,
					LineEnd:   endLine,
					Comment:   comment,
					Parent:    parent,
				}
				syms = append(syms, s)
			}
			pendingComment = ""
			pendingDecorators = nil
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
			!strings.HasPrefix(stripped, "def ") && !strings.HasPrefix(stripped, "class ") {

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

// extractDecoratorName extracts the decorator name from a decorator line.
// E.g., "@property" -> "property", "@foo.bar" -> "foo"
func extractDecoratorName(line string) string {
	// Remove leading @ and whitespace
	line = strings.TrimSpace(strings.TrimPrefix(line, "@"))
	if line == "" {
		return ""
	}
	// Get the first identifier before any parentheses or dots
	endIdx := strings.IndexAny(line, "(.)")
	if endIdx > 0 {
		line = line[:endIdx]
	}
	return strings.TrimSpace(line)
}

// extractDocstringAfter looks ahead from the current line (def/class line)
// to find a docstring (triple-quoted string) in the function/class body.
// Returns the docstring content (without quotes) if found, empty string otherwise.
func extractDocstringAfter(lines []string, defLineIdx int, defIndent int) string {
	// Start looking at the next line after def/class
	for i := defLineIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		
		// Skip empty lines
		if trimmed == "" {
			continue
		}
		
		indent := leadingSpaces(line)
		
		// Stop if we hit a line at the same or lower indentation (end of body)
		if indent <= defIndent {
			break
		}
		
		// Check for docstring (triple-quoted string)
		if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, "'''") {
			return extractDocstringContent(lines, i, defIndent)
		}
		
		// Any other non-empty, indented line means no docstring at the start
		break
	}
	return ""
}

// extractDocstringContent extracts the content of a docstring that starts at docstringLineIdx.
// Handles both single-line and multi-line docstrings.
func extractDocstringContent(lines []string, docstringLineIdx int, defIndent int) string {
	line := strings.TrimSpace(lines[docstringLineIdx])
	
	// Determine quote style
	var quoteStyle string
	if strings.HasPrefix(line, `"""`) {
		quoteStyle = `"""`
	} else if strings.HasPrefix(line, "'''") {
		quoteStyle = "'''"
	} else {
		return ""
	}
	
	// Remove opening quotes
	content := strings.TrimPrefix(line, quoteStyle)
	
	// Check if it's a single-line docstring (opening and closing quotes on same line)
	if strings.HasSuffix(content, quoteStyle) && len(content) > len(quoteStyle) {
		// Single-line docstring
		content = strings.TrimSuffix(content, quoteStyle)
		return strings.TrimSpace(content)
	}
	
	// Multi-line docstring: keep collecting until we find the closing quotes
	var docLines []string
	docLines = append(docLines, content)
	
	for i := docstringLineIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		
		indent := leadingSpaces(line)
		if indent <= defIndent {
			// Shouldn't happen if the docstring is well-formed
			break
		}
		
		// Check for closing quotes
		if strings.Contains(trimmed, quoteStyle) {
			// Found closing quotes
			endIdx := strings.Index(trimmed, quoteStyle)
			docLines = append(docLines, trimmed[:endIdx])
			break
		}
		
		docLines = append(docLines, trimmed)
	}
	
	// Join lines and trim
	result := strings.Join(docLines, " ")
	result = strings.TrimSpace(result)
	// Collapse multiple spaces
	result = strings.Join(strings.Fields(result), " ")
	return result
}
