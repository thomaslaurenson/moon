# C++ release workflows

How a shipped binary becomes a published release. Applies to any tier that ships one: an application, or a library with a bundled CLI. A plain library releases a tagged commit and nothing else; see workflows-lib.

Assumes cpp/workflows-app, which owns the paths filter, the caller wiring and the `build.yml` that produces the artefacts these workflows consume. The gpipe fragment covers the config surface and the action inputs.

## `release.yml`

Publishes a GitHub release: downloads every build artefact, generates install scripts and checksums with gpipe, signs them, and creates the release with changelog notes.

This is the build -> gpipe -> release pattern; the gpipe fragment covers what gpipe writes and how it is configured. The C++ specific part is that the build step is `build.yml` rather than a single builder, so the binaries arrive as downloaded artefacts.

```yaml
name: Release

on:
  workflow_call:

# contents: write to create the release and upload assets
# id-token: write to obtain the OIDC token for cosign keyless signing
permissions:
  contents: write
  id-token: write

jobs:
  release:
    runs-on: ubuntu-24.04
    steps:
      # Default depth. Nothing here reads git history: get_changelog reads
      # CHANGELOG.md from the working tree, and gh release create uses the API.
      - uses: actions/checkout@vN

      - name: Download all build artefacts
        uses: actions/download-artifact@vN
        with:
          path: dist
          merge-multiple: true

      - name: Extract release notes from CHANGELOG.md
        run: make get_changelog TAG="${GITHUB_REF_NAME}" > /tmp/release-notes.md

      - uses: thomaslaurenson/gpipe@vN
        with:
          cosign_sign: true

      # Name every asset. A dist/myproj-* glob is shorter and wrong: with
      # merge-multiple every artefact lands flat in dist/, so the Docker image
      # tar matches too and is attached to the release as if it were a binary.
      - name: Create release
        run: |
          gh release create "${GITHUB_REF_NAME}" \
            dist/myproj-linux-x86_64 \
            dist/myproj-linux-aarch64 \
            dist/myproj-windows-x86_64.exe \
            install.sh install.ps1 \
            checksums.txt checksums.txt.sigstore.json \
            --title "${GITHUB_REF_NAME}" \
            --notes-file /tmp/release-notes.md
        env:
          GH_TOKEN: ${{ github.token }}

  # Gated on the release: a registry outage then leaves a complete release with
  # no image, rather than an image with no release. See tools/docker.
  release_docker:
    needs: release
    runs-on: ubuntu-24.04
    steps:
      - name: Download image artefact
        uses: actions/download-artifact@vN
        with:
          name: myproj-docker

      - name: Load image
        run: docker load -i myproj-docker.tar

      - name: Push to ghcr
        env:
          GH_TOKEN: ${{ github.token }}
          ACTOR: ${{ github.actor }}
          IMAGE: ghcr.io/${{ github.repository }}
        run: |
          echo "$GH_TOKEN" | docker login ghcr.io -u "$ACTOR" --password-stdin
          docker tag myproj "$IMAGE:${GITHUB_REF_NAME}"
          docker tag myproj "$IMAGE:latest"
          docker push "$IMAGE:${GITHUB_REF_NAME}"
          docker push "$IMAGE:latest"
```

A project publishing an image adds `packages: write` to this workflow's `permissions` and to the caller job in `tag.yml`.

The `release_docker` job is the Docker fragment's publishing pattern, repeated here rather than referenced so that `release.yml` is complete: the bytes pushed are the ones `build.yml` saved, the tags are the ones its table gives, and the job is separate and gated so a registry outage cannot leave a release half published.

### `.gpipe.yml`

The gpipe fragment covers the config surface and the action inputs. What is C++ specific is that the `path` entries must match where `download-artifact` puts the binaries: with `path: dist` and `merge-multiple: true` every artefact lands flat in `dist/`, so the paths are `./dist/<asset>`.

```yaml
binary: myproj

platforms:
  linux_amd64:
    path: ./dist/myproj-linux-x86_64
    name: myproj-linux-x86_64
  linux_arm64:
    path: ./dist/myproj-linux-aarch64
    name: myproj-linux-aarch64
  windows_amd64:
    path: ./dist/myproj-windows-x86_64.exe
    name: myproj-windows-x86_64.exe
```

One platform key maps to exactly one binary, and that is the other reason Linux ships a single static musl build per architecture: there is no way to express "glibc or musl, reader's choice", so the installer has to be given the one that runs everywhere.

## `prerelease.yml`

A single rolling GitHub prerelease under the literal tag `dev`, rebuilt on every push to main: raw binaries from `build.yml` only, no install scripts, no checksums, no changelog notes.

**gpipe does not appear here, and cannot**; see the gpipe fragment for why.

This needs no separate `git tag -f`/`git push --force` step: deleting the old release with `--cleanup-tag` removes its git tag too, so the following `gh release create dev --target <sha>` creates a fresh `dev` tag at the built commit on its own. Pass `--target ${{ github.sha }}` explicitly rather than letting `gh` default it to the current default-branch head, which can already have moved on by the time the job publishes.

Reuse the same `path: dist, merge-multiple: true` download and the same explicit asset list as `release.yml`. Naming them is deliberate in both places: release assets are a public interface, and a glob attaches whatever happens to match, which is how an image tar sharing the directory ends up published as a binary. Adding a platform is a decision, so let it be an edit.

Existence-check the delete: three outcomes, not two. Never write `gh release delete dev --yes --cleanup-tag || true`; that collapses "no dev release exists yet" and "the API could not tell me" into the same branch, and the job then publishes over a release state it never established.

```yaml
name: Prerelease

on:
  workflow_call:

permissions:
  contents: write

jobs:
  prerelease:
    runs-on: ubuntu-24.04
    # This job never checks out, so gh has no remote to infer the repository
    # from and GH_REPO has to name it. See github/actions.
    env:
      GH_TOKEN: ${{ github.token }}
      GH_REPO: ${{ github.repository }}
    steps:
      - name: Download all build artefacts
        uses: actions/download-artifact@vN
        with:
          path: dist
          merge-multiple: true

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

      - name: Create prerelease
        run: |
          gh release create dev --prerelease \
            --target "${GITHUB_SHA}" \
            --title "Dev (Pre-release)" \
            --notes "Built from commit ${GITHUB_SHA}" \
            dist/myproj-linux-x86_64 \
            dist/myproj-linux-aarch64 \
            dist/myproj-windows-x86_64.exe

  prerelease_docker:
    needs: prerelease
    runs-on: ubuntu-24.04
    steps:
      - name: Download image artefact
        uses: actions/download-artifact@vN
        with:
          name: myproj-docker

      - name: Load image
        run: docker load -i myproj-docker.tar

      # dev only, never latest: latest tracks releases, so pointing it at a
      # rolling build makes an untagged docker pull return whatever last landed
      # on the default branch.
      - name: Push dev tag to ghcr
        env:
          GH_TOKEN: ${{ github.token }}
          ACTOR: ${{ github.actor }}
          IMAGE: ghcr.io/${{ github.repository }}
        run: |
          echo "$GH_TOKEN" | docker login ghcr.io -u "$ACTOR" --password-stdin
          docker tag myproj "$IMAGE:dev"
          docker push "$IMAGE:dev"
```

As with `release.yml`, a project publishing an image adds `packages: write` to this workflow's `permissions` and to the `prerelease` job in `main.yml`.
