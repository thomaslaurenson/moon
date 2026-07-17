# C++ workflow conventions

Supplements the GitHub Actions fragment. Universal rules (runners, action versions, workflow structure, caller patterns, concurrency, permissions) apply unchanged. This file covers what's common to any C++ project's CI; the build and release pattern itself differs by tier (see workflows-app or workflows-lib).

## Paths filter

Use these entries in the `paths:` filter for `pr.yml` and `main.yml`:

```yaml
paths:
  - ".github/workflows/**"
  - "Makefile"
  - "include/**"
  - "src/**"
  - "app/**"
  - "test/**"
  - ".clang-format"
  - ".clang-tidy"
  - "CMakeLists.txt"
  - "extern/**"
```

Include `extern/**` only if the project uses git submodules for dependencies. Drop `include/**` in an application and `app/**` in a library; a path filter naming a directory the tier does not have is dead configuration that outlives the reason it was copied. A tier that ships a binary adds `Dockerfile*` to this list; see workflows-app.md.

## Checkout

Always check out with `submodules: true`. C++ projects use git submodules for all dependencies and the build will fail without them:

`@vN` means pin the current major of the action at authoring time (for example `@v5`); Dependabot keeps the pin current. Do not copy a version number from this document as the target to match.

```yaml
- uses: actions/checkout@vN
  with:
    submodules: true
```

## Clang tools

Install clang tools via the Makefile target before running any lint step:

```yaml
- uses: actions/checkout@vN
  with:
    submodules: true

- name: Install clang tools
  run: make install_clang_tools
```

`install_clang_tools` installs `clang-format-18` and `clang-tidy-18` at the pinned version. CMake 3.21+ ships with `ubuntu-24.04` so no CMake install step is needed.

## `lint.yml`

Installs clang tools and runs format check and clang-tidy. Requires `make configure` first so clang-tidy can resolve `compile_commands.json`.

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
      - uses: actions/checkout@vN
        with:
          submodules: true

      - name: Install clang tools
        run: make install_clang_tools

      - run: make configure
      - run: make fmt_check
      - run: make lint_cpp
```

## Changelog extraction

Never extract the changelog with inline awk or bash in a workflow step. Use the `get_changelog` Makefile target instead. Pass the tag explicitly; the target writes the matching entry to stdout:

```yaml
- name: Extract release notes from CHANGELOG.md
  run: make get_changelog TAG=${GITHUB_REF_NAME} > /tmp/release-notes.md
```

The target exits non-zero if TAG is empty or no matching entry is found, failing the release before it can publish with an empty changelog.
