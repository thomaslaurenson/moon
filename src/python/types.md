# Python Type Hints

- All function and method signatures include type hints, parameters and return types.
- Class attributes are annotated.
- Use `from __future__ import annotations` at the top of every module for deferred evaluation.
- Use `X | Y` union syntax, not `Union[X, Y]`.
- Use `list[X]`, `dict[K, V]`, not `List[X]`, `Dict[K, V]`.
- Only import from `typing` when no built-in equivalent exists.
- Always annotate `__init__` with `-> None`.

```python
from __future__ import annotations

def get_instance(instance_id: str) -> Instance:
    ...

class InstanceClient:
    timeout: int = 30

    def list(self, project_id: str) -> list[Instance]:
        ...
```
