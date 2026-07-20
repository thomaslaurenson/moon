# Go project scaffolding

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
- `fmt_check`: capture `gofmt -l .` and fail if non-empty (`out="$(gofmt -l .)"; test -z "$out"`). Do not write `gofmt -l . && git diff --exit-code`: `gofmt -l` never changes files and always exits 0, so that form can never fail.
- `mod_check`: `go mod tidy && git diff --exit-code go.mod go.sum`
- `vet`: `go vet ./...`
- `test`: `go test -race -count=1 ./...`
- `test_coverage`: run `go test -race -count=1 -coverpkg=./internal/... -coverprofile=coverage.out ./...`, then `go tool cover -func=coverage.out` to print the per-function table ending in the aggregate `total:` line, then `rm coverage.out`.
- `build`: `go build -ldflags="..." -o dist/<binary> .`
- `snapshot`: `goreleaser release --snapshot --clean`
- `get_changelog`: extract release notes for a tag from `CHANGELOG.md` to stdout (strip the `v` prefix; git tags use `v1.0.0`, CHANGELOG uses `1.0.0`). Fail non-zero when no entry matches, so a release never publishes empty notes.
- `check`: validate embedded content if the binary embeds any (see the tooling fragment); omit for a plain Go project with nothing embedded.
- `ci`: `fmt_check mod_check vet test`

Tests always include `-race -count=1`, including coverage. Coverage is measured over `./internal/...` only; `cmd/` and the root package are excluded as wiring-only. The per-package percentages `go test` prints are each measured against the whole `-coverpkg` set, so they read low and do not sum; the real figure is the `total:` line from `go tool cover -func`, which is also the number used for the coverage badge.
