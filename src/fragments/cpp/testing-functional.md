# C++ functional testing (applications)

The subprocess/functional testing layer, which spawns the compiled binary and verifies its CLI behaviour end-to-end. Assumes cpp/testing.md and cpp/cmake-app.md. Only applies to applications; a library has no compiled binary to spawn.

## Structure addition

```
test/
  functional/             # Catch2 functional tests against the compiled binary
    test_create.cpp       # mirrors the create subcommand
    test_list.cpp         # mirrors the list subcommand
  subprocess_helper.h      # cross-platform subprocess runner
  subprocess_helper.cpp
extern/
  subprocess.h/           # submodule - pinned to a specific commit
```

- Functional tests mirror CLI subcommands or user-facing behaviour, not internal source files

## Subprocess helper

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

The binary path is baked in at CMake configure time via `target_compile_definitions`; see the cmake-app fragment for the pattern.

### subprocess.h dependency

The subprocess helper is built on top of `subprocess.h`, a small cross-platform C library. Pin it as a git submodule under `extern/subprocess.h`:

```bash
git submodule add https://github.com/sheredom/subprocess.h.git extern/subprocess.h
cd extern/subprocess.h && git checkout <commit-hash>
```

Always pin to an immutable reference: a release tag or a commit hash, never a moving branch name.

## Fixtures

Fixtures are plain C++ structs that set up and tear down state for tests. They are function-scoped; constructed at the start of each test and destroyed at the end. Each fixture lives in its own header in `test/fixtures/`.

`test_environment.h` is the one exception: it is a singleton that holds the CMake-baked paths (`MYAPP_BINARY_PATH`, `MYAPP_TEST_DIR`). It provides a single access point for values that are constant across the entire test run and do not vary per test:

```cpp
// test/fixtures/test_environment.h
#pragma once
#include <filesystem>
#include <string>

namespace fs = std::filesystem;

/// Singleton that exposes CMake-baked build and source paths to functional tests.
struct TestEnvironment {
    static const TestEnvironment &instance() {
        static TestEnvironment env;
        return env;
    }

    const fs::path binary_path { MYAPP_BINARY_PATH };
    const fs::path test_dir    { MYAPP_TEST_DIR };

private:
    TestEnvironment() = default;
};
```

All other fixtures (e.g. `TestFiles`) are function-scoped structs that include `test_environment.h` when they need the binary or test directory paths.

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

## Asserting on CLI output

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

## Platform differences

Use `#ifdef _WIN32` for expected values that differ between Windows and POSIX; for example file sizes that differ due to CRLF vs LF line endings. Never use runtime platform detection in tests:

```cpp
#ifdef _WIN32
    std::string expected_size = "1383";
#else
    std::string expected_size = "1381";
#endif
REQUIRE(result.stdout_output_.find(expected_size) != std::string::npos);
```

## Skipping tests with optional dependencies

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

## Tags

Tag each functional `TEST_CASE` with the subcommand or feature under test:

```cpp
TEST_CASE("create versions", "[create]") { ... }
TEST_CASE("list with filter", "[list]") { ... }
```

Run a subset during development:

```bash
./build/bin/myapp_functional_tests [create]
```

## Makefile targets

```makefile
.PHONY: test
test: test_unit test_functional ## Run all tests

.PHONY: test_functional
test_functional: ## Run Catch2 functional tests
	cmake --build build && cd build && ctest --output-on-failure -R functional
```

This replaces cpp/testing.md's standalone `test_unit`-only target with a combined `test` target that runs both layers.
