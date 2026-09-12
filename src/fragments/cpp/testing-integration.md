# C++ integration testing

The layer that tests against something real that the repository does not contain: a handful of genuine input files, an installed product, a live server, a populated API. Assumes cpp/testing.md and the tier fragment.

An integration test links the library target exactly as a unit test does. What separates it is not that it touches the filesystem (a unit test may do that freely with its own fixtures) but that it depends on an environment it cannot construct. A test is an integration test when the point of it is that the input is real: a file produced by the software this one has to interoperate with rather than one a fixture assembled, a running server's real handshake rather than a recorded blob.

That dependency is also why the layer is opt-in. CI has none of those things, so integration tests are excluded from the build entirely unless asked for, and skip cleanly at runtime when what they need is missing.

## Which tiers need it

Any tier can have an integration layer; whether it does depends on whether the project has real data worth testing against. A library reading a proprietary format needs one. A CLI wrapping that library may not: its functional layer already drives real files through the binary.

Never use this layer as a dumping ground for tests that are awkward to write. If the input can be synthesised, the test belongs in the unit layer with a fixture; see cpp/testing.md.

## Options

The root `CMakeLists.txt` declares a project-scoped option to build the layer, plus a cache variable naming the data it needs:

```cmake
option(MYPROJ_INTEGRATION "Build integration tests (requires the real inputs)" OFF)
set(MYPROJ_INTEGRATION_DATA "" CACHE PATH "Path to the real inputs (file or directory) for integration tests")
```

Default `OFF`: the integration binary is not built at all in a normal configure, so a developer without the data never sees it and never has to skip it. Project-scoped names, as always, so a consumer's own `INTEGRATION` flag cannot reach in here.

## Test target

Guard the whole target on the option. It links the library through its alias, the same as the unit binary:

```cmake
if(MYPROJ_INTEGRATION)
    add_executable(myproj_integration_tests
        integration/test_archive_reader.cpp
        integration/test_record_reader.cpp
    )

    target_include_directories(myproj_integration_tests PRIVATE
        "${CMAKE_CURRENT_SOURCE_DIR}/fixtures"
    )

    target_link_libraries(myproj_integration_tests PRIVATE myproj::myproj Catch2::Catch2WithMain)

    # Bake the data path in at configure time when given. If absent, the tests
    # fall back to an environment variable at runtime and skip cleanly if
    # neither is set.
    if(MYPROJ_INTEGRATION_DATA)
        target_compile_definitions(myproj_integration_tests PRIVATE
            MYPROJ_INTEGRATION_DATA="${MYPROJ_INTEGRATION_DATA}")
    endif()

    catch_discover_tests(myproj_integration_tests
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

namespace myproj::testing {

/// Returns the path to the real integration dataset, or an empty optional
///
/// Checks in order:
///   1. MYPROJ_INTEGRATION_DATA compile-time define (from -DMYPROJ_INTEGRATION_DATA=...)
///   2. MYPROJ_INTEGRATION_DATA environment variable at runtime
///
/// The path may be a single file or a directory; each test decides how to use it.
/// @return The dataset path, or nullopt if neither source points at something that exists.
inline std::optional<fs::path> IntegrationDataPath() {
#ifdef MYPROJ_INTEGRATION_DATA
    {
        fs::path p{MYPROJ_INTEGRATION_DATA};
        if (fs::exists(p)) {
            return p;
        }
    }
#endif
    if (const char *env = std::getenv("MYPROJ_INTEGRATION_DATA")) {
        fs::path p{env};
        if (fs::exists(p)) {
            return p;
        }
    }
    return std::nullopt;
}

} // namespace myproj::testing
```

Two sources rather than one because they serve different people: the CMake define suits a developer who configures once and forgets, the environment variable suits a machine where the path is already exported.

Whether that variable has a default, and what it is, is the project's decision. The test that matters is whether the default means the same thing on every machine. A path inside the repository does: a conventional directory the project sets aside for inputs it cannot commit is the same path for everyone who clones it, so defaulting to it makes the layer work with no configuration once the files are in place. A path outside the repository does not: an install location varies by machine, by operating system and by how the thing was installed, so a default pointing there is a guess that is wrong more often than right. Guess at neither, and never hardcode a path in a test.

`fs::exists` rather than `fs::is_directory`, so the one resolver accepts a dataset that is a single file or a whole directory; a test that needs a particular shape asserts it itself. A dependency that is a live service rather than data on disk follows the same shape with a separate variable: an `MYPROJ_INTEGRATION_ENDPOINT` holding a URL instead of a path, resolved from the same compile-time-define-then-environment order and skipped the same way when unset.

## Skipping

Every integration test opens by resolving the environment and skipping if it is absent. The message says what is missing and how to supply it:

```cpp
#include <catch2/catch_test_macros.hpp>

#include "integration_data.h"

TEST_CASE("reads entries from a real dataset archive", "[archive]") {
    auto data = myproj::testing::IntegrationDataPath();
    if (!data) {
        SKIP("No integration data found - set MYPROJ_INTEGRATION_DATA");
    }

    auto index = myproj::BuildIndex(*data);
    REQUIRE(index.Contains("records/main.dat"));
}
```

Skip on a missing environment, never on a missing *feature*: a test that skips because the code under test is broken is a test that never runs. Once the data is present, the test is a normal test and a failure is a failure.

Resolve the whole set of inputs a test needs before asserting on any of them, and skip once if any is missing rather than degrading to a smaller check. A test that quietly verifies less when half its inputs are absent reports success for a run that proved almost nothing, which is worse than a skip because nothing in the output says so.

## Running it

Integration needs its own configure, because the option is off by default. `configure_integration` in the Makefile targets fragment turns it on in `build/dev` and bakes in `INTEGRATION_DATA`, and `test_integration` builds through it and runs the layer. `test_all`, defined in cpp/testing.md, then runs whatever the current configure contains, which is the unit layer alone unless integration was configured in.

`configure_integration` fails with an actionable message when `INTEGRATION_DATA` is empty. That guard belongs in a project whose variable has no default: it stops the configure rather than producing a build whose integration tests all skip. A project that does default the variable drops the guard, because the condition can never be true and a check that cannot fire is one more thing to read.

## CI

Integration tests do not run in CI. The data is proprietary, large, or a live service, and none of that belongs in a workflow. CI runs `make test`, which is the unit layer; the integration layer is a local tool for the developer who has the data.

Never work around this by committing the inputs, fetching them in a workflow, or standing up the service in a container. If a behaviour needs covering in CI, synthesise the input and write a unit test.

## Obtaining the inputs

How the inputs reach a developer's machine is the project's business, and it differs every time: placed by hand, downloaded from somewhere public, copied off a device, generated by another tool, or a service that has to be running. What moon requires is that the project **says** which it is. A layer nobody outside the author can run is a layer that rots, and the symptom is a skip that every contributor sees and nobody can act on.

Document it in the README, next to how to run the tests. Where the inputs can be fetched without a human deciding anything, put that behind a Makefile target named `fetch_integration_data` (see the Makefile targets fragment), so the answer is the same command in every project that has one.

That target is offered, not required. A project whose inputs are proprietary, licensed, or simply not downloadable has nothing to put in it, and the README carries the whole answer instead. Note that this is a developer running a command deliberately, which is a different thing from the workflow fetch ruled out above: the ban is on CI reaching for the data, not on a person doing so.
