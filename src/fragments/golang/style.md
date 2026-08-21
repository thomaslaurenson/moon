# Go style

Go-specific style. Assumes the core conventions.

## Doc comments

- Every package has a package comment immediately before `package`, starting `// Package <name>` followed by a summary sentence.
- A `package main` starts `// Command <binary>` instead. It documents a program to run rather than a package to import, and `// Package main` names something nobody can import.
- Exported functions, methods, and types have a doc comment starting with the identifier name, as a complete sentence on a single opening line. Add a blank comment line before further paragraphs.
- Unexported functions do not require doc comments; add one when the purpose is not obvious.
- For interfaces, describe what the type represents rather than listing its methods.
- Put the package comment in `doc.go` when it runs to more than a couple of lines, so it is not buried above an unrelated file's code.

```go
// Package parser reads and validates configuration files.
package parser

// ParseConfig reads a configuration file from path and returns a Config.
func ParseConfig(path string) (*Config, error) {
```

```go
// Command mytool converts configuration files between formats.
package main
```

## Naming

- Exported identifiers use `CamelCase`; unexported use `mixedCase`. Never underscores (except `_test` package suffixes).
- Avoid stutter: do not repeat the package name in an exported name (`http.Client`, not `http.HTTPClient`).
- Acronyms are consistent case, all upper or all lower: `userID`, `parseURL`, `HTTPClient` (exported), `httpClient` (unexported). Never `userId` or `HttpClient`.
- Single-method interfaces take an `-er` suffix (`Reader`, `Closer`).

## Environment and process state

Read environment variables once at the wiring boundary in `cmd/`, and pass the values down as parameters. Packages under `internal/` take what they need as arguments and never call `os.Getenv` themselves. Treat the working directory the same way: resolve paths in `cmd/` rather than calling `os.Chdir`.

The point is to keep process-wide state out of the code that does the work. A package that reads the environment directly can only be exercised by mutating it, which forces `t.Setenv` on its tests and costs that whole test tree its parallelism (see the testing fragment).

## Constants

Package-level constants use `CamelCase` when exported, `mixedCase` when unexported. Never `UPPER_SNAKE_CASE` in Go.

```go
const DefaultTimeout = 30 * time.Second
const maxRetries = 3
```
