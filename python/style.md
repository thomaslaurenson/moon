# Python Style Guide

Style conventions for Python code in this project.

## Contents

- [Unusual characters](#unusual-characters) — em dash and other characters to avoid
- [Spelling](#spelling) — British English rules
- [Inline comments](#inline-comments) — formatting and capitalisation rules
- [Constants](#constants) — UPPER_SNAKE_CASE, public vs private
- [Comment hygiene](#comment-hygiene) — step narration, human comments
- [Lazy instantiation](#lazy-instantiation) — preferred pattern for deferred setup
- [Config classes](#config-classes) — os.getenv defaults, no hardcoded values

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

---

## Inline Comments

- Start with `#` followed by a single space.
- First word is capitalised.
- Never use a full stop, unless a multiline comment
- For continuation lines (when a comment wraps to a second line), capitalisation is not required.
- No decorative styles — avoid `# ---`, `# ===`, `# ***`, or similar dividers.

```python
# Good: single-line comment
x = compute_value()

# Good: multi-line comment where only the first line is capitalised,
# continuation lines do not need to start with a capital letter.
y = complex_operation()

# Bad: decorative divider — avoid this style
# --- section name ---
```

## Comment Hygiene

- Do not write step narration comments that describe the next line of code.
  Bad: `# Loop through results`, `# Open the file`
- Preserve comments that explain why something is done, not what.
  Good: `# Offset by 1 because the API returns 1-indexed page numbers`
- Do not over-annotate type hints with redundant inline comments.
- Do not inject `TODO` or `FIXME` comments unless they refer to a real, known issue.

## Constants

Module-level constants use `UPPER_SNAKE_CASE`. Private constants (not part of the
public API) use a leading underscore: `_UPPER_SNAKE_CASE`.

```python
# Public constant
DEFAULT_TIMEOUT = 300

# Private constant — internal implementation detail
_IGNORED_EVENT_TYPES = {"verbose", "playbook_on_start"}
```

## Lazy Instantiation

Prefer deferred setup — instantiate objects on first access, not at construction time.
Use a private attribute initialised to `None` and a property that creates it on demand.

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
