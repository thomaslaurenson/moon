# Python Unit Testing

```
tests/
  __init__.py
  conftest.py          # shared fixtures
  unit/
    __init__.py
    test_client.py     # mirrors <package>/client.py
```

- Every test file mirrors a source module: `<package>/client.py` -> `tests/unit/test_client.py`.
- Test functions follow `test_<what>_<condition>_<expected>`, e.g. `test_load_from_file_missing_path_raises_file_not_found`.
- No network calls, no real filesystem access (use `tmp_path`), no real credentials.
- Mock external calls with `unittest.mock.patch` or `monkeypatch`.
- Every unit test is fast; if it takes more than a second, it is not a unit test.
- Use `pytest.raises` for exception testing; assert the message where meaningful.

```python
def test_load_from_file_missing_path_raises_file_not_found(tmp_path):
    client = MyClient()
    with pytest.raises(FileNotFoundError, match="File not found"):
        client.load_from_file(tmp_path / "nonexistent.dat")
```

The `integration` marker must be registered in the pytest configuration (defined in the tooling fragment) so `--strict-markers` does not reject it.
