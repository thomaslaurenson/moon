# Go project scaffolding

CLI application layout:

```text
cmd/           # cobra commands (root.go, version.go, etc.)
internal/      # private packages
main.go        # minimal entry point, delegates to cmd/
.github/
  workflows/
  dependabot.yml
.goreleaser.yml
.goreleaser.prerelease.yml
.gpipe.yml
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

`SilenceErrors` and `SilenceUsage` are both set because the entry point owns error output (see the errors fragment).

`SilenceUsage` is load-bearing rather than a matter of taste, and the reason is worth knowing before anyone switches it off to get a usage hint back. Cobra prints usage-on-error through the *out* writer, not the error writer, so with `SetOut(out)` given `os.Stdout` a failed command would dump its usage onto stdout. That breaks the contract that stdout carries the answer and nothing else (see the logging fragment). The cost is real: a wrong argument produces an error line and no usage. The alternative is worse.

A root whose bare invocation means nothing still has to fail. Cobra's default for a root with subcommands and no `RunE` is to print help to stdout and exit zero, which tells a script the invocation succeeded:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    fmt.Fprint(cmd.ErrOrStderr(), cmd.UsageString())
    return &ExitCodeError{Code: 1}
},
```

Usage on stderr, nothing on stdout, exit 1. A root that does have a bare behaviour, such as a tool that prompts when given no argument, has its own `RunE` and needs none of this. `SetOut` and `SetErr` are what make `cmd.OutOrStdout()` work inside subcommands, so nothing has to reach for the `os` globals (see the logging fragment).

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

## Flags

- Long names are kebab-case and lowercase: `--dry-run`, `--log-file`, never `--dryRun` or `--DryRun`.
- Help text is a lowercase fragment with no trailing full stop, matching the `help for <binary>` line cobra generates beside it.
- A shorthand means one thing across the whole binary. Giving `-l` to `--long` under one subcommand and `--list` under another teaches a user something that then misfires, and the compiler cannot see it because the two are registered on different commands.
- Add a shorthand only for a flag typed often enough to earn one. Adding it later costs nothing; removing one breaks whoever learned it.

Where a value can come from more than one place, the order is flag, then environment, then config file, then the built-in default. Resolve it in `cmd/`, where the environment is already read (see the style fragment), and pass the settled value down. A package under `internal/` should never be able to tell which layer an argument came from.

## Shell completion

Cobra adds a `completion` subcommand to every root, generating scripts for bash, zsh, fish and powershell. Keep it. It costs nothing and it is the only way a user can obtain the script.

What gets completed is up to the commands. One taking a dynamic argument gets a `ValidArgsFunction` returning the candidates and `ShellCompDirectiveNoFileComp`, so the shell offers those names rather than falling back to filenames:

```go
// completeTargets offers configured target names for a <target> argument.
func (a *App) completeTargets(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    names, err := a.e.Targets()
    if err != nil {
        return nil, cobra.ShellCompDirectiveError
    }
    return filterByPrefix(names, toComplete), cobra.ShellCompDirectiveNoFileComp
}
```

These sit together in `cmd/completion.go` rather than beside each command: they are one concern and each is a few lines. Where a bare argument resolves against a user-defined name, take the reserved names off the command tree rather than listing them by hand, since a hand-written list goes stale as soon as someone adds an alias.

Installing the generated script is the user's job. The release tooling deliberately does not do it for them (see the gpipe fragment).

## Versioning

Follow semantic versioning strictly: `go install` and the module proxy key off the tag directly, so the tag is the interface.

For a v2 or later major version, the module path itself must gain the version suffix, per Go's module rules:

```text
module github.com/x/mytool/v2
```

Every import within the module must include `/v2`, and so must the ldflags path (`-X <module>/v2/cmd.Version=...`). This is not optional once a v2+ tag is pushed; omitting it breaks module resolution, and `go install <module>@latest` keeps resolving to the newest v1.
