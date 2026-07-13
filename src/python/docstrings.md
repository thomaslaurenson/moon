# Python docstrings

rST/Sphinx format, enforced by ruff `D` rules with `convention = "pep257"`.

```python
def sign_ssh_key(public_key: str, ttl: int = 3600) -> str:
    """Sign an SSH public key using the Vault CA.

    :param public_key: The SSH public key to sign.
    :param ttl: Certificate validity in seconds.
    :raises ValueError: If the public key is malformed.
    :return: The signed certificate as a string.
    """
```

- Every public function, method, and class requires a docstring.
- One-line docstrings are acceptable for trivial properties and getters.
- `:param:`, `:raises:`, and `:return:` are included only when they add information beyond the type hints.
- Use `:return:` not `:returns:`. Omit `:return:` from `__init__` entirely.
- `tests/**` files are exempt from docstring rules.

## Ruff enforcement

Docstrings are enforced by ruff's `D` rule family, which is not in the shared tooling baseline (see the tooling fragment). Tiers that require docstrings add it to the ruff config:

```toml
[tool.ruff.lint]
# extend the shared select with the D family
select = ["E", "W", "F", "I", "B", "C4", "SIM", "UP", "PTH", "FA", "RET", "D"]

[tool.ruff.lint.pydocstyle]
convention = "pep257"

[tool.ruff.lint.per-file-ignores]
"tests/**" = ["D"]
```

Only tiers that mandate docstrings enable `D`; application and script tiers leave it out.
