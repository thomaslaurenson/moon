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

The root command is built by a constructor in `cmd/root.go`, never a package-level variable configured by `init()` (see the style fragment). The constructor takes the dependencies the command tree needs, which is what lets a test supply its own writers and a fake filesystem:

```go
// App holds the dependencies shared by every subcommand.
type App struct {
    e *engine.Engine
}

// NewRootCmd builds the command tree, writing output to out and errw.
func NewRootCmd(fsys fs.FS, out, errw io.Writer) *cobra.Command {
    a := &App{e: engine.New(fsys)}

    root := &cobra.Command{
        Use:           "<binary>",
        SilenceErrors: true,
        SilenceUsage:  true,
        Version:       Version,
    }
    root.SetOut(out)
    root.SetErr(errw)

    root.AddCommand(a.newShowCmd())
    return root
}
```

`SilenceErrors` and `SilenceUsage` are both set because the entry point owns error output (see the errors fragment). `SetOut` and `SetErr` are what make `cmd.OutOrStdout()` work inside subcommands, so nothing has to reach for the `os` globals (see the logging fragment).

`main` calls `NewRootCmd` directly and executes the result; there is no package-level `Execute` wrapper. The parameters are whatever the tree actually needs, and the writers are always among them: a tool with no embedded assets drops `fsys` and takes `(out, errw io.Writer)` alone. This is the same constructor the functional tests call, which is the point of it (see the errors and functional testing fragments).

Subcommands are constructors too, as methods on `App` where they need shared dependencies and plain functions where they do not. Flags are registered inside the constructor, bound to a field or a local, never to a package-level variable.

Every CLI includes a `version` subcommand in `cmd/version.go`. `Version` is declared there with a `"dev"` fallback and injected at build time via ldflags; this is the canonical location, never `internal/`. The ldflags path must match: `-X <module>/cmd.Version={{.Version}}`. If the root uses `PersistentPreRunE`, override it on the version command so `version` never triggers that logic.

`go install <module>@<tag>` runs the toolchain rather than the Makefile, so nothing stamps `Version` and a tagged build reports itself as `dev`. The toolchain records the module version in the binary, so read it back when ldflags left the default in place. Use this implementation verbatim:

```go
const devVersion = "dev"

// Version is injected at build time via ldflags, falling back to devVersion.
var Version = devVersion

// init fills in the version for a build that set no ldflags.
func init() {
    if info, ok := debug.ReadBuildInfo(); ok {
        Version = versionFrom(Version, info.Main.Version)
    }
}

// versionFrom picks the version to report, given the one ldflags stamped in and
// the one the toolchain recorded. An injected version always wins; the module
// version only rescues a "dev" that came from "go install". The leading "v" is
// dropped so both routes report the same string.
func versionFrom(injected, module string) string {
    if injected != devVersion || !isReleaseVersion(module) {
        return injected
    }
    return strings.TrimPrefix(module, "v")
}

// pseudoVersion matches the timestamp and commit tail the toolchain appends when
// a build has no tag to name itself after. The separator is "-" when there was
// no earlier tag ("v0.0.0-<stamp>-<commit>") and "." above an existing one
// ("v1.2.4-0.<stamp>-<commit>"), so both have to match.
var pseudoVersion = regexp.MustCompile(`[-.][0-9]{14}-[0-9a-f]{12}$`)

// isReleaseVersion reports whether module names a published tag. A working tree
// is recorded as "(devel)", or as a pseudo-version with a "+dirty" suffix once
// the toolchain has VCS information. Reporting one of those would make an
// ordinary local build announce itself as a published version.
func isReleaseVersion(module string) bool {
    switch {
    case module == "", module == "(devel)":
        return false
    case strings.Contains(module, "+"):
        return false
    default:
        return !pseudoVersion.MatchString(module)
    }
}
```

This has to be `init()` rather than an initialiser on `Version` itself. The linker writes an `-X` value into the data segment, and a variable with a runtime initialiser has that value overwritten the moment the initialiser runs, with no build error to say so (see the style fragment). Nothing else about versioning changes: the ldflags path, the goreleaser config and the `build` target all keep their current form, and a local build with no ldflags still reports `dev`.

## Versioning

Follow semantic versioning strictly: `go install` and the module proxy key off the tag directly, so the tag is the interface.

For a v2 or later major version, the module path itself must gain the version suffix, per Go's module rules:

```
module github.com/x/mytool/v2
```

Every import within the module must include `/v2`, and so must the ldflags path (`-X <module>/v2/cmd.Version=...`). This is not optional once a v2+ tag is pushed; omitting it breaks module resolution, and `go install <module>@latest` keeps resolving to the newest v1.
