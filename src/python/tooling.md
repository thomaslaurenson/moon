# Python Tooling

| Tool | Purpose |
|---|---|
| `uv` | Package manager: install, sync, lock, build |
| `ruff` | Linter and formatter |
| `pytest` | Test runner |

Not used: `pip`, `poetry`, `black`, `isort`, `flake8`, or any overlapping tool. Never run `pip install` directly; use `uv sync`, `uv run <cmd>`, `uv lock --upgrade`.

```toml
[tool.ruff]
line-length = 100
target-version = "py310"

[tool.ruff.lint]
select = ["E", "W", "F", "I", "D", "B", "C4", "SIM", "UP", "PTH", "FA", "RET"]

[tool.ruff.lint.pydocstyle]
convention = "pep257"

[tool.ruff.lint.per-file-ignores]
"tests/**" = ["D"]
```

```toml
[tool.pytest.ini_options]
addopts = "-v --strict-markers"
testpaths = ["tests"]
markers = [
    "integration: marks tests as integration tests requiring live credentials or environment",
]
```

Never add `--cov` to `addopts`; it slows every test run and coverage is a separate step. Never put `uv`, `build`, or `twine` in optional dependency groups.
