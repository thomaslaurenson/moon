# Python Makefile Targets

Targets common to every Python project (see the Makefile conventions fragment for structure):

- `test`: `uv run pytest -m "not integration"`
- `test_integration`: `uv run pytest -m integration`
- `lint`: `uv run ruff check .`
- `fmt_check`: `uv run ruff format --check .`
- `fix`: `uv run ruff check --fix . && uv run ruff format .`
- `get_ruff_version`: `grep -oP 'ruff>=\K[0-9.]+' pyproject.toml`

The `ci` target and any build, coverage, or type-check targets are project-tier specific and defined in the project fragment.
