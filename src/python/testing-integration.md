# Python Integration Testing

- Require a real environment: live files, credentials, or external services.
- Always marked `@pytest.mark.integration`, applied at class or function level.
- Skipped in all CI runs via the Makefile flag `pytest -m "not integration"`; never implement skip logic in `conftest.py` for the marker itself.
- The developer runs them locally with `make test_integration`.
- Gate fixtures on an environment variable to skip gracefully:

```python
@pytest.fixture(scope="session")
def data_dir():
    d = os.getenv("TEST_DATA_DIR")
    if not d:
        pytest.skip("TEST_DATA_DIR not set")
    return Path(d)
```
