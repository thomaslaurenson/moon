# Go CLI release process

How a CLI project builds and publishes release binaries. Applies to every Go project in these specs, since all of them ship a compiled binary.

## GoReleaser config

Every project has two configs: `.goreleaser.yml` (versioned releases via `release.yml`) and `.goreleaser.prerelease.yml` (dev snapshot via `prerelease.yml`).

- GoReleaser builds binaries only; it does not create checksums, sign, or publish.
- Always inject version via `ldflags`. Default matrix is `linux`/`darwin`/`windows` x `amd64`/`arm64`, excluding `windows/arm64`.
- Windows on ARM runs x64 binaries under emulation, so the excluded `windows/arm64` build is a performance optimisation rather than a compatibility requirement. A project that ships `install.ps1` may include it: `Get-Platform` reports `windows_arm64`, and if no such asset exists the installer fails outright rather than falling back to the emulated x64 build. Include it in both goreleaser configs and in `.gpipe.yml`, or in neither.
- Always set `no_unique_dist_dir: true` so binaries land flat in `dist/`.
- Prefer `CGO_ENABLED=0` and `mod_timestamp` for reproducible static builds.
- Checksums, install scripts, and signing are gpipe's job, not goreleaser's; see the gpipe fragment.

```yaml
# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
version: 2
project_name: <name>
builds:
  - env: [CGO_ENABLED=0]
    mod_timestamp: "{{ .CommitTimestamp }}"
    binary: >-
      <name>-{{ .Os }}-{{ if eq .Arch "amd64" }}x86_64{{ else if eq .Arch "arm64" }}aarch64{{ else }}{{ .Arch }}{{ end }}
    no_unique_dist_dir: true
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ignore:
      - {goos: windows, goarch: arm64}
    ldflags:
      - -s -w -X <module>/cmd.Version={{.Version}}
git:
  ignore_tags: ["*-dev", "dev"]
```

`git.ignore_tags` belongs in both configs, not just the prerelease one. The prerelease channel publishes under a moving `dev` tag (see below), so once a developer has fetched tags, goreleaser's own tag discovery names that tag as the last release. A tagged CI release never sees it, because `GORELEASER_CURRENT_TAG` overrides discovery outright, and neither does the prerelease build, which already carried the setting. What it fixes is the local preview:

```text
$ make snapshot        # without it
dev-SNAPSHOT-120e4d7
$ make snapshot        # with it
1.2.3-SNAPSHOT-704f5bc
```

The `make build` path reaches the same tag through `git describe` and closes it with `--match 'v*'` instead (see the Makefile targets fragment).

The prerelease config is this file plus `snapshot.version_template: "{{ incpatch .Version }}-dev"`, so the ignored `dev` tag is never picked up as "the last tag" when computing `incpatch`. The binary naming template maps `amd64` to `x86_64` and `arm64` to `aarch64`; the `.gpipe.yml` platform paths must match exactly.

## CI wiring

Add `.goreleaser*.yml` to the `pr.yml`/`main.yml` paths filter alongside the shared Go entries. The gpipe fragment covers `.gpipe.yml`.

goreleaser is the build step of the three-step release (see the gpipe fragment): `args: build --clean`, which builds and does not publish.

```yaml
# release.yml
name: Release

on:
  workflow_call:

jobs:
  release:
    runs-on: ubuntu-24.04
    permissions:
      contents: write
      id-token: write
    steps:
      - uses: actions/checkout@vN
        with:
          fetch-depth: 0

      - uses: actions/setup-go@vN
        with:
          go-version-file: go.mod
          cache: true

      - name: Build binaries
        uses: goreleaser/goreleaser-action@<sha> # v<version>
        with:
          version: "~> v2"
          args: build --clean
        env:
          GORELEASER_CURRENT_TAG: ${{ github.ref_name }}

      - uses: thomaslaurenson/gpipe@vN
        with:
          cosign_sign: true

      - name: Get changelog
        run: make get_changelog TAG=${{ github.ref_name }} > /tmp/release-notes.md

      - name: Create release
        run: |
          gh release create "${{ github.ref_name }}" \
            dist/<name>-* \
            install.sh install.ps1 \
            checksums.txt checksums.txt.sigstore.json \
            --title "${{ github.ref_name }}" \
            --notes-file /tmp/release-notes.md
        env:
          GH_TOKEN: ${{ github.token }}
```

