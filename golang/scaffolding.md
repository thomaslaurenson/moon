# Go Project Scaffolding

Standards and conventions for Go projects. Use this as a reference when creating or refactoring a Go repository.

## Project Structure

```
cmd/           # cobra command definitions (root.go, run.go, version.go, etc.)
internal/      # private packages not imported by other modules
main.go        # entry point — minimal, delegates to cmd/
.github/
  workflows/
  dependabot.yml
.gitignore
.goreleaser.yml
go.mod
go.sum
Makefile
CHANGELOG.md
README.md
```

- Business logic lives in `internal/`, not `cmd/`
- `cmd/` only handles CLI parsing and wiring
- No `pkg/` directory — use `internal/` unless the package is intentionally public
- Do **not** add a named subdirectory under `internal/` (e.g. `internal/myapp/`) — this is only warranted in a monorepo with multiple binaries that need scoped internals. For a single-binary project, packages sit directly under `internal/`

## Tools

The only tooling used is what ships with Go itself. No third-party linters.

| Tool | Purpose |
|---|---|
| `gofmt` | Format source files |
| `go vet` | Static analysis |
| `go test` | Run tests |
| `go mod tidy` | Keep go.mod/go.sum clean |

**Not used:** `golangci-lint`, `govulncheck`, or any other third-party analysis tools.

## Makefile

All CI steps are Makefile targets. GitHub Actions call `make <target>` — never raw Go commands directly in workflows.

Key targets:

```makefile
fmt:          gofmt -w .
fmt_check:    gofmt -l . (exits 1 if unformatted files found)
mod_check:    go mod tidy && git diff --exit-code go.mod go.sum
vet:          go vet ./...
test:         go test -race -count=1 ./...
test_verbose: go test -race -count=1 -v ./...
test_coverage: go test -race -count=1 -coverpkg=./internal/... -coverprofile=coverage.out ./...
ci:           fmt_check mod_check vet test
build:        go build -ldflags="..." -o bin/<binary> .
install:      go install -ldflags="..." .
clean:        rm -rf bin/ dist/
snapshot:     goreleaser release --snapshot --clean
release_check: goreleaser check
```

Rules:
- Tests always include `-race -count=1`, including coverage runs
- The binary is built to `bin/` not the repo root
- `ci` is the local equivalent of the full CI pipeline

## go.mod

- Use the latest stable Go version
- No `replace` directives in committed code
- Run `go mod tidy` before committing

## GoReleaser

File: `.goreleaser.yml`

```yaml
# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
# GoReleaser v2
version: 2

dist: dist

builds:
  - env:
      - CGO_ENABLED=0
    mod_timestamp: "{{ .CommitTimestamp }}"  # reproducible builds
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w -X <module>/cmd.Version={{.Version}}

archives:
  - name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    formats: [tar.gz]
    format_overrides:
      - goos: windows
        formats: [zip]

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

changelog:
  sort: desc
  filters:
    exclude: ["^docs:", "^test:", "^chore:", "^Merge "]
```

- No `release.github` block — GoReleaser auto-detects owner/repo from the git remote
- `CGO_ENABLED=0` for static binaries
- `mod_timestamp` makes builds reproducible
- Changelog sorted `desc` (newest commits first)

## .gitignore

```gitignore
# CUSTOM
bin/
dist/
```

- Use `bin/` not the binary name directly — works for any project

## Dependabot

File: `.github/dependabot.yml`

```yaml
version: 2
updates:
  - package-ecosystem: github-actions
    directory: /
    schedule:
      interval: weekly
    assignees:
      - "<username>"

  - package-ecosystem: gomod
    directory: /
    schedule:
      interval: weekly
    open-pull-requests-limit: 0  # security updates only
    assignees:
      - "<username>"
```

- GitHub Actions: weekly bumps
- Go modules: security updates only (`open-pull-requests-limit: 0`)

## README Badges

Place at the top of README.md. All badges use `style=flat`.

```markdown
![Build Status](https://img.shields.io/github/actions/workflow/status/<owner>/<repo>/tag.yml?style=flat)
![Test Status](https://img.shields.io/github/actions/workflow/status/<owner>/<repo>/tag.yml?style=flat&label=test)

![Release Version](https://img.shields.io/github/v/release/<owner>/<repo>?style=flat)
![Release downloads](https://img.shields.io/github/downloads/<owner>/<repo>/total?label=downloads)

![Go Version](https://img.shields.io/github/go-mod/go-version/<owner>/<repo>)
![Code Coverage](https://img.shields.io/badge/coverage-XX%25-blue)
```

- Build/test badges point to `tag.yml` (reflects last release health)
- Coverage badge is updated manually after running `make test_coverage`
