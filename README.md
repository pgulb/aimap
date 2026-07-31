# aimap — AI-friendly project map generator

Scans Go and Python projects and generates `MAP.md` — a symbol map of functions, methods, types, variables, and constants with line numbers and doc comments. Designed to help AI agents (and humans) navigate codebases quickly.

## Why aimap?

Instead of full-repository searches, AI agents (and contributors) can read `MAP.md` to locate symbols instantly — faster, cheaper, and more predictable. Think of it as a compiled index for your codebase.

## Installation

### Quick install (recommended)

**Linux / macOS:**
```bash
curl -sfL https://raw.githubusercontent.com/pgulb/aimap/main/install.sh | sh
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/pgulb/aimap/main/install.ps1 | iex
```

The script downloads the latest pre-built binary for your platform and installs it to `/usr/local/bin` (Linux/macOS) or `%LOCALAPPDATA%\Programs` (Windows).

### With Go installed

```bash
go install github.com/pgulb/aimap/cmd/aimap@latest
```

### Build from source

```bash
git clone https://github.com/pgulb/aimap.git
cd aimap
go build -o aimap ./cmd/aimap/
```

## Usage

### Basic

```bash
aimap
```

Scans the current directory and writes `MAP.md`.

### Flags

| Flag | Default | Description |
|---|---|---|
| `-path` | `.` | Project root directory |
| `-output` | `MAP.md` | Output file path (overrides `--dev` default) |
| `-v` | false | Verbose output to stderr |
| `-dev` | false | Use `.aimapignore_dev` and `MAP_dev.md` |

### Examples

Scan a specific project:

```bash
aimap -path ~/projects/my-app
```

Development mode (separate files, safe to gitignore):

```bash
aimap --dev
```

Custom output:

```bash
aimap -output DOCS.md
```

## How it works

```
CLI → config loading → file tree scan → source parsing → markdown output
```

1. Configures ignores from `.aimapignore` (optional) plus built-in patterns (`node_modules`, `.venv`, `.git`, `__pycache__`, etc.).
2. Walks the directory tree, collecting `.go` and `.py` files.
3. Parses Go source with `go/ast` (stdlib) and Python with a line-based state machine.
4. Renders every symbol into `MAP.md` — grouped by language, then by file, with line ranges and doc comments.

## Output structure

`MAP.md` has two sections:

1. **Architectural overview** (top) — manually maintained. Describes project structure, packages, key types, and cross-module dependencies.
2. **Symbol map** (below `<!-- SYMBOL MAP -->`) — auto-generated. Lists every symbol grouped by language and file, with line ranges and doc comments.

The auto-generated section is safe to regenerate at any time. The architectural overview is preserved across runs.

## .aimapignore

Optional file in the project root using gitignore syntax. Add patterns to exclude files and directories from the scan:

```
# .aimapignore
*.pb.go
build/
tmp/
```

Built-in ignores already cover `node_modules`, `.venv`, `.git`, `__pycache__`, `.DS_Store`, `.idea`, vendor directories, and common binary/object file extensions.

## Development

```bash
aimap --dev              # use MAP_dev.md and .aimapignore_dev
go test ./...            # run tests
go vet ./...             # lint
go test -race ./...      # race detection
```

## Contributing

- `AGENTS.md` — AI agent guidance and MAP.md maintenance workflow.
- `BEST_PRACTICES.md` — coding conventions, file organization, naming, error handling.
- `TESTING.md` — testing framework, table-driven tests, fixture organization, CI expectations.

## License

MIT
