package symbol

// SymbolKind identifies the kind of a declared symbol.
type SymbolKind int

const (
	Function   SymbolKind = iota // Top-level function
	Method                       // Method on a type
	Struct                       // Go struct type
	Interface                    // Go interface type
	Type                         // Go type alias or defined type
	Variable                     // Variable declaration
	Constant                     // Constant declaration
	Class                        // Python class
)

// Symbol represents a named declaration in source code.
type Symbol struct {
	Name      string
	Kind      SymbolKind
	FilePath  string
	LineStart int
	LineEnd   int
	Comment   string   // Doc comment text, if any
	Parent    string   // Parent type name (for methods)
}
