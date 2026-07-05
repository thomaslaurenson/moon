# Python Exceptions

Every library defines a clear exception hierarchy so consumers can catch errors at the right level.

Define a single root exception `<LibraryName>Error` inheriting from `Exception`. All other exceptions inherit from it, grouped by domain.

```python
class MyLibraryError(Exception):
    """Base exception for all mylib errors."""

class DataError(MyLibraryError):
    """Base exception for data-related errors."""

class DataLoadError(DataError):
    """Raised when a data file cannot be loaded or parsed."""
```

- Every exception class requires a docstring explaining when it is raised.
- Never raise bare `ValueError` or `TypeError` from a public API method; wrap it in a library exception so it is catchable at the root.
- Always use `raise X from exc` when wrapping, to preserve the chain.
- Export all commonly caught exceptions from the top-level `__init__.py` so consumers import them directly (`from <package> import DataLoadError`), never from internal modules.
