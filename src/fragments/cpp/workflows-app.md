# C++ shipped-binary workflows

Applies to any tier that ships a distributable binary: an application, or a library with a bundled CLI. A plain library has no artifact to build and ships nothing, so it builds and tests in one plain job instead; see workflows-lib.md.

`@vN` in the examples below means pin the current major of the action at authoring time (for example `@v5`); Dependabot keeps the pin current. Do not copy a version number from this document as the target to match.

## Paths filter addition

Add `Dockerfile*` to the shared paths filter (see cpp/workflows.md), plus `.gpipe.yml` where the release uses gpipe (see `release.yml` below).

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
      id-token: write                   # cosign keyless signing, see release.yml
  prerelease:                           # main.yml only
    uses: ./.github/workflows/prerelease.yml
    needs: [build, lint, test]
    permissions:
      contents: write
```

No `secrets: inherit`: every job here authenticates with the automatic `github.token`, so nothing needs to be forwarded. Add `secrets: inherit` only if a workflow genuinely reads a repository secret.

`id-token: write` has to appear on the caller's `release` job as well as inside `release.yml`. A reusable workflow cannot widen its own permissions beyond what the caller grants, so declaring it only in the callee fails at signing time with an OIDC token request error.

## Reusable workflow bodies

### `build.yml`

Builds one release binary per target platform and uploads each as an artifact for the test, release and prerelease workflows to consume. Linux builds go through Docker; macOS and Windows build natively on their own runners, because there is no container route to either.

#### Asset naming

Name every artifact `<app>-<os>-<arch>`, using `x86_64`/`aarch64` rather than `amd64`/`arm64`, and append `.exe` on Windows:

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
          - osx_arch: arm64
            asset: myapp-darwin-aarch64
          - osx_arch: x86_64
            asset: myapp-darwin-x86_64

    runs-on: macos-15
    steps:
      - uses: actions/checkout@vN
        with:
          submodules: true

      - name: Build (${{ matrix.osx_arch }})
        run: |
          cmake -B build \
            -DCMAKE_BUILD_TYPE=Release \
            -DMYAPP_BUILD_TESTING=OFF \
            -DCMAKE_OSX_ARCHITECTURES=${{ matrix.osx_arch }} \
            -DCMAKE_OSX_DEPLOYMENT_TARGET=11.0
          cmake --build build --config Release --parallel 3
          strip build/bin/myapp
          # strip invalidates the linker's ad-hoc signature; re-sign or the
          # binary is killed on launch on Apple Silicon. See below.
          codesign --force --sign - build/bin/myapp
          codesign --verify --verbose build/bin/myapp
          mv build/bin/myapp ./${{ matrix.asset }}

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
          cmake -B build -G "Visual Studio 17 2022" -A x64 \
            -DMYAPP_BUILD_TESTING=OFF
          cmake --build build --config Release
          mv build/bin/myapp.exe ./myapp-windows-x86_64.exe

      - name: Upload binary as artifact
        uses: actions/upload-artifact@vN
        with:
          name: myapp-windows-x86_64.exe
          path: myapp-windows-x86_64.exe
          retention-days: 1
```

Set `MYAPP_BUILD_TESTING=OFF` on the release builds: they ship the binary, and compiling Catch2 for an artifact nobody tests from is wasted runner time. The test workflow configures its own build with testing on.

`build/bin/myapp.exe` rather than `build/bin/Release/myapp.exe` depends on the per-config `CMAKE_RUNTIME_OUTPUT_DIRECTORY_<CFG>` settings being present in the root `CMakeLists.txt`; see cpp/cmake.md. Without them the Visual Studio generator writes to the per-config subdirectory and the `mv` fails.

#### Raw cmake on Windows

This is the one place CI does not go through `make`. The Makefile sets `SHELL := /bin/bash` and its targets rely on `sudo apt-get`, `grep -oP` and `find | xargs`, none of which a Windows runner provides. Calling `cmake` directly there is deliberate; do not add a second Windows-only Makefile to preserve the rule.

#### Stripping a macOS binary requires re-signing

On Apple Silicon every executable must carry at least an ad-hoc signature to run; the linker applies one automatically. `strip` rewrites the Mach-O and invalidates it, and the result is not a warning at build time but `zsh: killed` when anyone tries to run the published binary. `codesign --force --sign -` restores an ad-hoc signature, and the `--verify` line turns a silent regression into a failed build.

