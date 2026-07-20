# Python logging

Use the standard library `logging` module. Never use `print()` for diagnostics. Do not add `structlog` or any third-party logging library.

```python
import logging

logger = logging.getLogger(__name__)

def load_from_file(self, path: Path) -> None:
    logger.info("loading file", extra={"path": str(path)})
```

- Declare `logger = logging.getLogger(__name__)` at module level in every file that logs.
- Pass context via the `extra` keyword, not string formatting.
- `logger.debug` for internal detail, `logger.info` for lifecycle, `logger.warning`/`logger.error` for problems.
- Never configure logging handlers inside an importable module; configuration is a whole-application decision.

## Library and CLI tiers

Who calls `logging.basicConfig()` (or otherwise attaches handlers) depends on the tier:

- A **library** never configures the stack. Its modules only get a logger and log; handler configuration is the consumer's responsibility, so a consumer keeps full control of level, format, and destination.
- A **library that ships a CLI** still behaves as a library when imported, but the console-script entry point owns configuration: call `logging.basicConfig(...)` once in the entry point (`main`), before any logging happens, and never on import. A consumer who imports the package as a library is unaffected; someone who runs the command gets configured output.

In short: the entry point configures; the package code only logs.
