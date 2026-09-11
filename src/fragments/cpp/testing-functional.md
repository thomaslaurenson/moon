# C++ functional testing

The subprocess/functional testing layer, which spawns the compiled binary and verifies its CLI behaviour end-to-end. Assumes cpp/testing.md and the tier fragment.

Applies to any tier that ships a binary: an application, or a library with a bundled CLI. A plain library has no compiled binary to spawn and stops at the unit and integration layers.

## Structure addition

```
test/
  functional/             # Catch2 functional tests against the compiled binary
    test_create.cpp       # mirrors the create subcommand
    test_list.cpp         # mirrors the list subcommand
  subprocess_helper.h     # cross-platform subprocess runner
  subprocess_helper.cpp
extern/
  subprocess.h/           # submodule - pinned to a specific commit
```

- Functional tests mirror CLI subcommands or user-facing behaviour, not internal source files

Register the target with a `functional` label, and with `SKIP_RETURN_CODE` so an optional-data skip is not reported as a failure:

```cmake
catch_discover_tests(myapp_functional_tests
    PROPERTIES LABELS "functional" SKIP_RETURN_CODE 4)
```

See cpp/testing.md for why the label rather than `-R`, and why the skip code matters.

## Subprocess helper

Functional tests have to start a process and capture what it did, and the standard library gives you no way to do that. The layer therefore rests on a small helper, `test/subprocess_helper.h` and `test/subprocess_helper.cpp`.

**The interface is fixed; the implementation is not.** Every test in the layer is written against `Run()`, so that shape is the rule: it takes a binary path and an argument list, and returns a `RunResult` carrying `stdout_output`, `stderr_output`, `returncode` and `timed_out`, with a `RunOptions` for optional stdin input and a working directory. How those are obtained is a project's own business. These projects build the helper on `subprocess.h` (below), which is the part that handles the platform differences; anything presenting the same interface serves equally well.

Usage in a functional test:

```cpp
#include "subprocess_helper.h"

TEST_CASE("create: target does not exist", "[create]") {
    auto result = Run(MYAPP_BINARY_PATH, {"create", "no-such-file"});
    REQUIRE(result.returncode == 1);
}
```

The binary path is baked in at CMake configure time via `target_compile_definitions`; see the tier fragment for the pattern.

### subprocess.h dependency

The subprocess helper is built on top of `subprocess.h`, a small cross-platform C library. Pin it as a git submodule under `extern/subprocess.h`:

```bash
git submodule add https://github.com/sheredom/subprocess.h.git extern/subprocess.h
cd extern/subprocess.h && git checkout <commit-hash>
```

Always pin to an immutable reference: a release tag or a commit hash, never a moving branch name.

## The test environment fixture

The general fixture conventions live in cpp/testing.md and apply here unchanged; `test/fixtures/` is shared by every layer.

`test_environment.h` is the one fixture that is not function-scoped: it is a singleton holding the CMake-baked paths (`MYAPP_BINARY_PATH`, `MYAPP_TEST_DIR`). A singleton is right here and nowhere else, because these values are constant for the whole run and cannot vary per test:

```cpp
// test/fixtures/test_environment.h
#pragma once
#include <filesystem>
#include <string>

namespace fs = std::filesystem;

/// Singleton that exposes CMake-baked build and source paths to functional tests.
struct TestEnvironment {
    static const TestEnvironment &Instance() {
        static TestEnvironment env;
        return env;
    }

    const fs::path binary_path{MYAPP_BINARY_PATH};
    const fs::path test_dir{MYAPP_TEST_DIR};

private:
    TestEnvironment() = default;
};
```

All other fixtures (for example `TestFiles`) are ordinary function-scoped structs that include `test_environment.h` when they need the binary or test directory paths.

```cpp
// test/fixtures/test_files.h
#pragma once
#include <filesystem>

#include "test_environment.h"

namespace fs = std::filesystem;

/// Creates the static input files used across functional tests.
struct TestFiles {
    fs::path files_dir;

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
    auto result =
        Run(MYAPP_BINARY_PATH, {"add", (files.files_dir / "sample.txt").string(), "out.dat"});
    REQUIRE(result.returncode == 0);
}
```

## Asserting on CLI output

Use a `LinesToSet` helper to split stdout or stderr into a set of lines for order-independent comparison. Define it in an anonymous namespace at the top of each functional test file:

