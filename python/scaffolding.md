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

## pyproject.toml

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

### Optional dep groups

Group by concern. Always define `test` and `dev` at minimum:

```toml
[project.optional-dependencies]
test = [
    "pytest>=...",
    "coverage>=...",
]
dev = [
    "<package>[test]",
    "ruff>=...",
]
```

Add additional groups for optional integrations (e.g. `nectar`, `vault`). Each group should only pull in what it strictly needs.

### Ruff configuration

```toml
[tool.ruff]
line-length = 100
target-version = "py310"

[tool.ruff.lint]
select = [
    "E",  # pycodestyle errors
    "W",  # pycodestyle warnings
    "F",  # pyflakes
    "I",  # isort
    "D",  # pydocstyle
    "B",  # flake8-bugbear
]

[tool.ruff.lint.per-file-ignores]
"tasks/**" = ["D100", "D103", "E402"]
"docs/**" = ["D100"]
"tests/**" = ["D"]

[tool.ruff.lint.pydocstyle]
convention = "google"
```

### Pytest configuration

```toml
[tool.pytest.ini_options]
addopts = "-v"
testpaths = ["tests"]
markers = [
    "integration: mark test as integration test (requires live API credentials)",
]
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

### Build backend

```toml
[tool.hatch.build.targets.wheel]
packages = ["<package>"]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```

## Makefile

All CI steps are Makefile targets. GitHub Actions call `make <target>` - never raw `uv` or `python` commands directly in workflows.

Key targets:

```makefile
install:        uv sync
install_all:    uv sync --all-extras
update:         uv lock --upgrade && uv sync --all-extras
clean:          rm -rf .venv dist .pytest_cache .coverage + __pycache__
build:          uv build
lint:           uv run ruff check .
format_check:   uv run ruff format --check .
fix:            uv run ruff check --fix . && uv run ruff format .
test:           uv run pytest -m "not integration"
test_integration: uv run pytest -m integration
test_coverage:  uv run coverage run -m pytest -m "not integration" && uv run coverage report
```

Rules:
- `fix` is the one-shot local cleanup command (lint-fix + format)
- `test` always excludes integration tests - they require live credentials
- CI calls `make lint` and `make test`, never the underlying commands directly
