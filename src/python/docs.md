# Python Sphinx Documentation

Any installable library package gets a `docs/` directory and Sphinx setup. Scripts-only projects do not require Sphinx.

```toml
[project.optional-dependencies]
docs = [
    "sphinx>=7.3.7",
    "furo>=2024.5.6",
    "sphinx-autodoc-typehints>=2.3.0",
    "myst-parser>=4.0.0",
    "sphinx-copybutton>=0.5.2",
]
```

```
docs/
  conf.py
  index.rst
  api/
    index.rst        # committed
    generated/       # gitignored (autosummary)
  _build/            # gitignored
```

- Theme: `furo`. Extensions: autodoc, autosummary, napoleon, viewcode, intersphinx, sphinx-autodoc-typehints, sphinx-copybutton, myst-parser.
- `autosummary_generate = True`; `autosummary_ignore_module_all = True`.
- Read version from installed package metadata via `importlib.metadata.version()`; no pyproject fallback.
- `autodoc_mock_imports` must remain empty. If a docs build fails on a missing import, add the dependency to the `docs` extra; never mock an import.
- Makefile: `docs`, `docs_check` (`-W --keep-going`), `docs_linkcheck`, `docs_clean`.
