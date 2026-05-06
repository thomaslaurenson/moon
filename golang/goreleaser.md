# GoReleaser Conventions

Short rules for `.goreleaser.yml` in Go projects.

## Non-Negotiable Rules

No third-party linters or formatters are permitted. Specifically, DO NOT use `golangci-lint` or `govulncheck` under any circumstances. However, third-party release tools like `goreleaser` and `cosign` are explicitly permitted and required.

- Build binaries only, not archives (`formats: [binary]`)
- Always output artifacts to `bin` (`dist: bin`)
- Always generate `checksums.txt`
- Always sign `checksums.txt` with `cosign`
- Always inject version via `ldflags`
- Default build matrix is `linux`, `darwin`, `windows` x `amd64`, `arm64`
- Build matrix can be reduced or extended per project requirements
- Release notes/changelog always come from repository `CHANGELOG.md`

## Notes

- `CHANGELOG.md` release notes are extracted in workflow and passed to GoReleaser
- Prefer `CGO_ENABLED=0` and `mod_timestamp` for reproducible static builds
- Do not add `release.github` unless there is a project-specific reason

## Minimal Reference

```yaml
# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
version: 2

dist: bin

builds:
  - env:
      - CGO_ENABLED=0
    mod_timestamp: "{{ .CommitTimestamp }}"
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w -X <module>/cmd.Version={{ .Version }}

archives:
  - formats: [binary]

checksum:
  name_template: checksums.txt
  algorithm: sha256

signs:
  - artifacts: checksum
    cmd: cosign
    args:
      - sign-blob
      - --yes
      - --output-signature=${signature}
      - ${artifact}
    signature: ${artifact}.sig
```
