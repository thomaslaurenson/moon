# Bash CI workflows

Supplements the shared GitHub Actions conventions. Applies to a maintained Bash repo (an installer, a set of scripts with tests); a one-off script needs no CI.

Paths filter for `pr.yml` and `main.yml`:

```yaml
paths:
  - ".github/workflows/**"
  - "Makefile"
  - "**/*.sh"
  - "**/*.bash"
  - "test/**"
```

`@vN` in the examples below means pin the current major of the action at authoring time (for example `@v4`); Dependabot keeps the pin current. Do not copy a version number from this document as the target to match.

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

Bash projects ship no compiled artifact, so there is no `build.yml` and no `prerelease.yml`. A tagged release, when wanted, is a `release.yml` that publishes changelog notes only, mirroring the pattern a pure library uses.
