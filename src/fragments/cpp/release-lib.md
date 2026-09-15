# C++ library release workflow

How a library is released. Applies to a plain library: a tagged commit with changelog notes and no artefact, since consumers take a git ref and compile it themselves. A tier that ships a binary uses release-app instead.

Assumes cpp/workflows-lib, which owns the caller wiring that runs this workflow on a `v*` tag.

## `release.yml`

A library's release is a tagged commit; there is no compiled artefact to attach. `release.yml` creates a GitHub Release with changelog notes and nothing else:

```yaml
name: Release

on:
  workflow_call:

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-24.04
    steps:
      # Default depth: get_changelog reads CHANGELOG.md from the working tree
      # and gh release create uses the API, so neither needs git history.
      - uses: actions/checkout@vN

      - name: Extract release notes from CHANGELOG.md
        run: make get_changelog TAG="${GITHUB_REF_NAME}" > /tmp/release-notes.md

      - name: Publish release
        run: |
          gh release create "${GITHUB_REF_NAME}" \
            --title "${GITHUB_REF_NAME}" \
            --notes-file /tmp/release-notes.md
        env:
          GH_TOKEN: ${{ github.token }}
```

Two details match the application flow rather than diverging from it, and both are worth keeping in step. The notes go through a file and `--notes-file`, never `--notes "$(...)"`: command substitution strips trailing newlines and re-splits the changelog through the shell, so an entry containing a backtick or a `$` is mangled or executed. And the token is `GH_TOKEN`, the name `gh` documents; `GITHUB_TOKEN` also works today, which is exactly why a project ends up with both spellings in different workflows and nobody can say which is required. See github/actions.

## No prerelease

There is no `prerelease.yml` and no rolling `dev` release. A library has no artefact to roll, and a consumer wanting the tip of `main` points a submodule at it; `main.yml` is lint and test alone. See github/actions for the rule and cpp/workflows-lib for the file set.
