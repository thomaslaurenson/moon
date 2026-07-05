# C++ Application Workflows

Applies only to applications that ship a distributable binary. A library builds
and tests in one plain job instead; see workflows-lib.md.

## Paths filter addition

Add `Dockerfile*` to the shared paths filter (see cpp/workflows.md).

## Build Step in Caller Workflows

Applications add a `build.yml` reusable workflow that compiles release binaries before lint and test can run. Callers must add a `build` job and `needs: build` on `lint` and `test`:

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
