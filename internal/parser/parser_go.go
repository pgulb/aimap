package parser

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"

	"github.com/pgulb/aimap/internal/symbol"
)

// ParseGoFile parses a single Go source file and returns the symbols found in it.
func ParseGoFile(path string) ([]symbol.Symbol, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var syms []symbol.Symbol

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			syms = append(syms, parseGenDecl(d, fset, path)...)
		case *ast.FuncDecl:
			syms = append(syms, parseFuncDecl(d, fset, path))
		}
	}

	return syms, nil
}

// parseGenDecl handles var, const, type, and import declarations.
func parseGenDecl(d *ast.GenDecl, fset *token.FileSet, path string) []symbol.Symbol {
	var syms []symbol.Symbol
	comment := docCommentFromGroup(d.Doc)

	for _, spec := range d.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			kind := symbol.Type
			typeParams := formatTypeParams(fset, s.TypeParams)

			if st, isStruct := s.Type.(*ast.StructType); isStruct {
				kind = symbol.Struct
				syms = append(syms, parseStructEmbedded(st, fset, path, s.Name.Name, comment)...)
			} else if it, isInterface := s.Type.(*ast.InterfaceType); isInterface {
				kind = symbol.Interface
				syms = append(syms, parseInterfaceMethods(it, fset, path, s.Name.Name)...)
			}

			syms = append(syms, symbol.Symbol{
				Name:       s.Name.Name,
				TypeParams: typeParams,
				Kind:       kind,
				FilePath:   path,
				LineStart:  fset.Position(s.Pos()).Line,
				LineEnd:    fset.Position(s.End()).Line,
				Comment:    comment,
			})
		case *ast.ValueSpec:
			kind := symbol.Variable
			if d.Tok.String() == "const" {
				kind = symbol.Constant
			}

			for _, name := range s.Names {
				lineStart := fset.Position(s.Pos()).Line
				lineEnd := fset.Position(s.End()).Line
				if lineEnd < lineStart {
					lineEnd = lineStart
				}

				syms = append(syms, symbol.Symbol{
					Name:      name.Name,
					Kind:      kind,
					FilePath:  path,
					LineStart: lineStart,
					LineEnd:   lineEnd,
					Comment:   comment,
				})
			}
		}
	}
	return syms
}

// parseInterfaceMethods extracts method symbols from an interface type.
func parseInterfaceMethods(it *ast.InterfaceType, fset *token.FileSet, path, parent string) []symbol.Symbol {
	var syms []symbol.Symbol
	for _, field := range it.Methods.List {
		// Named methods.
		for _, name := range field.Names {
			syms = append(syms, symbol.Symbol{
				Name:      name.Name,
				Kind:      symbol.Method,
				FilePath:  path,
				LineStart: fset.Position(field.Pos()).Line,
				LineEnd:   fset.Position(field.End()).Line,
				Comment:   docCommentFromGroup(field.Doc),
				Parent:    parent,
			})
		}
		// Embedded interface — field with no names, just a type expression.
		if len(field.Names) == 0 && field.Type != nil {
			typeName := exprString(fset, field.Type)
			syms = append(syms, symbol.Symbol{
				Name:      typeName,
				Kind:      symbol.Type,
				FilePath:  path,
				LineStart: fset.Position(field.Pos()).Line,
				LineEnd:   fset.Position(field.End()).Line,
				Comment:   docCommentFromGroup(field.Doc),
				Parent:    parent,
			})
		}
	}
	return syms
}

// parseStructEmbedded extracts embedded type symbols from a struct.
func parseStructEmbedded(st *ast.StructType, fset *token.FileSet, path, parent, parentComment string) []symbol.Symbol {
	var syms []symbol.Symbol
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 && field.Type != nil {
			typeName := exprString(fset, field.Type)
			comment := docCommentFromGroup(field.Doc)
			if comment == "" {
				comment = parentComment
			}
			syms = append(syms, symbol.Symbol{
				Name:      typeName,
				Kind:      symbol.Type,
				FilePath:  path,
				LineStart: fset.Position(field.Pos()).Line,
				LineEnd:   fset.Position(field.End()).Line,
				Comment:   comment,
				Parent:    parent,
			})
		}
	}
	return syms
}

// parseFuncDecl handles function and method declarations.
func parseFuncDecl(d *ast.FuncDecl, fset *token.FileSet, path string) symbol.Symbol {
	kind := symbol.Function
	parent := ""
	if d.Recv != nil && len(d.Recv.List) > 0 {
		kind = symbol.Method
		parent = receiverName(fset, d.Recv.List[0].Type)
	}

	comment := docCommentFromGroup(d.Doc)

	var typeParams string
	if d.Type.TypeParams != nil {
		typeParams = formatTypeParams(fset, d.Type.TypeParams)
	}

	return symbol.Symbol{
		Name:       d.Name.Name,
		TypeParams: typeParams,
		Kind:       kind,
		FilePath:   path,
		LineStart:  fset.Position(d.Pos()).Line,
		LineEnd:    fset.Position(d.End()).Line,
		Comment:    comment,
		Parent:     parent,
	}
}

// receiverName extracts the full type name from a receiver expression,
// including type parameters for generic receivers (e.g. "List[T]").
func receiverName(fset *token.FileSet, expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverName(fset, t.X)
	case *ast.IndexExpr:
		// e.g. List[T]
		return exprString(fset, t.X) + "[" + exprString(fset, t.Index) + "]"
	case *ast.IndexListExpr:
		// e.g. List[T, U]
		var parts []string
		for _, idx := range t.Indices {
			parts = append(parts, exprString(fset, idx))
		}
		return exprString(fset, t.X) + "[" + strings.Join(parts, ", ") + "]"
	}
	return ""
}

// formatTypeParams formats an ast.FieldList of type parameters as a string like "[T any]".
func formatTypeParams(fset *token.FileSet, fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("[")
	for i, f := range fields.List {
		if i > 0 {
			sb.WriteString(", ")
		}
		for j, name := range f.Names {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(name.Name)
		}
		if f.Type != nil {
			sb.WriteString(" ")
			sb.WriteString(exprString(fset, f.Type))
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// exprString converts an ast.Expr to a string representation using go/printer.
func exprString(fset *token.FileSet, expr ast.Expr) string {
	var buf strings.Builder
	printer.Fprint(&buf, fset, expr)
	return buf.String()
}

func docCommentFromGroup(g *ast.CommentGroup) string {
	if g == nil {
		return ""
	}
	return g.Text()
}

// ParseFile detects the language from the file extension and parses accordingly.
func ParseFile(path string) ([]symbol.Symbol, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := fileExt(path)
	switch ext {
	case ".go":
		return ParseGoFile(path)
	case ".py":
		return ParsePythonFile(path, string(data))
	default:
		return nil, nil
	}
}

func fileExt(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i:]
		}
		if path[i] == '/' || path[i] == '\\' {
			break
		}
	}
	return ""
}
