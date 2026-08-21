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

## Versioning

Follow semantic versioning strictly: `go install` and the module proxy key off the tag directly, so the tag is the interface.

For a v2 or later major version, the module path itself must gain the version suffix, per Go's module rules:

```
module github.com/x/mytool/v2
```

Every import within the module must include `/v2`, and so must the ldflags path (`-X <module>/v2/cmd.Version=...`). This is not optional once a v2+ tag is pushed; omitting it breaks module resolution, and `go install <module>@latest` keeps resolving to the newest v1.
