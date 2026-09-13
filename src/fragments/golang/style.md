# Go style

Go-specific style. Assumes the core conventions.

## Doc comments

- Every package has a package comment immediately before `package`, starting `// Package <name>` followed by a summary sentence.
- A `package main` starts `// Command <binary>` instead. It documents a program to run rather than a package to import, and `// Package main` names something nobody can import.
- Exported functions, methods, and types have a doc comment whose first sentence starts on the opening line with the identifier name, wrapping as it needs to. Add a blank comment line before further paragraphs.
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

## Imports

Three groups, separated by a blank line and in this order: standard library, external modules, this module.

```go
import (
    "errors"
    "io"

    "github.com/spf13/cobra"

    "github.com/owner/mytool/internal/scanner"
)
```

`gofmt` will not do this. It sorts every import into one alphabetical block, so the grouping is invisible to `check_format` and drifts the moment two people edit the same file. `goimports -local <module>` produces the layout above and is what the `format` target runs (see the Makefile targets fragment).

## Naming

- Exported identifiers use `CamelCase`; unexported use `mixedCase`. Never underscores (except `_test` package suffixes).
- Avoid stutter: do not repeat the package name in an exported name (`http.Client`, not `http.HTTPClient`).
- Acronyms are consistent case, all upper or all lower: `userID`, `parseURL`, `HTTPClient` (exported), `httpClient` (unexported). Never `userId` or `HttpClient`.
- Single-method interfaces take an `-er` suffix (`Reader`, `Closer`).

## Interfaces and types

- Accept interfaces, return structs. A function takes the smallest interface it actually uses, and returns the concrete type so callers keep every method.
- Define an interface where it is consumed, not beside the type that satisfies it. The consumer knows which methods it needs; the implementation does not.
- Assert satisfaction at compile time when a type exists to satisfy an interface: `var _ Scanner = (*fileScanner)(nil)`. It costs nothing at runtime and turns a mismatch into a build error at the point of definition.
- Make the zero value useful where you can, so `var buf bytes.Buffer` works without a constructor. Where it cannot be, give the type a `New` function and keep the zero value obviously unusable rather than subtly wrong.
- Write `any`, never `interface{}`.
- Generics are for code whose implementation is identical across types; where behaviour varies, use an interface. Do not write a type parameter until you have written the same function twice with matching bodies. For slice and map helpers, use the stdlib `slices` and `maps` packages rather than writing your own.

## Control flow

Handle the error and return, rather than nesting the success path inside an `if`. The body of a function should read as the thing it does, with failures leaving from the left margin.

```go
// Good
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()
return parse(f)
```

## Package state and init

Read environment variables once at the wiring boundary in `cmd/`, and pass the values down as parameters. Packages under `internal/` take what they need as arguments and never call `os.Getenv` themselves. Treat the working directory the same way: resolve paths in `cmd/` rather than calling `os.Chdir`.

The point is to keep process-wide state out of the code that does the work. A package that reads the environment directly can only be exercised by mutating it, which forces `t.Setenv` on its tests and costs that whole test tree its parallelism (see the testing fragment).

The same reasoning rules out `init()` and package-level mutable variables. Build what a package needs in a constructor and return it, so a caller can build a second one with different dependencies. A package-level variable configured by `init()` can only ever hold one value per process, which is why a test that needs a different one cannot have it.

```go
// Good: dependencies are arguments, and a test can supply its own
func NewRootCmd(fsys fs.FS, out, errw io.Writer) *cobra.Command

// Bad: nothing can be injected, and flag state persists between runs
var rootCmd = &cobra.Command{Use: "mytool"}

func init() {
    rootCmd.Flags().BoolVar(&verbose, "verbose", false, "")
}
```

Package-level variables that are never written to are fine: a lookup table, a compiled regular expression, a sentinel error, an `embed.FS`. The rule is about mutable state, not about the keyword.

The one sanctioned `init()` is the version stamp in `cmd/version.go`, where the value has to be settled before any command runs and every alternative silently discards what the linker wrote (see the scaffolding fragment). It qualifies because it derives a single fact from the binary itself, not from configuration a test would want to vary.

Cobra's own generator emits the `init()` form, so it is common in examples. It predates the practice of injecting writers and contexts, and cannot support either.

## Constants

Package-level constants use `CamelCase` when exported, `mixedCase` when unexported. Never `UPPER_SNAKE_CASE` in Go.

```go
const DefaultTimeout = 30 * time.Second
const maxRetries = 3
```
