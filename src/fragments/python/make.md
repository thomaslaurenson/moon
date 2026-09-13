# Python Makefile targets

Targets common to every Python project (see the Makefile conventions fragment for structure):

- `test`: `uv run pytest -m "not integration"`
- `test_integration`: `uv run pytest -m integration`
- `check_lint`: `uv run ruff check .`
- `check_format`: `uv run ruff format --check .`
- `fix`: `uv run ruff check --fix . && uv run ruff format .`
- `get_ruff_version`: `grep -oP 'ruff>=\K[0-9.]+' pyproject.toml`
- `get_python_required_version`: `grep -oP 'requires-python\s*=\s*">=\K[0-9.]+' pyproject.toml`
- `get_changelog`: the shared recipe in the Makefile conventions fragment.

`uv run` syncs the project and its default `dev` dependency group before running, and `dev` includes the `test` group (see the project fragment), so `make test` works on a fresh clone with no separate install step. `ruff` is likewise in `dev`, so `check_lint`, `check_format`, and `fix` need no install either.

The `get_ruff_version` and `get_python_required_version` targets use GNU grep's `-oP` (Perl regex); it is present on the `ubuntu-24.04` runners CI uses. On macOS, where BSD grep lacks `-P`, run these targets in CI (Linux) only, not as a local prerequisite of `test`. `get_changelog` uses only POSIX `awk` and is portable. All three are `##@ GET` targets: workflows call them rather than embedding raw extraction logic (see `github/actions`).

The `ci` target and any build, coverage, or type-check targets are project-tier specific and defined in the project fragment.
