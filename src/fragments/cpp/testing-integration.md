# C++ integration testing

The layer that tests against real data: a game client install, a live server, a populated API. Assumes cpp/testing.md and the tier fragment.

An integration test links the library target exactly as a unit test does. What separates it is not that it touches the filesystem (a unit test may do that freely with its own fixtures) but that it depends on an environment it cannot construct. A test is an integration test when the point of it is that the data is real: the actual 1.12.1 archives rather than a synthetic one, a running server's real handshake rather than a recorded blob.

That dependency is also why the layer is opt-in. CI has no game install and no server, so integration tests are excluded from the build entirely unless asked for, and skip cleanly at runtime when the environment is missing.

## Which tiers need it

Any tier can have an integration layer; whether it does depends on whether the project has real data worth testing against. A library reading a proprietary format needs one. A CLI wrapping that library may not: its functional layer already drives real files through the binary.

Never use this layer as a dumping ground for tests that are awkward to write. If the input can be synthesised, the test belongs in the unit layer with a fixture; see cpp/testing.md.

## Options

The root `CMakeLists.txt` declares a project-scoped option to build the layer, plus a cache variable naming the data it needs:

```cmake
option(MYLIB_INTEGRATION "Build integration tests (requires a real client install)" OFF)
set(MYLIB_CLIENT_PATH "" CACHE PATH "Path to a client install for integration tests")
```

Default `OFF`: the integration binary is not built at all in a normal configure, so a developer without the data never sees it and never has to skip it. Project-scoped names, as always, so a consumer's own `INTEGRATION` flag cannot reach in here.

## Test target

Guard the whole target on the option. It links the library through its alias, the same as the unit binary:

```cmake
if(MYLIB_INTEGRATION)
    add_executable(mylib_integration_tests
        integration/test_mpq_client.cpp
        integration/test_dbc_client.cpp
    )

    target_include_directories(mylib_integration_tests PRIVATE
        "${CMAKE_CURRENT_SOURCE_DIR}/fixtures"
    )

    target_link_libraries(mylib_integration_tests PRIVATE mylib::mylib Catch2::Catch2WithMain)

    # Bake the data path in at configure time when given. If absent, the tests
    # fall back to an environment variable at runtime and skip cleanly if
    # neither is set.
    if(MYLIB_CLIENT_PATH)
        target_compile_definitions(mylib_integration_tests PRIVATE
            MYLIB_CLIENT_PATH="${MYLIB_CLIENT_PATH}")
    endif()

    catch_discover_tests(mylib_integration_tests
        PROPERTIES LABELS "integration" SKIP_RETURN_CODE 4)
endif()
```

`SKIP_RETURN_CODE 4` is mandatory here, not optional. This layer skips by design whenever the data is absent, and without that property CTest reports every skip as a failure; see cpp/testing.md.

## Finding the data

Resolve the environment in one place, in a fixture header, checking the compile-time define first and an environment variable second. Return an optional rather than throwing, so each test decides whether to skip:

```cpp
// test/fixtures/client_environment.h
#pragma once
#include <cstdlib>
#include <filesystem>
#include <optional>

namespace fs = std::filesystem;

namespace mylib::testing {

/// Returns the client install directory, or an empty optional
///
/// Checks in order:
///   1. MYLIB_CLIENT_PATH compile-time define (from -DMYLIB_CLIENT_PATH=...)
///   2. MYLIB_CLIENT_PATH environment variable at runtime
///
/// @return The install root, or nullopt if neither yields a real directory.
inline std::optional<fs::path> ClientPath() {
#ifdef MYLIB_CLIENT_PATH
    {
        fs::path p { MYLIB_CLIENT_PATH };
        if (fs::is_directory(p)) {
            return p;
        }
    }
#endif
    if (const char *env = std::getenv("MYLIB_CLIENT_PATH")) {
        fs::path p { env };
        if (fs::is_directory(p)) {
            return p;
        }
    }
    return std::nullopt;
}

}  // namespace mylib::testing
```

Two sources rather than one because they serve different people: the CMake define suits a developer who configures once and forgets, the environment variable suits a machine where the path is already exported. Never hardcode a path, and never guess at a default install location.

## Skipping

Every integration test opens by resolving the environment and skipping if it is absent. The message says what is missing and how to supply it:

```cpp
#include <catch2/catch_test_macros.hpp>
#include "client_environment.h"

TEST_CASE("reads entries from a real client archive", "[mpq]") {
    auto client = mylib::testing::ClientPath();
    if (!client) {
        SKIP("No client install found - set MYLIB_CLIENT_PATH");
    }

    auto chain = mylib::testing::BuildChain(*client);
    REQUIRE(chain.Contains("DBFilesClient\\Map.dbc"));
}
```

Skip on a missing environment, never on a missing *feature*: a test that skips because the code under test is broken is a test that never runs. Once the data is present, the test is a normal test and a failure is a failure.

Distinguish required from optional data inside a fixture. Data that every real install has is required, and its absence throws rather than skips: a chain that silently builds smaller lets tests pass while verifying less.

## Makefile targets

Integration needs its own configure, because the option is off by default:

```makefile
CLIENT_PATH ?= $(MYLIB_CLIENT_PATH)

.PHONY: configure_integration
configure_integration: ## Configure with integration tests (requires: CLIENT_PATH)
	@if [ -z "$(CLIENT_PATH)" ]; then \
	  echo "Error: set CLIENT_PATH=/path/to/client or MYLIB_CLIENT_PATH" >&2; exit 1; \
	fi
	cmake -B build \
	  -DCMAKE_BUILD_TYPE=Debug \
	  -DCMAKE_EXPORT_COMPILE_COMMANDS=ON \
	  -DMYLIB_INTEGRATION=ON \
	  -DMYLIB_CLIENT_PATH="$(CLIENT_PATH)"

.PHONY: test_integration
test_integration: ## Run integration tests (requires: configure_integration first)
	ctest --test-dir build --output-on-failure -L integration

.PHONY: test_all
test_all: ## Run every test layer built into the current configure
	ctest --test-dir build --output-on-failure --parallel $(JOBS)
```

The guard on `CLIENT_PATH` fails the configure with an actionable message rather than producing a build whose integration tests all skip. `test_all` runs whatever the current configure contains, which is the unit layer alone unless integration was configured in.

## CI

Integration tests do not run in CI. The data is proprietary, large, or a live service, and none of that belongs in a workflow. CI runs `make test`, which is the unit layer; the integration layer is a local tool for the developer who has the data.

Never work around this by committing client data, downloading it in a workflow, or standing up the service in a container. If a behaviour needs covering in CI, synthesise the input and write a unit test.
