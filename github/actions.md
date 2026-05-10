# GitHub Actions Workflows

## Core Rules

- Runner: always `ubuntu-24.04`. Never `ubuntu-latest` or any unversioned alias.
- Always prefer official GitHub actions (`actions/*`) before third-party alternatives.
- Always use the `gh` CLI for creating releases. Never use a third-party release action.
- All workflow steps call `make <target>`. Never run raw commands directly in a workflow.
- Never write raw bash or awk scripts in workflows to extract versions or changelogs. Use
  the repository's Makefile targets (e.g. `make get_changelog_entry`) and pass their
  standard output to the respective workflow steps.
- Minimal permissions: `contents: read` by default, `contents: write` only in release and
  prerelease workflows. Test workflows must never declare `contents: write`.
- No `fetch-depth: 0` except in release workflows where changelog extraction requires it.

---

## Action Versions

Always use these pinned versions. Never use `@latest` or unversioned aliases:

| Action | Version |
|---|---|
| `actions/checkout` | `v6` |
| `actions/upload-artifact` | `v6` |
| `actions/download-artifact` | `v6` |
| `actions/setup-go` | `v6` |
| `actions/setup-python` | `v5` |
| `astral-sh/setup-uv` | `v7` |

---

## Workflow Structure

Use reusable workflows (`workflow_call`) for all job logic. Caller workflows compose them.

```
.github/workflows/
  build.yml       # reusable: build release binaries (Docker-based for C++)
  lint.yml        # reusable: linting and format checks
  test.yml        # reusable: run tests against pre-built binary artifact
  release.yml     # reusable: create GitHub release from artifacts
  prerelease.yml  # reusable: create or replace rolling dev prerelease
  pr.yml          # caller: lint + test on pull requests
  tag.yml         # caller: build + lint + test + release on v* tags
  main.yml        # caller: build + lint + test + prerelease on push to main
```

No `push.yml` for any project. Use `pr.yml` for pull requests, `tag.yml` for versioned
releases, and `main.yml` for rolling dev prereleases.

Concurrency on `pr.yml`: always add a concurrency group to cancel stale runs when new
commits are pushed to a PR:

```yaml
concurrency:
  group: pr-${{ github.event.pull_request.number }}
  cancel-in-progress: true
```

Concurrency on `main.yml`: always add a concurrency group to cancel in-progress prerelease
runs when new commits land on main:

```yaml
concurrency:
  group: prerelease
  cancel-in-progress: true
```

---

## Caller Workflow: pr.yml

Triggers on pull requests. Runs lint and test in parallel. Uses a `paths:` filter so the
workflow only fires on relevant changes.

```yaml
name: Pull Request

on:
  pull_request:
    paths:
      - ".github/workflows/**"
      - "src/**"
      - "test/**"
      - "CMakeLists.txt"
      - "Makefile"

concurrency:
  group: pr-${{ github.event.pull_request.number }}
  cancel-in-progress: true

permissions:
  contents: read

jobs:
  build:
    uses: ./.github/workflows/build.yml
  lint:
    uses: ./.github/workflows/lint.yml
    needs: build
  test:
    uses: ./.github/workflows/test.yml
    needs: build
```

### Common paths entries by language

Include only the entries that apply to the project. Omit entries for paths that do not exist:

| Entry | When to include |
|---|---|
| `.github/workflows/**` | Always -- workflow changes must trigger CI |
| `src/**` | Always -- application source |
| `test/**` | Always -- test source |
| `Makefile` | Always -- Makefile changes affect all targets |
| `.clang-format` | C++ projects |
| `.clang-tidy` | C++ projects |
| `extern/**` | C++ projects with submodule dependencies |
| `CMakeLists.txt` | C++ projects |
| `Dockerfile*` | C++ projects that build via Docker |
| `go.mod` | Go projects |
| `go.sum` | Go projects |
| `pyproject.toml` | Python projects |

The `paths:` filter on `main.yml` must match `pr.yml` exactly -- the same changes that
trigger a PR check should trigger a prerelease when merged.

---

## Caller Workflow: tag.yml

Triggers on `v*` tags. Runs build first, then lint and test in parallel, then release.
Release only runs after all three pass.

