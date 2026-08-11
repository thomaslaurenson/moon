# C++ library workflows

Applies to libraries. A library isn't distributed as a prebuilt binary (consumers pull it in as a git submodule and compile it themselves), so CI is a single plain build-and-test job: no Docker, no libc/arch matrix, no separate build.yml.

`@vN` in the examples below means pin the current major of the action at authoring time (for example `@v5`); Dependabot keeps the pin current. Do not copy a version number from this document as the target to match.

## Workflow set

Five files, not seven. There is no `build.yml`, because `test.yml` builds what it tests, and no `prerelease.yml`, because a library has no artifact to roll into one (see github/actions.md):

```
.github/workflows/
  lint.yml        # reusable
  test.yml        # reusable
  release.yml     # reusable
  pr.yml          # caller
  main.yml        # caller
  tag.yml         # caller
```

## Callers

`pr.yml` and `main.yml` are identical but for the trigger, the concurrency group and `cancel-in-progress`, so only `pr.yml` is shown. Both carry the `paths:` filter from cpp/workflows.md, and the two filters must match exactly.

```yaml
name: PR

on:
  pull_request:
    paths: # see cpp/workflows.md

concurrency:
  group: pr-${{ github.event.pull_request.number }}
  cancel-in-progress: true

permissions:
  contents: read

jobs:
  lint:
    uses: ./.github/workflows/lint.yml
  test:
    uses: ./.github/workflows/test.yml
```

```yaml
name: Tag

on:
  push:
    tags:
      - "v*"

permissions:
  contents: read

jobs:
  lint:
    uses: ./.github/workflows/lint.yml
  test:
    uses: ./.github/workflows/test.yml
  release:
    uses: ./.github/workflows/release.yml
    needs: [lint, test]
    permissions:
      contents: write
```

Declare `permissions: contents: read` at the top of every caller and widen it on the single job that needs more. Without a top-level block the caller inherits the repository default, which may be read and write; a caller that has never said what it needs is one setting away from handing write access to every job it composes.

No `needs:` between `lint` and `test`. Neither consumes the other's output, so wiring them in series only delays the faster signal behind the slower one.

## `test.yml`

The job body is the two-compiler matrix in cpp/workflows.md, unchanged. A library needs nothing added to it: there is no downloaded artifact to point the tests at and no functional layer to run against a binary, so `configure`, `build` and `test` in one job is the whole workflow.

```yaml
name: Test

on:
  workflow_call

permissions:
  contents: read

jobs:
  test_linux:
    # matrix, compiler install and steps: see cpp/workflows.md
```

That self-containment is why there is no `build.yml` here for anything to wait on.

## Releases

A library's release is a tagged commit; there is no compiled artifact to attach. `release.yml` creates a GitHub Release with changelog notes and nothing else:

```yaml
name: Release

on:
  workflow_call

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-24.04
    steps:
      # Default depth: get_changelog reads CHANGELOG.md from the working tree
      # and gh release create uses the API, so neither needs git history.
      - uses: actions/checkout@vN

      - name: Publish release
        run: |
          gh release create "${{ github.ref_name }}" \
            --title "${{ github.ref_name }}" \
            --notes "$(make get_changelog TAG=${{ github.ref_name }})"
        env:
          GITHUB_TOKEN: ${{ github.token }}
```

If broader platform confidence is wanted later, add more runners to the `test.yml` job directly rather than reaching for the application's Docker/matrix pattern, which exists specifically for producing distributable binaries.
