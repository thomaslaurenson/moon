# GitHub Actions Workflows

## Core Rules

- Runner: always `ubuntu-24.04`. Never `ubuntu-latest` or any unversioned alias.
- Always prefer official GitHub actions (`actions/*`) before third-party alternatives.
- Always use the `gh` CLI for creating releases. Never use a third-party release action.
- All workflow steps call `make <target>`. Never run raw commands directly in a workflow.
- Never write raw bash or awk scripts in workflows to extract versions or changelogs. Execute the repository's Makefile targets (e.g., `make get_python_project_version`, `make get_changelog_entry`) and pass their standard output to the respective workflow steps.
- Minimal permissions: `contents: read` by default, `contents: write` only in release workflows.
- No `fetch-depth: 0` except in release workflows where GoReleaser or CHANGELOG extraction requires it.

## Workflow Structure

Use reusable workflows (`workflow_call`) for all job logic. Caller workflows compose them.

```
.github/workflows/
  lint.yml      # reusable: linting, format checks, type checking
  test.yml      # reusable: run tests (matrix across Python versions for Python projects)
  release.yml   # reusable: build + create GitHub release
  pr.yml        # caller: lint + test on pull requests
  tag.yml       # caller: lint + test + release on v* tags
```

No `push.yml` for any project. Use `pr.yml` to validate on pull requests and `tag.yml`
to validate and release on tags.

Concurrency on `pr.yml`: always add a concurrency group to cancel stale runs when new
commits are pushed to a PR.

```yaml
concurrency:
  group: pr-${{ github.event.pull_request.number }}
  cancel-in-progress: true
```

## Go-Specific Setup

Add these steps before any `make` call in Go workflows:

```yaml
- uses: actions/checkout@v4

- uses: actions/setup-go@v6
  with:
    go-version-file: go.mod
    cache: true
```

- Always use `go-version-file: go.mod`. Never hardcode a Go version.
- `fetch-depth: 0` only in `release.yml`, not in lint or test workflows.

## Python-Specific Setup

Extract the Python version from `pyproject.toml` at runtime. Never hardcode it.

```yaml
- uses: actions/checkout@v4

- name: Extract Python version
  id: python-version
  run: echo "version=$(make get_python_required_version)" >> $GITHUB_OUTPUT

- uses: actions/setup-python@v5
  with:
    python-version: ${{ steps.python-version.outputs.version }}

- uses: astral-sh/setup-uv@v7
```

For lint-only jobs, extract the ruff version and use the ruff action directly — no Python
setup required:

```yaml
- name: Extract ruff version
  id: ruff-version
  run: echo "version=$(grep -oP 'ruff>=\K[0-9.]+' pyproject.toml)" >> $GITHUB_OUTPUT

- uses: chartboost/ruff-action@v1
  with:
    version: ${{ steps.ruff-version.outputs.version }}
    args: check .
```

## Reference: tag.yml

Full example of the most complex caller workflow. Both Go and Python projects follow
this same structure.

```yaml
name: Tag

on:
  push:
    tags: ["v*"]

permissions:
  contents: write

jobs:
  lint:
    uses: ./.github/workflows/lint.yml

  test:
    uses: ./.github/workflows/test.yml

  release:
    needs: [lint, test]
    uses: ./.github/workflows/release.yml
    secrets: inherit
```

- Release only runs after both lint and test pass.
- `secrets: inherit` passes `GITHUB_TOKEN` to the release workflow.
- `permissions: contents: write` is declared at the caller level.
