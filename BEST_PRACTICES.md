# Best Practices — aimap

## Overall Architecture

`aimap` is a single-binary CLI tool with no external dependencies. It uses the Go standard library exclusively. The architecture follows a pipeline pattern:

```
CLI entry → config loading → file tree scan → source parsing → symbol output
```

## Module Organization

The project is organized into two top-level directories:

- **`cmd/`** — contains the CLI entry point (`cmd/aimap/main.go`). Only `main()` and flag parsing live here.
- **`internal/`** — all private packages. No public API is exposed; `internal/` enforces this at the compiler level.

Internal packages are organized by responsibility, not by file type:

| Package | Responsibility |
|---|---|
| `internal/scanner/` | Walk file tree, ignore patterns, collect file list |
| `internal/parser/` | Go and Python source parsing into symbols |
| `internal/symbol/` | Symbol model: `Symbol` struct, `SymbolKind` enum |
| `internal/ignore/` | `.aimapignore` parsing (gitignore-style pattern matching) |
| `internal/output/` | Render symbol list to `MAP.md` |
| `internal/config/` | Load optional configuration |

## File Organization

### File size limits

- `.go` files should stay at or under **500 lines**.
- No `.go` file should exceed **1000 lines** except where splitting is impractical.
- Exceptions (require explicit note in `BEST_PRACTICES.md`): generated code, large data/string tables, files with a strong technical reason to stay unified.
- **Current state**: the largest `.go` file is `parser_go.go` at 200 lines and `parser_python.go` at 213 lines. All files are well under the 500-line target. No exceptions documented.

### Splitting strategy

When a `.go` file approaches 500 lines, split it following these conventions:

- **By type**: if a file defines multiple large types, split into one file per type (e.g., `symbol.go`, `symbol_kind.go`).
- **By responsibility**: if a package has multiple distinct concerns, split into files per concern (e.g., `parser.go`, `parser_go.go`, `parser_python.go`).
- **By feature**: if a package handles multiple features, split into files per feature (e.g., `ignore.go`, `ignore_test.go`).
- **By layer**: keep CLI, logic, and model in separate packages (never in the same package).

### Naming within a package

- One primary type or responsibility per file.
- Test files mirror the source file: `foo.go` → `foo_test.go`.
- Helper files prefixed with the concept: `scanner.go`, `scanner_test.go`, `ignore.go`, `ignore_test.go`.

## Naming Conventions

- **Packages**: lowercase, single word, no underscores (`scanner`, `parser`, `symbol`, `ignore`, `output`, `config`).
- **Files**: lowercase snake_case (`main.go`, `parser_go.go`, `parser_python.go`).
- **Exported symbols**: PascalCase (`Symbol`, `SymbolKind`, `ParseFile`).
- **Unexported symbols**: camelCase (`parsedSymbols`, `ignorePatterns`).
- **Acronyms**: uppercase (`HTTP`, `ID`, `AST`) — all caps, not `Http` or `Id`.
- **Interfaces**: name with `-er` suffix when single method (`Scanner`, `Parser`, `Walker`). Use `I` prefix only for C++ interop (not needed here).
- **Errors**: `var ErrFoo = errors.New(...)` for package-level sentinel errors.
- **Constants**: PascalCase for exported (`MaxLineLength`), camelCase for unexported (`defaultIgnoreList`).

## Function Design

- **Single responsibility**: each function does one thing. If a function is doing multiple things, split it.
- **Return errors**: return errors rather than panicking. Panic only for truly unrecoverable states (e.g., failed assertions in tests).
- **Avoid `init()`**: do not use `init()` functions. Use explicit initialization in `main()` or `New*` constructors.
- **Receiver naming**: single-letter receiver name (`s *Scanner`, `p *Parser`). Consistent across all methods on the same type.
- **Method vs function**: use methods when the operation depends on receiver state. Use standalone functions otherwise.

## Class/Struct Design

- **Zero-value usability**: structs should be usable in their zero-value state when possible (e.g., `Scanner{}` is ready to use with defaults).
- **Constructor functions**: use `New*` functions for structs that require initialization (`NewScanner(root string) *Scanner`).
- **Functional options**: use the functional options pattern for optional configuration:

```go
type Option func(*Scanner)
func WithIgnorePatterns(patterns []string) Option { ... }
func NewScanner(root string, opts ...Option) *Scanner
```

- **No exported fields without getters/setters**: prefer unexported fields with exported methods.
- **No `interface{}`**: use `any` (Go 1.18+). But prefer concrete types or generics over `any`.

## Error Handling

- **Always check errors**: errors from file I/O, parsing, and path operations must always be checked.
- **Wrap errors with context**: use `fmt.Errorf("reading file %s: %w", path, err)` to add context.
- **Sentinel errors**: define `var Err*` for errors callers might need to check.
- **No silent swallowing**: never `_ = fn()` or `_, _ = fn()` to ignore errors.
- **Defer close with error check**: use a helper or explicit check:

