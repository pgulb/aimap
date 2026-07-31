# Security — aimap

## Security Goals

- **Read-only file system access** — aimap scans files but must never modify them.
- **No network access** — aimap is a local CLI tool with no network I/O.
- **No credential exposure** — aimap must never read, log, or transmit credentials, secrets, or environment variables.
- **Safe path handling** — all file paths must be handled safely to prevent path traversal outside the intended project root.

## Input Validation

- **File paths** — validate that scanned paths are within the project root. Reject symbolic links that escape the project root.
- **`.aimapignore` patterns** — parse and validate before use. Reject malformed patterns with a clear error.
- **Source code input** — parsers must handle malformed Go/Python source gracefully (error recovery, no panics).

## File Handling

- **Read-only operations** — all file access must be read-only. Never write to any file except `MAP.md` in the project root.
- **Path traversal** — use `filepath.Clean` and `filepath.Abs` on all paths. Verify the resolved path is within the project root before reading.
- **Symlink safety** — do not follow symlinks that point outside the project root.
- **Binary files** — detect and skip binary files to avoid processing garbage input.

## Secrets and Configuration Management

- **No secrets in code** — aimap has no API keys, tokens, or credentials. Do not introduce any.
- **`.aimapignore`** — treat as a configuration file, not a secrets file. Document that users should not store secrets in it.
- **Environment variables** — aimap should not read environment variables. If needed in the future, never log or expose them.

## Logging and Sensitive Data

- **No logging of file contents** — log only file paths and line counts, never the content of scanned files.
- **No personal data** — do not log user names, home directory paths, or system information.
- **Error messages** — sanitize error messages to avoid leaking file contents or system paths.

## Dependency Management

- **Zero external dependencies** — aimap uses only Go standard library. This eliminates supply-chain risks.
- **Go toolchain** — pin the Go version in `go.mod` and test against it. Update deliberately, not automatically.
- **No vendoring** — not needed with zero dependencies.

## API Security

- **No network API** — aimap is a local CLI. Do not introduce HTTP servers, network calls, or remote APIs.

## SQL/Database Safety

- **No database** — aimap has no database or SQL component. Not applicable.

## Authentication and Authorization

- **No authentication** — aimap is a local CLI tool with no multi-user or remote access features. Not applicable.

## Secure Coding Guidelines

- **Use `filepath.Clean`** on all user-supplied paths before use.
- **Use `filepath.WalkDir`** with `fs.SkipDir` for safe directory traversal.
- **Limit panic recovery** — use `defer/recover` only at the top-level entry point and in parser goroutines. Do not recover panics silently in general logic.
- **No `unsafe` package** — avoid `unsafe` unless absolutely necessary and reviewed.
- **No `os/exec`** — do not shell out to external tools.
- **No `ioutil`** — use `os` and `io` packages (Go 1.16+).
- **No reflection on user input** — avoid `reflect` on data derived from file contents.

## Common Mistakes to Avoid

- Using `os.Open` without resolving the path first, potentially allowing path traversal.
- Logging file contents for debugging, which could accidentally include secrets or credentials.
- Following symlinks without checking if they escape the project root.
- Using `ioutil.ReadAll` on arbitrarily large files (use `bufio.Scanner` with a size limit).
- Silently recovering panics in parser logic, hiding bugs.
- Making network calls or shelling out, which breaks the security model.

## Security Checklist for Contributors

- [ ] All file reads use `filepath.Clean` + bounds check.
- [ ] Symlinks are not followed outside the project root.
- [ ] No file contents are logged or written to any output other than `MAP.md`.
- [ ] No network calls or `os/exec` invocations.
- [ ] Malformed `.aimapignore` patterns produce an error, not a panic.
- [ ] Parsers handle malformed source code gracefully (no panics from untrusted input).
- [ ] No secrets, tokens, or credentials are hardcoded or read.
- [ ] No external dependencies added.
- [ ] Binary files are detected and skipped.
- [ ] `MAP.md` is only written to the project root, never to arbitrary paths.