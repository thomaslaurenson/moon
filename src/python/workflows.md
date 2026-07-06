# Python CI Workflows

Supplements the generic GitHub Actions rules. CI steps call `make <target>`, never raw commands. Extract versions from `pyproject.toml` at runtime; never hardcode.

Python projects do not use `prerelease.yml`; it is a Go/C++ compiled-binary concept. `main.yml` runs lint + test only on push to main, with no prerelease job.

Paths filter for `pr.yml` and `main.yml`. Include every source-bearing directory the project actually has, additively - a project with both a package and `tasks/` (or `docs/`) lists all of them, not just one:

```yaml
paths:
  - ".github/workflows/**"
  - "Makefile"
  - "pyproject.toml"
  - "<package>/**"
  - "tasks/**"
  - "docs/**"
  - "tests/**"
```

Lint job (no Python setup needed):

```yaml
- uses: actions/checkout@v6
- name: Extract ruff version
  id: ruff-version
  run: echo "version=$(make get_ruff_version)" >> $GITHUB_OUTPUT
- uses: astral-sh/ruff-action@v3
  with:
    version: ${{ steps.ruff-version.outputs.version }}
    args: check .
```

Test job:

```yaml
- uses: actions/checkout@v6
- name: Extract Python version
  id: python-version
  run: echo "version=$(make get_python_required_version)" >> $GITHUB_OUTPUT
- uses: actions/setup-python@v5
  with:
    python-version: ${{ steps.python-version.outputs.version }}
- uses: astral-sh/setup-uv@v7
- run: uv sync --all-extras
- run: make test
```
