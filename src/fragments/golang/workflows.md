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

## vuln.yml

A fourth workflow, scheduled rather than triggered by a change, runs `make vuln`. It is the exception to the reusable-workflow split: it has exactly one caller, so separating body from caller would add a file and no reuse.

```yaml
name: Vuln

on:
  schedule:
    - cron: "0 6 * * 1"
  workflow_dispatch:

jobs:
  vuln:
    runs-on: ubuntu-24.04
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@vN
      - uses: actions/setup-go@vN
        with:
          go-version-file: go.mod
          cache: true
      - run: make vuln
```

No `paths:` filter and no concurrency group: it is not responding to a change, and a vulnerability disclosed against unchanged code is the case it exists to catch.

`go-version-file: go.mod` matters here for the same reason it does in the release workflow. The scan must run against the standard library the release is built with, not whatever is newest.

Keep `workflow_dispatch` so the scan can be run by hand after bumping the `go` directive, without waiting for the next Monday.

GitHub disables scheduled workflows after 60 days without repository activity, so a dormant project stops scanning silently. Re-enable it from the Actions tab, or run it by hand, when returning to one.

