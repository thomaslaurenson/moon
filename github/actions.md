# GitHub Actions Workflows

## Contents

- [Core rules](#core-rules) — non-negotiable conventions that apply to all workflows
- [Workflow structure](#workflow-structure) — reusable vs caller pattern, file layout
- [CHANGELOG extraction](#changelog-extraction) — canonical awk snippet used in all release workflows
- [Go-specific setup](#go-specific-setup) — setup-go, go-version-file
- [Python-specific setup](#python-specific-setup) — setup-uv, extract python version
- [Reference: tag.yml](#reference-tagyml) — full example of the most complex caller workflow

## Core Rules

- Runner: always `ubuntu-24.04`. Never `ubuntu-latest` or any unversioned alias.
- Always prefer official GitHub actions (`actions/*`) before third-party alternatives.
- Always use the `gh` CLI for creating releases. Never use a third-party release action.
- All workflow steps call `make <target>`. Never run raw commands directly in a workflow.
- Minimal permissions: `contents: read` by default, `contents: write` only in release workflows.
- No `fetch-depth: 0` except in release workflows where GoReleaser or CHANGELOG extraction requires it.

## Workflow Structure

Use reusable workflows (`workflow_call`) for all job logic. Caller workflows compose them.

```
.github/workflows/
  lint.yml      # reusable: linting and format checks
  test.yml      # reusable: run tests
  release.yml   # reusable: build + create GitHub release
  pr.yml        # caller: lint + test on pull requests
  tag.yml       # caller: lint + test + release on v* tags
```

No `push.yml` for Go projects. Python projects may add a `push.yml` that runs lint only.

Concurrency on `pr.yml`: always add a concurrency group to cancel stale runs when new
commits are pushed to a PR.

```yaml
concurrency:
  group: pr-${{ github.event.pull_request.number }}
  cancel-in-progress: true
```

## CHANGELOG Extraction

Use this exact awk snippet in all release workflows. It extracts the section matching the
current tag from `CHANGELOG.md` and fails explicitly if no entry is found.

For Go projects the tag is read from `GITHUB_REF_NAME` directly:

```bash
TAG="${GITHUB_REF_NAME}"
awk "/^## ${TAG} /{found=1; next} found && /^## /{exit} found{print}" CHANGELOG.md \
  > /tmp/release-notes.md
if [ ! -s /tmp/release-notes.md ]; then
  echo "No CHANGELOG entry found for ${TAG}" >&2
  exit 1
fi
```

For Python projects the version is read from `pyproject.toml` first, then matched:

```bash
VERSION=$(python3 -c "import tomllib, pathlib; print(tomllib.loads(pathlib.Path('pyproject.toml').read_text())['project']['version'])")
awk -v ver="$VERSION" '
  /^## / {
    if (found) exit
    if (index($0, "## " ver " ") || $0 == "## " ver) { found=1 }
    next
  }
  found { lines[n++] = $0 }
  END {
    start=0
    while (start < n && lines[start] ~ /^[[:space:]]*$/) start++
    end=n-1
    while (end >= start && lines[end] ~ /^[[:space:]]*$/) end--
    for (i=start; i<=end; i++) print lines[i]
  }
' CHANGELOG.md > /tmp/release_notes.md
if [ ! -s /tmp/release_notes.md ]; then
  echo "No CHANGELOG entry found for $VERSION" >&2; exit 1
fi
```

Pass the output file to `gh release create` with `--notes-file`.

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
  run: echo "version=$(grep -oP 'requires-python.*>=\K[0-9.]+' pyproject.toml)" >> $GITHUB_OUTPUT

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
