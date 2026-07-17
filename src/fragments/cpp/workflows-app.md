# C++ shipped-binary workflows

Applies to any tier that ships a distributable binary: an application, or a library with a bundled CLI. A plain library has no artifact to build and ships nothing, so it builds and tests in one plain job instead; see workflows-lib.md.

`@vN` in the examples below means pin the current major of the action at authoring time (for example `@v5`); Dependabot keeps the pin current. Do not copy a version number from this document as the target to match.

## Paths filter addition

Add `Dockerfile*` to the shared paths filter (see cpp/workflows.md).

## Build step in caller workflows

These tiers add a `build.yml` reusable workflow that compiles release binaries before lint and test can run. Callers must add a `build` job and `needs: build` on `lint` and `test`:

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
    permissions:
      contents: write
```

No `secrets: inherit`: every job here authenticates with the automatic `github.token`, so nothing needs to be forwarded. Add `secrets: inherit` only if a workflow genuinely reads a repository secret.

## Reusable workflow bodies

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
      - uses: actions/checkout@vN
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
        uses: actions/upload-artifact@vN
        with:
          name: myapp-linux-${{ matrix.arch }}-${{ matrix.libc }}
          path: myapp-linux-${{ matrix.arch }}-${{ matrix.libc }}
          retention-days: 1
```

### `test.yml`

Runs the test suite against the release binary produced by `build.yml`. The application binary under test is the downloaded release artifact, not a fresh local build; the test binaries themselves are compiled on the runner, because the artifact contains only the shipped executable, not the Catch2 test executables.

The functional tests spawn the release binary as a subprocess, so its path is injected at configure time via `MYAPP_BINARY_PATH_OVERRIDE` (see the tier fragment). Check out with submodules: the test binaries link Catch2 and `subprocess.h`, which are submodules. Each amd64 variant runs on the amd64 runner (a static musl binary runs fine on a glibc host); arm64 variants run on the arm64 runner.

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
      - uses: actions/checkout@vN
        with:
          submodules: true

      - name: Install clang tools
        run: make install_clang_tools

      - name: Download release binary (${{ matrix.arch }} ${{ matrix.libc }})
        uses: actions/download-artifact@vN
        with:
          name: myapp-linux-${{ matrix.arch }}-${{ matrix.libc }}
          path: artifact

      - name: Stage downloaded binary
        run: |
          mv artifact/myapp-linux-${{ matrix.arch }}-${{ matrix.libc }} ./myapp-under-test
          chmod +x ./myapp-under-test

      - name: Configure with the release binary as the functional-test target
        run: make configure CMAKE_ARGS="-DMYAPP_BINARY_PATH_OVERRIDE=${{ github.workspace }}/myapp-under-test"

      - name: Build test binaries
        run: make build

      - run: make test
```

### `release.yml`

Publishes a GitHub release: downloads every build artifact, generates checksums, and creates the release with changelog notes. Signing is intentionally not part of the C++ flow (the cosign/gpipe pipeline is Go-only per the GitHub Actions fragment); if artifact signing is later required, mirror the Go three-step release pattern.

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
      - uses: actions/checkout@vN
        with:
          fetch-depth: 0

      - name: Download all build artifacts
        uses: actions/download-artifact@vN
        with:
          path: dist
          merge-multiple: true

      - name: Generate checksums
        run: cd dist && sha256sum myapp-* > checksums.txt

      - name: Extract release notes from CHANGELOG.md
        run: make get_changelog TAG=${GITHUB_REF_NAME} > /tmp/release-notes.md

      - name: Create release
        run: |
          gh release create "${GITHUB_REF_NAME}" \
            dist/myapp-* dist/checksums.txt \
            --title "${GITHUB_REF_NAME}" \
            --notes-file /tmp/release-notes.md
        env:
          GH_TOKEN: ${{ github.token }}
```
