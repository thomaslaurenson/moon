# Python tools project

A managed set of operational scripts that consume a library. Not an installable package: no `[build-system]`, no console entry point. If the project ships an installed command, it is a library-with-CLI (`python-lib-cli`), not a tools project.

Layout - scripts grouped by domain, not a single `tasks/` dump:

```
<area>/            # runnable scripts grouped by what they act on (e.g. client/, patches/)
<area>/
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

- No `src/` layout and no installable package. Scripts live in domain-named directories, named for what they operate on, not in a single catch-all `tasks/`.
- `data/` may hold local artefacts; never commit real data.
- Scripts print to stdout for their output; they do not configure a logging stack.

Development dependencies go in PEP 735 dependency groups, not optional-dependencies. Extras (`[project.optional-dependencies]`) are published, user-facing installs; a tools project is not installable, so a self-referential `<package>[test]` extra would be wrong (and risks resolving `<package>` from PyPI). Dependency groups are local-only and never published, and they need no `[build-system]`, which is exactly what a non-installable project wants:

```toml
[dependency-groups]
test = ["pytest>=..."]
dev = [{ include-group = "test" }, "ruff>=..."]
```

- `dev` includes `test`, and `uv sync` installs `dev` by default, so a fresh clone has pytest and ruff without extra flags.
- Omit build-system configuration; a tools project is not installable.
- Omit `pyright` and coverage entirely.
- Badges: build state only.

Makefile targets (in addition to the common Python targets): a tools project adds no build, coverage, or type-check targets.

- `ci`: `lint fmt_check test`
