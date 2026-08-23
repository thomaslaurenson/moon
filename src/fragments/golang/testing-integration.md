# Go integration testing

An integration test needs something CI cannot provide: elevated privileges, a real remote host, a mounted filesystem, or credentials. Everything else stays an ordinary test.

That line matters, because the tag is where tests go to stop running. Binding a port on localhost, spawning a child process, or serving HTTP through `httptest` all work on a GitHub runner, so they are normal tests and keep running on every pull request. Only reach for the tag when the thing genuinely cannot exist in CI.

## The build tag

An integration test file opens with the constraint, followed by a blank line before `package`:

```go
//go:build integration

package mount_test
```

Go excludes the file from an untagged build, so `go test ./...` ignores it with no flag and no marker to filter. Name the file `<subject>_integration_test.go` so the constraint is visible before the file is opened.

## Skipping when the environment is absent

A developer without the setup gets a skip, not a failure. Gate on the environment variable that names the resource, and say what is missing:

```go
func TestMountRemote(t *testing.T) {
    host := os.Getenv("MYTOOL_TEST_HOST")
    if host == "" {
        t.Skip("MYTOOL_TEST_HOST not set")
    }
    ...
}
```

Check at the top of the test, before any setup. A skip halfway through leaves whatever the test already created behind.

Integration tests still clean up after themselves. A test that mounts unmounts in a `t.Cleanup`, whether or not it passed, because the next run has to start from the same state as this one.

## Running them

- `make test_integration` runs them: `go test -race -count=1 -tags=integration ./...`.
- `make test` and the `ci` target exclude them. Nothing in CI runs an integration test.
- `make vet` covers them, and this is the point that keeps them alive. `go vet ./...` does not look at files behind a build tag, and neither does `go build ./...` or `go test ./...`, so a tagged file can stop compiling without anything noticing. Running `go vet -tags=integration ./...` compiles them without running them, which is safe in CI and turns silent rot into a build failure.

## Coverage

`test_coverage` runs with `-tags=integration`, so the figure reflects everything the project can exercise rather than the subset that runs unattended.

The consequence is that the number depends on the machine. A developer without the required environment gets skips and a lower percentage, which is not a regression. The badge figure is the one measured where every resource is available.
