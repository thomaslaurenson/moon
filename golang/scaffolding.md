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
    lint.yml
    test.yml
    release.yml
    prerelease.yml
    pr.yml
    main.yml
    tag.yml
.gitignore
.goreleaser.yml
.goreleaser.prerelease.yml
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
- `build`: `go build -ldflags="..." -o dist/<binary> .`
- `build_snapshot`: `goreleaser build --snapshot --clean`
- `install`: `go install -ldflags="..." .`
- `snapshot`: `goreleaser release --snapshot --clean`
- `release_check`: `goreleaser check`
- `get_changelog`: extract release notes for a given tag from `CHANGELOG.md`, output to stdout (used by the release workflow via `make get_changelog TAG=v1.0.0 > /tmp/release-notes.md`)
- `clean`: `rm -rf bin/ dist/`
- `ci`: `fmt_check mod_check vet test`

Tests always include `-race -count=1`, including coverage runs. The binary is built to `bin/`, not the repo root. Coverage is measured over `./internal/...` only; not `cmd/` or the root package.

## go.mod

- Use the latest stable Go version
- No `replace` directives in committed code
- Run `go mod tidy` before committing

## Cobra CLI Structure

### Root Command

The root command is defined in `cmd/root.go`. Always set `SilenceErrors` and `SilenceUsage` to prevent cobra printing errors and usage automatically; the main entrypoint handles all error output. Set `Version` on the root command so that `<binary> --version` works:

```go
var rootCmd = &cobra.Command{
    Use:           "<binary>",
    Short:         "Short description",
    SilenceErrors: true,
    SilenceUsage:  true,
    Version:       Version,
}
```

### Version Command

Every CLI project must include a `version` subcommand defined in `cmd/version.go`. The `Version` var is declared here with a fallback of `"dev"` and injected at build time via ldflags. This is the canonical location, do not declare `Version` in `internal/`:

```go
// Version is set at build time using:
// -ldflags "-X github.com/<owner>/<repo>/cmd.Version=...".
// It falls back to "dev" for local builds that do not inject a value.
var Version = "dev"

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the <binary> version",
    Args:  cobra.NoArgs,
    Run: func(_ *cobra.Command, _ []string) {
        fmt.Printf("<binary> version %s\n", Version)
    },
}
```

If the root command uses `PersistentPreRunE` (e.g. to load config on startup), override it on the version command so `<binary> version` never triggers that logic:

```go
PersistentPreRunE: func(_ *cobra.Command, _ []string) error { return nil },
```

The ldflags injection path must match where `Version` is declared. Because `Version` lives in `cmd/version.go`, the path is `-X <module>/cmd.Version={{.Version}}`. This is the template used in both `.goreleaser.yml` and `.goreleaser.prerelease.yml`.
