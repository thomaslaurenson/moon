# Python Testing Standards

Standards and conventions for testing Python projects.

## Structure

```
tests/
  __init__.py
  conftest.py          # shared fixtures
  unit/
    __init__.py
    conftest.py        # unit-specific fixtures if needed
    test_client.py     # mirrors <package>/client.py
    test_helpers.py    # mirrors <package>/helpers.py
  integration/
    __init__.py
    conftest.py        # integration fixtures, skip logic
    test_client.py
```

- Every test file mirrors a source module: `<package>/client.py` -> `tests/unit/test_client.py`
- Test functions use the pattern `test_<what>_<condition>_<expected>`. Example: `test_load_from_file_missing_path_raises_file_not_found`
- Test files are exempt from all docstring (`D`) ruff rules

## Markers

Register markers in `pyproject.toml`:

```toml
[tool.pytest.ini_options]
addopts = "-v --strict-markers"
testpaths = ["tests"]
markers = [
    "integration: marks tests as integration tests requiring live credentials or environment",
]
```

Apply at class or function level:

```python
@pytest.mark.integration
class TestClientIntegration:
    def test_connect_real_endpoint(self, data_dir):
        ...

# Or on individual functions
@pytest.mark.integration
def test_connect_real_endpoint(data_dir):
    ...
```

Integration tests are skipped natively via the Makefile using the flag `pytest -m "not integration"`. Never implement skip logic in `conftest.py`.

## Unit Tests

- No network calls, no real filesystem access (use `tmp_path` fixture for temp files)
- No real credentials or environment variables
- Mock external calls with `unittest.mock.patch` or `pytest` `monkeypatch`
- Every unit test must be fast; if a test takes more than a second, it is not a unit test
- Use `pytest.raises` for exception testing, always assert the message where meaningful

```python
def test_load_from_file_missing_path_raises_file_not_found(tmp_path):
    client = MyClient()
    with pytest.raises(FileNotFoundError, match="File not found"):
        client.load_from_file(tmp_path / "nonexistent.dat")
```

## Integration Tests

- Require a real environment; live files, credentials, or external services
- Always marked with `@pytest.mark.integration`
- Skipped automatically in all CI runs; never run in `lint.yml` or `test.yml`
- The developer runs them locally with `make test_integration`
- Gate on an environment variable to skip gracefully:

```python
import os
import pytest

@pytest.fixture(scope="session")
def data_dir():
    d = os.getenv("TEST_DATA_DIR")
    if not d:
        pytest.skip("TEST_DATA_DIR not set")
    return Path(d)
```

## Coverage

Never add `--cov` flags to `addopts` in `pyproject.toml`. Coverage must be an explicit separate step.

```toml
# Bad - slows every test run
[tool.pytest.ini_options]
addopts = "--cov=<package> --cov-report=term-missing"

# Good - plain and fast
[tool.pytest.ini_options]
addopts = "-v --strict-markers"
```

Coverage is run only via the dedicated Makefile target:

```makefile
test_coverage: ## Run tests with coverage report
    uv run coverage run -m pytest -m "not integration" && uv run coverage report
```

Coverage configuration lives in `pyproject.toml`:

```toml
[tool.coverage.run]
source = ["<package>"]

[tool.coverage.report]
show_missing = true
skip_empty = true
```
