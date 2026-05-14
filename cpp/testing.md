# C++ Testing Standards

Standards and conventions for testing C++ projects.

## Overview

Two testing layers are used in combination:

- **Catch2**: unit tests for pure logic functions with no external dependencies
- **Catch2**: functional tests that spawn the compiled binary as a subprocess and verify its CLI behaviour end-to-end

Both layers must pass before a change is considered complete. All tests are written in C++; there is no Python test layer.

---

## Structure

```
test/
  fixtures/               # shared fixture headers, one file per fixture
    test_environment.h
    test_files.h
  functional/             # Catch2 functional tests against the compiled binary
    test_create.cpp       # mirrors the create subcommand
    test_list.cpp         # mirrors the list subcommand
  unit/                   # Catch2 unit tests for internal logic
    test_helpers.cpp      # mirrors src/helpers.cpp
    test_config.cpp       # mirrors src/config.cpp
  subprocess_helper.h     # cross-platform subprocess runner
  subprocess_helper.cpp
  CMakeLists.txt
```

- One unit test file per source file, named `test_<source>.cpp`
- Functional tests mirror CLI subcommands or user-facing behaviour, not internal source files
- Fixtures live in `test/fixtures/`, one header per fixture

---

## Catch2 Setup

Pin Catch2 as a git submodule under `extern/Catch2` so the version is controlled and no system install is required:

```bash
git submodule add https://github.com/catchorg/Catch2.git extern/Catch2
cd extern/Catch2 && git checkout v3.6.0
```

Catch2 v3 is not a single-header library; it is a compiled library with multiple headers. Always use it as a submodule, never copy individual headers.

In the root `CMakeLists.txt`:

```cmake
option(BUILD_TESTING "Build tests" ON)

if(BUILD_TESTING)
    add_subdirectory(extern/Catch2)
    add_subdirectory(test)
endif()
```

In `test/CMakeLists.txt`, two binaries are defined; one for unit tests and one for functional tests:

```cmake
add_executable(myapp_unit_tests
    unit/test_helpers.cpp
    unit/test_config.cpp
)
target_link_libraries(myapp_unit_tests PRIVATE Catch2::Catch2WithMain)

add_executable(myapp_functional_tests
    subprocess_helper.cpp
    functional/test_create.cpp
    functional/test_list.cpp
)
target_link_libraries(myapp_functional_tests PRIVATE Catch2::Catch2WithMain)

include(CTest)
include(Catch)
catch_discover_tests(myapp_unit_tests)
catch_discover_tests(myapp_functional_tests)
```

---

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
TEST_CASE("next_power_of_two", "[helpers]") { … }
TEST_CASE("parse_config", "[config]") { … }
```

Run a subset during development:

```bash
./build/bin/myapp_unit_tests [helpers]
./build/bin/myapp_unit_tests [config]
```

### What to unit test

A function gets a unit test if it can be called without:

- A real file on disk
- A network connection
- A third-party library handle or session
- Any global system state

Pure logic functions (string manipulation, maths, parsing, pattern matching) always get unit tests. Functions tightly coupled to external libraries or the filesystem are covered by functional tests instead.

Every bug fix must include a unit test that reproduces the bug before the fix.

---

## Functional Tests

Functional tests spawn the compiled binary as a subprocess and assert on its stdout, stderr, and exit code. They are black-box tests; they never link against application source files.

### Subprocess helper

Every project that has functional tests includes a cross-platform subprocess helper: `test/subprocess_helper.h` and `test/subprocess_helper.cpp`. This helper is not written from scratch each time; copy it from an existing project that already uses this pattern.

The helper exposes a `run()` function that returns a `RunResult` containing `stdout_output_`, `stderr_output_`, `returncode_`, and `timed_out_`. A `RunOptions` struct controls optional stdin input and working directory.

Usage in a functional test:

```cpp
#include "../subprocess_helper.h"

TEST_CASE("create: target does not exist", "[create]") {
    auto result = run(MYAPP_BINARY_PATH, {"create", "/does/not/exist"});
    REQUIRE(result.returncode_ == 1);
}
```

The binary path is baked in at CMake configure time via `target_compile_definitions`; see `tools/cmake.md` for the pattern.

### subprocess.h dependency

The subprocess helper is built on top of `subprocess.h`, a small cross-platform C library. Pin it as a git submodule under `extern/subprocess.h`:

```bash
git submodule add https://github.com/sheredom/subprocess.h.git extern/subprocess.h
cd extern/subprocess.h && git checkout <commit-hash>
```

Always pin to a specific commit hash, never a branch name.

### Fixtures

Fixtures are plain C++ structs that set up and tear down state for tests. They are function-scoped; constructed at the start of each test and destroyed at the end. Each fixture lives in its own header in `test/fixtures/`.

```cpp
// test/fixtures/test_files.h
#pragma once
#include <filesystem>
#include "test_environment.h"

namespace fs = std::filesystem;

/// Creates the static input files used across functional tests.
struct TestFiles {
    fs::path files_dir_;

    TestFiles() {
        // create files, set timestamps etc.
    }

    ~TestFiles() = default; // or clean up if needed
};
```

Instantiate in a test:

```cpp
TEST_CASE("add file to archive", "[add]") {
    TestFiles files;
    auto result = run(MYAPP_BINARY_PATH,
                      {"add", (files.files_dir_ / "sample.txt").string(), "out.dat"});
    REQUIRE(result.returncode_ == 0);
}
```

### Asserting on CLI output

Use a `lines_to_set` helper to split stdout or stderr into a set of lines for order-independent comparison. Define it as a static function at the top of each functional test file:

```cpp
static std::set<std::string> lines_to_set(const std::string &output,
                                           bool skip_empty = false) {
    std::set<std::string> result;
    std::istringstream stream(output);
    std::string line;
    while (std::getline(stream, line)) {
        if (!line.empty() && line.back() == '\r') line.pop_back();
        if (skip_empty && line.empty()) continue;
        result.insert(line);
    }
    return result;
}
```

Usage:

```cpp
auto output = lines_to_set(result.stdout_output_);
std::set<std::string> expected = {"cats.txt", "dogs.txt"};
REQUIRE(output == expected);
```

### Platform differences

Use `#ifdef _WIN32` for expected values that differ between Windows and POSIX; for example file sizes that differ due to CRLF vs LF line endings. Never use runtime platform detection in tests:

```cpp
#ifdef _WIN32
    std::string expected_size = "1383";
#else
    std::string expected_size = "1381";
#endif
REQUIRE(result.stdout_output_.find(expected_size) != std::string::npos);
```

### Skipping tests with optional dependencies

Use Catch2's `SKIP()` macro when a test depends on a file or resource that may not be present in all environments:

```cpp
TEST_CASE("verify signature", "[verify]") {
    fs::path data = env.test_dir_ / "data" / "sample.dat";
    if (!fs::exists(data)) {
        SKIP("Test data not found - run scripts/download_test_data.sh");
    }
    // test body
}
```

### Tags

Tag each functional `TEST_CASE` with the subcommand or feature under test:

```cpp
TEST_CASE("create versions", "[create]") { … }
TEST_CASE("list with filter", "[list]") { … }
```

Run a subset during development:

```bash
./build/bin/myapp_functional_tests [create]
```

---

## Makefile Targets

```makefile
.PHONY: test test_unit test_functional

test: test_unit test_functional ## Run all tests

test_unit: ## Run Catch2 unit tests
	cmake --build build && cd build && ctest --output-on-failure -R unit

test_functional: ## Run Catch2 functional tests
	cmake --build build && cd build && ctest --output-on-failure -R functional
```
