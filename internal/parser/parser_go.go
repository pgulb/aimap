package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"github.com/pgulb/aimap/internal/symbol"
)

// ParseGoFile parses a single Go source file and returns the symbols found in it.
func ParseGoFile(path string) ([]symbol.Symbol, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	fileStart := fset.File(f.Pos()).Offset(f.Pos())
	_ = fileStart

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
			if _, isStruct := s.Type.(*ast.StructType); isStruct {
				kind = symbol.Struct
			} else if _, isInterface := s.Type.(*ast.InterfaceType); isInterface {
				kind = symbol.Interface
			}

			syms = append(syms, symbol.Symbol{
				Name:      s.Name.Name,
				Kind:      kind,
				FilePath:  path,
				LineStart: fset.Position(s.Pos()).Line,
				LineEnd:   fset.Position(s.End()).Line,
				Comment:   comment,
			})
		case *ast.ValueSpec:
			kind := symbol.Variable
			if d.Tok.String() == "const" {
				kind = symbol.Constant
			}

			for _, name := range s.Names {
				lineStart := fset.Position(s.Pos()).Line
				lineEnd := fset.Position(s.End()).Line
				if s.Type != nil {
					lineEnd = fset.Position(s.End()).Line
				}
				if s.Values != nil && len(s.Values) > 0 {
					lineEnd = fset.Position(s.End()).Line
				}
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

// parseFuncDecl handles function and method declarations.
func parseFuncDecl(d *ast.FuncDecl, fset *token.FileSet, path string) symbol.Symbol {
	kind := symbol.Function
	parent := ""
	if d.Recv != nil && len(d.Recv.List) > 0 {
		kind = symbol.Method
		parent = receiverName(d.Recv.List[0].Type)
	}

	comment := docCommentFromGroup(d.Doc)

	return symbol.Symbol{
		Name:      d.Name.Name,
		Kind:      kind,
		FilePath:  path,
		LineStart: fset.Position(d.Pos()).Line,
		LineEnd:   fset.Position(d.End()).Line,
		Comment:   comment,
		Parent:    parent,
	}
}

// receiverName extracts the type name from a receiver expression.
func receiverName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

func docCommentFromGroup(g *ast.CommentGroup) string {
	if g == nil {
		return ""
	}
	return g.Text()
}

// ParseGoDir parses all .go files in a directory.
// This is a convenience for testing; ParseGoFile is the primary entry point.
func ParseGoDir(dir string) ([]symbol.Symbol, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var syms []symbol.Symbol
	for _, pkg := range pkgs {
		for path, f := range pkg.Files {
			fileSyms, err := parseASTFile(fset, f, path)
			if err != nil {
				return nil, err
			}
			syms = append(syms, fileSyms...)
		}
	}
	return syms, nil
}

func parseASTFile(fset *token.FileSet, f *ast.File, path string) ([]symbol.Symbol, error) {
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
