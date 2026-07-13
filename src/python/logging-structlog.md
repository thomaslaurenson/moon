# Python logging (application)

Applications control the full stack and configure logging once, at the entry point. Use `structlog` for structured logging; never `print()`.

Configure structlog exactly once, in the CLI or main module, before any logging happens. Never configure it in an importable library module: configuration is a whole-application decision, and a module that configures logging on import imposes that choice on everything that imports it.

```python
import logging
import structlog

def configure_logging(level: int = logging.INFO) -> None:
    structlog.configure(
        processors=[
            structlog.contextvars.merge_contextvars,
            structlog.processors.add_log_level,
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.JSONRenderer(),
        ],
        wrapper_class=structlog.make_filtering_bound_logger(level),
        cache_logger_on_first_use=True,
    )
```

Call `configure_logging()` once from the entry point (`main`), then get a logger per module:

```python
import structlog

logger = structlog.get_logger(__name__)

def run(path: str) -> None:
    logger.info("loading file", path=path)
```

- Bind context as keyword arguments (`logger.info("event", key=value)`), never string formatting.
- Use `logger.debug` for internal detail, `logger.info` for lifecycle, `logger.warning`/`logger.error` for problems.
- For a JSON renderer in production and a console renderer in development, branch on an environment flag inside `configure_logging`, not at call sites.

## Library and CLI tiers

A package that ships a CLI (the lib-cli tier) still behaves as a library when imported: its internal modules log through their own logger and never configure the stack. structlog is configured only behind the console-script entry point, so a consumer who imports the package as a library keeps full control of logging. In short: the entry point configures; the package code only logs.
