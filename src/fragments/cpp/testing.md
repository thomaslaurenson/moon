# C++ testing standards

Standards and conventions for testing C++ projects, and the unit layer in full. Universal to every tier; the layers built on top live in the testing-integration and testing-functional fragments.

## The four layers

Every test in a C++ project belongs to exactly one of four layers. All four live under `test/`; a test binary is never built anywhere else.

- **unit** - tests logic. Links the library target. Free to use fixtures, temporary files, synthetic data and loopback sockets: the test creates whatever it needs. Always built, always run.
- **integration** - tests against real data the machine must already have: a game client install, a live server, a real API. Links the library target. Opt-in, and skips cleanly when the data is absent. See testing-integration.md.
- **functional** - tests the compiled binary, spawned as a subprocess and driven through its CLI. Links neither the library nor the binary. See testing-functional.md.
- **fuzz** - libFuzzer harnesses driving a parser with hostile input. Links the library target. Built on demand, Clang only. See testing-fuzz.md.

The dividing line between unit and integration is **not** whether a test touches the filesystem. A unit test that writes a synthetic archive to a temp directory and reads it back is still a unit test: it tests logic, and it brought its own data. What makes a test an integration test is depending on an environment it cannot construct: a real 1.12.1 client, a running server, a populated API. That is also what makes it opt-in, since CI has none of those things.

Deciding where a test goes:

- Tests logic, and can build whatever inputs it needs: **unit**.
- Needs real data or a real service that must already exist on the machine: **integration**.
- Needs argv, an exit code, or something on stdout: **functional**.
- Feeds arbitrary bytes to a parser looking for a crash: **fuzz**.

Which layers a project has follows from its tier. A library has unit, integration and fuzz, and cannot have functional, having no binary to spawn. An application and a lib-cli have unit and functional, and add integration and fuzz where they apply. See the tier fragment.

## Structure

Every layer lives under `test/`, including fuzz harnesses. A project with modules under `src/<module>/` mirrors that structure inside `test/unit/`:

```
test/
  CMakeLists.txt
  data/                   # checked-in static inputs; tests read, never write
  fixtures/               # shared fixture headers, one per fixture, used by every layer
    synthetic_archive.h
    temp_dir.h
  unit/                   # tests logic; mirrors src/
    mpq/
      test_archive.cpp    # mirrors src/mpq/archive.cpp
      test_crypto.cpp
    dbc/
      test_dbc_reader.cpp
  integration/            # tests against real data; opt-in
    test_mpq_client.cpp
  functional/             # tests the compiled binary; only where one ships
    test_create.cpp
  fuzz/                   # libFuzzer harnesses; built on demand
    fuzz_mpq.cpp
extern/
  Catch2/                 # submodule - pinned to v3.x
```

- One unit test file per source file, named `test_<source>.cpp`, in a directory mirroring the module. The other layers mirror behaviour rather than source files and do not follow this rule.
- Fixtures live in `test/fixtures/`, one header per fixture, and are shared by every layer.

## Catch2 setup

Pin Catch2 as a git submodule under `extern/Catch2` so the version is controlled and no system install is required:

```bash
git submodule add https://github.com/catchorg/Catch2.git extern/Catch2
cd extern/Catch2 && git checkout v3.6.0
```

Catch2 v3 is not a single-header library; it is a compiled library with multiple headers. Always use it as a submodule, never copy individual headers.

The CMake wiring (the project-scoped testing option, `enable_testing()`, and `include(Catch)`, all in the root `CMakeLists.txt`) is defined in the universal CMake fragment; each test executable's own definition is defined in the tier fragment.

### Registering tests

Every `catch_discover_tests` call sets two properties, and both are load-bearing:

```cmake
catch_discover_tests(mylib_unit_tests
    PROPERTIES LABELS "unit" SKIP_RETURN_CODE 4)

catch_discover_tests(mylib_integration_tests
    PROPERTIES LABELS "integration" SKIP_RETURN_CODE 4)
```

`LABELS` is what lets a layer be selected, with `ctest -L`:

```bash
ctest --test-dir build --output-on-failure --parallel $(nproc) -L unit
```

Never select a layer with `-R` instead. Catch2 registers each test under its `TEST_CASE` name, not the name of the binary it was compiled into, so `ctest -R unit` matches whichever test cases happen to have "unit" somewhere in their description and silently misses the rest. `-L` matches the label, which is exact.

`SKIP_RETURN_CODE 4` is what makes Catch2's `SKIP()` macro work. `SKIP()` exits the test binary with code 4; without this property CTest sees a non-zero exit and reports a **failure**. Any layer that can skip (integration always, unit where a test needs an optional file) must set it, or the first skip turns CI red.

Use `ctest --test-dir build` rather than `cd build && ctest`; it needs no subshell and works from any directory.

## Unit tests

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

A function gets a unit test if its behaviour can be provoked from data the test itself can build. That covers far more than pure logic: a parser gets a unit test driven by a synthetic file written to a temp directory, a socket layer gets one driven over loopback, an archive reader gets one against an archive the fixture assembled in memory. Reach for a fixture rather than reaching for the integration layer.

A function only escapes to the integration layer when the input cannot be synthesised: when the test is meaningful precisely because the data is real (an actual 1.12.1 client archive, a live server's handshake).

Every bug fix must include a test that reproduces the bug before the fix, in whichever layer the bug lives.

### Fixtures

Fixtures are plain C++ structs that set up and tear down state, function-scoped: constructed at the start of each test and destroyed at the end. Each lives in its own header under `test/fixtures/`, and they are shared across every layer that needs them.

A fixture that builds synthetic input is what keeps tests in the unit layer, so it earns its keep quickly:

```cpp
// test/fixtures/synthetic_archive.h
#pragma once

/// Builds a minimal in-memory archive for tests that need a real one to read
struct SyntheticArchive {
    std::vector<std::byte> bytes_;

    SyntheticArchive() { /* assemble header, table, entries */ }
};
```

Add the fixtures directory to each test target's include path so tests include them by name:

```cmake
target_include_directories(mylib_unit_tests PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}/fixtures")
```

A fixture that owns a temporary directory must remove it in its destructor and must swallow the cleanup error: a failure there must not throw out of a destructor and mask the assertion that actually failed.

### Assert on the project's own exceptions

Where code under test throws, assert on the concrete type from the library's hierarchy (see the error handling fragment), never on `std::exception`. A test that accepts any exception passes when the wrong thing goes wrong:

```cpp
// Good
REQUIRE_THROWS_AS(ParseVersion("not-a-version"), mylib::ParseError);

// Bad - a typo that throws std::bad_alloc would satisfy this
REQUIRE_THROWS(ParseVersion("not-a-version"));
```

## Makefile targets

```makefile
JOBS ?= $(shell nproc 2>/dev/null || echo 4)

.PHONY: test
test: ## Run Catch2 unit tests
	ctest --test-dir build --output-on-failure --parallel $(JOBS) -L unit

.PHONY: test_verbose
test_verbose: ## Run unit tests with verbose Catch2 output
	ctest --test-dir build --verbose -L unit
```

`test` is the everyday target and runs the unit layer alone, because that is the layer that always works: it needs no client data, no server, and no shipped binary. The other layers get their own targets, each named for what it needs, and a `test_all` where a project wants everything at once. See testing-integration.md and testing-functional.md.
