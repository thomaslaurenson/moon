# Python Style Guide

Style conventions for Python code in this project.

## Unusual Characters

- Never use em dash (—)

## Quotes

Use double quotes for all strings. Never use single quotes.

```python
# Good
message = "Hello, world"
path = "data/patches"

# Bad
message = 'Hello, world'
```

Ruff format enforces this automatically. Running `make fix` will correct any single-quoted strings.

## Spelling

Use British English spellings:

- `Initialise` not `Initialize`
- `Colour` not `Color`

## Inline Comments

- Start with `#` followed by a single space.
- First word is capitalised.
- Never use a full stop, unless a multiline comment.
- For continuation lines (when a comment wraps to a second line), capitalisation is not required.
- No decorative styles; avoid `# ---`, `# ===`, `# ***`, or similar dividers.

```python
# Good: single-line comment
x = compute_value()

# Good: multi-line comment where only the first line is capitalised,
# continuation lines do not need to start with a capital letter.
y = complex_operation()

# Bad: decorative divider - avoid this style
# --- section name ---
```

## Comment Hygiene

- Do not write step narration comments that describe the next line of code. Bad: `# Loop through results`, `# Open the file`
- Preserve comments that explain the 'why' (business logic, architecture). Aggressively delete and refactor comments that narrate the 'what' (step-by-step code narration).
- Do not over-annotate type hints with redundant inline comments.
- Do not inject `TODO` or `FIXME` comments unless they refer to a real, known issue.

## Constants

Module-level constants use `UPPER_SNAKE_CASE`. Private constants (not part of the public API) use a leading underscore: `_UPPER_SNAKE_CASE`.

```python
# Public constant
DEFAULT_TIMEOUT = 300

# Private constant - internal implementation detail
_IGNORED_EVENT_TYPES = {"verbose", "playbook_on_start"}
```

## Lazy Instantiation

Prefer deferred setup; instantiate objects on first access, not at construction time. Use a private attribute initialised to `None` and a property that creates it on demand.

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

- Env vars are read once at class load time.
- Never hardcode credentials, URLs, or environment-specific values.

Some config classes also carry **internal constants** that are not environment-specific (for example, a mapping of team member IDs to email addresses, a set of fixed API group IDs, or a set of named connection profiles). These are permitted as hardcoded class-level attributes. All class-level attributes, whether sourced from `os.getenv` or hardcoded, must have type annotations.

```python
import os

class FreshdeskConfig:
    timeout: int = int(os.getenv("FRESHDESK_TIMEOUT", "30"))

    # Hardcoded internal constants - not environment-specific
    prod: dict[str, int] = {
        "email_config_id": 6000071619,
        "group_id": 6000207769,
    }
```
