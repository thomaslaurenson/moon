# C++ Workflow Conventions

Supplements `github/actions.md`. Universal rules (runners, action versions, workflow structure, caller patterns, concurrency, permissions) apply unchanged. This file covers C++-specific workflow patterns only.

---

## Paths Filter

Use these entries in the `paths:` filter for `pr.yml` and `main.yml`:

```yaml
paths:
  - ".github/workflows/**"
  - "Makefile"
  - "src/**"
  - "test/**"
  - ".clang-format"
  - ".clang-tidy"
  - "CMakeLists.txt"
  - "Dockerfile*"
  - "extern/**"
```

Include `extern/**` only if the project uses git submodules for dependencies.

---

## Build Step in Caller Workflows

C++ projects add a `build.yml` reusable workflow that compiles binaries before lint and test can run. Callers must add a `build` job and `needs: build` on `lint` and `test`:

```yaml
jobs:
  build:
    uses: ./.github/workflows/build.yml
  lint:
    uses: ./.github/workflows/lint.yml
    needs: build
  test:
    uses: ./.github/workflows/test.yml
    needs: build
  release:                              # tag.yml only
    uses: ./.github/workflows/release.yml
    needs: [build, lint, test]
    secrets: inherit
```

---

## Checkout

Always check out with `submodules: true`. C++ projects use git submodules for all dependencies and the build will fail without them:

```yaml
- uses: actions/checkout@v6
  with:
    submodules: true
```

Test workflows that only run a pre-built binary artifact do not need submodules; use `submodules: false` to keep checkout fast.

---

## Clang Tools

Install clang tools via the Makefile target before running any lint step:

```yaml
- uses: actions/checkout@v6
  with:
    submodules: true

- name: Install clang tools
  run: make install_clang_tools
```

`install_clang_tools` installs `clang-format-18` and `clang-tidy-18` at the pinned version. CMake 3.21+ ships with `ubuntu-24.04` so no CMake install step is needed.

---

## Reusable Workflow Bodies

### `build.yml`

Builds release binaries via Docker. Uses a matrix across target architectures and libc variants. Binaries are extracted from the container and uploaded as artifacts for consumption by the test and release workflows.

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

### `test.yml`

Downloads the pre-built artifact and runs tests against it. Never rebuilds from source. Submodules are not needed.

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

      - run: make test
```

### `lint.yml`

Installs clang tools and runs format check and clang-tidy. Submodules are required because clang-tidy resolves headers from dependencies.

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

      - run: make fmt_check
      - run: make lint_cpp
```

---

## Changelog Extraction

Never extract the changelog with inline awk or bash in a workflow step. Use the `get_changelog` Makefile target instead. Pass the tag explicitly; the target writes the matching entry to stdout:

```yaml
- name: Extract release notes from CHANGELOG.md
  run: make get_changelog TAG=${GITHUB_REF_NAME} > /tmp/release-notes.md
```

The target exits non-zero if TAG is empty or no matching entry is found, failing the release before it can publish with an empty changelog.
