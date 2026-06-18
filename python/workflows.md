# Python Workflow Conventions

Supplements `github/actions.md`. Universal rules (runners, action versions, workflow structure, caller patterns, concurrency, permissions) apply unchanged. This file covers Python-specific setup steps only.

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

## Reusable Workflow Bodies

### `lint.yml`

```yaml
name: Lint

on:
  workflow_call

permissions:
  contents: read

jobs:
  python_lint:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6

      - name: Extract ruff version
        id: ruff-version
        run: echo "version=$(grep -oP 'ruff>=\K[0-9.]+' pyproject.toml)" >> $GITHUB_OUTPUT

      - uses: astral-sh/ruff-action@v3
        with:
          version: ${{ steps.ruff-version.outputs.version }}
          args: check .

      - uses: astral-sh/ruff-action@v3
        with:
          version: ${{ steps.ruff-version.outputs.version }}
          args: format --check .
```

### `test.yml`

```yaml
name: Test

on:
  workflow_call

permissions:
  contents: read

jobs:
  python_test:
    runs-on: ubuntu-24.04
    steps:
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

### `prerelease.yml`

Creates a rolling dev prerelease on every push to `main`. Replaces the previous dev release each time.

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
      - uses: actions/checkout@v6

      - name: Extract Python version
        id: python-version
        run: echo "version=$(make get_python_required_version)" >> $GITHUB_OUTPUT

      - uses: actions/setup-python@v5
        with:
          python-version: ${{ steps.python-version.outputs.version }}

      - uses: astral-sh/setup-uv@v7

      - run: uv build

      - name: Delete existing dev release
        run: |
          if gh release view "dev" > /dev/null 2>&1; then
            gh release delete "dev" --yes --cleanup-tag
          fi
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Create dev prerelease
        run: |
          gh release create "dev" \
            --title "Dev (Pre-release)" \
            --prerelease \
            --notes "[${{ github.sha }}](${{ github.server_url }}/${{ github.repository }}/commit/${{ github.sha }})" \
            dist/*
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### `release.yml`

Triggered by `tag.yml` on `v*.*.*` tags. Extracts release notes from `CHANGELOG.md` and publishes a GitHub release.

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
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: Extract Python version
        id: python-version
        run: echo "version=$(make get_python_required_version)" >> $GITHUB_OUTPUT

      - uses: actions/setup-python@v5
        with:
          python-version: ${{ steps.python-version.outputs.version }}

      - uses: astral-sh/setup-uv@v7

      - run: uv build

      - name: Extract release notes from CHANGELOG.md
        run: make get_changelog TAG=${GITHUB_REF_NAME} > /tmp/release-notes.md

      - name: Create GitHub release
        run: |
          gh release create "${{ github.ref_name }}" \
            --title "${{ github.ref_name }}" \
            --notes-file /tmp/release-notes.md \
            dist/*
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```
