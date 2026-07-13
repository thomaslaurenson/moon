# Go style

Go-specific style. Assumes the core conventions.

## Doc comments

- Every package has a package comment immediately before `package`, starting `// Package <name>` followed by a summary sentence.
- Exported functions, methods, and types have a doc comment starting with the identifier name, as a complete sentence on a single opening line. Add a blank comment line before further paragraphs.
- Unexported functions do not require doc comments; add one when the purpose is not obvious.
- For interfaces, describe what the type represents rather than listing its methods.

```go
// Package parser reads and validates configuration files.
package parser

// ParseConfig reads a configuration file from path and returns a Config.
func ParseConfig(path string) (*Config, error) {
```

## Naming

- Exported identifiers use `CamelCase`; unexported use `mixedCase`. Never underscores (except `_test` package suffixes).
- Avoid stutter: do not repeat the package name in an exported name (`http.Client`, not `http.HTTPClient`).
- Acronyms are consistent case, all upper or all lower: `userID`, `parseURL`, `HTTPClient` (exported), `httpClient` (unexported). Never `userId` or `HttpClient`.
- Single-method interfaces take an `-er` suffix (`Reader`, `Closer`).

## Errors

- Sentinel error variables are prefixed `Err`: `var ErrNotFound = errors.New("not found")`.
- Custom error types end in `Error` and implement `Error() string`.

## Constants

Package-level constants use `CamelCase` when exported, `mixedCase` when unexported. Never `UPPER_SNAKE_CASE` in Go.

```go
const DefaultTimeout = 30 * time.Second
const maxRetries = 3
```
