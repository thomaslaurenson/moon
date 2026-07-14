# Python packaging

## `__init__.py`

The top-level `__init__.py` re-exports the public API and stays minimal. Only symbols intended for external consumers belong here. Never import optional-dependency modules at the top level. Sub-package `__init__.py` files follow the same rule.

```python
from <package>.client import MyClient
from <package>.config import MyConfig

__all__ = ["MyClient", "MyConfig"]
```

## PEP 561

Include an empty `py.typed` marker in the package root. Without it, pyright and mypy treat the package as untyped and ignore your annotations when consumers use the library.

## Optional dependencies

Import optional packages inside the function that uses them, never at module top-level. Each optional integration is its own dep group in `pyproject.toml`, and the import error message names the extra to install.

```python
def create_session(self) -> None:
    try:
        import hvac
    except ImportError as e:
        raise ImportError("Install the 'vault' extra: pip install <package>[vault]") from e
    self._client = hvac.Client(url=self._url)
```

## `TYPE_CHECKING` guard

Pair a `TYPE_CHECKING` guard with `from __future__ import annotations` when annotating types from external packages not otherwise imported at runtime. Do not use it for packages already imported at the top of the module for runtime use.

## Build system

```toml
[tool.hatch.build.targets.wheel]
packages = ["<package>"]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```
