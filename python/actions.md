# Python GitHub Actions Workflows

Conventions for structuring GitHub Actions in Python projects.

## Design Principles

- Reusable workflows (`workflow_call`) for all job logic - callers just compose them
- All steps call `make <target>`, never raw `uv` or `python` commands
- Minimal permissions - `contents: read` by default, `contents: write` only for release
- Standard runner: `ubuntu-24.04`
- Python version extracted from `pyproject.toml` at runtime, never hardcoded

## Workflow Files

```
.github/workflows/
  lint.yml      # reusable: ruff check
  test.yml      # reusable: pytest (unit only)
  release.yml   # reusable: build + GitHub Release
  push.yml      # caller: lint on every push
  pr.yml        # caller: lint + test on pull requests
  tag.yml       # caller: lint + test + release on v* tags
```

## Reusable Workflows

### lint.yml

```yaml
name: Lint

on: workflow_call

jobs:
  python_lint:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6

      - name: Extract ruff version
        id: ruff-version
        run: echo "version=$(grep -oP 'ruff>=\K[0-9.]+' pyproject.toml)" >> $GITHUB_OUTPUT

      - uses: chartboost/ruff-action@v1
        with:
          version: ${{ steps.ruff-version.outputs.version }}
          args: check .
```

- Ruff version is pinned to whatever is in `pyproject.toml` - no drift between local and CI
- No Python setup required for lint-only jobs

### test.yml

```yaml
name: Test

on: workflow_call

jobs:
  python_test:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6

      - name: Extract Python version
        id: python-version
        run: echo "version=$(grep -oP 'requires-python.*>=\K[0-9.]+' pyproject.toml)" >> $GITHUB_OUTPUT

      - uses: actions/setup-python@v5
        with:
          python-version: ${{ steps.python-version.outputs.version }}

      - uses: astral-sh/setup-uv@v7

      - run: make install
      - run: make test
```

- Unit tests only - integration tests are never run in CI (require live credentials)
- Python version sourced from `requires-python` in `pyproject.toml`

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
      - uses: actions/checkout@v6

      - name: Read version from pyproject.toml
        id: version
        run: |
          python3 - <<'PY'
          import tomllib, pathlib, os
          data = tomllib.loads(pathlib.Path("pyproject.toml").read_text())
          version = data["project"]["version"]
          with open(os.environ["GITHUB_OUTPUT"], "a") as f:
              f.write(f"value={version}\n")
          PY

      - name: Verify version matches tag
        run: |
          VERSION="${{ steps.version.outputs.value }}"
          TAG="${{ github.ref_name }}"
          EXPECTED="${TAG#v}"
          if [[ "$VERSION" != "$EXPECTED" ]]; then
            echo "ERROR: pyproject.toml version ($VERSION) does not match tag ($TAG)"
            exit 1
          fi

      - uses: astral-sh/setup-uv@v7

      - run: make build

      - name: Extract changelog section
        run: |
          VERSION="${{ steps.version.outputs.value }}"
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

      - uses: softprops/action-gh-release@v2
        with:
          body_path: /tmp/release_notes.md
          files: dist/*
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- Fails explicitly if `pyproject.toml` version does not match the tag
- Fails explicitly if no CHANGELOG entry exists for the version
- Built artifacts from `dist/` are attached to the release

## Caller Workflows

### push.yml

```yaml
name: Push

on:
  push

jobs:
  lint:
    uses: ./.github/workflows/lint.yml
    secrets: inherit
```

### pr.yml

```yaml
name: PR

on:
  pull_request

jobs:
  lint:
    uses: ./.github/workflows/lint.yml
    secrets: inherit

  test:
    uses: ./.github/workflows/test.yml
    secrets: inherit
```

### tag.yml

```yaml
name: Tag

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  lint:
    uses: ./.github/workflows/lint.yml
    secrets: inherit

  test:
    uses: ./.github/workflows/test.yml
    secrets: inherit

  release:
    needs: [lint, test]
    uses: ./.github/workflows/release.yml
    secrets: inherit
```

- Release only runs after both lint and test pass
- `contents: write` is declared at the caller level for the release job
