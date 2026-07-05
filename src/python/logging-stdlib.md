# Python Logging (library)

Use the standard library `logging` module. Never use `print()`. Do not add `structlog` or any third-party logging library; a library must not force logging configuration on its consumers.

```python
import logging

logger = logging.getLogger(__name__)

def load_from_file(self, path: Path) -> None:
    logger.info("loading file", extra={"path": str(path)})
```

- Declare `logger = logging.getLogger(__name__)` at module level in every file that logs.
- Pass context via the `extra` keyword, not string formatting.
- `logger.debug` for internal detail, `logger.info` for lifecycle, `logger.warning`/`logger.error` for problems.
- Never configure logging handlers inside a library; that is the consumer's responsibility.
