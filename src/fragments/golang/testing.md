# Go testing

- Always run tests with `-race -count=1` (enforced via `make test`).
- No third-party test frameworks; use the standard `testing` package only.
- Table-driven tests are the default for unit tests: a slice of anonymous structs with a required `name` field, iterated with `t.Run`.
- Call `t.Parallel()` at the top of each test function and subtest, unless it relies on shared mutable state.

```go
func TestNormalisePath(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name  string
        input string
        want  string
    }{
        {name: "absolute path unchanged", input: "/etc/hosts", want: "/etc/hosts"},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            if got := NormalisePath(tc.input); got != tc.want {
                t.Errorf("NormalisePath(%q) = %q, want %q", tc.input, got, tc.want)
            }
        })
    }
}
```

- Mark helper functions that call `t.Fatal`/`t.Error` with `t.Helper()` so failures point to the call site.
- Use `t.TempDir()` for filesystem isolation. It is per-test, parallel-safe, and cleans up automatically.
- `t.Setenv()` and `t.Chdir()` mutate process-wide state, so Go panics when the test, or any ancestor of it, is parallel. A test calling either one omits `t.Parallel()`, and so does its parent when the subtests call them. Prefer removing the need entirely: read the environment in `cmd/` and pass values down (see the style fragment), so `internal/` tests stay parallel.
- Assert errors with `errors.Is()` (sentinels) and `errors.As()` (typed), never string matching.
- Use `httptest.NewServer`/`NewTLSServer` for HTTP tests; no external network calls.
- Packages holding singleton state expose `ResetForTesting()`; call it at the start of tests that create an instance.

Add `Example` functions in `<file>_test.go`, or a dedicated `example_test.go`, for anything whose use is easier to show than to describe. They compile, run as part of `go test`, and are rendered in the generated documentation, so unlike a code block in a comment they cannot fall out of date without failing:

```go
func ExampleParseConfig() {
    cfg, _ := ParseConfig("testdata/config.yaml")
    fmt.Println(cfg.Name)
    // Output: my-app
}
```

Coverage is measured over `./internal/...` only, with `cmd/` and the root package excluded as wiring; see the Makefile targets fragment.
