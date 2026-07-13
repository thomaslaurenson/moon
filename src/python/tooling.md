# Python tooling

| Tool | Purpose |
|---|---|
| `uv` | Package manager: install, sync, lock, build |
| `ruff` | Linter and formatter |
| `pytest` | Test runner |

Not used: `pip`, `poetry`, `black`, `isort`, `flake8`, or any overlapping tool. Never run `pip install` directly; use `uv sync`, `uv run <cmd>`, `uv lock --upgrade`.

```toml
[tool.ruff]
line-length = 100

[tool.ruff.lint]
select = ["E", "W", "F", "I", "B", "C4", "SIM", "UP", "PTH", "FA", "RET"]
```

Do not set `target-version`; ruff infers it from `requires-python`. Setting it duplicates a fact that already lives in one place and drifts from it.

Docstring linting (the `D` rule family and its pydocstyle convention) is not in this shared baseline. It belongs only to tiers that mandate docstrings; the docstrings fragment adds it. Application and script tiers, which do not require docstrings, must not enable `D` or they will fail lint on docstrings their own conventions never asked for.

```toml
[tool.pytest.ini_options]
addopts = "-v --strict-markers"
testpaths = ["tests"]
markers = [
    "integration: marks tests as integration tests requiring live credentials or environment",
]
```

Never add `--cov` to `addopts`; it slows every test run and coverage is a separate step. Never put `uv`, `build`, or `twine` in dependency groups or optional dependencies.
