# Go Style Guide

Style conventions for Go code in this project.

## Unusual Characters

- Never use em dash (—)

## Comments

### Package comments

Every package has a package comment immediately before the `package` declaration.
Start with `// Package <name>` followed by a summary sentence.

```go
// Package parser reads and validates configuration files.
package parser
```

For packages with more to say, extend with a blank comment line and further paragraphs:

```go
// Package retry provides helpers for retrying operations with backoff.
//
// It supports fixed and exponential backoff strategies and respects
// context cancellation on every attempt.
package retry
```

### Function and method comments

Exported functions and methods must have a doc comment. Start with the name of the
function as the first word. Use a complete sentence. Put the summary on a single opening
line — godoc shows this line in index views.

```go
// ParseConfig reads a configuration file from path and returns a Config.
func ParseConfig(path string) (*Config, error) {
```

For more detail, add a blank comment line after the summary:

```go
// Run executes the command and blocks until it exits.
//
// The context is used to cancel execution before it completes naturally.
// A non-zero exit code is returned as an *ExitError.
func Run(ctx context.Context, args []string) error {
```

Unexported functions do not require doc comments, but add one when the purpose is
not obvious.

### Type comments

Exported types follow the same pattern — start with the type name:

```go
// Config holds the resolved settings for a single run.
type Config struct {
    Timeout time.Duration
    Output  string
}
```

For interfaces, describe what the type represents rather than listing its methods:

```go
// Store is a persistent key-value backend.
type Store interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key, value string) error
}
```

---

## Inline Comments

- Start with `//` followed by a single space.
- First word is capitalised.
- Never use a full stop, unless a multiline comment.
- For continuation lines (when a comment wraps to a second line), capitalisation is not required.
- No decorative styles — avoid `// ---`, `// ===`, `// ***`, or similar dividers.

```go
// Good: single-line comment
x := computeValue()

// Good: multi-line comment where only the first line is capitalised,
// continuation lines do not need to start with a capital letter.
y := complexOperation()

// Bad: decorative divider — avoid this style
// --- section name ---
```

---

## Comment Hygiene

- Do not write step narration comments that describe the next line of code.
  Bad: `// Initialise the counter`, `// Check for nil`
- Preserve comments that explain why something is done, not what.
  Good: `// Use index 0 because the legacy API expects a 0-indexed fallback`
- Do not use trailing block comments (`// end if`, `// end for`) unless the block
  exceeds 50 lines.
- Do not inject `TODO` or `FIXME` comments unless they refer to a real, known issue.

---

## Spelling

Use British English spellings:

- `Initialise` not `Initialize`
- `Colour` not `Color`

---

## Naming

### General rules

- Exported identifiers use `CamelCase`.
- Unexported identifiers use `mixedCase` (lower camel case).
- Never use underscores in identifier names (except test package suffixes `_test`).
- Avoid stutter: do not repeat the package name in an exported name.

```go
// Bad: package is "http", name stutters
http.HTTPClient

// Good
http.Client
```

### Acronyms

Acronyms are written in consistent case — either all upper or all lower — never mixed:

```go
// Good
userID
parseURL
httpClient  // unexported
HTTPClient  // exported

// Bad
userId
parseUrl
HttpClient
```

### Interfaces

Single-method interfaces are typically named after the method with an `-er` suffix:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}
```

### Errors

Sentinel error variables are prefixed with `Err`:

```go
var ErrNotFound = errors.New("not found")
var ErrTimeout = errors.New("operation timed out")
```

Custom error types end in `Error`:

```go
type ParseError struct {
    Line int
    Msg  string
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
}
```

### Constants

Package-level constants use `CamelCase` when exported, `mixedCase` when unexported.
Do not use `UPPER_SNAKE_CASE` for constants in Go.

```go
// Exported constant
const DefaultTimeout = 30 * time.Second

// Unexported constant
const maxRetries = 3
```
