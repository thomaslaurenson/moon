# Python badge row

```markdown
![Python Version](https://img.shields.io/badge/python-3.x%2B-blue?logo=python) ![Code Coverage](https://img.shields.io/badge/Coverage-XX%25-blue?logo=python)
```

The Python-version badge is static by default, derived from `requires-python` (keep `3.x` in sync with the floor). This matches the baseline where a library ships as a GitHub release and PyPI publishing is optional (see `python/release-lib.md`). Only when the project is published to PyPI may you switch to the live badge: `![Python Version](https://img.shields.io/pypi/pyversions/{package}?logo=python)`. Replace `XX` with the coverage percentage on each release.