```go
f, err := os.Open(path)
if err != nil { return err }
defer func() {
    if cerr := f.Close(); cerr != nil { err = cerr }
}()
```

## Logging

- **No logging library**: use `fmt.Fprintf(os.Stderr, ...)` for user-facing messages and `fmt.Fprintf(os.Stderr, "debug: ...")` for debug output.
- **Use `-v` / `--verbose` flag**: controlled by a global verbose flag, not a log level.
- **No logging of file contents**: log only filenames, counts, and progress, never the content of scanned files.
- **Progress output**: write progress to stderr so stdout can be used for other purposes.

## Configuration Management

- **No config file required**: aimap works with zero configuration.
- **`.aimapignore`**: optional file in the project root. Uses gitignore syntax.
- **No JSON/YAML/TOML**: avoid structured config formats. Keep it simple.
- **CLI flags**: only essential flags. Use `flag` package (stdlib).

## Dependency Injection

- **Not used**. The project is small enough that explicit constructor parameters and functional options suffice.
- If tests need mocking, define interfaces in the consuming package and pass mock implementations in tests.

## Testing Conventions

See `TESTING.md` for full details.

- **Location**: test files alongside source files (`foo.go` → `foo_test.go`).
- **Framework**: standard `testing` package only. No third-party test frameworks.
- **Table-driven tests**: preferred pattern for testing multiple inputs/outputs.
- **Golden files**: for output comparison (MAP.md generation), use `testdata/golden/` files.
- **Coverage target**: aim for 80%+ coverage on core logic (parser, scanner, ignore).

## Documentation Style

- **Go doc comments**: full sentences, start with the symbol name. Example:

```go
// Scanner walks a directory tree and collects source files to scan.
type Scanner struct { ... }
```

- **Package doc**: a `doc.go` file in each package with a concise overview.
- **No comments inside functions**: prefer self-documenting code. Use comments only when the "why" is non-obvious.
- **Exported symbols always documented**: every exported type, function, method, and constant gets a doc comment.

## Performance Considerations

- **Single-pass file walking**: walk the tree once, collect files, then parse.
- **Parallel parsing**: use `sync.WaitGroup` or `errgroup` to parse multiple files concurrently.
- **Limit concurrency**: bound parallel parsing with a semaphore channel (`sem := make(chan struct{}, runtime.NumCPU())`).
- **No premature optimization**: write clear code first, profile if needed, then optimize.
- **Avoid `fmt.Sprintf` in hot paths**: use `strconv` or `strings.Builder` for line-number formatting.
- **Use `bufio.Scanner`**: for reading large files, not `ioutil.ReadAll` or `os.ReadFile` on large files.

## Code Review Expectations

- **Every PR must include or update tests**.
- **Every PR must update `MAP.md`** — update the architectural overview manually, then regenerate the symbol map with `go run ./cmd/aimap/`.
- **No external dependencies** — any proposed dependency must be justified in the PR description.
- **File size check** — verify no `.go` file exceeds 500 lines (soft) or 1000 lines (hard).
- **`go vet` must pass** — run `go vet ./...` before submitting.
- **No unused code** — run `go mod tidy` and ensure no unused imports or variables.

## Refactoring Guidelines

- **Split large files first**: when a `.go` file exceeds 500 lines, split it before adding new code.
- **Preserve package structure**: do not move symbols between packages as part of refactoring unless the package boundary is wrong.
- **One refactoring per PR**: avoid mixing refactoring with feature work.
- **Update `MAP.md`**: after any structural change, update the architectural overview and regenerate the symbol map with `go run ./cmd/aimap/`.
- **Run tests after refactoring**: ensure no regressions before submitting.

## Anti-patterns to Avoid

- **God objects**: do not create a single package or file that does everything. Split by responsibility.
- **Global state**: avoid package-level variables except for `Err*` sentinel errors and `Default*` values.
- **`init()` functions**: avoid them. They make testing and initialization order unpredictable.
- **`ioutil` usage**: deprecated since Go 1.16. Use `os` and `io` packages instead.
- **`io/ioutil`**: do not import. Use `os.ReadFile`, `os.WriteFile`, `os.CreateTemp` instead.
- **Silent errors**: never ignore errors from `Close()`, `Write()`, or `Seek()`.
- **Panic for errors**: never use `panic` for recoverable errors. Only panic for programmer mistakes (e.g., impossible switch cases).
- **Deep nesting**: avoid nesting beyond 3 levels. Extract helper functions or early-return.
- **Duplicate code**: do not copy-paste parser logic for Go and Python. Extract common parsing patterns into shared helpers.
- **Magic numbers**: define constants for magic numbers (e.g., `MaxLineLength = 1000`).
- **Unbounded reads**: always limit the size of files read into memory. Use `io.LimitReader` for large files.