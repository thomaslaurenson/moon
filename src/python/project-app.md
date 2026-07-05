# Python Application Project

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

Optional dependencies (no coverage, no pyright):

```toml
[project.optional-dependencies]
test = ["pytest>=..."]
dev = ["<package>[test]", "ruff>=..."]
```

- Omit build-system configuration; app projects are not installable.
- Omit `pyright` entirely.
- Badges: build state only.

Makefile targets (in addition to the common Python targets): app projects add no build, coverage, or type-check targets.

- `ci`: `lint fmt_check test`
