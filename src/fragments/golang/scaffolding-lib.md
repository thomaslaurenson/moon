# Go library project scaffolding

Library layout (no `cmd/`, no binary, nothing to release):

```
<package>/     # or the package files directly at the module root for a single-package library
internal/      # private helper packages, not part of the public API
.github/
  workflows/
  dependabot.yml
go.mod
go.sum
Makefile
CHANGELOG.md
README.md
doc.go         # package-level doc comment, if long enough to warrant its own file
```

- A single-package library puts its files directly at the module root; the package name matches the last path element of the module (module `github.com/x/mylib` -> package `mylib`).
- A multi-package library exposes each importable package as its own top-level directory. `internal/` is still for genuinely private implementation detail, not the whole library.
- No `cmd/`, no `main.go`, no cobra: nothing in a library is an entry point.

## Documentation

The package doc comment matters more here than for a CLI: it renders as the landing page on pkg.go.dev. Put it in `doc.go` when it needs more than a couple of lines, so it isn't buried above an unrelated file's code.

Add `Example` functions in `<file>_test.go` (or a dedicated `example_test.go`) for anything a consumer would want to see used, not just described. These compile, run as part of `go test`, and are rendered directly in the generated documentation:

```go
func ExampleParseConfig() {
    cfg, _ := ParseConfig("testdata/config.yaml")
    fmt.Println(cfg.Name)
    // Output: my-app
}
```

## Versioning

Version comes from git tags directly (`v1.2.3`); there is no ldflags-injected `Version` variable, because nothing is compiled into a distributable binary. Follow semantic versioning strictly, because `go get` and the module proxy key off it directly.

For a v2 or later major version, the module path itself must gain the version suffix, per Go's module rules:

```
module github.com/x/mylib/v2
```

Every import of the package elsewhere in the module (and by consumers) must include `/v2`. This is not optional once a v2+ tag is pushed; omitting it breaks module resolution.

## Makefile targets

Follow the shared Makefile conventions.

- `fmt`: `gofmt -w .`
- `fmt_check`: capture `gofmt -l .` and fail if non-empty (`out="$(gofmt -l .)"; test -z "$out"`). Do not write `gofmt -l . && git diff --exit-code`: `gofmt -l` never changes files and always exits 0, so that form can never fail.
- `mod_check`: `go mod tidy && git diff --exit-code go.mod go.sum`
- `vet`: `go vet ./...`
- `test`: `go test -race -count=1 ./...`
- `test_coverage`: run `go test -race -count=1 -coverprofile=coverage.out ./...`, then `go tool cover -func=coverage.out` to print the per-function table ending in the aggregate `total:` line, then `rm coverage.out`.
- `ci`: `fmt_check mod_check vet test`

No `build` or `snapshot` target: there is nothing to compile into a release artifact. Release is a tagged commit; see the workflows fragment.

Coverage runs over `./...`; unlike a CLI project there is no `cmd/` wiring layer to exclude. The aggregate figure is the `total:` line from `go tool cover -func`, which is also the number used for the coverage badge.
