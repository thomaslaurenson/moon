# Python CI workflows

Supplements the generic GitHub Actions rules. CI steps call `make <target>` rather than raw commands, with one exception: ruff runs through `astral-sh/ruff-action`, which is the canonical way to pin and run ruff in CI. Extract tool versions from `pyproject.toml` at runtime; never hardcode.

Python projects do not use `prerelease.yml`; it is a Go/C++ compiled-binary concept. `main.yml` runs lint and test only on push to main, with no prerelease job.

Paths filter for `pr.yml` and `main.yml`. Include every source-bearing directory the project actually has, additively: a project with both a package and `tasks/` (or `docs/`) lists all of them, not just one:

```yaml
paths:
  - ".github/workflows/**"
  - "Makefile"
  - "pyproject.toml"
  - "uv.lock"
  - "<package>/**"
  - "tasks/**"
  - "docs/**"
  - "tests/**"
```

`@vN` in the examples below means pin the current major of the action at authoring time (for example `@v6`); Dependabot keeps the pin current. Do not copy a version number from this document as the target to match.

Lint job. Ruff lint and format both run, so CI enforces formatting as well as linting (locally the equivalent is `make lint` and `make fmt_check`). No Python setup is needed:

```yaml
- uses: actions/checkout@vN
- name: Extract ruff version
  id: ruff-version
  run: echo "version=$(make get_ruff_version)" >> $GITHUB_OUTPUT
- uses: astral-sh/ruff-action@vN
  with:
    version: ${{ steps.ruff-version.outputs.version }}
    args: check .
- uses: astral-sh/ruff-action@vN
  with:
    version: ${{ steps.ruff-version.outputs.version }}
    args: format --check .
```

Test job:

```yaml
- uses: actions/checkout@vN
- name: Extract Python version
  id: python-version
  run: echo "version=$(make get_python_required_version)" >> $GITHUB_OUTPUT
- uses: actions/setup-python@vN
  with:
    python-version: ${{ steps.python-version.outputs.version }}
- uses: astral-sh/setup-uv@vN
- run: uv sync
- run: make test
```

`uv sync` installs the project plus the default `dev` dependency group, which is where the test tooling lives (see the tooling and project fragments). A library that keeps user-facing integration extras runs `uv sync --all-extras` instead, so optional-dependency imports resolve during tests.
