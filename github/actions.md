# GitHub Actions Workflows

## Core Rules

- Always prefer official GitHub actions (`actions/*`) before third-party alternatives.
- Use the `gh` CLI for creating releases by default. `goreleaser-action` is the only
  permitted exception, and only for Go projects; see `golang/workflows.md`.
- All workflow steps call `make <target>`. Never write raw multi-line bash, awk, or Python
  scripts inline in a workflow step. Simple one-line commands (e.g. `sudo apt-get install -y
  shellcheck`) are acceptable as raw `run:` commands when there is no corresponding Makefile
  target.
- Never write raw bash or awk scripts in workflows to extract versions or changelogs. Use
  the repository's Makefile targets (e.g. `make get_changelog TAG=v1.0.0`) and pass their
  standard output to the respective workflow steps.
- Minimal permissions: `contents: read` by default, `contents: write` only in release and
  prerelease workflows. Test workflows must never declare `contents: write`.
- No `fetch-depth: 0` except in release workflows where changelog extraction requires it.

---

## Runners

Always pin runners to a specific version. Never use unversioned aliases such as
`ubuntu-latest` or `macos-latest`.

| OS | Supported pinned versions |
|---|---|
| Linux | `ubuntu-24.04`, `ubuntu-24.04-arm` |
| macOS | `macos-14`, `macos-15` |
| Windows | `windows-2022`, `windows-2025` |

When GitHub releases a new runner version, update the pin deliberately. Do not rely on
`-latest` aliases to update automatically.

---

## Action Versions

Always use these pinned versions. Never use `@latest` or unversioned aliases:

| Action | Version |
|---|---|
| `actions/checkout` | `v6` |
| `actions/upload-artifact` | `v6` |
| `actions/download-artifact` | `v6` |
| `actions/setup-go` | `v6` |
| `actions/setup-python` | `v5` |
| `astral-sh/setup-uv` | `v7` |

---

## Workflow Structure

Use reusable workflows (`workflow_call`) for all job logic. Caller workflows compose them.

```
.github/workflows/
  lint.yml        # reusable: linting and format checks
  test.yml        # reusable: run tests
  release.yml     # reusable: create GitHub release
  prerelease.yml  # reusable: create or replace rolling dev prerelease
  pr.yml          # caller: lint + test on pull requests
  tag.yml         # caller: lint + test + release on v* tags
  main.yml        # caller: lint + test + prerelease on push to main
```

Languages with a compilation step that must run before lint and test (e.g. C++) add a
`build.yml` reusable workflow and a `needs: build` dependency in the callers. See
`cpp/workflows.md`.

No `push.yml` for any project. Use `pr.yml` for pull requests, `tag.yml` for versioned
releases, and `main.yml` for rolling dev prereleases.

Concurrency on `pr.yml`: always add a concurrency group to cancel stale runs when new
commits are pushed to a PR:

```yaml
concurrency:
  group: pr-${{ github.event.pull_request.number }}
  cancel-in-progress: true
```

Concurrency on `main.yml`: always add a concurrency group to cancel in-progress prerelease
runs when new commits land on main:

```yaml
concurrency:
  group: prerelease
  cancel-in-progress: true
```

---

## Caller Workflow: pr.yml

Triggers on pull requests. Runs lint and test. Uses a `paths:` filter so the workflow only
fires on relevant changes.

```yaml
name: Pull Request

on:
  pull_request:
    types: [opened, synchronize, reopened]
    paths:
      - ".github/workflows/**"
      - "Makefile"
      # language-specific paths (see table below)

concurrency:
  group: pr-${{ github.event.pull_request.number }}
  cancel-in-progress: true

permissions:
  contents: read

jobs:
  lint:
    uses: ./.github/workflows/lint.yml
  test:
    uses: ./.github/workflows/test.yml
```

### Common paths entries by language

Include only the entries that apply to the project:

| Entry | When to include |
|---|---|
| `.github/workflows/**` | Always; workflow changes must trigger CI |
| `Makefile` | Always; Makefile changes affect all targets |
| `**.go` | Go projects |
| `go.mod` | Go projects |
| `go.sum` | Go projects |
| `.goreleaser*.yml` | Go projects |
| `src/**` | C++ projects |
| `test/**` | C++ projects |
| `.clang-format` | C++ projects |
| `.clang-tidy` | C++ projects |
| `CMakeLists.txt` | C++ projects |
| `Dockerfile*` | C++ projects that build via Docker |
| `pyproject.toml` | Python projects |

The `paths:` filter on `main.yml` must match `pr.yml` exactly; the same changes that trigger a PR check should trigger a prerelease when merged.

---

## Caller Workflow: tag.yml

Triggers on `v*.*.*` tags. Runs lint and test, then release. Release only runs after both
pass.

```yaml
name: Tag and Release

on:
  push:
    tags:
      - "v*.*.*"

permissions:
  contents: write
  id-token: write

jobs:
  lint:
    uses: ./.github/workflows/lint.yml
  test:
    uses: ./.github/workflows/test.yml
  release:
    uses: ./.github/workflows/release.yml
    needs: [lint, test]
    secrets: inherit
```

- `permissions: contents: write` declared at the caller level.
- `id-token: write` required when the release workflow signs artifacts with cosign.
- `secrets: inherit` passes `GITHUB_TOKEN` to the release workflow.

---

## Caller Workflow: main.yml

Triggers on pushes to `main`. Mirrors `tag.yml` but calls `prerelease.yml` instead of
`release.yml`.

```yaml
name: Main

on:
  push:
    branches:
      - "main"
    paths:
      - ".github/workflows/**"
      - "Makefile"
      # language-specific paths (same as pr.yml)

concurrency:
  group: prerelease
  cancel-in-progress: true

permissions:
  contents: write
  id-token: write

jobs:
  lint:
    uses: ./.github/workflows/lint.yml
  test:
    uses: ./.github/workflows/test.yml
  prerelease:
    uses: ./.github/workflows/prerelease.yml
    needs: [lint, test]
    secrets: inherit
```

---

## Prerelease Workflow: prerelease.yml

The default prerelease pattern uses the `gh` CLI to create or replace a static `dev`
release on every passing main build. Go projects use goreleaser instead; see `golang/workflows.md`.

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

      # prepare artifacts (language-specific: build, download, etc.)

      - name: Delete existing dev release
        run: gh release delete dev --yes --cleanup-tag || true
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Create dev prerelease
        run: |
          gh release create dev \
            --title "Dev (Pre-release)" \
            --prerelease \
            --notes "**Commit:** ${{ github.sha }}" \
            <artifact files>
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- The `|| true` on the delete step is intentional; it prevents failure when no `dev` release exists yet (first run). The tag name is always the static string `dev`, so there is no ambiguity about which release is being deleted.
- Release notes contain only the commit SHA. Full changelog entries are for versioned
  releases only.
