# Python Workflow Conventions

Supplements `github/actions.md`. Universal rules (runners, action versions, workflow structure, caller patterns, concurrency, permissions) apply unchanged. This file covers Python-specific setup steps only.

---

## Paths Filter

Use these entries in the `paths:` filter for `pr.yml` and `main.yml`:

```yaml
paths:
  - ".github/workflows/**"
  - "Makefile"
  - "pyproject.toml"
  - "<package>/**"
  - "tests/**"
```

Replace `<package>` with the actual package directory name (the installable package at the repo root). For scripts-only projects that have no installable package, use `tasks/**` instead.

---

## Python Setup Steps

Extract the Python version from `pyproject.toml` at runtime. Never hardcode it:

```yaml
- uses: actions/checkout@v6

- name: Extract Python version
  id: python-version
  run: echo "version=$(make get_python_required_version)" >> $GITHUB_OUTPUT

- uses: actions/setup-python@v5
  with:
    python-version: ${{ steps.python-version.outputs.version }}

- uses: astral-sh/setup-uv@v7
```

---

## Lint: Ruff

For lint-only jobs, extract the ruff version and use the ruff action directly; no Python setup required:

```yaml
- name: Extract ruff version
  id: ruff-version
  run: echo "version=$(grep -oP 'ruff>=\K[0-9.]+' pyproject.toml)" >> $GITHUB_OUTPUT

- uses: astral-sh/ruff-action@v3
  with:
    version: ${{ steps.ruff-version.outputs.version }}
    args: check .
```
