# GitHub Actions Workflows

Conventions for structuring GitHub Actions workflows across any project type.

## Design Principles

- Reusable workflows (`workflow_call`) for all job logic - callers just compose them
- All steps call `make <target>`, never raw commands directly
- Minimal permissions - `contents: read` by default, `contents: write` only for release
- Standard runner: `ubuntu-24.04` - always pin the version, never use `ubuntu-latest`
- Prefer official GitHub actions (`actions/*`) over third-party actions
- Releases use the `gh` CLI (pre-installed on all runners) - no third-party release actions
- Workflow file names reflect their trigger: `pr.yml` for pull requests, `tag.yml` for tags

## Workflow Files

```
.github/workflows/
  lint.yml      # reusable: project linting/formatting checks
  test.yml      # reusable: run tests
  release.yml   # reusable: create GitHub release via gh CLI
  pr.yml        # caller: lint + test on pull requests
  tag.yml       # caller: lint + test + release on v* tags
```

No `push.yml` - linting and testing are only triggered by PRs and tags.

## Reusable Workflows

### lint.yml

```yaml
name: Lint

on: workflow_call

permissions:
  contents: read

jobs:
  lint:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4

      - run: make lint
```

- Steps use `make` targets, not raw commands
- Add language-specific setup steps (e.g. `actions/setup-node`, `actions/setup-python`) before the `make` call if needed

### test.yml

```yaml
name: Test

on: workflow_call

permissions:
  contents: read

jobs:
  test:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4

      - run: make test
```

- No `fetch-depth: 0` - not needed for tests
- Add language-specific setup steps before the `make` call if needed

### release.yml

```yaml
name: Release

on: workflow_call

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # required to read full git history for release context

      - name: Extract release notes from CHANGELOG.md
        run: |
          TAG="${GITHUB_REF_NAME}"
          awk "/^## ${TAG} /{found=1; next} found && /^## /{exit} found{print}" CHANGELOG.md \
            > /tmp/release-notes.md
          if [ ! -s /tmp/release-notes.md ]; then
            echo "No CHANGELOG entry found for ${TAG}" >&2
            exit 1
          fi

      - name: Build release artifacts
        run: make build

      - name: Create GitHub release
        run: |
          gh release create "${GITHUB_REF_NAME}" \
            --title "${GITHUB_REF_NAME}" \
            --notes-file /tmp/release-notes.md \
            bin/*
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- `fetch-depth: 0` required to ensure full git history is available
- Release notes are extracted from `CHANGELOG.md` matching the tag name - fails explicitly if no entry is found
- `gh release create` attaches build artifacts from `bin/`; adjust the glob to match your project's output
- `gh` is pre-installed on all GitHub-hosted runners - no third-party release action needed

## Caller Workflows

### pr.yml

```yaml
name: PR

on:
  pull_request:
    types: [opened, synchronize, reopened]

permissions:
  contents: read

concurrency:
  group: pr-${{ github.event.pull_request.number }}
  cancel-in-progress: true

jobs:
  lint:
    uses: ./.github/workflows/lint.yml

  test:
    uses: ./.github/workflows/test.yml
```

- `concurrency` cancels stale runs when new commits are pushed to a PR

### tag.yml

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

- Release only runs after both lint and test pass
- `secrets: inherit` passes `GITHUB_TOKEN` to the release workflow

## Conventions

- Always pin runner versions: `ubuntu-24.04`, not `ubuntu-latest`
- Prefer `actions/checkout`, `actions/setup-*` over third-party equivalents
- `fetch-depth: 0` only in `release.yml`, not in lint or test workflows
- `permissions: contents: write` only in `release.yml` and `tag.yml`
- Concurrency group on `pr.yml` to cancel stale runs
- No workflow runs on direct branch pushes - only PRs and tags
- All commands go through `make` targets - workflows stay language-agnostic

## CHANGELOG Format

The release workflow expects `CHANGELOG.md` entries in this format:

```markdown
## v1.2.3

### Added

- Add feature X

### Fixed

- Fix bug Y
```

The tag name (e.g. `v1.2.3`) must match the `## <tag>` heading exactly. The release workflow fails if no matching entry is found.
