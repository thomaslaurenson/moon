# Python Style Guide

Style conventions for Python code in this project.

---

## Unusual Characters

- Never use em dash (—)

## Docstrings

Docstrings use **reStructuredText (rST) / Sphinx** syntax, not Google-style sections.

### Module docstrings

Brief summary sentence, followed by a description of what the module does (numbered or
bulleted list), and a `Usage::` code block where applicable.

Use `Usage::` (double colon) — this is standard rST syntax and creates a rendered code
block in Sphinx. Do not use `Usage:` (single colon), `Example:`, or RST section underlines
(`--------`) inside module docstrings. Run commands use `uv run`, never `python`.

```python
"""Short description of the module.

Does the following:

1. First thing.
2. Second thing.

Usage::

    uv run tasks/some/script.py --arg value
"""
```

### Function and method docstrings

One-line summary sentence on the opening line. Optional extended description as a
paragraph. Then `:param:`, `:raises:`, and `:return:` fields.

```python
def my_function(arg1: str, arg2: int) -> list[dict]:
    """Do something useful.

    Extended description goes here when the summary alone is not sufficient.

    :param arg1: Description of arg1.
    :param arg2: Description of arg2.
    :raises ValueError: When arg1 is invalid.
    :return: List of result dicts.
    """
```

- Use `:class:\`~fully.qualified.ClassName\`` when referencing types in param descriptions.
- Omit `:param:` / `:return:` fields for trivial one-liners where the signature is self-documenting.
- Use `:return:` (not `:returns:`).
- Omit `:return:` entirely from `__init__` methods — never document `None` returns.
- Private helpers (prefixed `_`) follow the same conventions.

---

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
