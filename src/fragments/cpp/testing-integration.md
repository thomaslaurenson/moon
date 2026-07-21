# C++ integration testing

The layer that tests against real data: a real dataset install, a live server, a populated API. Assumes cpp/testing.md and the tier fragment.

An integration test links the library target exactly as a unit test does. What separates it is not that it touches the filesystem (a unit test may do that freely with its own fixtures) but that it depends on an environment it cannot construct. A test is an integration test when the point of it is that the data is real: the actual production dataset rather than a synthetic one, a running server's real handshake rather than a recorded blob.

That dependency is also why the layer is opt-in. CI has no dataset install and no server, so integration tests are excluded from the build entirely unless asked for, and skip cleanly at runtime when the environment is missing.

## Which tiers need it

Any tier can have an integration layer; whether it does depends on whether the project has real data worth testing against. A library reading a proprietary format needs one. A CLI wrapping that library may not: its functional layer already drives real files through the binary.

Never use this layer as a dumping ground for tests that are awkward to write. If the input can be synthesised, the test belongs in the unit layer with a fixture; see cpp/testing.md.

## Options

The root `CMakeLists.txt` declares a project-scoped option to build the layer, plus a cache variable naming the data it needs:

```cmake
option(MYLIB_INTEGRATION "Build integration tests (requires a real dataset)" OFF)
set(MYLIB_INTEGRATION_DATA "" CACHE PATH "Path to the real dataset (file or directory) for integration tests")
```

Default `OFF`: the integration binary is not built at all in a normal configure, so a developer without the data never sees it and never has to skip it. Project-scoped names, as always, so a consumer's own `INTEGRATION` flag cannot reach in here.

## Test target

Guard the whole target on the option. It links the library through its alias, the same as the unit binary:

```cmake
if(MYLIB_INTEGRATION)
    add_executable(mylib_integration_tests
        integration/test_archive_reader.cpp
        integration/test_record_reader.cpp
    )

    target_include_directories(mylib_integration_tests PRIVATE
        "${CMAKE_CURRENT_SOURCE_DIR}/fixtures"
    )

    target_link_libraries(mylib_integration_tests PRIVATE mylib::mylib Catch2::Catch2WithMain)

    # Bake the data path in at configure time when given. If absent, the tests
    # fall back to an environment variable at runtime and skip cleanly if
    # neither is set.
    if(MYLIB_INTEGRATION_DATA)
        target_compile_definitions(mylib_integration_tests PRIVATE
            MYLIB_INTEGRATION_DATA="${MYLIB_INTEGRATION_DATA}")
    endif()

    catch_discover_tests(mylib_integration_tests
        PROPERTIES LABELS "integration" SKIP_RETURN_CODE 4)
endif()
```

`SKIP_RETURN_CODE 4` is mandatory here, not optional. This layer skips by design whenever the data is absent, and without that property CTest reports every skip as a failure; see cpp/testing.md.

## Finding the data

Resolve the environment in one place, in a fixture header, checking the compile-time define first and an environment variable second. Return an optional rather than throwing, so each test decides whether to skip:

```cpp
// test/fixtures/integration_data.h
#pragma once
#include <cstdlib>
#include <filesystem>
#include <optional>

namespace fs = std::filesystem;

namespace mylib::testing {

/// Returns the path to the real integration dataset, or an empty optional.
///
/// Checks in order:
///   1. MYLIB_INTEGRATION_DATA compile-time define (from -DMYLIB_INTEGRATION_DATA=...)
///   2. MYLIB_INTEGRATION_DATA environment variable at runtime
///
/// The path may be a single file or a directory; each test decides how to use it.
/// @return The dataset path, or nullopt if neither source points at something that exists.
inline std::optional<fs::path> IntegrationDataPath() {
#ifdef MYLIB_INTEGRATION_DATA
    {
        fs::path p { MYLIB_INTEGRATION_DATA };
        if (fs::exists(p)) {
            return p;
        }
    }
#endif
    if (const char *env = std::getenv("MYLIB_INTEGRATION_DATA")) {
        fs::path p { env };
        if (fs::exists(p)) {
            return p;
        }
    }
    return std::nullopt;
}

}  // namespace mylib::testing
```

Two sources rather than one because they serve different people: the CMake define suits a developer who configures once and forgets, the environment variable suits a machine where the path is already exported. Never hardcode a path, and never guess at a default install location.

`fs::exists` rather than `fs::is_directory`, so the one resolver accepts a dataset that is a single file or a whole directory; a test that needs a particular shape asserts it itself. A dependency that is a live service rather than data on disk follows the same shape with a separate variable: an `MYLIB_INTEGRATION_ENDPOINT` holding a URL instead of a path, resolved from the same compile-time-define-then-environment order and skipped the same way when unset.

## Skipping

Every integration test opens by resolving the environment and skipping if it is absent. The message says what is missing and how to supply it:

```cpp
#include <catch2/catch_test_macros.hpp>
#include "integration_data.h"

TEST_CASE("reads entries from a real dataset archive", "[archive]") {
    auto data = mylib::testing::IntegrationDataPath();
    if (!data) {
        SKIP("No integration data found - set MYLIB_INTEGRATION_DATA");
    }

    auto index = mylib::BuildIndex(*data);
    REQUIRE(index.Contains("records/main.dat"));
}
```

Skip on a missing environment, never on a missing *feature*: a test that skips because the code under test is broken is a test that never runs. Once the data is present, the test is a normal test and a failure is a failure.

Distinguish required from optional data inside a fixture. Data that every real install has is required, and its absence throws rather than skips: a chain that silently builds smaller lets tests pass while verifying less.

## Makefile targets

Integration needs its own configure, because the option is off by default:

```makefile
INTEGRATION_DATA ?= $(MYLIB_INTEGRATION_DATA)

.PHONY: configure_integration
configure_integration: ## Configure with integration tests (requires: INTEGRATION_DATA)
	@if [ -z "$(INTEGRATION_DATA)" ]; then \
	  echo "Error: set INTEGRATION_DATA=/path/to/dataset or MYLIB_INTEGRATION_DATA" >&2; exit 1; \
	fi
	cmake -B build \
	  -DCMAKE_BUILD_TYPE=Debug \
	  -DCMAKE_EXPORT_COMPILE_COMMANDS=ON \
	  -DMYLIB_INTEGRATION=ON \
	  -DMYLIB_INTEGRATION_DATA="$(INTEGRATION_DATA)"

.PHONY: test_integration
test_integration: ## Run integration tests (requires: configure_integration first)
	ctest --test-dir build --output-on-failure -L integration

.PHONY: test_all
test_all: ## Run every test layer built into the current configure
	ctest --test-dir build --output-on-failure --parallel $(JOBS)
```

The guard on `INTEGRATION_DATA` fails the configure with an actionable message rather than producing a build whose integration tests all skip. `test_all` runs whatever the current configure contains, which is the unit layer alone unless integration was configured in.

## CI

Integration tests do not run in CI. The data is proprietary, large, or a live service, and none of that belongs in a workflow. CI runs `make test`, which is the unit layer; the integration layer is a local tool for the developer who has the data.

Never work around this by committing the dataset, downloading it in a workflow, or standing up the service in a container. If a behaviour needs covering in CI, synthesise the input and write a unit test.
