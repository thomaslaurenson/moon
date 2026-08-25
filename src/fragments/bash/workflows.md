# Bash CI workflows

Supplements the shared GitHub Actions conventions. Applies to a maintained Bash repo (an installer, a set of scripts with tests); a one-off script needs no CI.

`@vN` in the examples below means pin the current major of the action at authoring time (for example `@v4`); Dependabot keeps the pin current. Do not copy a version number from this document as the target to match.

## Paths filter

Base paths filter for `pr.yml` and `main.yml` (the two lists must stay identical):

```yaml
paths:
  - ".github/workflows/**"
  - "Makefile"
  - "src/**"
  - "test/**"
```

Extend the base list with every additional path the project lints, tests, or ships: directories such as `scripts/**`, `contrib/**`, `completion/**`, `man/**`, and top-level scripts such as `install.sh`. Use directory globs, not extension globs; bash projects ship extensionless files (completions, man pages, mock helpers) that `**/*.sh` misses.

Derivation rule: the filter is correct when every path referenced by the Makefile's lint, test, and release targets is covered. If `make lint` or `make test` reads a file the filter misses, the filter is wrong, not the Makefile.

## Lint and test jobs

Lint job. Runs the syntax check and ShellCheck via the Makefile (see the bash testing fragment). ShellCheck is preinstalled on the `ubuntu-24.04` runner, so no install step is needed:

```yaml
- uses: actions/checkout@vN
- run: make lint
```

Test job. bats is vendored as a git submodule, so check out with submodules and run the suite through the Makefile:

```yaml
- uses: actions/checkout@vN
  with:
    submodules: true
- run: make test
```

## Runner matrix

When the scripts support macOS, lint and test run on a matrix of `ubuntu-24.04` and `macos-15`:

```yaml
strategy:
  matrix:
    os: [ubuntu-24.04, macos-15]
runs-on: ${{ matrix.os }}
```

The macOS leg installs its own tooling behind `if: runner.os == 'macOS'` (`brew install bash shellcheck`): the system bash is 3.2, too old for the scripts and for bats, and ShellCheck is not preinstalled there. A Linux-only project skips the matrix and runs on `ubuntu-24.04` alone.

## Versioning

The version is embedded in exactly one source file as a `VERSION` variable; that is the single source of truth. Every other place a version appears (man page, changelog) is checked against it, never edited independently. Git tags are `v`-prefixed; the embedded version is bare.

Three Makefile targets carry the convention:

- `get_version` extracts the embedded version (all extraction goes through `##@ GET` targets; see the Makefile conventions).
- `check_version` verifies every file that states the version agrees with the embedded one. It runs in the lint workflow and in `ci`.
- `check_version_tag TAG=vX.Y.Z` verifies the embedded version matches the tag, leading `v` stripped. The release workflow runs it before building anything, so a release with a mismatched tag fails before an artefact exists.

## Releases

Bash has no compile step, so there is never a `build.yml` or a `prerelease.yml`.

Releasing is optional and decided per project:

- A project that does not release has no `tag.yml` and no `release.yml`; CI is `pr.yml` plus `main.yml` only.
- A project that releases adds `tag.yml` and `release.yml`. Artefacts are assembled, not compiled: `git archive` tarballs, version-baked scripts, `checksums.txt`. Each artefact is produced by a Makefile target so the release can be reproduced locally, then published with `gh release create` using the changelog notes.
- A project distributed by URL (a `curl | bash` installer) may add a publish job (for example GitHub Pages) alongside the release; it still builds its payload through make targets.
