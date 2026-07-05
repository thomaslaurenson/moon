# Python Type Checking

All library projects include pyright.

```toml
[tool.pyright]
include = ["<package>"]
exclude = ["tests"]
typeCheckingMode = "basic"
```

- Do not set `pythonVersion`; pyright infers it from `requires-python`.
- Add `pyright` to the `dev` optional dep group and a `typecheck` Makefile target (`uv run pyright`) included in `ci`.
- Start at `basic`; move to `strict` once fully annotated and clean.
- `tests/` is excluded from type checking.
