# Go tooling

No third-party linters or formatters. Do not use `golangci-lint` under any circumstances. `gofmt` and `go vet` are the standard toolchain, and a single-author project of this size gains nothing from a second opinion on style that it then has to configure and silence.

This is a rule about tooling, not about dependencies. Ordinary libraries are fine, starting with cobra for the command tree. What is banned is narrow, and each case is named in the fragment that owns it: linters and formatters here, test frameworks in the testing fragment, logging packages in the logging fragment. Every one of them replaces something the standard library or the standard toolchain already does well enough.

Treat `golang.org/x/*` as standard library for this purpose. Those modules are maintained by the Go team and are where capabilities that do not belong in the standard library proper live: `x/term` to ask whether a stream is a terminal, `x/sync` for goroutine coordination, `x/vuln` for the scan below.

| Tool | Purpose |
|---|---|
| `gofmt` | Format source files |
| `go vet` | Static analysis |
| `go test` | Run tests |
| `go mod tidy` | Keep go.mod/go.sum clean |

- No `replace` directives in committed code. Run `go mod tidy` before committing.
- Third-party release tools (`goreleaser`, `cosign`) are permitted, since they build and sign the release rather than judge the source.

## Embedded assets

A binary that needs data files at runtime embeds them with `//go:embed` rather than shipping files alongside it. The declaration lives in the package that owns the data, directly above an `embed.FS` variable, and the embedded tree must sit at or below that package's directory.

- Use the `all:` prefix when the tree contains files beginning with `_` or `.`, which the default pattern skips: `//go:embed all:src`. It skips them silently, so the build still succeeds and the file is simply absent at runtime.
- Expose a subdirectory with `fs.Sub` so callers use paths relative to it rather than repeating the embedded prefix at every call site.
- Add every embedded tree to the CI paths filter. A change under one is a change to the compiled binary; see the workflows fragment.

Embedded content needs a `check_embed` target in the Makefile. The compiler proves only that the files exist, never that their contents are valid, so nothing otherwise stops a broken asset being compiled in and failing at runtime. What the target runs is project-specific: validating an embedded document graph, or linting embedded shell and PowerShell templates with the tooling for those languages. A project that embeds nothing omits the target.

## The go directive

The workflows pass `go-version-file: go.mod` (see the workflows fragment), so the `go` directive decides which toolchain builds the release, and therefore which standard library the published binary carries. The form of the directive decides whether security patches reach users on their own:

| Directive | CI installs | Standard library patches |
|---|---|---|
| `go 1.27` | the newest 1.27.x available | arrive automatically |
| `go 1.27.0` | exactly 1.27.0 | frozen until the line is edited |

**Write the minor-only form.** A patch release of Go is security and bug fixes for the same language version, and pinning one holds a signed public binary on a standard library that is known to be superseded. Verify with `go version -m <binary>` against a published artifact rather than assuming; the directive alone does not tell you what shipped.

The trade-off is that rebuilding an old tag later may use a newer toolchain and produce a different binary. That is the right way round: a rebuild should pick up the fixes.

Locally the directive is only a minimum, and a newer installed toolchain always wins. A developer therefore never sees the pin, and the released artifact is the only place it takes effect.

Dependabot does not bump the `go` directive; it updates module requirements only. Raise the minor version deliberately when moving to a new Go release.

## Vulnerability scanning

`govulncheck` is the one permitted addition to the toolchain table above. It is maintained by the Go team, has no configuration, and answers a question no other tool here can.

Dependabot covers dependencies and cannot see the standard library, since the toolchain is not a go.mod requirement. On a project with a handful of dependencies the standard library is most of the attack surface, so the small dependency count argues for the scan rather than against it.

- Invoke it with `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`, never `go install`. Nothing lands on `PATH` and there is nothing to maintain across machines; the module cache makes runs after the first take a few seconds.
- Run it on a schedule, never as a release gate and never in `ci`. A vulnerability discovered while publishing forces a fix at the worst possible moment; the same finding on a Tuesday is ordinary work.
- Scan with the same Go version the release is built with, which `go-version-file: go.mod` gives for free. Scanning with a newer toolchain than the one that builds the binary reports the wrong standard library.
- Act on the `Your code is affected by` count only. The trailing counts for packages imported and modules required are vulnerabilities the analysis proved unreachable, and are the noise the reachability analysis exists to remove.

A standard library finding is usually fixed by bumping the `go` directive to the patched release and cutting a new release.
