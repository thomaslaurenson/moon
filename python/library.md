# Python Library Authoring

Standards and conventions for authoring installable Python library packages.

## Type Hints

- All function and method signatures must include type hints - parameters and return types
- Class attributes must be annotated
- Use `from __future__ import annotations` at the top of every module for deferred evaluation

```python
from __future__ import annotations

def get_instance(instance_id: str) -> Instance:
    ...

class InstanceClient:
    timeout: int = 30

    def list(self, project_id: str) -> list[Instance]:
        ...
```

- Use `X | Y` union syntax (not `Union[X, Y]`)
- Use `list[X]`, `dict[K, V]` (not `List[X]`, `Dict[K, V]`)
- Only import from `typing` when no built-in equivalent exists
- Always annotate `__init__` with `-> None`

## Docstrings

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

Rules:
- Every public function, method, and class requires a docstring.
- One-line docstrings are acceptable for trivial properties and getters.
- `:param:`, `:raises:`, and `:return:` fields are only included when they add information beyond the type hints.
- Use `:return:` not `:returns:`.
- Omit `:return:` from `__init__` methods entirely.
- `tests/**` files are exempt from docstring rules.

## `__init__.py`

The top-level `__init__.py` re-exports the public API. Keep it minimal.

```python
# <package>/__init__.py
from <package>.client import MyClient
from <package>.config import MyConfig

__all__ = ["MyClient", "MyConfig"]
```

- Only symbols intended for external consumers belong here
- Never import optional-dependency modules at the top level of `__init__.py`; this would break imports for users who don't have those extras installed
- Sub-package `__init__.py` files follow the same rule: only re-export what is public

## PEP 561

Include an empty `py.typed` marker file in the package root directory. This signals to type checkers like pyright and mypy that the package supports inline type annotations.

```
<package>/
    __init__.py
    py.typed       # empty file - signals PEP 561 compliance
    client.py
    ...
```

Without `py.typed`, pyright will treat your package as untyped and ignore all annotations when consumers use your library, even if your annotations are complete and correct.

## Optional Dependencies

Import guards prevent missing extras from breaking unrelated code. Always import optional packages inside the function or method that uses them, not at module top-level.

```python
# Good
def create_session(self) -> None:
    try:
        import hvac
    except ImportError as e:
        raise ImportError("Install the 'vault' extra: pip install <package>[vault]") from e
    self._client = hvac.Client(url=self._url)

# Bad - breaks at import time for users without this extra
import hvac
```

- Each optional integration is a separate dep group in `pyproject.toml`
- The import error message must name the extra the user needs to install

## `TYPE_CHECKING` Guard

Use a `TYPE_CHECKING` guard when annotating parameters or return types that come from
external packages not otherwise imported at runtime in that module. Combined with
`from __future__ import annotations`, the annotation is only evaluated by the type
checker, never at runtime.

```python
from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    import hvac
    from keystoneauth1.session import Session


def authenticate(client: hvac.Client, session: Session | None = None) -> None:
    ...
```

Rules:
- Always pair a `TYPE_CHECKING` guard with `from __future__ import annotations`.
- Use it for packages that are required dependencies but not imported elsewhere in the module.
- Use it for packages that are optional extras, to avoid breaking imports for users who do not have them installed.
- Do not use it for packages that are already imported at the top of the module for runtime use.

## Logging

Use the Python standard library `logging` module. Never use `print()`. Do not add `structlog` or any third-party logging library as a dependency; libraries must not force logging configuration on their consumers.

```python
import logging

logger = logging.getLogger(__name__)

def load_from_file(self, path: Path) -> None:
    logger.info("loading file", extra={"path": str(path)})
```

- Declare `logger = logging.getLogger(__name__)` at module level in every file that logs
- Use `__name__` so consumers can control output per-module via their own logging config
- Pass context via the `extra` keyword argument, not via string formatting
- Use `logger.debug` for internal detail, `logger.info` for lifecycle events, `logger.warning` / `logger.error` for problems
- Never configure logging handlers inside a library; that is the consumer's responsibility

Note: `structlog` is appropriate for applications where you control the full stack. For libraries, always use stdlib `logging`.

## Exceptions

Every library must define a clear exception hierarchy so consumers can catch errors at the appropriate level of specificity.

### Hierarchy

Always define a single root exception named `<LibraryName>Error` that inherits from `Exception`. All other library exceptions inherit from this root.

```python
class MyLibraryError(Exception):
    """Base exception for all mylib errors."""

class DataError(MyLibraryError):
    """Base exception for data-related errors."""

class DataLoadError(DataError):
    """Raised when a data file cannot be loaded or parsed."""
```

Group exceptions by domain under the root. A consumer who wants to catch everything catches `MyLibraryError`. A consumer who only cares about data errors catches `DataError`.

### Rules

- Every exception class requires a docstring explaining when it is raised
- Never raise bare `ValueError` or `TypeError` from a public API method; wrap them in a library exception so they are catchable at the `LibraryError` level
- Always use `raise X from exc` when wrapping an exception to preserve the chain
- Export all commonly caught exceptions from the top-level `__init__.py`

```python
# Good - consumer can catch at any level
raise exceptions.DataLoadError("Invalid version") from exc

# Bad - escapes the library exception hierarchy
raise ValueError("Invalid version")
```

### Export

All exceptions that a consumer is reasonably expected to catch must be importable directly from the top-level package:

```python
from <package> import DataLoadError, ConfigError
```

Never require consumers to reach into internal modules:

```python
# Bad - exposes internal structure
from <package>.exceptions import DataLoadError
```

## Type checking

All library projects must include pyright for static type checking.

### Configuration

Add to `pyproject.toml`:

```toml
[tool.pyright]
include = ["<package>"]
exclude = ["tests"]
typeCheckingMode = "basic"
```

Do not set `pythonVersion`; pyright infers it automatically from `requires-python` in `pyproject.toml`.

Add to the `dev` optional dep group:

```toml
dev = [
    "<package>[test]",
    "ruff>=...",
    "pyright>=1.1.0",
]
```

Add to the Makefile:

```makefile
.PHONY: typecheck
typecheck: ## Run pyright type checker
    uv run pyright
```

Include `typecheck` in the `ci` target:

```makefile
ci: lint fmt_check typecheck test ## Run all CI checks locally
```

### Modes

- Start with `typeCheckingMode = "basic"` for new or existing projects being migrated
- Move to `"strict"` once the codebase is fully annotated and all basic errors are resolved
- `tests/` is excluded; test files are exempt from type checking
