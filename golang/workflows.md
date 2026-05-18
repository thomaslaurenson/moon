# Go Workflow Conventions

Supplements `github/actions.md`. Universal rules (runners, action versions, workflow structure, caller patterns, concurrency, permissions) apply unchanged. This file covers Go-specific divergences and reusable workflow bodies only.

---

## Paths Filter

Use these entries in the `paths:` filter for `pr.yml` and `main.yml`:

```yaml
paths:
  - ".github/workflows/**"
  - "**.go"
  - go.mod
  - go.sum
  - .goreleaser*.yml
  - Makefile
```

---

## Go Setup Steps

Add these before any `make` call in reusable workflows. Always use `go-version-file: go.mod`; never hardcode a Go version:

```yaml
- uses: actions/setup-go@v6
  with:
    go-version-file: go.mod
    cache: true
```

`fetch-depth: 0` is required only in `release.yml` (goreleaser and changelog extraction need full history). Omit it in `lint.yml` and `test.yml`.

---

## Releases: Goreleaser Instead of `gh` CLI

Go projects use `goreleaser/goreleaser-action` for releases. This is the only exception to the `gh` CLI release rule in `github/actions.md`. Goreleaser handles building, signing, and publishing in a single step. See `golang/goreleaser.md` for config details.

Install cosign before running goreleaser:

```yaml
- uses: sigstore/cosign-installer@v4.1.2
```

Always set `GORELEASER_CURRENT_TAG`. Without it, goreleaser may pick up a `-dev` tag pointing at the same commit as the release tag and release the wrong version:

```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  GORELEASER_CURRENT_TAG: ${{ github.ref_name }}
```

---

## Reusable Workflow Bodies

### `lint.yml`

```yaml
name: Lint

on:
  workflow_call

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

### `test.yml`

```yaml
name: Test

on:
  workflow_call

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

### `release.yml`

Release notes are extracted from `CHANGELOG.md` via `make get_changelog_entry` and passed to goreleaser with `--release-notes`.

```yaml
name: Release

on:
  workflow_call

permissions:
  contents: write
  id-token: write

jobs:
  goreleaser:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true

      - uses: sigstore/cosign-installer@v4.1.2

      - name: Extract release notes from CHANGELOG.md
        run: make get_changelog_entry TAG=${GITHUB_REF_NAME} > /tmp/release-notes.md

      - uses: goreleaser/goreleaser-action@v7
        with:
          version: "~> v2"
          args: release --clean --release-notes /tmp/release-notes.md
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GORELEASER_CURRENT_TAG: ${{ github.ref_name }}
```

### `prerelease.yml`

Creates a rolling dev prerelease on every push to `main`. The dev tag is derived from the latest release tag with an incremented patch version and `-dev` suffix (e.g. `v0.5.1` becomes `v0.5.2-dev`). The tag and release are deleted and recreated on every run.

Delete the release by exact tag name. Do not use a generic grep pattern, and do not use `|| true`. Masking deletion failures allows goreleaser to find an existing release with stale assets, which causes all asset uploads to fail with `already_exists`.

```yaml
name: Prerelease

on:
  workflow_call

permissions:
  contents: write
  id-token: write

jobs:
  prerelease:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true

      - uses: sigstore/cosign-installer@v4.1.2

      - name: Compute dev tag
        id: dev_tag
        run: |
          LATEST=$(git tag --list 'v*.*.*' | grep -Ev -- '-' | sort -V | tail -1)
          if [ -z "$LATEST" ]; then
            echo "Error: no release tag found" >&2
            exit 1
          fi
          PATCH=$(echo "$LATEST" | cut -d. -f3)
          DEV_TAG="$(echo "$LATEST" | cut -d. -f1-2).$((PATCH + 1))-dev"
          echo "dev_tag=${DEV_TAG}" >> "$GITHUB_OUTPUT"
          echo "Computed dev tag: ${DEV_TAG} (from latest release: ${LATEST})"

      - name: Delete any existing dev release
        run: |
          DEV_TAG="${{ steps.dev_tag.outputs.dev_tag }}"
          if gh release view "${DEV_TAG}" > /dev/null 2>&1; then
            gh release delete "${DEV_TAG}" --yes --cleanup-tag
            echo "Deleted existing dev release: ${DEV_TAG}"
          else
            echo "No existing dev release for ${DEV_TAG}"
          fi
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Tag and push dev tag
        run: |
          git config user.name  "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git tag -f ${{ steps.dev_tag.outputs.dev_tag }}
          git push origin ${{ steps.dev_tag.outputs.dev_tag }} --force

      - name: Generate dev release notes
        run: echo "Built from commit ${{ github.sha }}" > /tmp/dev-release-notes.md

      - uses: goreleaser/goreleaser-action@v7
        with:
          version: "~> v2"
          args: >
            release
            --clean
            --config .goreleaser.prerelease.yml
            --release-notes /tmp/dev-release-notes.md
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GORELEASER_CURRENT_TAG: ${{ steps.dev_tag.outputs.dev_tag }}
```
