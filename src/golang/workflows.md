# Go CI Workflows

Supplements the shared GitHub Actions conventions. Universal to any Go project; a
CLI project's release-specific wiring lives in the release-cli fragment instead.

Paths filter for `pr.yml` and `main.yml`:

```yaml
paths:
  - ".github/workflows/**"
  - "**.go"
  - go.mod
  - go.sum
  - Makefile
```

Setup (before any `make` call). Always use `go-version-file: go.mod`; never hardcode a version:

```yaml
- uses: actions/setup-go@v6
  with:
    go-version-file: go.mod
    cache: true
```

`lint.yml` runs `make fmt_check`, `make mod_check`, `make vet`. `test.yml` runs `make test`. Neither needs `fetch-depth: 0`.
