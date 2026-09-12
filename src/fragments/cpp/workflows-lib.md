# C++ library workflows

Applies to libraries. A library isn't distributed as a prebuilt binary (consumers pull it in as a git submodule and compile it themselves), so CI is a single plain build-and-test job: no Docker, no libc/arch matrix, no separate build.yml. The release is in release-lib.md.

## Workflow set

Six files, not eight. There is no `build.yml`, because `test.yml` builds what it tests, and no `prerelease.yml`, because a library has no artefact to roll into one (see github/actions.md):

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

No `needs:` between `lint` and `test`. Neither consumes the other's output, so wiring them in series only delays the faster signal behind the slower one.

## `test.yml`

The job bodies are the shared `test_linux` matrix and `test_asan` from cpp/workflows.md, unchanged. A library needs nothing added to them: there is no functional layer to run against a binary, so `configure`, `build` and `test` is the whole of it.

A library that ships examples turns them on here, and nowhere else:

```yaml
      - run: |
          make configure CMAKE_ARGS="-DCMAKE_CXX_COMPILER=${{ matrix.cxx }} \
            -DMYPROJ_WERROR=ON -DMYPROJ_BUILD_EXAMPLES=ON"
```

`make build` then compiles them along with everything else, and nothing runs them. That is the whole of what cmake-lib.md asks for when it says to keep examples compiling: an example is the code a consumer copies, so one that no longer builds is worse than none, and the only way to notice is to build it. The option gates whether the targets exist rather than how they are built, so they belong in the same `build/dev` as everything else; see the build directory rules in cpp/cmake.md.

It costs one extra compile of a handful of small programs on a job that is already running, which is why this is a flag on the existing job rather than a job of its own.

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

If broader platform confidence is wanted later, add more runners to the `test.yml` job directly rather than reaching for the application's Docker/matrix pattern, which exists specifically for producing distributable binaries.
