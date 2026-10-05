# C++ library workflows

Applies to libraries. A library isn't distributed as a prebuilt binary (consumers pull it in as a git submodule and compile it themselves), so CI is plain build-and-test jobs: no Docker, no libc matrix, no separate build.yml. The only jobs beyond those in cpp/workflows are for an architecture the library is deployed to; see the end of this fragment. The release is in release-lib.

## Workflow set

Six files, not eight. There is no `build.yml`, because `test.yml` builds what it tests, and no `prerelease.yml`, because a library has no artefact to roll into one (see github/actions):

```text
.github/workflows/
  lint.yml        # reusable
  test.yml        # reusable
  release.yml     # reusable
  pr.yml          # caller
  main.yml        # caller
  tag.yml         # caller
```

## Callers

`pr.yml` and `main.yml` are identical but for the trigger, the concurrency group and `cancel-in-progress`, so only `pr.yml` is shown. Both carry the `paths:` filter from cpp/workflows, and the two filters must match exactly.

```yaml
name: PR

on:
  pull_request:
    paths: # see cpp/workflows

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

No `needs:` between `lint` and `test`. Neither consumes the other's output, so wiring them in series only delays the faster signal behind the slower one.

## `test.yml`

The job bodies are the shared `test_linux` matrix and `test_asan` from cpp/workflows, unchanged. A library needs nothing added to them: there is no functional layer to run against a binary, so `configure`, `build` and `test` is the whole of it.

A library that ships examples turns them on here, and nowhere else:

```yaml
      - run: |
          make configure CMAKE_ARGS="-DCMAKE_CXX_COMPILER=${{ matrix.cxx }} \
            -DMYPROJ_WERROR=ON -DMYPROJ_BUILD_EXAMPLES=ON"
```

`make build` then compiles them along with everything else, and nothing runs them. That is the whole of what cmake-lib asks for when it says to keep examples compiling: an example is the code a consumer copies, so one that no longer builds is worse than none, and the only way to notice is to build it. The option gates whether the targets exist rather than how they are built, so they belong in the same `build/dev` as everything else; see the build directory rules in cpp/cmake.

It costs one extra compile of a handful of small programs on a job that is already running, which is why this is a flag on the existing job rather than a job of its own.

```yaml
name: Test

on:
  workflow_call:

permissions:
  contents: read

jobs:
  test_linux:
    # matrix, compiler install and steps: see cpp/workflows
```

That self-containment is why there is no `build.yml` here for anything to wait on.

A library never reaches for the application's Docker/matrix pattern, which exists to produce distributable binaries. A library whose consumer runs on another architecture tests there too, because that is a deployment target rather than extra confidence (see github/actions). A library linked into a 32-bit process gets a `test_32` job running `make test_32 CMAKE_ARGS="-DMYPROJ_WERROR=ON"`, after installing `g++-multilib`. One linked into a 32-bit Windows binary built with MinGW gets a `test_mingw` job running `make test_mingw` with the same `CMAKE_ARGS`, which builds into `build/mingw` with the MinGW toolchain and runs the tests under Wine. That job runs `dpkg --add-architecture i386` and `apt-get update`, installs `g++-mingw-w64-i686`, `wine` and `wine32:i386`, sets `WINEDEBUG=-all`, and runs `wineboot` once before the tests so the first test does not pay for Wine's start-up. Each is a separate job in `test.yml`, not a matrix entry, because its steps differ.
