# Testing Guide — aimap

## Testing Framework

- **Standard library**: `testing` package only. No third-party test frameworks.
- **Test runner**: `go test` (standard Go toolchain).
- **Coverage tool**: `go test -cover` and `go test -coverprofile=coverage.out`.
- **Mocking**: no mocking library. Use handwritten mock implementations via interfaces.
- **Golden files**: for output comparison (MAP.md generation via `internal/output`), use `testdata/golden/` files.

## Running Tests

### All tests

```bash
go test ./...
```

### Single package

```bash
go test ./internal/parser/
```

### Single test

```bash
go test -run TestParseGoFunction ./internal/parser/
```

### With coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out   # HTML report
```

### Verbose mode

```bash
go test -v ./...
```

### Race detection

```bash
go test -race ./...
```

### Lint + tests

```bash
go vet ./... && go test ./...
```

## Test Organization

### Location

- Test files live alongside the source files they test.
- `foo.go` → `foo_test.go`
- `internal/parser/parser.go` → `internal/parser/parser_test.go`

### Naming

- **Test functions**: `Test<FunctionName>(t *testing.T)`.
- **Table test functions**: `Test<FunctionName>(t *testing.T)` with a `tests` slice inside.
- **Helper functions**: `helper<Description>(t *testing.T)`.
- **Example functions**: `Example<FunctionName>()` for runnable examples.

### Fixture organization

```
internal/
└── parser/
    ├── testdata/
    │   ├── src/           # Sample source files for parsing
    │   │   ├── hello.go
    │   │   ├── hello.py
    │   │   ├── malformed.go
    │   │   └── malformed.py
    │   └── golden/        # Expected output for generated MAP.md
    │       └── full_project.md
    ├── parser_test.go
    └── parser.go
```

### Helper utilities

- Shared test helpers go in `internal/<package>/export_test.go` (for accessing unexported symbols).
- Package-level test helpers: `internal/<package>/testhelper_test.go` (filename suffix `_test.go` keeps them in `_test` build tag).
- Cross-package test helpers: `internal/testutil/` (if shared across packages).

### Mock organization

- Mock implementations live in the same package as the test (e.g., `scanner_test.go` defines `mockWalker`).
- If a mock is shared across packages, put it in `internal/testutil/`.

## Writing New Tests

### Where to place new tests

- **New function/method**: add tests in the existing `*_test.go` file for that package.
- **New file**: create `newfile_test.go` alongside `newfile.go`.
- **New package**: create `package_test.go` (or `doc_test.go`) in the new package directory.

### How to name test files

- `foo_test.go` for tests of `foo.go`.
- `export_test.go` to export unexported symbols for testing (white-box tests).

### How to structure test cases

Use **table-driven tests** as the primary pattern:

```go
func TestParseFile(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    []symbol.Symbol
        wantErr bool
    }{
        {
            name:  "empty file",
            input: "",
            want:  nil,
        },
        {
            name:  "go function",
            input: "func Foo() {}",
            want:  []symbol.Symbol{{Name: "Foo", Kind: symbol.Function, LineStart: 1, LineEnd: 1}},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseFile("test.go", strings.NewReader(tt.input))
            if (err != nil) != tt.wantErr {
                t.Fatalf("ParseFile() error = %v, wantErr = %v", err, tt.wantErr)
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseFile() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### How to test new modules

- Create `*_test.go` in the module's package directory.
- Define interfaces for dependencies that the module accepts.
- Create minimal test fixtures in `testdata/` under the package.

### How to test bug fixes

- Write a **regression test** that reproduces the bug before fixing it.
- Name the test case after the bug: `"issue 42: malformed python file"`.
- Include the minimal input that triggers the bug.
- Verify the fix makes the test pass.

### How to test regressions

- Add a test case to the existing table-driven test.
- If the regression is a complex scenario, add a fixture file to `testdata/src/` and reference it.
- Run the full test suite before submitting.

### How to add fixtures or mocks

- **Fixtures**: put source files in `testdata/src/`. Reference them by a relative path from the test file.
- **Mocks**: define interfaces for dependencies (e.g., `FileSystem` interface in `scanner/`). Write a `mockFileSystem` in the test file that implements the interface.
- **Golden files**: put expected output in `testdata/golden/`. Use `flag.Update()` to allow `go test -update` to rewrite golden files.

### Assertion style

- Use `t.Fatal`/`t.Fatalf` for setup failures.
- Use `t.Error`/`t.Errorf` for non-fatal assertion failures.
- Use `t.Cleanup` for teardown.
- Use `cmp.Diff` from `github.com/google/go-cmp` only if the project eventually adds it. For now, use `reflect.DeepEqual` (stdlib).
- Prefer specific error checks over string matching: use `errors.Is` for sentinel errors, not `strings.Contains`.

### Setup and teardown conventions

- **Setup**: use `t.Setenv` for environment variables (auto-restored after test).
- **Teardown**: use `t.Cleanup(func() { ... })` instead of `defer` in test functions.
- **Temp directories**: use `t.TempDir()` for temporary files (auto-cleaned after test).
- **Parallel tests**: call `t.Parallel()` at the start of the test function, not inside subtables.

## AI Contributor Guidance

1. **Every new feature must include tests**. Add tests in the same PR as the feature.
2. **Every bug fix must include a regression test**. Write the test that reproduces the bug, then fix it.
3. **Follow existing conventions**. Look at existing tests in the same package before writing new ones.
4. **Reuse existing infrastructure**. Before creating new helpers, fixtures, or mocks, check if something suitable already exists in `internal/testutil/` or the package's `export_test.go`.
5. **Place tests consistently**. Follow the organizational pattern used by similar tests in the codebase.
6. **Run the smallest relevant test set** during development, then run the full suite before submitting.

## CI and Quality Gates

### Current state

No CI configuration exists yet. The following are planned quality gates:

- `go vet ./...` must pass.
- `go test -race ./...` must pass.
- `go test -coverprofile=coverage.out ./...` — minimum 80% coverage for core packages (`parser`, `scanner`, `ignore`).
- No external dependencies.
- All tests pass on Go 1.26.x.

### Commands to run before submitting

```bash
go vet ./...
go test -race ./...
go test -cover ./...
```

### Formatting

- `gofmt` (or `go fmt ./...`) must be run on all `.go` files.
- No trailing whitespace.
- Standard Go formatting: tabs for indentation, `gofmt` style.