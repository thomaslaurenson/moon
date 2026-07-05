# C++ Library Workflows

Applies to libraries. A library isn't distributed as a prebuilt binary (consumers
pull it in as a git submodule and compile it themselves), so CI is a single plain
build-and-test job: no Docker, no libc/arch matrix, no separate build.yml.

## `test.yml`

```yaml
name: Test

on:
  workflow_call

permissions:
  contents: read

jobs:
  test:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6
        with:
          submodules: true

      - run: make configure
      - run: make build
      - run: make test
```

No `needs: build` wiring in the caller: this workflow configures, builds, and tests in one job, unlike an application's separate `build.yml`.

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
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: Publish release
        run: |
          gh release create "${{ github.ref_name }}" \
            --title "${{ github.ref_name }}" \
            --notes "$(make get_changelog TAG=${{ github.ref_name }})"
        env:
          GITHUB_TOKEN: ${{ github.token }}
```

If broader platform confidence is wanted later, add more runners to the `test.yml` job directly rather than reaching for the application's Docker/matrix pattern, which exists specifically for producing distributable binaries.
