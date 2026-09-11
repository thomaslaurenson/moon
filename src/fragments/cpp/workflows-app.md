# C++ shipped-binary workflows

Applies to any tier that ships a distributable binary: an application, or a library with a bundled CLI. A plain library has no artifact to build and ships nothing, so it builds and tests in one plain job instead; see workflows-lib.md.

`@vN` in the examples below means pin the current major at authoring time (for example `@v5`); Dependabot keeps it current. Do not copy a version number from this document as the target to match. Majors are for `actions/*` and for actions you publish yourself; every other action pins to a full SHA (see the GitHub Actions fragment).

## Paths filter addition

Add `Dockerfile*` to the shared paths filter (see cpp/workflows.md), plus `.gpipe.yml` where the release uses gpipe (see cpp/release-app.md).

## Depth is a function of the trigger

These tiers add a `build.yml` that produces the distributable artifacts. Running the whole of it on every pull request is the expensive default and buys the least: a pull request needs to know the code compiles and the tests pass, not that a shippable artifact for every platform is correct. A tag needs the second, and only a tag can act on it.

So the caller decides the depth, and `build.yml` takes an input:

| Trigger | Runs | Why |
|---|---|---|
| `pr.yml` | lint, test, and the one artifact job a test cannot cover | Cheapest signal that catches a real break |
| `main.yml` | the above plus the full artifact set and `prerelease` | The rolling channel has to contain everything a release would |
| `tag.yml` | everything, plus `release` | This is the run whose output people install |

Concretely, `inputs.artifacts` gates `build_windows` and `build_docker`; `build_linux` runs on every trigger. That split is the whole cost control, and it is worth stating why it falls where it does rather than the other way round. `build_linux` builds through `Dockerfile.musl`, so running it is the only thing that proves the release container recipe still works, and it runs on the cheapest runner there is. The other three are either a second compile of a platform `test.yml` already covers, or an image tar nothing consumes until a release. On a private repository the gated jobs are also the more expensive ones: a Windows runner bills at twice a Linux one, so leaving it ungated means a pull request pays for the platform it least needs.

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
      id-token: write                   # cosign keyless signing, see cpp/release-app.md
  prerelease:                           # main.yml only
    uses: ./.github/workflows/prerelease.yml
    needs: [build, lint, test]
    permissions:
      contents: write
