# Go Project Scaffolding

Standards and conventions for Go projects. Use this as a reference when creating or refactoring a Go repository.

## Contents

- [Project structure](#project-structure) — directory layout for a single-binary Go project
- [Tools](#tools) — only what ships with Go, no third-party linters
- [Makefile targets](#makefile) — canonical target names and what they do
- [go.mod rules](#gomod) — versioning and tidy rules

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

The only tooling used is what ships with Go itself. No third-party linters.

| Tool | Purpose |
|---|---|
| `gofmt` | Format source files |
| `go vet` | Static analysis |
| `go test` | Run tests |
| `go mod tidy` | Keep go.mod/go.sum clean |

**Not used:** `golangci-lint`, `govulncheck`, or any other third-party analysis tools.

## Makefile

All CI steps are Makefile targets. GitHub Actions call `make <target>` - never raw Go commands directly in workflows.

Key targets:

```makefile
fmt:           gofmt -w .
fmt_check:     gofmt -l . (exits 1 if unformatted files found)
mod_check:     go mod tidy && git diff --exit-code go.mod go.sum
vet:           go vet ./...
test:          go test -race -count=1 ./...
test_verbose:  go test -race -count=1 -v ./...
test_coverage: go test -race -count=1 -coverpkg=./internal/... -coverprofile=coverage.out ./...
ci:            fmt_check mod_check vet test
build:         go build -ldflags="..." -o bin/<binary> .
install:       go install -ldflags="..." .
clean:         rm -rf bin/ dist/
snapshot:      vgoreleaser release --snapshot --clean
release_check: goreleaser check
```

Rules:
- Tests always include `-race -count=1`, including coverage runs
- The binary is built to `bin/` not the repo root
- `ci` is the local equivalent of the full CI pipeline

## go.mod

- Use the latest stable Go version
- No `replace` directives in committed code
- Run `go mod tidy` before committing


