# Go Project Scaffolding

CLI application layout:

```
cmd/           # cobra commands (root.go, version.go, etc.)
internal/      # private packages
main.go        # minimal entry point, delegates to cmd/
.github/
  workflows/
  dependabot.yml
.goreleaser.yml
.goreleaser.prerelease.yml
go.mod
go.sum
Makefile
CHANGELOG.md
README.md
```

- Business logic lives in `internal/`, not `cmd/`. `cmd/` only handles CLI parsing and wiring.
- No `pkg/` directory; use `internal/`. For a single-binary project, packages sit directly under `internal/` (no named subdirectory).

## Cobra

The root command (`cmd/root.go`) sets `SilenceErrors` and `SilenceUsage` (the entry point handles error output) and `Version`:

```go
var rootCmd = &cobra.Command{
    Use:           "<binary>",
    SilenceErrors: true,
    SilenceUsage:  true,
    Version:       Version,
}
```

Every CLI includes a `version` subcommand in `cmd/version.go`. `Version` is declared there with a `"dev"` fallback and injected at build time via ldflags; this is the canonical location, never `internal/`. The ldflags path must match: `-X <module>/cmd.Version={{.Version}}`. If the root uses `PersistentPreRunE`, override it on the version command so `version` never triggers that logic.

## Makefile targets

Follow the shared Makefile conventions. Standard targets:

- `fmt`: `gofmt -w .`
- `fmt_check`: `gofmt -l . && git diff --exit-code`
- `mod_check`: `go mod tidy && git diff --exit-code go.mod go.sum`
- `vet`: `go vet ./...`
- `test`: `go test -race -count=1 ./...`
- `test_coverage`: `go test -race -count=1 -coverpkg=./internal/... -coverprofile=coverage.out ./...`
- `build`: `go build -ldflags="..." -o dist/<binary> .`
- `snapshot`: `goreleaser release --snapshot --clean`
- `get_changelog`: extract release notes for a tag from `CHANGELOG.md` to stdout (strip the `v` prefix; git tags use `v1.0.0`, CHANGELOG uses `1.0.0`)
- `ci`: `fmt_check mod_check vet test`

Tests always include `-race -count=1`, including coverage. Coverage is measured over `./internal/...` only; `cmd/` and the root package are excluded as wiring-only.
