# C++ shipped-binary workflows

Applies to any tier that ships a distributable binary: an application, or a library with a bundled CLI. A plain library has no artifact to build and ships nothing, so it builds and tests in one plain job instead; see workflows-lib.md.

`@vN` in the examples below means pin the current major at authoring time (for example `@v5`); Dependabot keeps it current. Do not copy a version number from this document as the target to match. Majors are for `actions/*` and for actions you publish yourself; every other action pins to a full SHA (see the GitHub Actions fragment).

## Paths filter addition

Add `Dockerfile*` to the shared paths filter (see cpp/workflows.md), plus `.gpipe.yml` where the release uses gpipe (see `release.yml` below).

## Depth is a function of the trigger

These tiers add a `build.yml` that produces the distributable artifacts. Running the whole of it on every pull request is the expensive default and buys the least: a pull request needs to know the code compiles and the tests pass, not that a shippable artifact for every platform is correct. A tag needs the second, and only a tag can act on it.

So the caller decides the depth, and `build.yml` takes an input:

| Trigger | Runs | Why |
|---|---|---|
| `pr.yml` | lint, test, and the artifact jobs that a test cannot cover | Cheapest signal that catches a real break |
| `main.yml` | the above plus the full artifact set and `prerelease` | The rolling channel has to contain everything a release would |
| `tag.yml` | everything, plus `release` | This is the run whose output people install |

"An artifact job a test cannot cover" is a narrow set. A container image built from the same Dockerfile the release uses is one: nothing else exercises that file, and a break in it is invisible until release day. A second compile of a platform the test job already compiles is not: it is the same sources through the same compiler, at that platform's billing rate, for no new information.

```yaml
# build.yml
on:
  workflow_call:
    inputs:
      artifacts:
        description: Build the full release artifact set, not just the cheap subset
        type: boolean
        default: false
```

```yaml
# pr.yml and main.yml
jobs:
  lint:
    uses: ./.github/workflows/lint.yml
  test:
    uses: ./.github/workflows/test.yml
  build:
    uses: ./.github/workflows/build.yml
    with:
      artifacts: true                   # main.yml and tag.yml only; omit on pr.yml
  release:                              # tag.yml only
    uses: ./.github/workflows/release.yml
    needs: [build, lint, test]
    permissions:
      contents: write
      id-token: write                   # cosign keyless signing, see release.yml
  prerelease:                           # main.yml only
    uses: ./.github/workflows/prerelease.yml
    needs: [build, lint, test]
    permissions:
      contents: write
```

**No `needs: build` on `lint` or `test`.** Neither consumes a build artifact: lint configures its own tree, and test builds the binaries it runs. Wiring them behind `build` holds the fastest signal in the pipeline behind the slowest job and buys nothing. Only `release` and `prerelease` genuinely need the artifacts, and they say so.

Declare `permissions: contents: read` at the top of every caller and widen it on the jobs that need more. A caller with no top-level block inherits the repository default, which may be read and write.

No `secrets: inherit`: every job here authenticates with the automatic `github.token`, so nothing needs to be forwarded. Add `secrets: inherit` only if a workflow genuinely reads a repository secret.

`id-token: write` has to appear on the caller's `release` job as well as inside `release.yml`. A reusable workflow cannot widen its own permissions beyond what the caller grants, so declaring it only in the callee fails at signing time with an OIDC token request error.

## Reusable workflow bodies

### `build.yml`

Builds one release binary per target platform and uploads each as an artifact for the test, release and prerelease workflows to consume. Linux builds go through Docker; macOS and Windows build natively on their own runners, because there is no container route to either.

**Which platforms is a decision, not a default.** The jobs below are the full set; a project takes the ones matching where its users actually run the binary. macOS in particular is opt-in: it costs ten times a Linux job on a private repository and proves nothing about a tool nobody runs there. The same test applies as in cpp/workflows.md, that a platform earns a job by being a deployment target rather than by adding confidence.

Once a platform is in, it stays consistent all the way through: a job in `build.yml`, a matching entry in `test.yml`, an asset in `.gpipe.yml`, and a line in the release asset list. A platform built but not tested ships an artifact nothing has run.

#### Asset naming

Name every artifact `<app>-<os>-<arch>`, using `x86_64`/`aarch64` rather than `amd64`/`arm64`, and append `.exe` on Windows. The table is the naming for whichever platforms a project builds, not a list of platforms it must:

