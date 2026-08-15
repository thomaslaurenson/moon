# C++ testing standards

Standards and conventions for testing C++ projects, and the unit layer in full. Universal to every tier; the layers built on top live in the testing-integration and testing-functional fragments.

## The four layers

Every test in a C++ project belongs to exactly one of four layers. All four live under `test/`; a test binary is never built anywhere else.

- **unit** - tests logic. Links the library target. Free to use fixtures, temporary files, synthetic data and loopback sockets: the test creates whatever it needs. Always built, always run.
- **integration** - tests against real data the machine must already have: a real dataset install, a live server, a real API. Links the library target. Opt-in, and skips cleanly when the data is absent. See testing-integration.md.
- **functional** - tests the compiled binary, spawned as a subprocess and driven through its CLI. Links neither the library nor the binary. See testing-functional.md.
- **fuzz** - libFuzzer harnesses driving a parser with hostile input. Links the library target. Built on demand, Clang only. See testing-fuzz.md.

The dividing line between unit and integration is **not** whether a test touches the filesystem. A unit test that writes a synthetic archive to a temp directory and reads it back is still a unit test: it tests logic, and it brought its own data. What makes a test an integration test is depending on an environment it cannot construct: a real dataset install, a running server, a populated API. That is also what makes it opt-in, since CI has none of those things.

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
    archive/
      test_archive.cpp    # mirrors src/archive/archive.cpp
      test_crypto.cpp
    record/
      test_reader.cpp
  integration/            # tests against real data; opt-in
    test_archive_reader.cpp
  functional/             # tests the compiled binary; only where one ships
    test_create.cpp
  fuzz/                   # libFuzzer harnesses; built on demand
    fuzz_archive.cpp
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
ctest --test-dir build/dev --output-on-failure -L unit
```

Never select a layer with `-R` instead. Catch2 registers each test under its `TEST_CASE` name, not the name of the binary it was compiled into, so `ctest -R unit` matches whichever test cases happen to have "unit" somewhere in their description and silently misses the rest. `-L` matches the label, which is exact.

`SKIP_RETURN_CODE 4` is what makes Catch2's `SKIP()` macro work. `SKIP()` exits the test binary with code 4; without this property CTest sees a non-zero exit and reports a **failure**. Any layer that can skip (integration always, unit where a test needs an optional file) must set it, or the first skip turns CI red.

Use `ctest --test-dir build/dev` rather than `cd build/dev && ctest`; it needs no subshell and works from any directory.

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
./build/dev/bin/mytarget_unit_tests [helpers]
./build/dev/bin/mytarget_unit_tests [config]
```

### What to unit test

A function gets a unit test if its behaviour can be provoked from data the test itself can build. That covers far more than pure logic: a parser gets a unit test driven by a synthetic file written to a temp directory, a socket layer gets one driven over loopback, an archive reader gets one against an archive the fixture assembled in memory. Reach for a fixture rather than reaching for the integration layer.

A function only escapes to the integration layer when the input cannot be synthesised: when the test is meaningful precisely because the data is real (an actual production dataset, a live server's handshake).

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

`JOBS` is declared once in the CMake fragment's `build` target and reused here.

```makefile
.PHONY: test
test: build ## Run Catch2 unit tests
	ctest --test-dir build/dev --output-on-failure --parallel $(JOBS) -L unit

.PHONY: test_verbose
test_verbose: build ## Run unit tests with verbose Catch2 output
	ctest --test-dir build/dev --verbose -L unit

.PHONY: test_all
test_all: build ## Run every test layer built into the current configure
	ctest --test-dir build/dev --output-on-failure --parallel $(JOBS)
```

Every test target depends on `build`, and `build` depends on `configure`, so any of them works from a fresh clone with no prior step. That chain is what makes `make ci` runnable on a clean checkout; without it `ctest` fails on a missing `build/dev` with a message about the directory rather than about the missing build. Reconfiguring an existing tree is a no-op that costs under a second, and CMake keeps every cached `-D` from the original `make configure CMAKE_ARGS=...`, so the repeat does not discard a CI job's compiler or `-Werror` setting.

`test_all` is defined here, in the fragment every tier includes, rather than in the layer fragments. It runs whatever the current configure contains, which is the unit layer alone unless another was configured in, so it belongs with the universal targets and not with any one layer.

`test` is the everyday target and runs the unit layer alone, because that is the layer that always works: it needs no external data, no server, and no shipped binary. The other layers get their own targets, each named for what it needs, and a `test_all` where a project wants everything at once. See testing-integration.md and testing-functional.md.

## Coverage

Coverage is measured over the unit layer using clang's source-based instrumentation. It gets its own `build/coverage` directory: the instrumentation changes code generation, and the compiler is pinned to clang whatever the everyday build uses. `CLANG_CXX`, `JOBS`, `LLVM_PROFDATA` and `LLVM_COV` are declared once in the CMake fragment and reused here.

```makefile
.PHONY: configure_coverage
configure_coverage: ## Configure build/coverage with clang source-based coverage
	cmake -B build/coverage \
	  -DCMAKE_BUILD_TYPE=Debug \
	  -DMYLIB_BUILD_TESTING=ON \
	  -DCMAKE_CXX_COMPILER=$(CLANG_CXX) \
	  -DCMAKE_CXX_FLAGS="-fprofile-instr-generate -fcoverage-mapping" \
	  -DCMAKE_EXE_LINKER_FLAGS="-fprofile-instr-generate" \
	  $(CMAKE_ARGS)

.PHONY: test_coverage
test_coverage: configure_coverage ## Report unit-test coverage
	cmake --build build/coverage --parallel $(JOBS) --target mylib_unit_tests
	LLVM_PROFILE_FILE=build/coverage/unit.profraw ./build/coverage/bin/mylib_unit_tests
	$(LLVM_PROFDATA) merge -o build/coverage/unit.profdata build/coverage/unit.profraw
	$(LLVM_COV) report ./build/coverage/bin/mylib_unit_tests \
	  -instr-profile=build/coverage/unit.profdata \
	  --ignore-filename-regex="extern/|test/"
```

- The unit layer alone, for the same reason `test_asan` uses it: that layer needs no external data and runs anywhere, so the number means the same thing on every machine and in every checkout. A figure that moves depending on whether the developer happens to have the integration dataset is not a figure worth publishing.
- `--ignore-filename-regex` keeps vendored code and the tests themselves out of the report. A project that counts its own test files reports a number that climbs as tests are added and says nothing about how well the library is covered.
- The report goes to stdout. Nothing publishes it automatically: the percentage is copied by hand into the static coverage badge on each release, which is what cpp/badges.md asks for. This target is where that number comes from.
