# Go CLI release process

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

The prerelease config is identical plus `snapshot.version_template: "{{ incpatch .Version }}-dev"` and `git.ignore_tags: ["*-dev", "dev"]`, so the moving `dev` tag (see below) is never picked up as "the last tag" when computing `incpatch`. The binary naming template maps `amd64` to `x86_64` and `arm64` to `aarch64`; the `.gpipe.yml` platform paths must match exactly.

## CI wiring

Add `.goreleaser*.yml` and `.gpipe.yml` to the `pr.yml`/`main.yml` paths filter alongside the shared Go entries.

Three-step release pattern: goreleaser builds binaries (`args: build --clean`, does not publish), `gpipe-action` generates install scripts + checksums + cosign bundle, `gh release create` publishes. `fetch-depth: 0` is required in `release.yml`. Always set `GORELEASER_CURRENT_TAG: ${{ github.ref_name }}` so goreleaser does not pick up a `-dev` tag at the same commit. `id-token: write` is required on the workflow and its caller for cosign OIDC signing.

## Prerelease process

The prerelease channel is a single rolling GitHub release under the literal tag `dev`, rebuilt on every push to main: raw binaries only, no install scripts, checksums, or cosign signing (those are release-only, via `gpipe-action`). `id-token: write` is not needed for `prerelease.yml`, only `contents: write`.

`dev` is a real git tag that moves. `prerelease.yml` runs four steps in order:

1. Delete any existing `dev` release, existence-checked; see below.
2. Force the tag to the current commit and push it: `git tag -f dev` then `git push origin dev --force`, after setting a bot `user.name`/`user.email`.
3. Build: `goreleaser build --snapshot` with `.goreleaser.prerelease.yml`. The version comes from `snapshot.version_template`, so no `GORELEASER_CURRENT_TAG` is needed.
4. Publish: `gh release create dev --prerelease dist/<name>-*`.

Require `fetch-depth: 0` and `fetch-tags: true` on checkout so the existing `dev` tag is visible to `git tag -f`.

### Deleting the existing dev release

The delete is existence-checked, and the check separates "there is no release" from "the API could not tell me":

```yaml
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
    GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

Three outcomes, three behaviours: the release exists, so delete it; `gh` reports it missing, so carry on; `gh` failed for any other reason, so stop. That third branch is the point of the whole step. A rate limit, an expired token or a flaky API all fail the `view`, and treating that as "no release exists" means the job skips the delete and then publishes on top of a release whose state it never established.

`2>&1 >/dev/null` captures stderr while discarding stdout, and the order matters. Redirections apply left to right, so this points stderr at the still-open capture and only then sends stdout to `/dev/null`. Writing `>/dev/null 2>&1` sends both to `/dev/null`, leaves `err` empty, and the not-found branch can never match.

Never write this as `gh release delete dev --yes --cleanup-tag || true`. That collapses all three outcomes into one and continues regardless. `--cleanup-tag` deletes the git tag along with the release, which step 2 then recreates at the new commit.