| Platform | Asset |
|---|---|
| Linux x86_64 | `myapp-linux-x86_64` |
| Linux ARM64 | `myapp-linux-aarch64` |
| macOS x86_64 | `myapp-darwin-x86_64` |
| macOS ARM64 | `myapp-darwin-aarch64` |
| Windows x86_64 | `myapp-windows-x86_64.exe` |

This is gpipe's platform vocabulary, so the names map onto `.gpipe.yml` with no translation; the identifiers are the `platforms` keys shown under `release.yml` below, and `gpipe validate` checks a config against them. Pick the naming before the first release: the assets are a public interface, and renaming them later breaks anyone's install script.

#### One Linux binary, not two

Build Linux as a **single statically linked musl binary per architecture**, not a glibc/musl pair. A static musl binary runs on every distribution with no libc version floor; a glibc binary built on `ubuntu-24.04` needs glibc >= ~2.39 and will not start on RHEL 8/9 or Debian 11/12. Shipping both is not "portable plus compatible", it is portable plus a strictly narrower duplicate, and gpipe can only point the installer at one of them anyway.

The usual objection is musl's slower allocator. Measure before believing it applies: for a tool that allocates a few large buffers up front and then computes, the difference is single-digit percent. Static musl is the default; deviate only with a benchmark showing it costs something real.

```yaml
name: Build

on:
  workflow_call

permissions:
  contents: read

jobs:
  build_linux:
    strategy:
      matrix:
        include:
          - arch: amd64
            asset: myapp-linux-x86_64
            runner: ubuntu-24.04
          - arch: arm64
            asset: myapp-linux-aarch64
            runner: ubuntu-24.04-arm

    runs-on: ${{ matrix.runner }}
    steps:
      - uses: actions/checkout@vN
        with:
          submodules: true

      - name: Build Docker image (${{ matrix.arch }})
        run: |
          docker build --platform linux/${{ matrix.arch }} \
            -t myapp-${{ matrix.arch }} \
            -f Dockerfile.musl .

      - name: Extract binary from Docker image
        run: |
          CONTAINER_ID=$(docker create myapp-${{ matrix.arch }})
          docker cp "$CONTAINER_ID":/myapp ./${{ matrix.asset }}
          docker rm "$CONTAINER_ID"

      - name: Upload binary as artifact
        uses: actions/upload-artifact@vN
        with:
          name: ${{ matrix.asset }}
          path: ${{ matrix.asset }}
          retention-days: 1

  build_macos:
    strategy:
      matrix:
        include:
          - runner: macos-15
            asset: myapp-darwin-aarch64
          - runner: macos-15-intel
            asset: myapp-darwin-x86_64

    runs-on: ${{ matrix.runner }}
    steps:
      - uses: actions/checkout@vN
        with:
          submodules: true

      - name: Build
        run: |
          cmake -B build/release \
            -DCMAKE_BUILD_TYPE=Release \
            -DMYAPP_BUILD_TESTING=OFF \
            -DCMAKE_OSX_DEPLOYMENT_TARGET=11.0
          cmake --build build/release --config Release --parallel 3
          strip build/release/bin/myapp
          # strip invalidates the linker's ad-hoc signature; re-sign or the
          # binary is killed on launch on Apple Silicon. See below.
          codesign --force --sign - build/release/bin/myapp
          codesign --verify --verbose build/release/bin/myapp
          mv build/release/bin/myapp ./${{ matrix.asset }}

      - name: Upload binary as artifact
        uses: actions/upload-artifact@vN
        with:
          name: ${{ matrix.asset }}
          path: ${{ matrix.asset }}
          retention-days: 1

  build_windows:
    runs-on: windows-2022
    steps:
      - uses: actions/checkout@vN
        with:
          submodules: true

      - name: Build
        shell: bash
        run: |
          cmake -B build/release -G "Visual Studio 17 2022" -A x64 \
            -DMYAPP_BUILD_TESTING=OFF
          cmake --build build/release --config Release
          mv build/release/bin/myapp.exe ./myapp-windows-x86_64.exe

      - name: Upload binary as artifact
        uses: actions/upload-artifact@vN
        with:
          name: myapp-windows-x86_64.exe
          path: myapp-windows-x86_64.exe
          retention-days: 1

  # Only where the project publishes an image. The tar is what release and
  # prerelease push, so the bytes that ship are the bytes that were built here.
  build_docker:
    if: inputs.artifacts
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@vN
        with:
          submodules: true

      - name: Build image
        run: docker build --platform linux/amd64 -t myapp -f Dockerfile.musl .

      - name: Save image as a tar
        run: docker save myapp -o myapp-docker.tar

      - name: Upload image as artifact
        uses: actions/upload-artifact@vN
        with:
          name: myapp-docker
          path: myapp-docker.tar
          retention-days: 1
```