```

**No `needs:` on `lint`, `test` or `build`.** None of the three consumes another's output: lint configures its own tree, test builds the binaries it runs, and build produces artifacts that nothing reads until a release. All three start at once, so the fastest signal is never held behind the slowest job. Only `release` and `prerelease` genuinely need the artifacts, and they say so.

This is what the decoupling in `test.yml` buys, and it is worth protecting. The moment `test` downloads something `build` produced, that edge has to exist, and gating a platform off a pull request then breaks the test matrix rather than just saving money.

Declare `permissions: contents: read` at the top of every caller and widen it on the jobs that need more. A caller with no top-level block inherits the repository default, which may be read and write.

No `secrets: inherit`: every job here authenticates with the automatic `github.token`, so nothing needs to be forwarded. Add `secrets: inherit` only if a workflow genuinely reads a repository secret.

`id-token: write` has to appear on the caller's `release` job as well as inside `release.yml`. A reusable workflow cannot widen its own permissions beyond what the caller grants, so declaring it only in the callee fails at signing time with an OIDC token request error.

## Reusable workflow bodies

### `build.yml`

Builds one release binary per target platform and uploads each as an artifact for the release and prerelease workflows to consume. Linux builds go through Docker; Windows builds natively on its own runner, because there is no container route to it.

**Which platforms is a decision, not a default.** The jobs below are the full set for the platforms these projects target, and a project takes the ones matching where its users actually run the binary. The test is the one in cpp/workflows.md: a platform earns a job by being a deployment target, not by adding confidence.

**No C++ project here targets macOS.** There is no macOS job, no `darwin` asset and no `darwin` entry in `.gpipe.yml`, and that is a decision rather than an omission: nobody runs these tools there, and a macOS runner bills at ten times a Linux one. Adding it later means more than a build job, which is why the absence is recorded here rather than left to be noticed: macOS needs a matching `test.yml` job, `darwin_amd64` and `darwin_arm64` entries in `.gpipe.yml`, two more lines in the release asset list, and a `codesign --force --sign -` step after any `strip`, because stripping invalidates the ad-hoc signature the linker applies and the binary is then killed on launch rather than failing to build.

Once a platform is in, it stays consistent all the way through: a job in `build.yml`, a matching entry in `test.yml`, an asset in `.gpipe.yml`, and a line in the release asset list. A platform built but not tested ships an artifact nothing has run.

#### Asset naming

Name every artifact `<app>-<os>-<arch>`, using `x86_64`/`aarch64` rather than `amd64`/`arm64`, and append `.exe` on Windows. The table is the naming for whichever platforms a project builds, not a list of platforms it must:

| Platform | Asset |
|---|---|
| Linux x86_64 | `myapp-linux-x86_64` |
| Linux ARM64 | `myapp-linux-aarch64` |
| Windows x86_64 | `myapp-windows-x86_64.exe` |

This is gpipe's platform vocabulary, so the names map onto `.gpipe.yml` with no translation; the identifiers are the `platforms` keys shown in cpp/release-app.md, and `gpipe validate` checks a config against them. Pick the naming before the first release: the assets are a public interface, and renaming them later breaks anyone's install script.

#### One Linux binary, not two

Build Linux as a **single statically linked musl binary per architecture**, not a glibc/musl pair. A static musl binary runs on every distribution with no libc version floor; a glibc binary built on `ubuntu-24.04` needs glibc >= ~2.39 and will not start on RHEL 8/9 or Debian 11/12. Shipping both is not "portable plus compatible", it is portable plus a strictly narrower duplicate, and gpipe can only point the installer at one of them anyway.

The usual objection is musl's slower allocator. Measure before believing it applies: for a tool that allocates a few large buffers up front and then computes, the difference is single-digit percent. Static musl is the default; deviate only with a benchmark showing it costs something real.

```yaml
name: Build

on:
  workflow_call:
    inputs:
      artifacts:
        description: Build the full release artifact set, not just the cheap subset
        type: boolean
        default: false

permissions:
  contents: read

jobs:
  build_linux:
    strategy:
      # Report both architectures independently; one failing tells you nothing
      # about the other, and cancelling hides half the answer.
      fail-fast: false
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

      # Run the bytes that will ship. See Smoke-run every artifact.
      - name: Smoke-run the artifact
        run: |
          chmod +x ./${{ matrix.asset }}
          ./${{ matrix.asset }} --version

      - name: Upload binary as artifact
        uses: actions/upload-artifact@vN
        with:
          name: ${{ matrix.asset }}
          path: ${{ matrix.asset }}
          retention-days: 1

  build_windows:
    if: inputs.artifacts
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

      - name: Smoke-run the artifact
        shell: bash
        run: ./myapp-windows-x86_64.exe --version

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

#### Where CI does not go through `make`

`build.yml` is the exception to the rule that CI calls `make <target>` and never `cmake` directly, and it is the exception on both platforms rather than only Windows. The Linux jobs build inside a container, so the recipe lives in the Dockerfile; the Windows job invokes `cmake` directly.

Windows has the strongest reason: the Makefile sets `SHELL := /bin/bash` and its targets rely on `find | xargs` and GNU-only flags, none of which a Windows runner provides. A release build also wants an explicit build type and output path rather than the everyday `build/dev` configuration the Makefile is built around, which is the reason that holds on any platform.

Do not add a second, platform-specific Makefile to preserve the rule. `lint.yml` and `test.yml` do go through `make`, and those are the workflows the rule is really about, because they run the same checks a developer runs.

#### Smoke-run every artifact

