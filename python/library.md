# Python Library Authoring

Standards and conventions for authoring installable Python library packages.

## Contents

- [Type hints](#type-hints) — required on all signatures, modern union syntax
- [Docstrings](#docstrings) — rST/Sphinx format, when to include param/return fields
- [__init__.py](#initpy) — public API exports, what not to import at top level
- [Optional dependencies](#optional-dependencies) — import guards, error messages
- [Logging](#logging) — structlog, never print(), context as keyword args
- [Testing](#testing) — unit vs integration, pytest.mark.integration, conftest

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

## Docstrings

rST/Sphinx format, enforced by ruff `D` rules with `convention = "sphinx"`.

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
- `:param:`, `:raises:`, and `:return:` fields are only included when they add
  information beyond the type hints.
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
- Never import optional-dependency modules at the top level of `__init__.py` - this would break imports for users who don't have those extras installed
- Sub-package `__init__.py` files follow the same rule: only re-export what is public

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

## Logging

Use `structlog`. Never use `print()`.

```python
import structlog

logger = structlog.get_logger()

def create_session(self) -> None:
    logger.info("creating session", url=self._url)
```

- Declare `logger` at module level
- Pass context as keyword arguments to log calls, not via string formatting
- Use `logger.debug` for internal detail, `logger.info` for lifecycle events, `logger.warning` / `logger.error` for problems

## Testing

- **Unit tests**: no network or credentials; use `unittest.mock.patch` and `pytest` `monkeypatch`
- **Integration tests**: require live credentials; always mark with `@pytest.mark.integration`
- Integration tests are never run in CI - ask the user to run `make test_integration` locally
- Fixtures for live clients are defined in `conftest.py`

```python
@pytest.mark.integration
def test_list_instances(nectar_client):
    instances = nectar_client.instances.list(project_id="abc123")
    assert len(instances) > 0
```