Set `MYAPP_BUILD_TESTING=OFF` on the release builds: they ship the binary, and compiling Catch2 for an artifact nobody tests from is wasted runner time. The test workflow configures its own build with testing on.

`build/release/bin/myapp.exe` rather than `build/release/bin/Release/myapp.exe` depends on the per-config `CMAKE_RUNTIME_OUTPUT_DIRECTORY_<CFG>` settings being present in the root `CMakeLists.txt`; see cpp/cmake.md. Without them the Visual Studio generator writes to the per-config subdirectory and the `mv` fails.

#### Raw cmake on Windows

This is the one place CI does not go through `make`. The Makefile sets `SHELL := /bin/bash` and its targets rely on `sudo apt-get`, `grep -oP` and `find | xargs`, none of which a Windows runner provides. Calling `cmake` directly there is deliberate; do not add a second Windows-only Makefile to preserve the rule.

#### Stripping a macOS binary requires re-signing

On Apple Silicon every executable must carry at least an ad-hoc signature to run; the linker applies one automatically. `strip` rewrites the Mach-O and invalidates it, and the result is not a warning at build time but `zsh: killed` when anyone tries to run the published binary. `codesign --force --sign -` restores an ad-hoc signature, and the `--verify` line turns a silent regression into a failed build.

Either strip and re-sign, or do neither. What must not happen is stripping without re-signing, because the build stays green and only the shipped artifact is broken.

#### macOS architectures

Both architectures build natively, on `macos-15` for Apple Silicon and `macos-15-intel` for Intel. Pick the runner, not `CMAKE_OSX_ARCHITECTURES`: a cross-compiled binary cannot run on the machine that produced it, so its tests can only be skipped, and a build-verified-only artifact is the one most likely to be broken on arrival.

Set `CMAKE_OSX_DEPLOYMENT_TARGET` explicitly so the binary is not accidentally floored to the runner's current macOS version.

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
          - asset: myapp-linux-x86_64
            runner: ubuntu-24.04
          - asset: myapp-linux-aarch64
            runner: ubuntu-24.04-arm

    runs-on: ${{ matrix.runner }}
    steps:
      - uses: actions/checkout@vN
        with:
          submodules: true

      - name: Download release binary
        uses: actions/download-artifact@vN
        with:
          name: ${{ matrix.asset }}
          path: artifact

      - name: Stage downloaded binary
        run: |
          mv artifact/${{ matrix.asset }} ./myapp-under-test
          chmod +x ./myapp-under-test

      - name: Configure with the release binary as the functional-test target
        run: make configure CMAKE_ARGS="-DMYAPP_BINARY_PATH_OVERRIDE=${{ github.workspace }}/myapp-under-test"

      - name: Build test binaries
        run: make build

      - run: make test
      - run: make test_functional

  test_macos:
    strategy:
      matrix:
        include:
          - runner: macos-15
            asset: myapp-darwin-aarch64
          - runner: macos-15-intel
            asset: myapp-darwin-x86_64

    runs-on: ${{ matrix.runner }}
    steps:
      - uses: actions/checkout@vN
        with:
          submodules: true

      - name: Download release binary
        uses: actions/download-artifact@vN
        with:
          name: ${{ matrix.asset }}
          path: artifact

      - name: Stage downloaded binary
        run: |
          mv artifact/${{ matrix.asset }} ./myapp-under-test
          chmod +x ./myapp-under-test

      - name: Configure with the release binary as the functional-test target
        run: make configure CMAKE_ARGS="-DMYAPP_BINARY_PATH_OVERRIDE=${{ github.workspace }}/myapp-under-test"

      - name: Build test binaries
        run: make build

      - run: make test
      - run: make test_functional

  test_windows:
    runs-on: windows-2022
    steps:
      - uses: actions/checkout@vN
        with:
          submodules: true

      - name: Download release binary
        uses: actions/download-artifact@vN
        with:
          name: myapp-windows-x86_64.exe
          path: artifact

      - name: Stage downloaded binary
        shell: bash
        run: mv artifact/myapp-windows-x86_64.exe ./myapp-under-test.exe

      - name: Configure and build test binaries
        shell: bash
        run: |
          workspace="${GITHUB_WORKSPACE//\\//}"
          cmake -B build/dev -G "Visual Studio 17 2022" -A x64 \
            -DMYAPP_BINARY_PATH_OVERRIDE="${workspace}/myapp-under-test.exe"
          cmake --build build/dev --config Release

      - name: Run tests
        shell: bash
        run: |
          ctest --test-dir build/dev --output-on-failure -C Release -L unit
          ctest --test-dir build/dev --output-on-failure -C Release -L functional
