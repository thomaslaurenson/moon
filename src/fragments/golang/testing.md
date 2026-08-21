# Go testing

- Always run tests with `-race -count=1` (enforced via `make test`).
- No third-party test frameworks; use the standard `testing` package only.
- Table-driven tests are the default for unit tests: a slice of anonymous structs with a required `name` field, iterated with `t.Run`.
- Call `t.Parallel()` at the top of each test function and subtest. The only exception is a test that mutates process-wide state, covered below; shared package state is not a reason, because the style fragment rules it out.

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
- Assert errors with `errors.Is()` (sentinels) and `errors.As()` (typed), never string matching; see the errors fragment for how they are constructed.
- Use `httptest.NewServer`/`NewTLSServer` for HTTP tests; no external network calls.
- Build a fresh instance per test with the package's constructor. A package that instead shares one instance forces tests to reset it between runs, which makes them order-dependent and rules out `t.Parallel()`; the fix is the constructor, not a `ResetForTesting()` helper to wipe the shared copy (see the style fragment).

No fuzz targets. Fuzzing earns its place against a hand-rolled parser of untrusted input, and these tools delegate parsing to the standard library. Add a `FuzzXxx` when one of them grows a parser of its own.

## Test inputs and golden files

Prefer inputs the test builds and throws away. A file constructed a few lines above the assertion is visible to the reader, cannot go stale, and cannot be broken by an edit somewhere else in the tree.

- `fstest.MapFS` for a read-only tree the code only reads.
- `t.TempDir()` for real files on disk, cleaned up automatically.

Commit an input only when constructing it is not reasonable: a binary blob, a large sample from the real world that you did not author, or a fuzz seed corpus. When that happens the file goes in `testdata/` beside the package that reads it, never in a `fixtures/` or `test/` directory invented per project. The Go toolchain ignores `testdata/`, and tests run with the working directory set to the package, so `os.ReadFile("testdata/sample.bin")` works with no path juggling.

Golden files follow the same rule. If a test compares against a large expected output, that output belongs in `testdata/`, committed and reviewed in the diff, because an unexpected change to it is the test failing. Reserve them for output big enough that inlining it would bury the assertion; a handful of lines stays in the table.

- Regenerate goldens behind an explicit flag: `go test ./... -update`.
- A test must never write to `testdata/` during an ordinary run. That dirties the working tree and breaks parallel tests. Scratch output goes to `t.TempDir()`; only `-update` writes a golden.
- Do not create an empty `testdata/`, and do not add a `.gitkeep`. The directory appears when there is a file to put in it, and `go test -fuzz` creates `testdata/fuzz/` itself when it finds a failing input.

Coverage is measured over `./internal/...` only, with `cmd/` and the root package excluded as wiring; see the Makefile targets fragment.
