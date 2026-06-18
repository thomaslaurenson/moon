# Go Workflow Conventions

Supplements `github/actions.md`. Universal rules (runners, action versions, workflow structure, caller patterns, concurrency, permissions) apply unchanged. This file covers Go-specific divergences and reusable workflow bodies only.

## Paths Filter

Use these entries in the `paths:` filter for `pr.yml` and `main.yml`:

```yaml
paths:
  - ".github/workflows/**"
  - "**.go"
  - go.mod
  - go.sum
  - .goreleaser*.yml
  - .gpipe.yml
  - Makefile
```

## Go Setup Steps

Add these before any `make` call in reusable workflows. Always use `go-version-file: go.mod`; never hardcode a Go version:

```yaml
- uses: actions/setup-go@v6
  with:
    go-version-file: go.mod
    cache: true
```

`fetch-depth: 0` is required only in `release.yml` (goreleaser and changelog extraction need full history). Omit it in `lint.yml` and `test.yml`.

## Releases: Build, gpipe, Publish

Go projects use a three-step release pattern:

1. **Build** - `goreleaser/goreleaser-action` with `args: build --clean` to produce platform binaries in `dist/`. Goreleaser does not publish.
2. **gpipe** - `thomaslaurenson/gpipe-action@v1` to generate `install.sh`, `install.ps1`, `checksums.txt`, and the cosign bundle `checksums.txt.sigstore.json`.
3. **Publish** - `gh release create` to create the GitHub release and upload all assets.

Always set `GORELEASER_CURRENT_TAG`. Without it, goreleaser may pick up a `-dev` tag pointing at the same commit as the release tag and build the wrong version:

```yaml
env:
  GORELEASER_CURRENT_TAG: ${{ github.ref_name }}
```

`id-token: write` is required on the workflow and its caller (`tag.yml`) for cosign OIDC signing inside `gpipe-action`.

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

Release notes are extracted from `CHANGELOG.md` via `make get_changelog TAG=...`. The `id-token: write` permission is required for cosign OIDC signing inside `gpipe-action`. Replace `<name>` with the project binary name.

```yaml
name: Release

on:
  workflow_call

permissions:
  contents: write
  id-token: write

jobs:
  release:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true

      - name: Extract release notes from CHANGELOG.md
        run: make get_changelog TAG=${GITHUB_REF_NAME} > /tmp/release-notes.md

      - name: Build binaries
        uses: goreleaser/goreleaser-action@v7
        with:
          version: "~> v2"
          args: build --clean
        env:
          GORELEASER_CURRENT_TAG: ${{ github.ref_name }}

      - uses: thomaslaurenson/gpipe-action@v1
        with:
          cosign_sign: true

      - name: Create GitHub release
        run: |
          gh release create "${{ github.ref_name }}" \
            --title "${{ github.ref_name }}" \
            --notes-file /tmp/release-notes.md \
            dist/<name>-* \
            install.sh \
            install.ps1 \
            checksums.txt \
            checksums.txt.sigstore.json
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### `prerelease.yml`

Creates a rolling dev prerelease on every push to `main`. The GitHub release always uses the static `dev` tag so there is only ever one dev release on the releases page. GoReleaser runs in `--snapshot` mode, which bypasses the semver tag requirement entirely. The version embedded in the binary is set via `snapshot.version_template` in the goreleaser config (e.g. `{{ incpatch .Version }}-dev`), which walks existing release tags in history. No `GORELEASER_CURRENT_TAG` is needed.

`fetch-depth: 0` and `fetch-tags: true` are still required so GoReleaser can walk the full tag history to resolve `{{ .Version }}` inside the snapshot template.

Delete the `dev` release by exact tag name. Do not use a generic grep pattern, and do not use `|| true`. Masking deletion failures allows `gh release create` to find an existing release with stale assets, which causes all asset uploads to fail with `already_exists`.

The `.goreleaser.prerelease.yml` `git.ignore_tags` must list both `"*-dev"` and `"dev"` so GoReleaser never treats either as a base tag when resolving `{{ .Version }}`.

Replace `<name>` with the project binary name.

```yaml
name: Prerelease

on:
  workflow_call

permissions:
  contents: write

jobs:
  prerelease:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0
          fetch-tags: true

      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
          cache: true

      - name: Delete any existing dev release
        run: |
          if gh release view "dev" > /dev/null 2>&1; then
            gh release delete "dev" --yes --cleanup-tag
          fi
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Tag and push dev tag
        run: |
          git config user.name  "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git tag -f dev
          git push origin dev --force

      - name: Build binaries
        uses: goreleaser/goreleaser-action@v7
        with:
          version: "~> v2"
          args: build --snapshot --clean --config .goreleaser.prerelease.yml

      - name: Create dev release
        run: |
          gh release create "dev" \
            --title "Dev (Pre-release)" \
            --notes "Built from commit [${{ github.sha }}](${{ github.server_url }}/${{ github.repository }}/commit/${{ github.sha }})" \
            --prerelease \
            dist/<name>-*
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### `main.yml`

Triggers on every push to `main`. Runs lint and test, then calls `prerelease.yml` if both pass.

```yaml
name: Main

on:
  push:
    branches: [main]
    paths:
      - ".github/workflows/**"
      - "**.go"
      - go.mod
      - go.sum
      - .goreleaser*.yml
      - .gpipe.yml
      - Makefile

permissions:
  contents: write

concurrency:
  group: main-${{ github.ref }}
  cancel-in-progress: false

jobs:
  lint:
    uses: ./.github/workflows/lint.yml

  test:
    uses: ./.github/workflows/test.yml

  prerelease:
    needs: [lint, test]
    uses: ./.github/workflows/prerelease.yml
    secrets: inherit
```