```

Three details in there are load-bearing:

- **No clang tools installed.** `test.yml` does not lint, so clang-format and clang-tidy are not needed to build or run tests. Installing them here would also mean a per-runner branch, since the apt packages exist only on the Linux runner.
- **`GITHUB_WORKSPACE` is rewritten with forward slashes on Windows.** The raw value is a backslash path (`D:\a\repo\repo`), and the binary path is baked into the test binary as a compile definition, where `\a` and friends are read as C escape sequences. `${GITHUB_WORKSPACE//\\//}` is pure bash and needs no `cygpath`.
- **The layers run as separate steps** (`make test`, which is the unit layer alone, then `make test_functional`; or two `ctest -L` calls) rather than one `make test_all`, so a failure names the layer that broke. Use the target names the Makefile fragments actually define - `test`, `test_functional`, `test_all` - and do not invent a `test_unit`.

#### Every artifact runs its own tests

Each matrix entry tests on the runner its binary was built on, so the functional layer always spawns a binary the runner can execute and there is no skip branch anywhere in this workflow.

That is the argument for building on a native runner rather than cross-compiling. A skip is not a weaker test, it is no test: the artifact goes out having been compiled and nothing more, and the failures it hides are exactly the ones that only appear at runtime. A macOS binary stripped without re-signing, for instance, builds cleanly and is killed on launch, which no build-only check can catch.

### `release.yml`

Publishes a GitHub release: downloads every build artifact, generates install scripts and checksums with gpipe, signs them, and creates the release with changelog notes.

This is the build -> gpipe -> release pattern; the gpipe fragment covers what gpipe writes and how it is configured. The C++ specific part is that the build step is `build.yml` rather than a single builder, so the binaries arrive as downloaded artifacts.

```yaml
name: Release

on:
  workflow_call

# contents: write to create the release and upload assets
# id-token: write to obtain the OIDC token for cosign keyless signing
permissions:
  contents: write
  id-token: write

jobs:
  release:
    runs-on: ubuntu-24.04
    steps:
      # Default depth. Nothing here reads git history: get_changelog reads
      # CHANGELOG.md from the working tree, and gh release create uses the API.
      - uses: actions/checkout@vN

      - name: Download all build artifacts
        uses: actions/download-artifact@vN
        with:
          path: dist
          merge-multiple: true

      - name: Extract release notes from CHANGELOG.md
        run: make get_changelog TAG=${GITHUB_REF_NAME} > /tmp/release-notes.md

      - uses: thomaslaurenson/gpipe@vN
        with:
          cosign_sign: true

      # Name every asset. A dist/myapp-* glob is shorter and wrong: with
      # merge-multiple every artifact lands flat in dist/, so the Docker image
      # tar matches too and is attached to the release as if it were a binary.
      - name: Create release
        run: |
          gh release create "${GITHUB_REF_NAME}" \
            dist/myapp-linux-x86_64 \
            dist/myapp-linux-aarch64 \
            dist/myapp-windows-x86_64.exe \
            install.sh install.ps1 \
            checksums.txt checksums.txt.sigstore.json \
            --title "${GITHUB_REF_NAME}" \
            --notes-file /tmp/release-notes.md
        env:
          GH_TOKEN: ${{ github.token }}

  # Gated on the release: a registry outage then leaves a complete release with
  # no image, rather than an image with no release. See tools/docker.md.
  release_docker:
    needs: release
    runs-on: ubuntu-24.04
    steps:
      - name: Download image artifact
        uses: actions/download-artifact@vN
        with:
          name: myapp-docker

      - name: Load image
        run: docker load -i myapp-docker.tar

      - name: Push to ghcr
        env:
          GH_TOKEN: ${{ github.token }}
          ACTOR: ${{ github.actor }}
          IMAGE: ghcr.io/${{ github.repository }}
        run: |
          echo "$GH_TOKEN" | docker login ghcr.io -u "$ACTOR" --password-stdin
          docker tag myapp "$IMAGE:${GITHUB_REF_NAME}"
          docker tag myapp "$IMAGE:latest"
          docker push "$IMAGE:${GITHUB_REF_NAME}"
          docker push "$IMAGE:latest"
```

A project publishing an image adds `packages: write` to this workflow's `permissions` and to the caller job in `tag.yml`.

#### `.gpipe.yml`

The gpipe fragment covers the config surface and the action inputs. What is C++ specific is that the `path` entries must match where `download-artifact` puts the binaries: with `path: dist` and `merge-multiple: true` every artifact lands flat in `dist/`, so the paths are `./dist/<asset>`.

```yaml
binary: myapp

platforms:
  linux_amd64:
    path: ./dist/myapp-linux-x86_64
    name: myapp-linux-x86_64
  linux_arm64:
    path: ./dist/myapp-linux-aarch64
    name: myapp-linux-aarch64
  darwin_amd64:
    path: ./dist/myapp-darwin-x86_64
    name: myapp-darwin-x86_64
  darwin_arm64:
    path: ./dist/myapp-darwin-aarch64
    name: myapp-darwin-aarch64
  windows_amd64:
    path: ./dist/myapp-windows-x86_64.exe
    name: myapp-windows-x86_64.exe
```

One platform key maps to exactly one binary, and that is the other reason Linux ships a single static musl build per architecture: there is no way to express "glibc or musl, reader's choice", so the installer has to be given the one that runs everywhere.

### `prerelease.yml`

A single rolling GitHub prerelease under the literal tag `dev`, rebuilt on every push to main: raw binaries from `build.yml` only, no install scripts, no checksums, no changelog notes.

**gpipe does not appear here, and cannot**; see the gpipe fragment for why.

This needs no separate `git tag -f`/`git push --force` step: deleting the old release with `--cleanup-tag` removes its git tag too, so the following `gh release create dev --target <sha>` creates a fresh `dev` tag at the built commit on its own. Pass `--target ${{ github.sha }}` explicitly rather than letting `gh` default it to the current default-branch head, which can already have moved on by the time the job publishes.

Reuse the same `path: dist, merge-multiple: true` download and the same explicit asset list as `release.yml`. Naming them is deliberate in both places: release assets are a public interface, and a glob attaches whatever happens to match, which is how an image tar sharing the directory ends up published as a binary. Adding a platform is a decision, so let it be an edit.

Existence-check the delete exactly as in the Go prerelease pattern: three outcomes, not two. Never write `gh release delete dev --yes --cleanup-tag || true`; that collapses "no dev release exists yet" and "the API could not tell me" into the same branch, and the job then publishes over a release state it never established.

```yaml
name: Prerelease

on:
  workflow_call

permissions:
  contents: write

jobs:
  prerelease:
    runs-on: ubuntu-24.04
    # This job never checks out, so gh has no remote to infer the repository
    # from and GH_REPO has to name it. See github/actions.md.
    env:
      GH_TOKEN: ${{ github.token }}
      GH_REPO: ${{ github.repository }}
    steps:
      - name: Download all build artifacts
        uses: actions/download-artifact@vN
        with:
          path: dist
          merge-multiple: true

      - name: Delete any existing dev release
        run: |
          if err=$(gh release view "dev" 2>&1 >/dev/null); then
            gh release delete "dev" --yes --cleanup-tag
          elif grep -qi "release not found" <<<"$err"; then
            echo "No existing dev release"
          else
            echo "::error::could not determine whether a dev release exists: ${err}"
            exit 1
          fi

      - name: Create prerelease
        run: |
          gh release create dev --prerelease \
            --target "${{ github.sha }}" \
            --title "dev" \
            --notes "Rolling build of ${{ github.sha }}" \
            dist/myapp-linux-x86_64 \
            dist/myapp-linux-aarch64 \
            dist/myapp-windows-x86_64.exe

  prerelease_docker:
    needs: prerelease
    runs-on: ubuntu-24.04
    steps:
      - name: Download image artifact
        uses: actions/download-artifact@vN
        with:
          name: myapp-docker

      - name: Load image
        run: docker load -i myapp-docker.tar

      # dev only, never latest: latest tracks releases, so pointing it at a
      # rolling build makes an untagged docker pull return whatever last landed
      # on the default branch.
      - name: Push dev tag to ghcr
        env:
          GH_TOKEN: ${{ github.token }}
          ACTOR: ${{ github.actor }}
          IMAGE: ghcr.io/${{ github.repository }}
        run: |
          echo "$GH_TOKEN" | docker login ghcr.io -u "$ACTOR" --password-stdin
          docker tag myapp "$IMAGE:dev"
          docker push "$IMAGE:dev"
```

As with `release.yml`, a project publishing an image adds `packages: write` to this workflow's `permissions` and to the `prerelease` job in `main.yml`.
