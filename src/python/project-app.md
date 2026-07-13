# Python application project

Scripts-only project layout (no installable package):

```
tasks/             # all runnable scripts live here
tests/
.github/
  workflows/
  dependabot.yml
pyproject.toml
uv.lock
Makefile
CHANGELOG.md
README.md
```

- No `src/` layout. Scripts live in `tasks/`, never in a package.
- `data/` may hold local artefacts; never commit real data.

Development dependencies go in PEP 735 dependency groups, not optional-dependencies. Extras (`[project.optional-dependencies]`) are published, user-facing installs; an app is not installable, so a self-referential `<package>[test]` extra would be wrong (and risks resolving `<package>` from PyPI). Dependency groups are local-only and never published, which is exactly what dev tooling is:

```toml
[dependency-groups]
test = ["pytest>=..."]
dev = [{ include-group = "test" }, "ruff>=..."]
```

- `dev` includes `test`, and `uv sync` installs `dev` by default, so a fresh clone has pytest and ruff without extra flags.
- Omit build-system configuration; app projects are not installable.
- Omit `pyright` and coverage entirely.
- Badges: build state only.

Makefile targets (in addition to the common Python targets): app projects add no build, coverage, or type-check targets.

- `ci`: `lint fmt_check test`
