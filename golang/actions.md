# Go GitHub Actions Workflows

Conventions for structuring GitHub Actions in Go projects.

## Design Principles

- Reusable workflows (`workflow_call`) for all job logic - callers just compose them
- All steps call `make <target>`, never raw Go commands
- Minimal permissions - `contents: read` by default, `contents: write` only for release
- Standard runner: `ubuntu-24.04`
- Go version sourced from `go.mod` via `go-version-file`, never hardcoded

## Workflow Files

```
.github/workflows/
  lint.yml      # reusable: fmt_check, mod_check, vet
  test.yml      # reusable: test
  release.yml   # reusable: goreleaser
  pr.yml        # caller: lint + test on pull requests
  tag.yml       # caller: lint + test + release on v* tags
```

No `push.yml` - linting and testing are only triggered by PRs and tags.

## Reusable Workflows

### lint.yml

```yaml
name: Lint

on: workflow_call

permissions:
  contents: read

jobs:
  go_lint:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6

      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true

      - run: make fmt_check

      - run: make mod_check

      - run: make vet
```

- No `fetch-depth: 0` - not needed for linting
- Steps use `make` targets, not raw commands

### test.yml

```yaml
name: Test

on: workflow_call

permissions:
  contents: read

jobs:
  go_test:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6

      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true

      - run: make test
```

- No `fetch-depth: 0` - not needed for tests

### release.yml

```yaml
name: Release

on: workflow_call

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0  # required by GoReleaser for changelog/versioning

      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true

      - name: Extract release notes from CHANGELOG.md
        run: |
          TAG="${GITHUB_REF_NAME}"
          awk "/^## ${TAG} /{found=1; next} found && /^## /{exit} found{print}" CHANGELOG.md \
            > /tmp/release-notes.md
          if [ ! -s /tmp/release-notes.md ]; then
            echo "No CHANGELOG entry found for ${TAG}" >&2
            exit 1
          fi

      - uses: goreleaser/goreleaser-action@v7
        with:
          args: release --clean --release-notes /tmp/release-notes.md
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- `fetch-depth: 0` is required here for GoReleaser to read git history
- Release notes are extracted from `CHANGELOG.md` matching the tag name
- Fails explicitly if no CHANGELOG entry exists for the tag

## Caller Workflows

### pr.yml

```yaml
name: PR

on:
  pull_request:
    types: [opened, synchronize, reopened]

permissions:
  contents: read

concurrency:
  group: pr-${{ github.event.pull_request.number }}
  cancel-in-progress: true

jobs:
  lint:
    uses: ./.github/workflows/lint.yml

  test:
    uses: ./.github/workflows/test.yml
```

- `concurrency` cancels stale runs when new commits are pushed to a PR

### tag.yml

```yaml
name: Tag

on:
  push:
    tags: ["v*"]

permissions:
  contents: write

jobs:
  lint:
    uses: ./.github/workflows/lint.yml

  test:
    uses: ./.github/workflows/test.yml

  release:
    needs: [lint, test]
    uses: ./.github/workflows/release.yml
    secrets: inherit
```

- Release only runs after both lint and test pass
- `secrets: inherit` passes `GITHUB_TOKEN` to the release workflow

## Conventions

- `fetch-depth: 0` only in `release.yml`, not in lint or test workflows
- `permissions: contents: write` only in `release.yml` and `tag.yml`
- Concurrency group on `pr.yml` to avoid redundant runs
- No workflow runs on direct branch pushes - only PRs and tags
