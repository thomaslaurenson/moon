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

## Docstrings

Google format, enforced by ruff `D` rules with `convention = "google"`.

```python
def sign_ssh_key(public_key: str, ttl: int = 3600) -> str:
    """Sign an SSH public key using the Vault CA.

    Args:
        public_key: The SSH public key to sign.
        ttl: Certificate validity in seconds.

    Returns:
        The signed certificate as a string.

    Raises:
        ValueError: If the public key is malformed.
    """
```

Rules:
- Every public function, method, and class requires a docstring
- One-line docstrings are acceptable for trivial properties and getters
- `Args`, `Returns`, and `Raises` sections are only included when they add information beyond the type hints
- `tests/**` files are exempt from docstring rules - configure via `per-file-ignores` in `pyproject.toml`

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

## Client Pattern

Every integration client follows the same lifecycle - no I/O at construction time:

```python
client = MyClient()       # __init__: store config only, no network calls
client.creds_from_env()   # read env vars, raise ValueError if missing
client.create_session()   # establish API connection
```

- `creds_from_env()` raises `ValueError` with a descriptive message if a required env var is absent
- `create_session()` is where all network I/O happens

## Lazy Sub-clients

Expose sub-clients as `@property` on the parent client. Instantiate on first access only.

```python
@property
def instances(self) -> InstanceClient:
    if self._instances is None:
        self._instances = InstanceClient(self._session)
    return self._instances
```

## Config Classes

Use class-level attributes with `os.getenv` defaults. No `__init__` required.

```python
import os

class MyConfig:
    timeout: int = int(os.getenv("MY_TIMEOUT", "30"))
    base_url: str = os.getenv("MY_BASE_URL", "https://api.example.com")
```

- Env vars are read once at class load time
- Never hardcode credentials, URLs, or environment-specific values

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
