# Go CI workflows

Supplements the shared GitHub Actions conventions. Universal to any Go project; a CLI project's release-specific wiring lives in the release-cli fragment instead.

Paths filter for `pr.yml` and `main.yml`:

```yaml
paths:
  - ".github/workflows/**"
  - "**.go"
  - go.mod
  - go.sum
  - Makefile
```

If the binary embeds non-Go assets (via `go:embed`), add those trees to the filter too. A change under an embedded directory is a change to the compiled binary, so CI must run for it even though no `.go` file changed. For example, a tool that embeds `src/` and `bundles/` adds `"src/**"` and `"bundles/**"`.

Setup (before any `make` call). Always use `go-version-file: go.mod`; never hardcode a version:

```yaml
- uses: actions/setup-go@vN
  with:
    go-version-file: go.mod
    cache: true
```

`@vN` means pin the current major of the action at authoring time (for example `@v6`); Dependabot keeps the pin current from there. Do not copy a version number from this document as the target to match.

`lint.yml` runs `make fmt_check`, `make mod_check`, `make vet`. `test.yml` runs `make test`. Neither needs `fetch-depth: 0`.
