# C++ Testing Standards

Standards and conventions for testing C++ projects. Universal to both applications
and libraries; the subprocess/functional testing layer that applications add on top
lives in the testing-functional fragment.

## Structure

```
test/
  fixtures/               # shared fixture headers, one file per fixture
  unit/                   # Catch2 unit tests for internal logic
    test_helpers.cpp      # mirrors src/helpers.cpp
    test_config.cpp       # mirrors src/config.cpp
  CMakeLists.txt
extern/
  Catch2/                 # submodule - pinned to v3.x
```

- One unit test file per source file, named `test_<source>.cpp`
- Fixtures live in `test/fixtures/`, one header per fixture

## Catch2 Setup

Pin Catch2 as a git submodule under `extern/Catch2` so the version is controlled and no system install is required:

```bash
git submodule add https://github.com/catchorg/Catch2.git extern/Catch2
cd extern/Catch2 && git checkout v3.6.0
```

Catch2 v3 is not a single-header library; it is a compiled library with multiple headers. Always use it as a submodule, never copy individual headers.

The CMake wiring (the `BUILD_TESTING` option) is defined in the CMake fragment; the test executable's own definition (linking application source directly, or the library target) is defined in the cmake-app or cmake-lib fragment.

## Unit Tests

### Test structure

Use one `TEST_CASE` per function under test, with `SECTION` blocks for individual scenarios. The `TEST_CASE` name is the function name:

```cpp
#include <catch2/catch_test_macros.hpp>
#include "helpers.h"

TEST_CASE("next_power_of_two", "[helpers]") {
    SECTION("returns the same value for exact powers of two") {
        REQUIRE(next_power_of_two(1) == 1);
        REQUIRE(next_power_of_two(32) == 32);
    }

    SECTION("rounds up to the next power for non-powers") {
        REQUIRE(next_power_of_two(5) == 8);
        REQUIRE(next_power_of_two(33) == 64);
    }
}
```

### Tags

Tag each `TEST_CASE` with the name of the source file under test:

```cpp
TEST_CASE("next_power_of_two", "[helpers]") { ... }
TEST_CASE("parse_config", "[config]") { ... }
```

Run a subset during development:

```bash
./build/bin/mytarget_unit_tests [helpers]
./build/bin/mytarget_unit_tests [config]
```

### What to unit test

A function gets a unit test if it can be called without:

- A real file on disk
- A network connection
- A third-party library handle or session
- Any global system state

Pure logic functions (string manipulation, maths, parsing, pattern matching) always get unit tests. Functions tightly coupled to external libraries or the filesystem are covered by functional tests instead, where those exist (see testing-functional.md).

Every bug fix must include a unit test that reproduces the bug before the fix.

## Makefile Targets

```makefile
.PHONY: test_unit
test_unit: ## Run Catch2 unit tests
	cmake --build build && cd build && ctest --output-on-failure -R unit
```

A project with a functional test layer adds `test_functional` and a combined `test` target; see testing-functional.md.