Either strip and re-sign, or do neither. What must not happen is stripping without re-signing, because the build stays green and only the shipped artifact is broken.

#### macOS architectures

`macos-14` and `macos-15` are both Apple Silicon, and `macos-13` (Intel) is outside the supported runner list and being retired. So the x86_64 macOS binary is **cross-compiled** via `CMAKE_OSX_ARCHITECTURES=x86_64` and cannot be exercised natively; see `test.yml` for how that is handled. Set `CMAKE_OSX_DEPLOYMENT_TARGET` explicitly so the binary is not accidentally floored to the runner's current macOS version.

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
          - asset: myapp-darwin-aarch64
            native: true
          - asset: myapp-darwin-x86_64
            native: false

    runs-on: macos-15
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

      # A native binary that will not run is a broken release artifact and must
      # fail the job. Only the cross-compiled one may legitimately be unable to
      # run here, and only then is skipping the right answer.
      - name: Check the binary can execute on this runner
        id: exec_check
        run: |
          if ./myapp-under-test --version > /dev/null 2>&1; then
            echo "runnable=true" >> "$GITHUB_OUTPUT"
          elif [ "${{ matrix.native }}" = "true" ]; then
            echo "::error::${{ matrix.asset }} is native to this runner but will not execute."
            echo "A stripped macOS binary whose ad-hoc signature was not restored fails exactly this way;"
            echo "check the codesign step in build.yml. Diagnostics follow."
            codesign --verify --verbose ./myapp-under-test || true
            ./myapp-under-test --version || true
            exit 1
          else
            echo "runnable=false" >> "$GITHUB_OUTPUT"
            echo "::warning::${{ matrix.asset }} is cross-compiled and cannot execute here (no Rosetta 2); functional tests skipped"
          fi

      - name: Configure with the release binary as the functional-test target
        run: make configure CMAKE_ARGS="-DMYAPP_BINARY_PATH_OVERRIDE=${{ github.workspace }}/myapp-under-test"

      - name: Build test binaries
        run: make build

      - run: make test

      - if: steps.exec_check.outputs.runnable == 'true'
        run: make test_functional

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
          cmake -B build -G "Visual Studio 17 2022" -A x64 \
            -DMYAPP_BINARY_PATH_OVERRIDE="${workspace}/myapp-under-test.exe"
          cmake --build build --config Release

      - name: Run tests
        shell: bash
        run: |
          ctest --test-dir build --output-on-failure -C Release -L unit
          ctest --test-dir build --output-on-failure -C Release -L functional
