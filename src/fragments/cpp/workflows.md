# C++ workflow conventions

Supplements the GitHub Actions fragment. Universal rules (runners, action versions, workflow structure, caller patterns, concurrency, permissions) apply unchanged. This file covers what's common to any C++ project's CI; the build and release pattern itself differs by tier (see workflows-app or workflows-lib).

## Paths filter

Use these entries in the `paths:` filter for `pr.yml` and `main.yml`:

```yaml
paths:
  - ".github/workflows/**"
  - "Makefile"
  - "CMakeLists.txt"
  - "cmake/**"
  - "include/**"
  - "src/**"
  - "app/**"
  - "test/**"
  - "extern/**"
  - ".clang-format"
  - ".clang-tidy"
```

`cmake/**` is not optional. Every project has that directory, because `version.h.in` lives there (see cpp/cmake.md), and it also holds helper modules such as `mark_system.cmake`. A filter that omits it lets a change to the version template or a helper module merge with no CI run at all.

Include `extern/**` only if the project uses git submodules for dependencies. Drop `include/**` in an application and `app/**` in a library; a path filter naming a directory the tier does not have is dead configuration that outlives the reason it was copied. A tier that ships a binary adds `Dockerfile*` to this list; see workflows-app.md.

The two failure modes are not symmetric. A stale entry is inert: it names a path that never changes, so it never triggers anything. A missing entry fails silently in the dangerous direction, letting a real change skip CI entirely, which is why a directory every project has belongs in the list rather than being left to each project to remember.

## Checkout

Always check out with `submodules: true`. C++ projects use git submodules for all dependencies and the build will fail without them:

`@vN` means pin the current major of the action at authoring time (for example `@v5`); Dependabot keeps the pin current. Do not copy a version number from this document as the target to match.

```yaml
- uses: actions/checkout@vN
  with:
    submodules: true
```

## Compilers

`test.yml` builds under both GCC and clang on Linux, with warnings as errors. They disagree at the same warning level, so a tree that is clean under one is not necessarily clean under the other, and whichever one a developer happens to have locally is the one CI adds nothing by repeating.

```yaml
jobs:
  test_linux:
    name: test_linux (${{ matrix.compiler }}, ${{ matrix.build_type }})
    strategy:
      # Report both independently. Cancelling one on the other's failure hides
      # half the findings on exactly the runs where they matter.
      fail-fast: false
      matrix:
        include:
          - compiler: gcc
            cxx: g++
            build_type: Debug
          # Release paired with one compiler rather than a third job: NDEBUG and
          # the optimiser are what ships, and nothing else exercises them.
          - compiler: clang
            cxx: clang++-18
            build_type: Release

    runs-on: ubuntu-24.04
    # BUILD_TYPE goes in the environment, not on the configure line. `build` and
    # every test target re-run `configure` through their prerequisite chain, and
    # an invocation without it would reset the cache to the Debug default, so the
    # Release entry would quietly test Debug and report green.
    env:
      BUILD_TYPE: ${{ matrix.build_type }}
    steps:
      - uses: actions/checkout@vN
        with:
          submodules: true

      - name: Install clang
        if: matrix.compiler == 'clang'
        run: sudo apt-get install -y clang-18

      - run: make configure CMAKE_ARGS="-DCMAKE_CXX_COMPILER=${{ matrix.cxx }} -D<PROJECT>_WERROR=ON"
      - run: make build
      - run: make test

  # A dedicated job, not a step in the one above: sanitizers change code
  # generation and roughly double runtime, so folding them into the main job
  # slows every run and leaves a failure ambiguous between the instrumented and
  # ordinary builds. `test_asan` configures build/asan itself and runs the unit
  # layer alone, which is the layer that needs no external data.
  test_asan:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@vN
        with:
          submodules: true

      - run: make test_asan
```

This is the whole of `test.yml` for a library. A tier that ships a binary adds `make test_functional` to `test_linux` and the native platform jobs; see workflows-app.md. Both tiers get `test_asan` unchanged.

`test_asan` is the CI job the sanitizer section of cpp/cmake.md asks for. It is a Linux job at the lowest billing rate and it runs on every trigger, because a memory error is exactly the kind of defect that is cheapest to find on the pull request that introduced it. A project whose unit layer is slow enough for the doubled runtime to matter can gate it to `main.yml` and `tag.yml`, at the cost of learning about the finding later.

Two compilers on one platform is a better use of a budget than one compiler on two platforms. Both run on Linux at the lowest billing rate, where a second platform costs two to ten times as much and mostly re-runs the same compiler. Reach for another platform when it is a deployment target, not for extra confidence in the code.

`<PROJECT>_WERROR` is off by default so a developer upgrading a compiler is not blocked by new warnings, and on in CI so those warnings are never merged; see cpp/cmake.md.

## Clang tools

Install the clang toolchain as a workflow step before running any lint step:

```yaml
- uses: actions/checkout@vN
  with:
    submodules: true

- name: Install clang tools
  run: sudo apt-get install -y clang-18 clang-format-18 clang-tidy-18
```

`clang-18` itself, not only the two tools: `make check_all` configures its own clang build directory so clang-tidy can resolve libstdc++ headers (see cpp/cmake.md), which needs the compiler present.

Installing the toolchain is a workflow step, not a Makefile target: it is specific to the runner image, and a `make` target doing it would fail on the macOS and Windows runners. Pin the major version here, so a runner image bump cannot silently change formatting output. CMake 3.21+ ships with `ubuntu-24.04`, so no CMake install step is needed.

## `lint.yml`

Installs the clang toolchain and runs format check and clang-tidy. No `make configure` step: `check_lint` depends on `configure_lint`, which produces the `compile_commands.json` clang-tidy reads (see cpp/cmake.md).

```yaml
name: Lint

on:
  workflow_call

permissions:
  contents: read

jobs:
  check_lint:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@vN
        with:
          submodules: true

      - name: Install clang tools
        run: sudo apt-get install -y clang-18 clang-format-18 clang-tidy-18

      - run: make check_all
      - run: make check_all
```

## Changelog extraction

Never extract the changelog with inline awk or bash in a workflow step. Use the `get_changelog` Makefile target instead. Pass the tag explicitly; the target writes the matching entry to stdout:

```yaml
- name: Extract release notes from CHANGELOG.md
  run: make get_changelog TAG="${GITHUB_REF_NAME}" > /tmp/release-notes.md
```

The target exits non-zero if TAG is empty or no matching entry is found, failing the release before it can publish with an empty changelog. `TAG` is `${GITHUB_REF_NAME}`, a `v`-prefixed tag (`v1.2.3`), but changelog headers are bare (`## 1.2.3 - ...`, see `github/changelog.md`), so the target strips the leading `v` from `TAG` before matching.
