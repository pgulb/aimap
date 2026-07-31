# AGENTS.md — AI Agent Guidance for aimap

## Project Overview

aimap is a CLI tool written in Go that scans Go and Python projects and generates `MAP.md` — a symbol map of functions, objects, methods, variables, line ranges, and comments. It is designed to help AI agents (and human contributors) locate code symbols quickly without full-repository searches.

- **Primary usage**: run without parameters to regenerate `MAP.md` in the project root.
- **Target languages**: Go and Python source code, plus popular config/settings files (e.g., `pyproject.toml`, `Taskfile.yml`, `Makefile`).
- **Implementation**: modern Go, using the standard library primarily.
- **No external dependencies**: the tool relies on Go's rich standard library.
- **Ignored paths**: automatically skips `node_modules`, `.venv`, `venv`, `.git`, `.gitignore`, `.dockerignore`, `__pycache__`, etc.
- **Configurable ignores**: `.aimapignore` file with gitignore syntax for user-defined exclusions.
- **Self-hosting**: aimap will be used by AI agents to implement aimap itself, once its features are sufficient.

## Using MAP.md

`MAP.md` is the primary navigation document for this repository. It has two sections:

1. **Architectural overview** (top) — manually maintained by contributors. Describes project structure, packages, key types, and cross-module dependencies.
2. **Symbol map** (below the overview) — **automatically generated** by running `aimap` in the project root. Lists every function, method, type, variable, and constant with file paths and line numbers.

**AI agents must consult `MAP.md` before exploring the repository or searching for code.**

Its purpose is to allow contributors to quickly locate the relevant parts of the codebase without performing unnecessary repository-wide searches.

`MAP.md` provides a concise, high-level architectural overview of the project rather than implementation details.

It answers questions such as:

- Where is a feature implemented?
- Which module owns a responsibility?
- Which package exposes a public API?
- Where are important classes, functions, or types located?
- Which modules interact with each other?
- Where should new functionality be added?
- Which files are the primary entry points?

### Workflow

1. **Read `MAP.md`** before searching the repository.
2. **Use `MAP.md`** to determine which files and modules are relevant.
3. **Restrict exploration** to those locations whenever practical.
4. **Perform broader searches** only if `MAP.md` is incomplete or outdated.
5. **Update `MAP.md`** if broader searches reveal missing or incorrect architectural information.
6. **Treat `MAP.md`** as the canonical architectural index of the repository.

### Regenerating the symbol listing

Run `aimap` from the project root to regenerate the symbol map section of `MAP.md`:

```bash
go run ./cmd/aimap/
# or, after building:
aimap
```

This scans all `.go` and `.py` files (excluding ignored paths) and rewrites the symbol listing. The architectural overview at the top of `MAP.md` is preserved.

## Maintaining MAP.md

`MAP.md` has two parts that are maintained differently:

- **Architectural overview** (top section) — **manually maintained** by contributors. Update it whenever the project's architecture changes.
- **Symbol map** (below the overview) — **automatically generated** by `aimap`. Regenerate it after adding, renaming, or deleting source files.

### When to update the architectural overview

Update the architectural overview after:

- creating files
- deleting files
- renaming files
- moving files
- adding or removing directories
- adding or removing major modules or packages
- adding or removing significant public APIs
- adding or removing important Python functions or classes
- adding or removing important Go functions or types
- adding or removing important services, controllers, handlers, commands, or entry points
- introducing new architectural layers
- changing ownership or responsibilities of modules
- adding or removing important cross-module dependencies
- significant project restructuring
- architectural changes affecting project organization
- splitting a `.go` or `.py` file into multiple files for size or organizational reasons

### When to regenerate the symbol map

Run `aimap` to regenerate the symbol map after:

- creating, deleting, renaming, or moving any `.go` or `.py` file
- adding, removing, or renaming any function, method, type, variable, or constant

### How to update MAP.md

- Preserve existing useful content in the architectural overview.
- Update only the sections affected by the current task.
- Do not rewrite the entire document unless necessary.
- Remove obsolete information.
- Keep descriptions concise and focused on helping contributors navigate the project.
- Prefer architectural information over implementation details.
- Include important directories, packages, modules, entry points, public APIs, major classes/functions/types, and significant relationships between components.
- Reference the primary location of important functionality rather than every implementation detail.
- Keep formatting and ordering stable whenever possible.
- Avoid unnecessary wording, formatting, or organizational changes unrelated to the current task.
- Before completing any coding task, verify that `MAP.md` accurately reflects the current repository structure.
- After making structural changes, run `aimap` to regenerate the symbol map, then review and update the architectural overview.

## File Size Limits for .go and .py Files

AI agents must follow this rule whenever creating or editing `.go` or `.py` files:

- Keep `.go` and `.py` files at roughly **500–1000 lines maximum**. Treat 500 lines as the soft target and 1000 lines as a hard ceiling.
- Before finishing a task that adds substantial code to a `.go` or `.py` file, check its resulting line count. If it is approaching or exceeding the limit, split it into smaller files along logical boundaries (by type, class, feature, handler group, or responsibility), consistent with the project's existing organizational conventions.
- Prefer splitting proactively while writing new code rather than writing a large file and refactoring afterward.
- When splitting a file, update all internal imports/references and update `MAP.md` to reflect the new file locations.
- If a large file cannot reasonably be split, leave it as-is but note the exception in `BEST_PRACTICES.md`.

## Additional Guidance

- **Security**: follow `SECURITY.md` — read-only access, no network calls, safe path handling.
- **Best practices**: follow `BEST_PRACTICES.md` — naming, error handling, testing, file organization conventions.
- **Testing**: follow `TESTING.md` — always write tests, use table-driven tests, include regression tests for bug fixes.
- **No external dependencies**: aimap uses Go standard library only. Do not add third-party packages.
- **Documentation**: every exported symbol must have a Go doc comment. Update `MAP.md` when architecture changes.