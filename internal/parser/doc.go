// Package parser scans Go and Python source files and extracts symbol information.
// For Go files it uses go/parser and go/ast from the standard library.
// For Python files it uses a line-based state machine.
package parser