```yaml
name: Tag and Release

on:
  push:
    tags:
      - "v*.*.*"

permissions:
  contents: write
  packages: write

jobs:
  build:
    uses: ./.github/workflows/build.yml
  lint:
    uses: ./.github/workflows/lint.yml
    needs: build
  test:
    uses: ./.github/workflows/test.yml
    needs: build
  release:
    uses: ./.github/workflows/release.yml
    needs: [build, lint, test]
    secrets: inherit
```

- `permissions: contents: write` and `packages: write` are declared at the caller level.
- `secrets: inherit` passes `GITHUB_TOKEN` to the release workflow.

---

## Caller Workflow: main.yml

Triggers on pushes to `main`. Mirrors `tag.yml` but calls `prerelease.yml` instead of
`release.yml`. Creates or replaces a rolling `dev` prerelease on every passing main build.

```yaml
name: Main

on:
  push:
    branches:
      - "main"
    paths:
      - ".github/workflows/**"
      - "src/**"
      - "test/**"
      - "CMakeLists.txt"
      - "Makefile"

concurrency:
  group: prerelease
  cancel-in-progress: true

permissions:
  contents: write
  packages: write

jobs:
  build:
    uses: ./.github/workflows/build.yml
  lint:
    uses: ./.github/workflows/lint.yml
    needs: build
  test:
    uses: ./.github/workflows/test.yml
    needs: build
  prerelease:
    uses: ./.github/workflows/prerelease.yml
    needs: [build, lint, test]
    secrets: inherit
```

---

## Prerelease Workflow: prerelease.yml

Creates or replaces a rolling `dev` release. Always deletes the existing `dev` release and
tag before recreating it so the entry stays current without accumulating stale releases.

```yaml
name: Prerelease

on:
  workflow_call

permissions:
  contents: write
  packages: write

jobs:
  prerelease_binaries:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6

      - name: Download artifacts
        # download-artifact steps per project, matching build.yml output names

      - name: Delete existing dev release
        run: gh release delete dev --yes --cleanup-tag || true
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Create prerelease
        run: |
          gh release create dev \
            --title "Dev (Pre-release)" \
            --prerelease \
            --notes "**Commit:** ${{ github.sha }}" \
            <artifact files>
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- The `|| true` on the delete step is intentional -- it prevents failure when no `dev`
  release exists yet (first run).
- The prerelease notes contain only the commit SHA. Full changelog entries are for
  versioned releases only.

---

## Go-Specific Setup

Add these steps before any `make` call in Go workflows:

```yaml
- uses: actions/checkout@v6

- uses: actions/setup-go@v6
  with:
    go-version-file: go.mod
    cache: true
```

- Always use `go-version-file: go.mod`. Never hardcode a Go version.
- `fetch-depth: 0` only in `release.yml`, not in lint or test workflows.

---

## Python-Specific Setup

Extract the Python version from `pyproject.toml` at runtime. Never hardcode it.

```yaml
- uses: actions/checkout@v6

- name: Extract Python version
  id: python-version
  run: echo "version=$(make get_python_required_version)" >> $GITHUB_OUTPUT

- uses: actions/setup-python@v5
  with:
    python-version: ${{ steps.python-version.outputs.version }}

- uses: astral-sh/setup-uv@v7
```

For lint-only jobs, extract the ruff version and use the ruff action directly -- no Python
setup required:

```yaml
- name: Extract ruff version
  id: ruff-version
  run: echo "version=$(grep -oP 'ruff>=\K[0-9.]+' pyproject.toml)" >> $GITHUB_OUTPUT

- uses: chartboost/ruff-action@v1
  with:
    version: ${{ steps.ruff-version.outputs.version }}
    args: check .
```

---

## C++-Specific Setup

C++ workflows install clang tools via the Makefile and rely on CMake being available on
the runner. No language-specific GitHub Action is needed.

### Checkout

Always check out with `submodules: true`. C++ projects use git submodules for all
dependencies and the build will fail without them:

```yaml
- uses: actions/checkout@v6
  with:
    submodules: true
```

Test workflows that only run a pre-built binary artifact do not need submodules and should
use `submodules: false` to keep checkout fast.

### Clang tools

Install clang tools via the Makefile target before running any lint step:

```yaml
- uses: actions/checkout@v6
  with:
    submodules: true

- name: Install clang tools
  run: make install_clang_tools
```

`install_clang_tools` installs `clang-format-19` and `clang-tidy-19` at the pinned
version. CMake 3.21+ ships with `ubuntu-24.04` so no CMake install step is needed.

### Build

C++ projects build release binaries via Docker. The build workflow uses a matrix across
target architectures and libc variants. Binaries are extracted from the container and
uploaded as artifacts for consumption by the test and release workflows:

```yaml
name: Build

