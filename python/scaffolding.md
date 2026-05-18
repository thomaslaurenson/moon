# Python Project Scaffolding

Standards and conventions for Python projects. Use this as a reference when creating or refactoring a Python repository.

## Project Structure

### Library project

```
<package>/         # the installable library package
tests/             # pytest test suite
docs/              # Sphinx documentation (libraries only)
tasks/             # operational scripts (if needed alongside the library)
.github/
  workflows/
  dependabot.yml
.gitignore
pyproject.toml
uv.lock
Makefile
CHANGELOG.md
README.md
```

### Scripts-only project

```
tasks/             # all runnable scripts live here
tests/             # pytest test suite
.github/
  workflows/
  dependabot.yml
.gitignore
pyproject.toml
uv.lock
Makefile
CHANGELOG.md
README.md
```

- No `src/` layout - package sits at the repo root
- Scripts live in `tasks/`, never in the package itself
- `data/` may be added for local artefacts; never commit real data

## Tools

| Tool | Purpose |
|---|---|
| `uv` | Package manager - install, sync, lock, build |
| `ruff` | Linter and formatter |
| `pytest` | Test runner |

**Not used:** `pip`, `poetry`, `black`, `isort`, `flake8`, or any other overlapping tools.

## uv

- `uv sync` - install deps from lock file
- `uv sync --all-extras` - install all optional dep groups
- `uv run <cmd>` - run a command in the project environment
- `uv lock --upgrade` - upgrade all locked dependencies
- Never run `pip install` directly

## Common pyproject.toml

### Project metadata

```toml
[project]
name = "<package>"
version = "0.1.0"
description = "..."
authors = [{name = "...", email = "..."}]
readme = "README.md"
requires-python = ">=3.10"
dependencies = []
```

### Ruff configuration

```toml
[tool.ruff]
line-length = 100
target-version = "py310"

[tool.ruff.lint]
select = [
    "E",    # pycodestyle errors
    "W",    # pycodestyle warnings
    "F",    # pyflakes
    "I",    # isort
    "D",    # pydocstyle
    "B",    # flake8-bugbear
    "C4",   # flake8-comprehensions
    "SIM",  # flake8-simplify
    "UP",   # pyupgrade
    "PTH",  # use pathlib over os.path
    "FA",   # enforce from __future__ import annotations
    "RET",  # return statement best practices
]

[tool.ruff.lint.pydocstyle]
convention = "pep257"

[tool.ruff.lint.per-file-ignores]
"docs/**" = ["D100"]  # library projects only; omit for scripts-only projects
"tests/**" = ["D"]
```

### Pytest configuration

```toml
[tool.pytest.ini_options]
addopts = "-v --strict-markers"
testpaths = ["tests"]
markers = [
    "integration: marks tests as integration tests requiring live credentials or environment",
]
```

Never add `--cov` to `addopts`. Coverage is a separate explicit Makefile target (`make test_coverage`).

Never put `uv`, `build`, or `twine` in optional dependency groups. `uv` is a package manager, not a project dependency. `build` and `twine` are invoked via `uv run` in the Makefile.

## Libraries

### Optional dependencies

```toml
[project.optional-dependencies]
test = [
    "pytest>=...",
    "pytest-cov>=...",
    "coverage>=...",
]
dev = [
    "<package>[test]",
    "ruff>=...",
    "pyright>=...",
]
```

Add additional groups for optional integrations (e.g. `nectar`, `vault`). Each group should only pull in what it strictly needs.

### Type checking

Must include `pyright` in the `dev` dep group. Configure it in `pyproject.toml`:

```toml
[tool.pyright]
include = ["<package>"]
```

Never hardcode `pythonVersion` in `[tool.pyright]`. It must infer the version automatically from `requires-python`.

### Build system

```toml
[tool.hatch.build.targets.wheel]
packages = ["<package>"]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```

### Coverage configuration

```toml
[tool.coverage.run]
source = ["<package>"]
omit = ["tasks/**"]

[tool.coverage.report]
show_missing = true
skip_empty = true
```

### Logging

Use the Python standard library `logging` module. Never use `structlog` or any third-party logger in a library. Libraries must not force logging configuration on their consumers.

### Badges

Include the following badges in the README:

| Badge | Source |
|---|---|
| Build state | GitHub Actions `tag.yml` workflow status |
| Release state | GitHub Actions `tag.yml` release status |
| Release version | Latest GitHub release tag |
| Release downloads | GitHub release total download count |
| Python version | Dynamically extracted from `pyproject.toml` `requires-python` |
| Test coverage | Manually-maintained static badge, updated on each release |

### Makefile

Adhere to the global Makefile structure established in `tools/makefile.md`. Use the following commands for your standard targets:

- `build`: `uv build`
- `test`: `uv run pytest -m "not integration"`
- `test_integration`: `uv run pytest -m integration`
- `test_coverage`: `uv run coverage run -m pytest -m "not integration" && uv run coverage report`
- `lint`: `uv run ruff check .`
- `fmt_check`: `uv run ruff format --check .`
- `typecheck`: `uv run pyright`
- `fix`: `uv run ruff check --fix . && uv run ruff format .`
- `ci`: `lint fmt_check typecheck test`

Include the `# GET` section targets from `tools/makefile.md`.

## Apps/Scripts

### Optional dependencies

```toml
[project.optional-dependencies]
test = [
    "pytest>=...",
]
dev = [
    "<package>[test]",
    "ruff>=...",
]
```

No `pytest-cov`, `coverage`, or `pyright` in app/script projects.

### Type checking

Omit `pyright` entirely from app/script projects.

### Build system

Omit build system configuration entirely. App/script projects are not installable packages.

### Logging

Use `structlog` for structured logging. Apps control the full stack and may configure logging as needed.

### Badges

Include only build state badges. Omit coverage and release badges.

### Makefile

Adhere to the global Makefile structure established in `tools/makefile.md`. Use the following commands for your standard targets:

- `test`: `uv run pytest -m "not integration"`
- `test_integration`: `uv run pytest -m integration`
- `lint`: `uv run ruff check .`
- `fmt_check`: `uv run ruff format --check .`
- `fix`: `uv run ruff check --fix . && uv run ruff format .`
- `ci`: `lint fmt_check test`