```

Three details in there are load-bearing:

- **No `install_clang_tools` step.** `test.yml` does not lint, so clang-format and clang-tidy are not needed to build or run tests. The target is also `sudo apt-get`, which fails outright on the macOS and Windows runners.
- **`GITHUB_WORKSPACE` is rewritten with forward slashes on Windows.** The raw value is a backslash path (`D:\a\repo\repo`), and the binary path is baked into the test binary as a compile definition, where `\a` and friends are read as C escape sequences. `${GITHUB_WORKSPACE//\\//}` is pure bash and needs no `cygpath`.
- **The layers run as separate steps** (`make test`, which is the unit layer alone, then `make test_functional`; or two `ctest -L` calls) rather than one `make test_all`. That is what allows the functional layer alone to be skipped on a platform where the artifact cannot execute. Use the target names the Makefile fragments actually define - `test`, `test_functional`, `test_all` - and do not invent a `test_unit`.

#### The cross-compiled macOS binary

The x86_64 macOS artifact is cross-compiled on an Apple Silicon runner (see `build.yml`), so the runner may not be able to execute it. The unit layer is compiled natively on the runner and so verifies the library on the runner's own architecture for both matrix entries; only the functional layer actually spawns the downloaded artifact.

The `native` matrix flag is what keeps the skip honest, and it is the reason the check is not a plain `if runnable`. "This binary will not run" has two very different causes: the runner lacks Rosetta 2 and cannot be expected to run a foreign-architecture binary, or the binary itself is broken. Without the flag both take the skip branch, and a macOS release binary that was stripped without re-signing sails through CI green with a warning that blames Rosetta. Skipping is only ever correct for the cross-compiled artifact; for the native one, refusing to run *is* the test failing.

Whether the skip branch is ever reached depends on Rosetta 2 being present on the image, which is not guaranteed and changes between runner image releases. Confirm it empirically on the first run rather than assuming: if the check reports `runnable=false`, the x86_64 macOS binary is build-verified only, and that is worth stating in the project's README rather than leaving implicit.

### `release.yml`

Publishes a GitHub release: downloads every build artifact, generates install scripts and checksums with gpipe, signs them, and creates the release with changelog notes.

The pattern is the same three steps as the Go flow (see golang/release-cli.md): **build -> gpipe -> release**. gpipe is language-agnostic, not Go-only; it consumes built binaries from paths named in `.gpipe.yml` and neither knows nor cares what produced them. `gpipe` generates `install.sh`, `install.ps1` and `checksums.txt`, and with `cosign_sign: true` also emits `checksums.txt.sigstore.json`. Do not hand-roll `sha256sum`: gpipe's `checksums.txt` covers the installer scripts themselves as well as the platform binaries, which is what lets a cautious user verify `install.sh` before piping it to a shell.

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
      - uses: actions/checkout@vN
        with:
          fetch-depth: 0

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

      - name: Create release
        run: |
          gh release create "${GITHUB_REF_NAME}" \
            dist/myapp-* \
            install.sh install.ps1 \
            checksums.txt checksums.txt.sigstore.json \
            --title "${GITHUB_REF_NAME}" \
            --notes-file /tmp/release-notes.md
        env:
          GH_TOKEN: ${{ github.token }}
```

`gpipe` needs no `version` or `repo` inputs here: they default to `github.ref_name` and `github.repository`, and `tag.yml` only ever fires on a `v*` tag, so `ref_name` is already the semantic version gpipe expects. `id-token: write` must also be granted by the caller job in `tag.yml`, not just declared here.

The action builds gpipe from its own checkout, so the ref pinned in `uses:` is the gpipe that runs and there is no separate version input to keep current (see golang/release-cli.md, which uses the same action). It needs Go on the runner, which GitHub-hosted runners provide; a self-hosted runner without Go must add `actions/setup-go` before it.

#### `.gpipe.yml`

Lives at the project root, and its `path` entries must match where `download-artifact` puts the binaries. With `path: dist` and `merge-multiple: true` every artifact lands flat in `dist/`, so the paths are `./dist/<asset>`:

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

One platform key maps to exactly one binary, which is the other reason Linux ships a single static musl build per architecture: there is no way to express "glibc or musl, reader's choice" here, and the installer has to pick the one that runs everywhere.

`binary`, `platforms` and `hooks` are the whole config surface. Shell completions are not gpipe's job: a binary that ships them installs them from a `post-sh` hook, which has the installed location in `INSTALL_DIR` and the name in `BINARY`.

Validate the config before relying on it in CI: `gpipe validate --repo <owner/repo> --version v0.0.0` checks the schema, the platform identifiers and any hooks without needing the binaries present.

### `prerelease.yml`

A single rolling GitHub prerelease under the literal tag `dev`, rebuilt on every push to main: raw binaries from `build.yml` only, no install scripts, no checksums, no changelog notes.

**gpipe does not appear here, and cannot.** It validates `--version` as a semantic version, so the literal string `dev` is rejected outright; and the installers it generates hardcode `releases/download/<version>/<asset>`, so passing a semver-shaped stand-in like `v1.2.4-dev` produces a script whose every download URL 404s against a release actually tagged `dev`. Install scripts are a release-only artifact. A rolling channel that is deleted and recreated on every push is the wrong thing to hang a stable `curl | bash` URL off anyway.

This mirrors the Go rolling-dev channel (see golang/release-cli.md) but needs no separate `git tag -f`/`git push --force` step: deleting the old release with `--cleanup-tag` removes its git tag too, so the following `gh release create dev --target <sha>` creates a fresh `dev` tag at the built commit on its own. Pass `--target ${{ github.sha }}` explicitly rather than letting `gh` default it to the current default-branch head, which can already have moved on by the time the job publishes.

Reuse the same `path: dist, merge-multiple: true` download and `dist/myapp-*` glob as `release.yml`, rather than downloading and naming each `build.yml` matrix entry individually. A build matrix entry added later needs no matching change here; a prerelease workflow that lists each artifact by name has to be edited every time the matrix does, and silently omits new targets in the meantime.

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
        env:
          GH_TOKEN: ${{ github.token }}

      - name: Create prerelease
        run: |
          gh release create dev --prerelease \
            --target "${{ github.sha }}" \
            --title "dev" \
            --notes "Rolling build of ${{ github.sha }}" \
            dist/myapp-*
        env:
          GH_TOKEN: ${{ github.token }}
```
