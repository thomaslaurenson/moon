# Go Project Scaffolding

Standards and conventions for Go projects. Use this as a reference when creating or refactoring a Go repository.

## Project Structure

```
cmd/           # cobra command definitions (root.go, run.go, version.go, etc.)
internal/      # private packages not imported by other modules
main.go        # entry point - minimal, delegates to cmd/
.github/
  workflows/
  dependabot.yml
.gitignore
.goreleaser.yml
go.mod
go.sum
Makefile
CHANGELOG.md
README.md
```

- Business logic lives in `internal/`, not `cmd/`
- `cmd/` only handles CLI parsing and wiring
- No `pkg/` directory - use `internal/` unless the package is intentionally public
- Do **not** add a named subdirectory under `internal/` (e.g. `internal/myapp/`) - this is only warranted in a monorepo with multiple binaries that need scoped internals. For a single-binary project, packages sit directly under `internal/`

## Tools

No third-party linters or formatters are permitted. Specifically, DO NOT use `golangci-lint` or `govulncheck` under any circumstances. However, third-party release tools like `goreleaser` and `cosign` are explicitly permitted and required.

| Tool | Purpose |
|---|---|
| `gofmt` | Format source files |
| `go vet` | Static analysis |
| `go test` | Run tests |
| `go mod tidy` | Keep go.mod/go.sum clean |

**Not used:** `golangci-lint`, `govulncheck`, or any other third-party analysis tools.

## Makefile

Adhere to the global Makefile structure established in `tools/makefile.md`. Use the following commands for your standard targets:

- `fmt`: `gofmt -w .`
- `fmt_check`: `gofmt -l . && git diff --exit-code` (exits 1 if unformatted files found)
- `mod_check`: `go mod tidy && git diff --exit-code go.mod go.sum`
- `vet`: `go vet ./...`
- `test`: `go test -race -count=1 ./...`
- `test_verbose`: `go test -race -count=1 -v ./...`
- `test_coverage`: `go test -race -count=1 -coverpkg=./internal/... -coverprofile=coverage.out ./...`
- `build`: `go build -ldflags="..." -o bin/<binary> .`
- `install`: `go install -ldflags="..." .`
- `snapshot`: `goreleaser release --snapshot --clean`
- `release_check`: `goreleaser check`
- `clean`: `rm -rf bin/ dist/`
- `ci`: `fmt_check mod_check vet test`

Tests always include `-race -count=1`, including coverage runs. The binary is built to `bin/`, not the repo root.

## go.mod

- Use the latest stable Go version
- No `replace` directives in committed code
- Run `go mod tidy` before committing