Every build job runs the binary it just produced, on the runner that produced it, before uploading it. `--version` is enough: the point is not to test behaviour, which `test.yml` already does, but to prove the artifact starts at all.

This is the whole class of failure a build-only job cannot see, and each member of it ships silently:

- a `scratch` image binary that was not statically linked, so there is no loader and the container exits with `no such file or directory` on a file that plainly exists
- a binary linked against a library version the runner has and a user does not

It costs seconds on a runner already holding the binary, and it is the reason `test.yml` can build its own binary rather than downloading this one: between them, `test.yml` proves the code is correct and `build.yml` proves the artifact runs.

Where a project's binary has no `--version`, use the cheapest subcommand that exits zero without arguments. A binary with no such entry point should still be executed with `--help`.

### `test.yml`

Builds and tests on each platform the project supports. It does not download anything from `build.yml`: it compiles the library, the binary and the test binaries itself, and the functional layer spawns the binary it just built.

**`test.yml` never consumes a build artifact, and that is a deliberate decoupling.** Downloading the shipped binary and testing that instead sounds stronger, and it ties the test matrix to whatever `build.yml` happened to produce on this trigger. The moment a platform is gated off a pull request to save runner minutes, its test job has nothing to download and fails, so the test matrix has to start varying by trigger too. Building what it tests keeps `test.yml` identical on every trigger, and leaves gating a decision that only `build.yml` has to know about.

What that gives up is running the functional suite against the exact bytes that ship. `build.yml` covers the part of that which actually breaks by executing each artifact where it was built; see Smoke-run every artifact below.

Check out with submodules: the test binaries link Catch2 and `subprocess.h`, which are submodules.

```yaml
name: Test

on:
  workflow_call

permissions:
  contents: read

jobs:
  # test_linux and test_asan are the shared jobs from cpp/workflows.md, with one
  # line added: a tier that ships a binary runs the functional layer too.
  #
  #   - run: make test_functional
  #
  # Only where the project ships a Windows binary: on a private repository a
  # Windows runner bills at twice a Linux one.
  test_windows:
    runs-on: windows-2022
    steps:
      - uses: actions/checkout@vN
        with:
          submodules: true

      # Raw cmake: the Makefile needs a POSIX shell. See Where CI does not go
      # through make.
      - name: Configure and build
        shell: bash
        run: |
          cmake -B build/dev -G "Visual Studio 17 2022" -A x64
          cmake --build build/dev --config Release --parallel

      # Separate steps so a failure names the layer that broke.
      - name: Run tests
        shell: bash
        run: |
          ctest --test-dir build/dev --output-on-failure -C Release -L unit
          ctest --test-dir build/dev --output-on-failure -C Release -L functional
```

Three details in there are load-bearing:

- **No clang tools installed.** `test.yml` does not lint, so clang-format and clang-tidy are not needed to build or run tests. Installing them here would also mean a per-runner branch, since the apt packages exist only on the Linux runner.
- **Both build types are covered, without a third job.** Pairing `Release` with one compiler and `Debug` with the other costs nothing extra and stops `NDEBUG` and the optimiser from being exercised for the first time by a release. Testing only `Debug` leaves the shipped configuration untested; testing both under both compilers doubles the matrix for very little.
- **The layers run as separate steps** (`make test`, which is the unit layer alone, then `make test_functional`; or two `ctest -L` calls) rather than one `make test_all`, so a failure names the layer that broke. Use the target names the Makefile fragments actually define - `test`, `test_functional`, `test_all` - and do not invent a `test_unit`.

#### Every platform tests natively

Each job tests on the runner it built on, so the functional layer always spawns a binary the runner can execute and there is no skip branch anywhere in this workflow.

That is also the argument for building release artifacts on a native runner rather than cross-compiling. A skip is not a weaker test, it is no test: the artifact goes out having been compiled and nothing more, and the failures it hides are exactly the ones that only appear at runtime.

The workflows that publish these artifacts, `release.yml` and `prerelease.yml`, live in cpp/release-app.md.
