# Python style

Python-specific style. Assumes the core conventions.

## Quotes

Use double quotes for all strings; never single quotes. Where the project uses ruff (every tier except one-off scripts), `ruff format` enforces this automatically and `make fix` corrects any single-quoted strings.

## Constants

Module-level constants use `UPPER_SNAKE_CASE`. Private constants (not part of the public API) use a leading underscore: `_UPPER_SNAKE_CASE`.

```python
DEFAULT_TIMEOUT = 300
_IGNORED_EVENT_TYPES = {"verbose", "playbook_on_start"}
```

## Config classes

Use class-level attributes with `os.getenv` defaults; no `__init__` required. Env vars are read once at class load time. Never hardcode credentials, URLs, or environment-specific values. Fixed internal constants (id maps, group ids, named profiles) may be hardcoded class attributes. Every class-level attribute, whether from `os.getenv` or hardcoded, must have a type annotation.

```python
import os

class FreshdeskConfig:
    timeout: int = int(os.getenv("FRESHDESK_TIMEOUT", "30"))

    # Hardcoded internal constants - not environment-specific
    prod: dict[str, int] = {"email_config_id": 6000071619, "group_id": 6000207769}
```

Env vars are read once, when the class body is evaluated at import time. A test that sets `FRESHDESK_TIMEOUT` with `monkeypatch.setenv` after import will not change an already-read attribute. To vary config in a test, monkeypatch the class attribute directly (`monkeypatch.setattr(FreshdeskConfig, "timeout", 5)`), or reload the module after setting the environment. This is a deliberate trade-off: reading once keeps config cheap and predictable, at the cost of import-time binding.

## Lazy instantiation

Prefer deferred setup: instantiate objects on first access, not at construction time. Use a private attribute initialised to `None` and a property that creates it on demand.

```python
@property
def instances(self) -> InstanceClient:
    if self._instances is None:
        self._instances = InstanceClient(self._session)
    return self._instances
```
