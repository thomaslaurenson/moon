# GoReleaser Conventions

Short rules for `.goreleaser.yml` in Go projects.

## Non-Negotiable Rules

- Goreleaser builds binaries only; it does not create checksums, sign, or publish releases
- Always inject version via `ldflags`
- Default build matrix is `linux`, `darwin`, `windows` x `amd64`, `arm64`, excluding `windows/arm64`
- Always set `GORELEASER_CURRENT_TAG` in the release workflow env to pin the tag explicitly; not required for prerelease (snapshot mode resolves version from history)
- Use the OS/arch naming template so binary names match the `.gpipe.yml` platform map exactly
- Always set `no_unique_dist_dir: true` so all binaries land flat in `dist/`
- Checksums, install scripts, and signing are handled by `gpipe` via `.gpipe.yml`, not goreleaser

## Two Config Files

Every project requires two goreleaser configs:

| File | Used by | Purpose |
|---|---|---|
| `.goreleaser.yml` | `release.yml` workflow | Official versioned releases triggered by `v*.*.*` tags |
| `.goreleaser.prerelease.yml` | `prerelease.yml` workflow | Dev snapshot built on every push to `main` |

## Release Config (`.goreleaser.yml`)

```yaml
# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
version: 2

project_name: <name>

builds:
  - env:
      - CGO_ENABLED=0
    mod_timestamp: "{{ .CommitTimestamp }}"
    binary: >-
      <name>-{{ .Os }}-{{ if eq .Arch "amd64" }}x86_64{{ else if eq .Arch "arm64" }}aarch64{{ else }}{{ .Arch }}{{ end }}
    no_unique_dist_dir: true
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w -X <module>/cmd.Version={{.Version}}
```

## Prerelease Config (`.goreleaser.prerelease.yml`)

Identical to the release config with two additions: `snapshot.version_template` sets the version string embedded in the binary (using `incpatch` on the latest release tag in history), and `git.ignore_tags` prevents GoReleaser treating a previous `dev` or `-dev` tag as the base version when walking history. Used with `goreleaser build --snapshot`; no `GORELEASER_CURRENT_TAG` is required.

```yaml
# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
version: 2

project_name: <name>

builds:
  - env:
      - CGO_ENABLED=0
    mod_timestamp: "{{ .CommitTimestamp }}"
    binary: >-
      <name>-{{ .Os }}-{{ if eq .Arch "amd64" }}x86_64{{ else if eq .Arch "arm64" }}aarch64{{ else }}{{ .Arch }}{{ end }}
    no_unique_dist_dir: true
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w -X <module>/cmd.Version={{.Version}}

snapshot:
  version_template: "{{ incpatch .Version }}-dev"

git:
  ignore_tags:
    - "*-dev"
    - "dev"
```

## `.gpipe.yml`

Every project also requires a `.gpipe.yml` that maps platform identifiers to the binary paths and asset names goreleaser produces. This file is consumed by `gpipe-action` in the release workflow to produce `install.sh`, `install.ps1`, and `checksums.txt`.

The `path` values mirror the flat `dist/` output from goreleaser. The `name` values are the GitHub release asset filenames.

```yaml
binary: <name>

platforms:
  linux_amd64:
    path: ./dist/<name>-linux-x86_64
    name: <name>-linux-x86_64
  linux_arm64:
    path: ./dist/<name>-linux-aarch64
    name: <name>-linux-aarch64
  darwin_amd64:
    path: ./dist/<name>-darwin-x86_64
    name: <name>-darwin-x86_64
  darwin_arm64:
    path: ./dist/<name>-darwin-aarch64
    name: <name>-darwin-aarch64
  windows_amd64:
    path: ./dist/<name>-windows-x86_64.exe
    name: <name>-windows-x86_64.exe
```

## Notes

- `CHANGELOG.md` release notes are extracted in the workflow via `make get_changelog TAG=...` and written to `/tmp/release-notes.md`
- Prefer `CGO_ENABLED=0` and `mod_timestamp` for reproducible static builds
- The binary naming template maps `amd64` to `x86_64` and `arm64` to `aarch64`; the `.gpipe.yml` paths must match exactly
- If the project entrypoint is not the module root, add `main: <path>` to the builds entry explicitly (e.g. `main: ./cmd/<name>`)
