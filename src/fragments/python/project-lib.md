# Python library project

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

## Dependencies

A library distinguishes two kinds of non-runtime dependency, and they live in different tables:

- User-facing optional integrations go in `[project.optional-dependencies]` (extras). These are published with the package; a consumer installs them with `pip install <package>[vault]`.
- Development tooling (tests, linters, type checker, coverage) goes in PEP 735 `[dependency-groups]`. These are local-only and never published, so they never leak into a consumer's install.

```toml
[project.optional-dependencies]
# published extras: optional integrations a consumer can opt into
vault = ["hvac>=..."]

[dependency-groups]
test = ["pytest>=...", "pytest-cov>=...", "coverage>=..."]
docs = ["sphinx>=...", "furo>=..."]
dev  = [{ include-group = "test" }, { include-group = "docs" }, "ruff>=...", "pyright>=..."]
```

- `uv sync` installs the project plus the default `dev` group (which includes `test` and `docs`), so a fresh clone can run tests, lints, and the docs build immediately.
- `uv sync --all-extras` additionally installs the published extras, which tests that exercise optional integrations need.
- Add further extras for each optional integration; each pulls in only what it needs.

Badges: build state, release state, release version, release downloads, Python version (from `requires-python`), test coverage (static, updated per release).

Makefile targets (in addition to the common Python targets):

- `build`: `uv build`
- `test_coverage`: `uv run coverage run -m pytest -m "not integration" && uv run coverage report`
- `typecheck`: `uv run pyright`
- `ci`: `lint fmt_check typecheck test`
