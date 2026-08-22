# Go error handling

How a failure travels from `internal/` to a user. Assumes the Go style fragment.

## The boundary

Packages under `internal/` report failure and stop there. They never decide what a failure means to a user:

- Never call `os.Exit` or `log.Fatal`. The caller owns the process lifetime.
- Never write to stdout or stderr to report an error. The caller owns the streams.
- Carry the facts (the operation, the input, the underlying error) and let `cmd/` do the wording.

`cmd/` and `main` are where a failure becomes an exit code and a line on stderr, and that translation lives in exactly one place. This is the same boundary the style fragment draws for environment variables, for the same reason: the process belongs to the entry point, not to the code doing the work.

## Wrapping

- Wrap with `%w` when adding context: `fmt.Errorf("read config %q: %w", path, err)`. This keeps `errors.Is` and `errors.As` working for every caller above.
- Use `%v` only to break the chain deliberately, where the underlying error is an implementation detail no caller should be able to match on. Breaking the chain is a decision, and a rare one.
- Add a layer only when it adds a fact the caller does not already have. An error you are passing straight back is returned unchanged, not wrapped in a restatement of itself.
- One wrap per boundary. A path that wraps at every frame produces `report: read: open: no such file or directory`, which says the same thing four times.

## Error strings

- Lowercase, no trailing punctuation. An error string is a fragment that ends up inside a larger sentence once something wraps it.
- No `failed to` or `error` prefix. The type already says it is an error; the string says what was being attempted.
- Name the input, quoted with `%q`, so the message identifies which one of many failed.

```go
// Good
fmt.Errorf("parse config %q: %w", path, err)

// Bad: capitalised, punctuated, drops the chain, and the prefix carries nothing
fmt.Errorf("Failed to parse config: %v.", err)
```

## Sentinels and typed errors

- Sentinel error variables are prefixed `Err`: `var ErrNotFound = errors.New("not found")`.
- Custom error types end in `Error` and implement `Error() string`.
- Export a sentinel only when callers outside the package need to match it. An unexported one is still worth having so tests in the package can match precisely.
- Implement `Error()` on the pointer receiver and return `&T{...}`, so `errors.As` behaves the way callers expect.
- Match with `errors.Is` for sentinels and `errors.As` for types. Never match on the message text: the string is for humans and is free to change.

## Panics

Never panic for a runtime condition. A missing file, malformed input, or a failed network call is an error value however unlikely it looks.

The one sanctioned exception is programmer error detected during package initialisation, where there is no caller to return to and the binary cannot function at all:

```go
var templateFS = func() fs.FS {
    sub, err := fs.Sub(embeddedTemplates, "templates")
    if err != nil {
        panic("embedded templates directory not found: " + err.Error())
    }
    return sub
}()
```

That failure means the binary was built wrong, so every invocation is broken and no user input could avoid it. Anything that depends on what the user typed, or on what is on disk at runtime, is an error return instead.

Recover only at a goroutine boundary you own, and only to turn a panic into an error. Never to carry on.

## The top-level error path

The root command sets `SilenceErrors` and `SilenceUsage` (see the scaffolding fragment), so cobra prints nothing and `main` owns all error output. `main` builds the command tree, runs it, and turns whatever comes back into an exit code:

```go
func main() {
    root := cmd.NewRootCmd(os.Stdout, os.Stderr)
    if err := root.Execute(); err != nil {
        var ec *cmd.ExitCodeError
        if errors.As(err, &ec) {
            os.Exit(ec.Code)
        }
        fmt.Fprintf(os.Stderr, "<binary>: %v\n", err)
        os.Exit(1)
    }
}
```

A command that needs a context adds signal handling above this and calls `ExecuteContext`; see the contexts fragment. The error handling is the same either way.

- `main` builds the tree through `NewRootCmd` rather than a package-level `Execute` wrapper, so production and tests construct it exactly one way (see the scaffolding and functional testing fragments). A wrapper that only forwards its arguments is a second construction path that can drift from the one the tests use.
- `main` imports `cmd` and the standard library, nothing else. A sentinel it has to match on, such as one meaning the user backed out of a prompt, is exported from `cmd` rather than reached for in `internal/`.
- The prefix is the binary name, never `error:`. It tells someone reading a wall of shell output which program spoke.
- Exit 1 for every ordinary failure.
- `ExitCodeError` covers the cases where 1 is the wrong code: propagating a wrapped process's exit status, or signalling findings from a scan. Its `Error()` returns the empty string, so a command that has already written its own output returns `&ExitCodeError{Code: 1}` and exits non-zero without a second message.

```go
// ExitCodeError is returned by a command that has already produced its output
// and needs to set the process exit code itself.
type ExitCodeError struct{ Code int }

func (e *ExitCodeError) Error() string { return "" }
```

Add the type only once a command returns one. A CLI whose every failure is exit 1 omits both the type and the `errors.As` branch, and gains them when it first needs them.

## Closing and deferred errors

`defer f.Close()` is fine for a file opened for reading: there is no buffered state to lose. Anything written to is different, because `Close` is where a buffered write can still fail, and discarding that error silently truncates output. Capture it with a named return:

```go
func writeReport(path string, b []byte) (err error) {
    f, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("create %q: %w", path, err)
    }
    defer func() {
        if cerr := f.Close(); cerr != nil && err == nil {
            err = fmt.Errorf("close %q: %w", path, cerr)
        }
    }()
    _, err = f.Write(b)
    return err
}
```

The `err == nil` guard matters: a close error must not mask the failure that came first.
