# Python library release

How an installable library publishes a release. A library's release is a git tag; on that tag, CI builds the distribution, then creates a GitHub release with the changelog notes and the built artifacts attached. Publishing to PyPI is an optional add-on layered on top of this baseline (see the optional section below), not a requirement. Scripts-only application projects do not release and have no equivalent.

`@vN` in the examples below means pin the current major of the action at authoring time (for example `@v6`); Dependabot keeps the pin current. Do not copy a version number from this document as the target to match.

## `release.yml` (baseline: GitHub release)

Reusable, called from `tag.yml` after lint and test pass. A single job builds the distribution and creates the GitHub release, using the changelog section for the tag as the release notes and attaching the build output.

```yaml
name: Release

on:
  workflow_call

jobs:
  release:
    runs-on: ubuntu-24.04
    permissions:
      contents: write        # create the GitHub release
    steps:
      - uses: actions/checkout@vN
        with:
          fetch-depth: 0

      - uses: astral-sh/setup-uv@vN

      - run: uv build

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

`fetch-depth: 0` is required so `get_changelog` can read the tagged history. `get_changelog` strips the leading `v` from the tag before matching the bare changelog header (see `python/make.md` and `github/changelog.md`). The version comes from the tag via the build backend's tag-based versioning, or from `[project]` in `pyproject.toml`; either way, never inject it by hand.

## Caller wiring

`tag.yml` adds the release job after lint and test. The baseline needs only `contents: write`:

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
```

## Optional: publish to PyPI (trusted publishing)

Publishing to PyPI is opt-in. A library consumed only from git or a private index does not need it, and by default the badge row uses the static Python badge (see `python/badges.md`) rather than any PyPI-derived badge.

To publish, first register the repository and workflow as a trusted publisher in the PyPI project settings, then use PyPI trusted publishing (OIDC) never a long-lived API token in a secret. Restructure `release.yml` into two jobs:

- `build`: run `uv build`, then upload `dist/` with `actions/upload-artifact@vN`.
- `publish` (`needs: build`): download the artifact with `actions/download-artifact@vN` into `dist/`, run `pypa/gh-action-pypi-publish@vN`, then the same checkout (`fetch-depth: 0`), `get_changelog`, and `gh release create dist/*` steps as the baseline job. Grant this job `id-token: write` alongside `contents: write`.

The caller (`tag.yml`) must then also grant `id-token: write` on the `release` job. When publishing to PyPI, you may switch the Python-version badge to the live PyPI badge (see `python/badges.md`).