on:
  workflow_call

jobs:
  build_linux:
    strategy:
      matrix:
        include:
          - arch: amd64
            libc: glibc
            runner: ubuntu-24.04
            dockerfile: Dockerfile.glibc
            binary_path: /usr/local/bin/myapp
          - arch: amd64
            libc: musl
            runner: ubuntu-24.04
            dockerfile: Dockerfile.musl
            binary_path: /myapp
          - arch: arm64
            libc: glibc
            runner: ubuntu-24.04-arm
            dockerfile: Dockerfile.glibc
            binary_path: /usr/local/bin/myapp
          - arch: arm64
            libc: musl
            runner: ubuntu-24.04-arm
            dockerfile: Dockerfile.musl
            binary_path: /myapp

    runs-on: ${{ matrix.runner }}
    steps:
      - uses: actions/checkout@v6
        with:
          submodules: true

      - name: Build Docker image (${{ matrix.arch }} ${{ matrix.libc }})
        run: |
          docker build --platform linux/${{ matrix.arch }} \
            -t myapp-${{ matrix.arch }}-${{ matrix.libc }} \
            -f ${{ matrix.dockerfile }} .

      - name: Extract binary from Docker image
        run: |
          CONTAINER_ID=$(docker create myapp-${{ matrix.arch }}-${{ matrix.libc }})
          docker cp $CONTAINER_ID:${{ matrix.binary_path }} \
            ./myapp-linux-${{ matrix.arch }}-${{ matrix.libc }}
          docker rm $CONTAINER_ID

      - name: Upload binary as artifact
        uses: actions/upload-artifact@v6
        with:
          name: myapp-linux-${{ matrix.arch }}-${{ matrix.libc }}
          path: myapp-linux-${{ matrix.arch }}-${{ matrix.libc }}
          retention-days: 1
```

### Test

The test workflow downloads the pre-built artifact and runs ctest against it. It never
rebuilds from source. Submodules are not needed:

```yaml
name: Test

on:
  workflow_call

permissions:
  contents: read

jobs:
  test_linux:
    strategy:
      matrix:
        include:
          - arch: amd64
            libc: glibc
            runner: ubuntu-24.04
          - arch: amd64
            libc: musl
            runner: ubuntu-24.04
          - arch: arm64
            libc: glibc
            runner: ubuntu-24.04-arm
          - arch: arm64
            libc: musl
            runner: ubuntu-24.04-arm

    runs-on: ${{ matrix.runner }}
    steps:
      - uses: actions/checkout@v6
        with:
          submodules: false

      - name: Download binary (${{ matrix.arch }} ${{ matrix.libc }})
        uses: actions/download-artifact@v6
        with:
          name: myapp-linux-${{ matrix.arch }}-${{ matrix.libc }}
          path: build/bin

      - name: Make binary executable
        run: chmod +x build/bin/myapp-linux-${{ matrix.arch }}-${{ matrix.libc }}

      - name: Rename binary
        run: mv build/bin/myapp-linux-${{ matrix.arch }}-${{ matrix.libc }} build/bin/myapp

      - name: Run tests
        run: make test
```

### Lint

The lint workflow installs clang tools and runs format check and clang-tidy. Submodules
are required because clang-tidy resolves headers from dependencies:

```yaml
name: Lint

on:
  workflow_call

permissions:
  contents: read

jobs:
  lint_cpp:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6
        with:
          submodules: true

      - name: Install clang tools
        run: make install_clang_tools

      - name: Check formatting
        run: make format_check

      - name: Run clang-tidy
        run: make lint_cpp
```

### Changelog extraction

Never extract the changelog with inline awk or bash in a workflow step. Use the
`get_changelog_entry` Makefile target instead. The target reads the version from
`CMakeLists.txt` and writes the matching entry to `/tmp/release_notes.md`:

```yaml
- name: Extract changelog entry
  id: changelog
  run: |
    make get_changelog_entry
    {
      echo "content<<EOF"
      cat /tmp/release_notes.md
      echo "EOF"
    } >> $GITHUB_OUTPUT
```

See `tools/makefile.md` for the `get_changelog_entry` target definition for C++ projects.
The target exits non-zero if no entry is found for the current version, failing the
release before it can create an empty changelog release.
