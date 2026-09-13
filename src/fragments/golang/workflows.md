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

If the binary embeds non-Go assets (via `go:embed`), add those trees to the filter too. A change under an embedded directory is a change to the compiled binary, so CI must run for it even though no `.go` file changed. For example, a tool that embeds `templates/` and `schemas/` adds `"templates/**"` and `"schemas/**"`.

Setup (before any `make` call). Always use `go-version-file: go.mod`; never hardcode a version:

```yaml
- uses: actions/setup-go@vN
  with:
    go-version-file: go.mod
    cache: true
```

The two reusable workflows fill the skeleton from the GitHub Actions fragment with the Go setup step. Calling the aggregate rather than listing its members is what keeps the workflow and the `ci` target from drifting apart:

```yaml
# lint.yml
name: Lint

on:
  workflow_call:

jobs:
  lint:
    runs-on: ubuntu-24.04
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@vN

      - uses: actions/setup-go@vN
        with:
          go-version-file: go.mod
          cache: true

      - run: make check_all
```

`test.yml` is the same file with `name: Test`, the job named `test`, and `make test` as its final step. Neither needs `fetch-depth: 0`.

## Cross-platform coverage

CI runs on `ubuntu-24.04` only, while releases ship linux, darwin, and windows binaries. `make check_cross` closes the gap that matters: it type-checks for windows and darwin on the Linux runner, so a file behind `//go:build windows` cannot stop compiling without CI noticing. It vets rather than builds, because `go build` never looks at test files and a platform-tagged test rots just as quietly as a platform-tagged source file. Platform-specific source is where this bites, since a Linux-only build never looks at it.

Tests themselves stay on Linux. A macOS or Windows leg would run the same unit tests against the same pure logic, and anything genuinely platform-dependent (mounting, ssh, subprocess behaviour) is integration-tagged and excluded from CI regardless. The leg would cost queue time and catch almost nothing.

### If real cross-platform test execution is needed

Should a platform bug ever escape, add a matrix leg to `tag.yml` rather than `pr.yml`, so pull requests stay fast and only releases pay; What a job costs in the GitHub Actions fragment is why the leg is gated rather than added everywhere.

- **macOS** is straightforward: add `macos-15` to the matrix and it runs `make test` unchanged, because the image has GNU make.
- **Windows needs a carve-out**, and the two obvious fixes both fail. The image has no GNU make. Installing it via Chocolatey gives a native Windows binary that cannot resolve the Makefile's `SHELL := /bin/bash`, since Git Bash lives at `C:\Program Files\Git\bin\bash.exe`. Installing MSYS2's make does resolve `/bin/bash`, but then Go, a native Windows binary, receives MSYS paths such as `/c/Users/...` and mishandles them. The workable route is to call `go test -race -count=1 ./...` directly on the Windows leg, accepting that the command then exists both there and in the Makefile.

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

`go-version-file: go.mod` matters here for the reason the tooling fragment gives: the scan has to run against the standard library the release is built with.

Keep `workflow_dispatch` so the scan can be run by hand after bumping the `go` directive, without waiting for the next Monday.

GitHub disables scheduled workflows after 60 days without repository activity, so a dormant project stops scanning silently. Re-enable it from the Actions tab, or run it by hand, when returning to one.

