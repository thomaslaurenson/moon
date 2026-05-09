# C++ Testing Standards

Standards and conventions for testing C++ projects.

## Overview

Two testing layers are used in combination:

- **Catch2** — unit tests for pure logic functions with no external dependencies
- **pytest** — functional tests against the compiled binary

Both layers must pass before a change is considered complete. Python functional
tests follow the conventions in `python/testing.md`.

---

## Structure

```
test/
  unit/                   # Catch2 unit tests
    test_helpers.cpp      # mirrors src/helpers.cpp
    test_gamerules.cpp    # mirrors src/gamerules.cpp
  functional/             # pytest functional tests against the binary
    conftest.py
    test_extract.py
    test_create.py
  integration/            # when needed — pytest or Catch2
```

- One unit test file per source file, named `test_<source>.cpp`
- Functional tests mirror the CLI commands or user-facing behaviour, not
  internal source files
- An `integration/` directory is added only when a project has tests that
  require a real environment but do not test the full binary end-to-end

---

## Catch2

### Setup

Pin Catch2 as a git submodule under `extern/Catch2` so the version is
controlled and no system install is required:

```bash
git submodule add https://github.com/catchorg/Catch2.git extern/Catch2
```

In the root `CMakeLists.txt`:

```cmake
option(BUILD_TESTING "Build tests" ON)

if(BUILD_TESTING)
    add_subdirectory(extern/Catch2)
    add_subdirectory(test)
endif()
```

In `test/CMakeLists.txt`:

```cmake
add_executable(unit_tests
    unit/test_helpers.cpp
    unit/test_gamerules.cpp
)

target_link_libraries(unit_tests PRIVATE
    Catch2::Catch2WithMain
)

include(CTest)
include(Catch)
catch_discover_tests(unit_tests)
```

### Test structure

Use one `TEST_CASE` per function under test, with `SECTION` blocks for
individual scenarios. The `TEST_CASE` name is the function name. Each
`SECTION` describes the specific scenario being tested:

```cpp
#include <catch2/catch_test_macros.hpp>
#include "helpers.h"

TEST_CASE("next_power_of_two", "[helpers]") {
    SECTION("returns the same value for exact powers of two") {
        REQUIRE(next_power_of_two(1) == 1);
        REQUIRE(next_power_of_two(32) == 32);
        REQUIRE(next_power_of_two(64) == 64);
    }

    SECTION("rounds up to the next power for non-powers") {
        REQUIRE(next_power_of_two(5) == 8);
        REQUIRE(next_power_of_two(33) == 64);
        REQUIRE(next_power_of_two(100) == 128);
    }

    SECTION("handles zero") {
        REQUIRE(next_power_of_two(0) == 0);
    }
}
```

### Tags

Tag each `TEST_CASE` with the name of the source file under test:

```cpp
TEST_CASE("next_power_of_two", "[helpers]") { … }
TEST_CASE("match_file_mask", "[helpers]") { … }
TEST_CASE("parse_game_rules", "[gamerules]") { … }
```

This allows running a subset of tests during development:

```bash
./unit_tests [helpers]     # run only helpers tests
./unit_tests [gamerules]   # run only gamerules tests
```

### What to unit test

A function gets a unit test if it can be called without:

- A real file on disk
- A network connection
- A third-party library handle or session
- Any global system state

Pure logic functions — string manipulation, maths, parsing, pattern matching —
always get unit tests. Functions that are tightly coupled to external libraries
or the filesystem are covered by functional tests instead.

Every bug fix must include a unit test that reproduces the bug before the fix.
This prevents regressions and documents the expected behaviour.

### Running unit tests

Run the unit test suite via the Makefile:

```makefile
.PHONY: test_unit
test_unit: ## Run Catch2 unit tests
	cmake --build build && cd build && ctest --output-on-failure
```

---

## Functional Tests

Functional tests are written in Python using pytest. They test the compiled
binary through its CLI interface, verifying real user-facing behaviour across
a range of scenarios.

Follow all conventions in `python/testing.md` for structure, markers, fixtures,
and test naming.

The binary under test is built before running functional tests. Gate on the
binary path via a fixture in `conftest.py`:

```python
import subprocess
import pytest
from pathlib import Path

@pytest.fixture(scope="session")
def binary(tmp_path_factory):
    path = Path("build") / "myapp"
    if not path.exists():
        pytest.skip("Binary not built — run make build first")
    return path

def run(binary, *args, **kwargs):
    return subprocess.run(
        [binary, *args],
        capture_output=True,
        text=True,
        **kwargs,
    )
```

Run functional tests via the Makefile:

```makefile
.PHONY: test_functional
test_functional: ## Run pytest functional tests
	uv run pytest test/functional -v
```

---

## Makefile Targets

Projects provide individual targets for each layer and a combined target that
runs both:

```makefile
.PHONY: test test_unit test_functional

test: test_unit test_functional ## Run all tests

test_unit: ## Run Catch2 unit tests
	cmake --build build && cd build && ctest --output-on-failure

test_functional: ## Run pytest functional tests
	uv run pytest test/functional -v
```
