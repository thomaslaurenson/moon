# Python Makefile targets

Targets common to every Python project (see the Makefile conventions fragment for structure):

- `test`: `uv run pytest -m "not integration"`
- `test_integration`: `uv run pytest -m integration`
- `lint`: `uv run ruff check .`
- `fmt_check`: `uv run ruff format --check .`
- `fix`: `uv run ruff check --fix . && uv run ruff format .`
- `get_ruff_version`: `grep -oP 'ruff>=\K[0-9.]+' pyproject.toml`
- `get_python_required_version`: `grep -oP 'requires-python\s*=\s*">=\K[0-9.]+' pyproject.toml`

`uv run` syncs the project and its default `dev` dependency group before running, and `dev` includes the `test` group (see the project fragment), so `make test` works on a fresh clone with no separate install step. `ruff` is likewise in `dev`, so `lint`, `fmt_check`, and `fix` need no install either.

Both `get_*` targets use GNU grep's `-oP` (Perl regex); it is present on the `ubuntu-24.04` runners CI uses. On macOS, where BSD grep lacks `-P`, run these targets in CI (Linux) only, not as a local prerequisite of `test`.

The `ci` target and any build, coverage, or type-check targets are project-tier specific and defined in the project fragment.
