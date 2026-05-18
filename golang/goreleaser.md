# GoReleaser Conventions

Short rules for `.goreleaser.yml` in Go projects.

## Non-Negotiable Rules- Build binaries only, not archives (`formats: [binary]`)
- Always generate `checksums.txt` with SHA256
- Always sign `checksums.txt` with `cosign` using bundle format (`--bundle`)
- Always inject version via `ldflags`
- Default build matrix is `linux`, `darwin`, `windows` x `amd64`, `arm64`, excluding `windows/arm64`
- Always set `name_template` on archives explicitly
- Disable goreleaser changelog generation; release notes always come from `CHANGELOG.md`
- Always set `GORELEASER_CURRENT_TAG` in the workflow env to pin the tag explicitly

## Two Config Files

Every project requires two goreleaser configs:

| File | Used by | Purpose |
|---|---|---|
| `.goreleaser.yml` | `release.yml` workflow | Official versioned releases triggered by `v*.*.*` tags |
| `.goreleaser.prerelease.yml` | `prerelease.yml` workflow | Dev snapshot released on every push to `main` |

## Release Config (`.goreleaser.yml`)

```yaml
# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
version: 2

project_name: <name>

builds:
  - env:
      - CGO_ENABLED=0
    mod_timestamp: "{{ .CommitTimestamp }}"
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

archives:
  - formats: [binary]
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: checksums.txt
  algorithm: sha256

signs:
  - artifacts: checksum
    cmd: cosign
    signature: ${artifact}.sig
    args:
      - sign-blob
      - "--bundle=${signature}"
      - "${artifact}"
      - "--yes"

release:
  prerelease: auto
  draft: false

changelog:
  disable: true
```

## Prerelease Config (`.goreleaser.prerelease.yml`)

Identical to the release config except for the `release` and `git` sections:

```yaml
# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
version: 2

project_name: <name>

builds:
  - env:
      - CGO_ENABLED=0
    mod_timestamp: "{{ .CommitTimestamp }}"
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

archives:
  - formats: [binary]
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: checksums.txt
  algorithm: sha256

signs:
  - artifacts: checksum
    cmd: cosign
    signature: ${artifact}.sig
    args:
      - sign-blob
      - "--bundle=${signature}"
      - "${artifact}"
      - "--yes"

release:
  prerelease: true
  draft: false
  name_template: "Dev (Pre-release)"

changelog:
  disable: true

git:
  ignore_tags:
    - "*-dev"
```

The `git.ignore_tags` prevents goreleaser treating a previous `-dev` tag as the previous release when computing context.

## Notes

- `CHANGELOG.md` release notes are extracted in the workflow via `make get_changelog_entry TAG=...` and passed to goreleaser with `--release-notes`
- Prefer `CGO_ENABLED=0` and `mod_timestamp` for reproducible static builds
- Do not add `release.github` unless there is a project-specific reason
- cosign uses `--bundle` (not `--output-signature`); the bundle format includes the certificate and signature in a single file