Three details in there are load-bearing. `fetch-depth: 0` is needed because goreleaser reads tags. `GORELEASER_CURRENT_TAG` must always be set, so goreleaser does not pick up a `-dev` tag sitting on the same commit. And `version: "~> v2"` pins the goreleaser binary, which is a separate thing from the `@<sha>` pinning the action.

The `actions/setup-go` here is for goreleaser, and gpipe rides on it. gpipe installs no Go of its own and builds with whatever is on `PATH`, so this step is what settles the version both of them get (see the gpipe fragment).

## Release artefacts

goreleaser writes the binaries into `dist/`. Everything else a release publishes is written by gpipe into the repository root; the gpipe fragment lists the files and explains why they all belong in `clean`.

## Prerelease process

The prerelease channel is a single rolling GitHub release under the literal tag `dev`, rebuilt on every push to main: raw binaries only, no install scripts, checksums or signing, because gpipe cannot run against a `dev` tag at all (see the gpipe fragment). `prerelease.yml` therefore needs `contents: write` and not `id-token: write`.

`dev` is a real git tag that moves, which is why checkout needs both `fetch-depth: 0` and `fetch-tags: true`: `git tag -f` has to see the existing tag.

```yaml
# prerelease.yml
name: Prerelease

on:
  workflow_call:

jobs:
  prerelease:
    runs-on: ubuntu-24.04
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@vN
        with:
          fetch-depth: 0
          fetch-tags: true

      - uses: actions/setup-go@vN
        with:
          go-version-file: go.mod
          cache: true

      - name: Delete any existing dev release
        run: |
          if err=$(gh release view "dev" 2>&1 >/dev/null); then
            gh release delete "dev" --yes --cleanup-tag
          elif grep -qi "release not found" <<<"$err"; then
            echo "No existing dev release"
          else
            echo "::error::could not determine whether a dev release exists: ${err}"
            exit 1
          fi
        env:
          GH_TOKEN: ${{ github.token }}

      - name: Tag and push dev
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git tag -f dev
          git push origin dev --force

      - name: Build binaries
        uses: goreleaser/goreleaser-action@<sha> # v<version>
        with:
          version: "~> v2"
          args: build --snapshot --clean --config .goreleaser.prerelease.yml

      - name: Create dev release
        run: |
          gh release create dev \
            --title "Dev (Pre-release)" \
            --notes "Built from commit ${{ github.sha }}" \
            --prerelease \
            dist/<name>-*
        env:
          GH_TOKEN: ${{ github.token }}
```

No `GORELEASER_CURRENT_TAG` here: the version comes from `snapshot.version_template`.

### Why the delete is existence-checked

Three outcomes, three behaviours: the release exists, so delete it; `gh` reports it missing, so carry on; `gh` failed for any other reason, so stop. That third branch is the point of the whole step. A rate limit, an expired token or a flaky API all fail the `view`, and treating that as "no release exists" means the job skips the delete and then publishes on top of a release whose state it never established.

`2>&1 >/dev/null` captures stderr while discarding stdout, and the order matters. Redirections apply left to right, so this points stderr at the still-open capture and only then sends stdout to `/dev/null`. Writing `>/dev/null 2>&1` sends both to `/dev/null`, leaves `err` empty, and the not-found branch can never match.

Never write this as `gh release delete dev --yes --cleanup-tag || true`. That collapses all three outcomes into one and continues regardless. `--cleanup-tag` deletes the git tag along with the release, which step 2 then recreates at the new commit.
