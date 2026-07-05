# Go CLI Release Process

How a CLI project builds and publishes release binaries. Applies only to projects that ship a compiled binary; a pure library has no equivalent (see the library scaffolding fragment).

## GoReleaser config

Every project has two configs: `.goreleaser.yml` (versioned releases via `release.yml`) and `.goreleaser.prerelease.yml` (dev snapshot via `prerelease.yml`).

- GoReleaser builds binaries only; it does not create checksums, sign, or publish.
- Always inject version via `ldflags`. Default matrix is `linux`/`darwin`/`windows` x `amd64`/`arm64`, excluding `windows/arm64`.
- Always set `no_unique_dist_dir: true` so binaries land flat in `dist/`.
- Prefer `CGO_ENABLED=0` and `mod_timestamp` for reproducible static builds.
- Checksums, install scripts, and signing are handled by `thomaslaurenson/gpipe-action` (`cosign_sign: true` for signing) via `.gpipe.yml`, not goreleaser.

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
```

The prerelease config is identical plus `snapshot.version_template: "{{ incpatch .Version }}-dev"` and `git.ignore_tags` set to the rolling prerelease's own tag (see below), so that leftover tag is never picked up as "the last tag" when computing `incpatch`. The binary naming template maps `amd64` to `x86_64` and `arm64` to `aarch64`; the `.gpipe.yml` platform paths must match exactly.

## CI wiring

Add `.goreleaser*.yml` to the `pr.yml`/`main.yml` paths filter alongside the shared Go entries.

Three-step release pattern: goreleaser builds binaries (`args: build --clean`, does not publish), `gpipe-action` generates install scripts + checksums + cosign bundle, `gh release create` publishes. `fetch-depth: 0` is required in `release.yml`. Always set `GORELEASER_CURRENT_TAG: ${{ github.ref_name }}` so goreleaser does not pick up a `-dev` tag at the same commit. `id-token: write` is required on the workflow and its caller for cosign OIDC signing.

Prerelease uses `goreleaser build --snapshot` with `.goreleaser.prerelease.yml`; version comes from `snapshot.version_template`, so no `GORELEASER_CURRENT_TAG` is needed. Require `fetch-depth: 0` and `fetch-tags: true`. Delete the rolling prerelease by exact tag name with an existence check, never `|| true`.

`gpipe-action` always validates in strict mode (there is no `--dry-run` passthrough) and only accepts a bare `v?X.Y.Z` with no suffix; it also bakes `version` directly into `install.sh`'s download URL, so whatever is passed must equal the actual release tag. That rules out tagging the rolling prerelease `dev` or `X.Y.Z-dev`: use a fixed placeholder tag instead (e.g. `v0.0.0`), with `--title dev` on `gh release create` for a human-readable label, and match it in `git.ignore_tags` above.
