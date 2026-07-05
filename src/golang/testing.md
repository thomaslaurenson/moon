# Go Testing

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
- Use `t.TempDir()` and `t.Setenv()` for isolation; both clean up automatically.
- Assert errors with `errors.Is()` (sentinels) and `errors.As()` (typed), never string matching.
- Use `httptest.NewServer`/`NewTLSServer` for HTTP tests; no external network calls.
- Packages holding singleton state expose `ResetForTesting()`; call it at the start of tests that create an instance.

The scope `test_coverage` measures (all of `./...`, or `./internal/...` with `cmd/` excluded) depends on project shape; see the scaffolding fragment for this project's tier.
