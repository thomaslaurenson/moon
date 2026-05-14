# Go Testing Conventions

Standards and patterns for writing tests in Go projects.

## Non-Negotiable Rules

- Always run tests with `-race -count=1` (enforced via `make test`)
- No third-party test frameworks; use the standard `testing` package only
- No test helpers that hide assertion failures without `t.Helper()`

## Table-Driven Tests

The preferred pattern for all unit tests. Define a slice of test cases as an anonymous struct, then iterate with `t.Run`:

```go
func TestNormalisePath(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name  string
        input string
        want  string
    }{
        {name: "absolute path unchanged", input: "/etc/hosts", want: "/etc/hosts"},
        {name: "home dir expanded", input: "~/config", want: "/home/user/config"},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            got := NormalisePath(tc.input)
            if got != tc.want {
                t.Errorf("NormalisePath(%q) = %q, want %q", tc.input, got, tc.want)
            }
        })
    }
}
```

- The `name` field is required on every test case
- Call `t.Parallel()` at the top of the test function and at the top of each subtest

## Parallelism

Call `t.Parallel()` at the start of every test function and subtest where the test does not rely on shared mutable state. This is the default; only omit it when there is a specific reason (e.g. the test modifies a global singleton that cannot be isolated).

## Helper Functions

Mark any function that calls `t.Fatal`, `t.Error`, or similar with `t.Helper()`. This ensures failure output points to the call site, not inside the helper:

```go
func buildCatalog(t *testing.T, entries []Entry) *Catalog {
    t.Helper()
    c, err := NewCatalog(entries)
    if err != nil {
        t.Fatalf("NewCatalog: %v", err)
    }
    return c
}
```

## Isolation

Use the standard library helpers for test isolation; both clean up automatically when the test ends:

- `t.TempDir()`: creates a temporary directory; use for any test that reads or writes files
- `t.Setenv(key, value)`: sets an environment variable for the duration of the test; use to override `HOME` or other env-dependent config

```go
func TestConfigLoad(t *testing.T) {
    t.Parallel()
    home := t.TempDir()
    t.Setenv("HOME", home)
    // test proceeds with isolated HOME
}
```

## Error Assertions

Use `errors.Is()` for sentinel errors, never string matching:

```go
if !errors.Is(err, ErrNotFound) {
    t.Fatalf("expected ErrNotFound, got %v", err)
}
```

Use `errors.As()` when you need to inspect a typed error's fields.

## HTTP / Integration Tests

Use `httptest.NewServer` or `httptest.NewTLSServer` for tests that exercise HTTP handlers or clients. No external network calls in tests.

```go
func TestProxyIntegration(t *testing.T) {
    t.Parallel()
    backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    t.Cleanup(backend.Close)
    // test against backend.URL
}
```

## Singleton State

Packages that hold singleton state (e.g. a package-level instance initialised once) must expose a `ResetForTesting()` function that returns the package to its zero state. Call it at the start of any test that creates an instance:

```go
func TestNewProxy(t *testing.T) {
    t.Parallel()
    proxy.ResetForTesting()
    // ...
}
```

## Coverage

Coverage is measured over `./internal/...` only:

```
make test_coverage
```

This corresponds to `go test -race -count=1 -coverpkg=./internal/... -coverprofile=coverage.out ./...`. The `cmd/` package and root package are excluded because they contain only wiring and CLI parsing with no testable logic.
