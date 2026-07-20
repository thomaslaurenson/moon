# Python Makefile targets

Targets common to every Python project (see the Makefile conventions fragment for structure):

- `test`: `uv run pytest -m "not integration"`
- `test_integration`: `uv run pytest -m integration`
- `lint`: `uv run ruff check .`
- `fmt_check`: `uv run ruff format --check .`
- `fix`: `uv run ruff check --fix . && uv run ruff format .`
- `get_ruff_version`: `grep -oP 'ruff>=\K[0-9.]+' pyproject.toml`
- `get_python_required_version`: `grep -oP 'requires-python\s*=\s*">=\K[0-9.]+' pyproject.toml`
- `get_changelog`: print the `CHANGELOG.md` section for a release, given `TAG`. Tags are `v`-prefixed (`v1.2.3`) but changelog headers are bare (`## 1.2.3 - ...`, see `github/changelog.md`), so the target strips a leading `v` from `TAG` before matching. It exits non-zero when `TAG` is empty or no entry matches, so a release never publishes empty notes:

```make
get_changelog:
	@test -n "$(TAG)" || { echo "TAG is required" >&2; exit 2; }
	@awk -v raw="$(TAG)" '\
	  BEGIN { v = raw; sub(/^v/, "", v) } \
	  /^## / { if (found) exit; if ($$2 == v) { found = 1; print; next } } \
	  found { print } \
	  END { if (!found) exit 1 }' CHANGELOG.md
```

`uv run` syncs the project and its default `dev` dependency group before running, and `dev` includes the `test` group (see the project fragment), so `make test` works on a fresh clone with no separate install step. `ruff` is likewise in `dev`, so `lint`, `fmt_check`, and `fix` need no install either.

The `get_ruff_version` and `get_python_required_version` targets use GNU grep's `-oP` (Perl regex); it is present on the `ubuntu-24.04` runners CI uses. On macOS, where BSD grep lacks `-P`, run these targets in CI (Linux) only, not as a local prerequisite of `test`. `get_changelog` uses only POSIX `awk` and is portable. All three are `# GET` targets: workflows call them rather than embedding raw extraction logic (see `github/actions.md`).

The `ci` target and any build, coverage, or type-check targets are project-tier specific and defined in the project fragment.