```cpp
namespace {

std::set<std::string> LinesToSet(const std::string &output, bool skip_empty = false) {
    std::set<std::string> result;
    std::istringstream stream(output);
    std::string line;
    while (std::getline(stream, line)) {
        if (!line.empty() && line.back() == '\r')
            line.pop_back();
        if (skip_empty && line.empty())
            continue;
        result.insert(line);
    }
    return result;
}

} // namespace
```

Usage:

```cpp
auto output = LinesToSet(result.stdout_output);
std::set<std::string> expected = {"cats.txt", "dogs.txt"};
REQUIRE(output == expected);
```

## Platform differences

Inputs and expected values are handled differently. Prefer a portable input; conditionalise only an expectation that genuinely cannot be unified.

### Inputs: choose a value that means the same thing everywhere

**Never pass a POSIX-absolute path as a command-line argument.** On Windows a leading `/` is option syntax, not a path, and argument parsers honour it: CLI11 defaults `allow_windows_style_options` to `true` there, so `/does/not/exist` is consumed as a flag and never reaches the code under test. The test then fails with a parse error that looks nothing like the file error it was written to assert:

```cpp
// Wrong - parsed as an option on Windows, exits 106 (CLI11 RequiredError)
Run(binary, {"info", "/does/not/exist.bin"});

// Right - means the same thing on every platform
Run(binary, {"info", "no-such-file.bin"});
```

A relative path needs no `#ifdef`, which is the point: a conditional here would compile two tests that assert different things, when one value works for both. Reach for a portable input first and a branch only when there is not one.

An application can opt out of the collision with `app.allow_windows_style_options(false)`, and should if it has no `/x` style options of its own: the convention then costs a path shape and buys nothing. That is a decision about the CLI, though, not a substitute for portable test inputs.

### Expectations: `#ifdef` only where the value really differs

Use `#ifdef _WIN32` for expected values that differ between Windows and POSIX; for example file sizes that differ due to CRLF vs LF line endings. Never use runtime platform detection in tests:

```cpp
#ifdef _WIN32
    std::string expected_size = "1383";
#else
    std::string expected_size = "1381";
#endif
REQUIRE(result.stdout_output.find(expected_size) != std::string::npos);
```

## Skipping tests with optional dependencies

Use Catch2's `SKIP()` macro when a test depends on a file or resource that may not be present in all environments:

```cpp
TEST_CASE("verify signature", "[verify]") {
    fs::path data = TestEnvironment::Instance().test_dir / "data" / "sample.dat";
    if (!fs::exists(data)) {
        SKIP("Test data not found - see the README for how to obtain it");
    }
    // test body
}
```

This is why the target sets `SKIP_RETURN_CODE 4`: `SKIP()` exits the binary with code 4, and without that property CTest reports the skip as a failure.

## Tags

Tag each functional `TEST_CASE` with the subcommand or feature under test:

```cpp
TEST_CASE("create versions", "[create]") { ... }
TEST_CASE("list with filter", "[list]") { ... }
```

Run a subset during development:

```bash
./build/dev/bin/myapp_functional_tests [create]
```

## Asserting on the CLI contract

The functional layer owns the exit-code contract, because it is the only layer that can see it. Assert the code, not just the output:

```cpp
TEST_CASE("create: target does not exist", "[create]") {
    auto result = Run(MYAPP_BINARY_PATH, {"create", "no-such-file"});
    REQUIRE(result.returncode == 1);
    REQUIRE_THAT(result.stderr_output, Catch::Matchers::ContainsSubstring("does not exist"));
}
```

An expected failure the user can act on exits 1 and explains itself on stderr; an unexpected one exits 2. See the error handling fragment for where those codes come from. A test asserting only that the command failed would pass if the binary crashed instead.

## Makefile targets

```makefile
.PHONY: test_functional
test_functional: build ## Run Catch2 functional tests against the built binary
	ctest --test-dir build/dev --output-on-failure --parallel $(JOBS) -L functional
```

`test_functional` depends on `build`: the layer spawns the compiled binary, so a stale or absent one is a failure with a confusing message rather than a test result.

This layer does not change cpp/testing.md's `test` target, which stays the unit layer alone and remains the everyday command. `test_all` is what runs both in one go, and it is defined once in cpp/testing.md rather than here, so a tier with a functional layer and a tier without see the same target.
