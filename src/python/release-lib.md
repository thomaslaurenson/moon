# Python library release

How an installable library publishes a release. A library's release is a git tag; on that tag, CI builds the distribution and publishes it to PyPI, then creates a GitHub release with changelog notes. Scripts-only application projects do not publish and have no equivalent.

`@vN` in the examples below means pin the current major of the action at authoring time (for example `@v6`); Dependabot keeps the pin current. Do not copy a version number from this document as the target to match.

## Trusted publishing

Publish with PyPI trusted publishing (OIDC), never a long-lived API token in a secret. The workflow needs `id-token: write`; register the repository and workflow as a trusted publisher in the PyPI project settings first.

## `release.yml`

Reusable, called from `tag.yml` after lint and test pass. Two jobs: build the distribution, then publish and create the GitHub release.

```yaml
name: Release

on:
  workflow_call

jobs:
  build:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@vN
      - uses: astral-sh/setup-uv@vN
      - run: uv build
      - uses: actions/upload-artifact@vN
        with:
          name: dist
          path: dist/

  publish:
    needs: build
    runs-on: ubuntu-24.04
    permissions:
      contents: write        # create the GitHub release
      id-token: write        # PyPI trusted publishing (OIDC)
    steps:
      - uses: actions/checkout@vN
        with:
          fetch-depth: 0

      - uses: actions/download-artifact@vN
        with:
          name: dist
          path: dist/

      - uses: pypa/gh-action-pypi-publish@vN

      - name: Extract release notes from CHANGELOG.md
        run: make get_changelog TAG=${GITHUB_REF_NAME} > /tmp/release-notes.md

      - name: Create GitHub release
        run: |
          gh release create "${GITHUB_REF_NAME}" \
            dist/* \
            --title "${GITHUB_REF_NAME}" \
            --notes-file /tmp/release-notes.md
        env:
          GH_TOKEN: ${{ github.token }}
```

`fetch-depth: 0` is required so `get_changelog` can read the tagged history. The version comes from the tag via the build backend's tag-based versioning, or from `[project]` in `pyproject.toml`; either way, never inject it by hand.

## Caller wiring

`tag.yml` adds the release job after lint and test:

```yaml
jobs:
  lint:
    uses: ./.github/workflows/lint.yml
  test:
    uses: ./.github/workflows/test.yml
  release:
    needs: [lint, test]
    uses: ./.github/workflows/release.yml
    permissions:
      contents: write
      id-token: write
```

If the project is deliberately not published to PyPI, drop the `build`/`publish` split and the `id-token` permission, and make `release.yml` a GitHub-release-only workflow that attaches `uv build` output and changelog notes, mirroring the C++ library pattern. In that case also remove the release-downloads and PyPI-derived version badges from the badge row.
