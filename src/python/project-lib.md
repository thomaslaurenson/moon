# Python Library Project

Installable library layout:

```
<package>/         # the installable library package
tests/
docs/              # Sphinx documentation
tasks/             # operational scripts (if needed)
.github/
  workflows/
  dependabot.yml
pyproject.toml
uv.lock
Makefile
CHANGELOG.md
README.md
```

- No `src/` layout; the package sits at the repo root.

## CLI entry points (optional)

A library may also ship a console script:

```toml
[project.scripts]
mycli = "<package>.cli.__main__:main"
```

CLI code lives inside the package (e.g. `<package>/cli/`), not in `tasks/`. `tasks/` is for standalone operational scripts that are not installed or importable by consumers.

Optional dependencies:

```toml
[project.optional-dependencies]
test = ["pytest>=...", "pytest-cov>=...", "coverage>=..."]
dev  = ["<package>[test]", "ruff>=...", "pyright>=..."]
```

Add further groups for optional integrations; each pulls in only what it needs.

Badges: build state, release state, release version, release downloads, Python version (from `requires-python`), test coverage (static, updated per release).

Makefile targets (in addition to the common Python targets):

- `build`: `uv build`
- `test_coverage`: `uv run coverage run -m pytest -m "not integration" && uv run coverage report`
- `typecheck`: `uv run pyright`
- `ci`: `lint fmt_check typecheck test`
