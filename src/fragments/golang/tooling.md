# Go tooling

No third-party linters or formatters. Do not use `golangci-lint` or `govulncheck` under any circumstances.

| Tool | Purpose |
|---|---|
| `gofmt` | Format source files |
| `go vet` | Static analysis |
| `go test` | Run tests |
| `go mod tidy` | Keep go.mod/go.sum clean |

- Use the latest stable Go version. No `replace` directives in committed code. Run `go mod tidy` before committing.
- Third-party release tools (`goreleaser`, `cosign`) are permitted, since they build and sign the release rather than judge the source.

## Embedded assets

A binary that needs data files at runtime embeds them with `//go:embed` rather than shipping files alongside it. The declaration lives in the package that owns the data, directly above an `embed.FS` variable, and the embedded tree must sit at or below that package's directory.

- Use the `all:` prefix when the tree contains files beginning with `_` or `.`, which the default pattern skips: `//go:embed all:src`. It skips them silently, so the build still succeeds and the file is simply absent at runtime.
- Expose a subdirectory with `fs.Sub` so callers use paths relative to it rather than repeating the embedded prefix at every call site.
- Add every embedded tree to the CI paths filter. A change under one is a change to the compiled binary; see the workflows fragment.

Embedded content needs a `check` target in the Makefile. The compiler proves only that the files exist, never that their contents are valid, so nothing otherwise stops a broken asset being compiled in and failing at runtime. What the target runs is project-specific: validating an embedded document graph, or linting embedded shell and PowerShell templates with the tooling for those languages. A project that embeds nothing omits the target.
