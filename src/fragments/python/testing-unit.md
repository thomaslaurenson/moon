# Python unit testing

```
tests/
  __init__.py
  conftest.py          # shared fixtures
  unit/
    __init__.py
    test_client.py     # mirrors <package>/client.py
```

- Every test file mirrors a source module: `<package>/client.py` -> `tests/unit/test_client.py`.
- Test functions follow `test_<what>_<condition>_<expected>`, e.g. `test_load_from_file_missing_path_raises_data_load_error`.
- No network calls, no real filesystem access (use `tmp_path`), no real credentials.
- Mock external calls with `unittest.mock.patch` or `monkeypatch`.
- Every unit test is fast; if it takes more than a second, it is not a unit test.
- Use `pytest.raises` for exception testing; assert the message where meaningful. A public API method raises the library's own exception hierarchy (see the exceptions fragment), not a bare stdlib error, so tests assert the library exception:

```python
def test_load_from_file_missing_path_raises_data_load_error(tmp_path):
    client = MyClient()
    with pytest.raises(DataLoadError, match="could not load"):
        client.load_from_file(tmp_path / "nonexistent.dat")
```

The `integration` marker must be registered in the pytest configuration (defined in the tooling fragment) so `--strict-markers` does not reject it.
